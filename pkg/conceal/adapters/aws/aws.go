// Package aws prepares a bounded, read-only S3 ListObjectsV2 request. Signing and
// credential lookup belong to the host; this package never calls the AWS SDK.
package aws

import (
	"encoding/json"
	"fmt"
	"net"
	"net/url"
	"regexp"
	"slices"
	"strings"

	"github.com/infamousjoeg/conceal/pkg/conceal/adapter"
)

const ListObjects = "aws.s3.objects.list"

type Adapter struct{}

// Initial regions are intentionally explicit. Adding an endpoint is a reviewed
// adapter change, which must also increment the adapter version.
var regions = []string{"us-east-1", "us-west-2", "eu-west-1"}

type Bucket struct {
	Name   string  `json:"name"`
	Prefix *string `json:"prefix"`
}

type Scope struct {
	Region  string   `json:"region"`
	Buckets []Bucket `json:"buckets"`
}

var bucketName = regexp.MustCompile(`^[a-z0-9][a-z0-9.-]{1,61}[a-z0-9]$`)

func (Adapter) Descriptor() adapter.Descriptor {
	hosts := make([]string, 0, len(regions))
	for _, region := range regions {
		hosts = append(hosts, "s3."+region+".amazonaws.com")
	}
	return adapter.Descriptor{
		APIVersion: adapter.APIVersion, ID: "aws", Version: "1", Name: "Amazon S3",
		AuthScheme: "aws-sigv4", SigningService: "s3", AllowedHosts: hosts,
		Credentials: []adapter.Credential{
			{Name: "access_key_id", Description: "AWS access key ID", Required: true},
			{Name: "secret_access_key", Description: "AWS secret access key", Required: true},
			{Name: "session_token", Description: "AWS session token (required for temporary credentials)"},
		},
		ScopeSchema: json.RawMessage(`{"type":"object","required":["region","buckets"],"additionalProperties":false,"properties":{"region":{"type":"string","enum":["us-east-1","us-west-2","eu-west-1"]},"buckets":{"type":"array","minItems":1,"items":{"type":"object","required":["name","prefix"],"additionalProperties":false,"properties":{"name":{"type":"string"},"prefix":{"type":"string"}}}}}}`),
		Operations: []adapter.Operation{{
			Name: ListObjects, Title: "List S3 object names", ReadOnly: true,
			Description: "Read one page of object names under an explicitly allowed bucket prefix. Does not read object contents.",
			InputSchema: json.RawMessage(`{"type":"object","additionalProperties":false,"properties":{"prefix":{"type":"string"},"limit":{"type":"integer","minimum":1,"maximum":100,"default":30}}}`),
		}},
	}
}

func (Adapter) ValidateScope(raw json.RawMessage) error {
	var scope Scope
	if err := adapter.Decode(raw, &scope); err != nil {
		return fmt.Errorf("S3 scope: %w", err)
	}
	if !slices.Contains(regions, scope.Region) {
		return fmt.Errorf("supported S3 regions are us-east-1, us-west-2, and eu-west-1")
	}
	if len(scope.Buckets) == 0 {
		return fmt.Errorf("explicit S3 buckets and prefixes are required")
	}
	names := make(map[string]bool)
	for _, bucket := range scope.Buckets {
		if !bucketName.MatchString(bucket.Name) || net.ParseIP(bucket.Name) != nil ||
			strings.Contains(bucket.Name, "..") || strings.HasSuffix(bucket.Name, "--x-s3") || names[bucket.Name] {
			return fmt.Errorf("bucket names must be unique, explicit general-purpose S3 bucket names")
		}
		if bucket.Prefix == nil || strings.ContainsAny(*bucket.Prefix, "\x00\r\n") || len(*bucket.Prefix) > 1024 {
			return fmt.Errorf("invalid bucket prefix")
		}
		names[bucket.Name] = true
	}
	return nil
}

func (a Adapter) Prepare(raw json.RawMessage, operation, resource string, input json.RawMessage) (adapter.Plan, error) {
	if err := a.ValidateScope(raw); err != nil {
		return adapter.Plan{}, err
	}
	if operation != ListObjects {
		return adapter.Plan{}, fmt.Errorf("unsupported AWS operation")
	}
	var scope Scope
	if err := adapter.Decode(raw, &scope); err != nil {
		return adapter.Plan{}, err
	}
	var selected *Bucket
	for _, bucket := range scope.Buckets {
		if bucket.Name == resource {
			selected = &bucket
			break
		}
	}
	if selected == nil {
		return adapter.Plan{}, fmt.Errorf("bucket %q is not allowed by this contract", resource)
	}
	parameters := struct {
		Prefix string `json:"prefix"`
		Limit  int    `json:"limit"`
	}{Prefix: *selected.Prefix, Limit: 30}
	if len(input) == 0 {
		input = json.RawMessage(`{}`)
	}
	if err := adapter.Decode(input, &parameters); err != nil {
		return adapter.Plan{}, fmt.Errorf("S3 input: %w", err)
	}
	if !strings.HasPrefix(parameters.Prefix, *selected.Prefix) || strings.ContainsAny(parameters.Prefix, "\x00\r\n") || len(parameters.Prefix) > 1024 {
		return adapter.Plan{}, fmt.Errorf("requested prefix is outside the contract or invalid")
	}
	if parameters.Limit < 1 || parameters.Limit > 100 {
		return adapter.Plan{}, fmt.Errorf("limit must be between 1 and 100")
	}
	query := url.Values{
		"list-type": {"2"}, "encoding-type": {"url"},
		"max-keys": {fmt.Sprint(parameters.Limit)}, "prefix": {parameters.Prefix},
	}
	prefix := parameters.Prefix
	if prefix == "" {
		prefix = "(entire bucket)"
	}
	return adapter.Plan{
		HTTP: adapter.HTTPRequest{
			Method: "GET", URL: "https://s3." + scope.Region + ".amazonaws.com/" + resource + "?" + query.Encode(),
		},
		Authentication: adapter.Authentication{Scheme: "aws-sigv4", Service: "s3", Region: scope.Region},
		Review: adapter.Review{
			Title:       "List S3 object names",
			Consequence: "Returns one page of object names and metadata, not file contents. Names may be private. AWS request charges may apply.",
			Fields: []adapter.Field{
				{Label: "Bucket", Value: resource}, {Label: "AWS region", Value: scope.Region},
				{Label: "Prefix", Value: prefix}, {Label: "Maximum results", Value: fmt.Sprint(parameters.Limit)},
				{Label: "Pagination", Value: "One page; no automatic follow-up requests"},
			},
		},
	}, nil
}

// Package adapter defines the versioned contribution API for reviewed, compiled
// action adapters. Adapters prepare data; the host owns credentials and dispatch.
// An interface is not a sandbox: installed adapter code must be trusted.
package adapter

import (
	"encoding/json"
	"fmt"
	"maps"
	"net/http"
	"net/url"
	"regexp"
	"slices"
	"strings"
)

const APIVersion = 1

type Credential struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Required    bool   `json:"required"`
}

type Operation struct {
	Name        string          `json:"name"`
	Title       string          `json:"title"`
	Description string          `json:"description"`
	ReadOnly    bool            `json:"read_only"`
	InputSchema json.RawMessage `json:"input_schema"`
}

type Descriptor struct {
	APIVersion     int             `json:"api_version"`
	ID             string          `json:"id"`
	Version        string          `json:"version"`
	Name           string          `json:"name"`
	AuthScheme     string          `json:"auth_scheme"`
	SigningService string          `json:"signing_service,omitempty"`
	AllowedHosts   []string        `json:"allowed_hosts"`
	Credentials    []Credential    `json:"credentials"`
	Operations     []Operation     `json:"operations"`
	ScopeSchema    json.RawMessage `json:"scope_schema"`
}

type Field struct {
	Label string `json:"label"`
	Value string `json:"value"`
}

type Review struct {
	Title       string  `json:"title"`
	Consequence string  `json:"consequence"`
	Fields      []Field `json:"fields"`
}

type HTTPRequest struct {
	Method  string            `json:"method"`
	URL     string            `json:"url"`
	Body    json.RawMessage   `json:"body,omitempty"`
	Headers map[string]string `json:"headers,omitempty"`
}

type Authentication struct {
	Scheme  string `json:"scheme"`
	Service string `json:"service,omitempty"`
	Region  string `json:"region,omitempty"`
}

// Plan contains all stable behavior covered by approval. Volatile authentication
// headers are added by the host only after approval and credential lookup.
type Plan struct {
	Adapter        string         `json:"adapter"`
	AdapterVersion string         `json:"adapter_version"`
	Operation      string         `json:"operation"`
	ReadOnly       bool           `json:"read_only"`
	HTTP           HTTPRequest    `json:"http"`
	Authentication Authentication `json:"authentication"`
	Review         Review         `json:"review"`
}

// Adapter implementations must be deterministic and perform no credential,
// filesystem, environment, clock, or network access while preparing requests.
// ValidateScope and Prepare must reject unknown/ambiguous input using Decode.
type Adapter interface {
	Descriptor() Descriptor
	ValidateScope(json.RawMessage) error
	Prepare(scope json.RawMessage, operation, resource string, input json.RawMessage) (Plan, error)
}

type Registry struct {
	adapters map[string]Adapter
}

var identifier = regexp.MustCompile(`^[a-z][a-z0-9-]*$`)

func NewRegistry(adapters ...Adapter) (*Registry, error) {
	r := &Registry{adapters: make(map[string]Adapter)}
	for _, a := range adapters {
		d := a.Descriptor()
		if d.APIVersion != APIVersion || !identifier.MatchString(d.ID) || strings.TrimSpace(d.Version) == "" || d.Name == "" {
			return nil, fmt.Errorf("invalid adapter identity or API version")
		}
		if _, exists := r.adapters[d.ID]; exists {
			return nil, fmt.Errorf("duplicate adapter %q", d.ID)
		}
		if len(d.AllowedHosts) == 0 || len(d.Operations) == 0 || !json.Valid(d.ScopeSchema) {
			return nil, fmt.Errorf("adapter %q needs hosts, operations, and a scope schema", d.ID)
		}
		for _, host := range d.AllowedHosts {
			parsed, err := url.Parse("https://" + host)
			if err != nil || host == "" || parsed.Host != host || parsed.Hostname() != host || parsed.Path != "" ||
				strings.ContainsAny(host, "*\\ \r\n") || strings.ToLower(host) != host {
				return nil, fmt.Errorf("adapter hosts must be explicit lowercase hostnames")
			}
		}
		if d.AuthScheme != "bearer" && d.AuthScheme != "aws-sigv4" {
			return nil, fmt.Errorf("unsupported authentication scheme %q", d.AuthScheme)
		}
		if d.AuthScheme == "aws-sigv4" && d.SigningService != "s3" {
			return nil, fmt.Errorf("only S3 signing is currently supported")
		}
		names := make(map[string]bool)
		for _, op := range d.Operations {
			if !strings.HasPrefix(op.Name, d.ID+".") || names[op.Name] || op.Title == "" || !json.Valid(op.InputSchema) {
				return nil, fmt.Errorf("invalid or duplicate operation %q", op.Name)
			}
			names[op.Name] = true
		}
		names = make(map[string]bool)
		for _, credential := range d.Credentials {
			if credential.Name == "" || credential.Description == "" || names[credential.Name] {
				return nil, fmt.Errorf("invalid or duplicate credential role")
			}
			names[credential.Name] = true
		}
		required := map[string]bool{"token": true}
		if d.AuthScheme == "aws-sigv4" {
			required = map[string]bool{"access_key_id": true, "secret_access_key": true}
		} else if d.SigningService != "" {
			return nil, fmt.Errorf("bearer adapters cannot declare a signing service")
		}
		for _, credential := range d.Credentials {
			if required[credential.Name] && credential.Required {
				delete(required, credential.Name)
			} else if d.AuthScheme != "aws-sigv4" || credential.Name != "session_token" {
				return nil, fmt.Errorf("credential roles must match the authentication scheme")
			}
		}
		if len(required) != 0 {
			return nil, fmt.Errorf("required authentication credential roles are missing")
		}
		r.adapters[d.ID] = a
	}
	return r, nil
}

func (r *Registry) Lookup(id string) (Adapter, error) {
	if r != nil {
		if a, ok := r.adapters[id]; ok {
			return a, nil
		}
	}
	return nil, fmt.Errorf("adapter %q is not installed", id)
}

func (r *Registry) Descriptors() []Descriptor {
	var result []Descriptor
	for _, a := range r.adapters {
		result = append(result, a.Descriptor())
	}
	slices.SortFunc(result, func(a, b Descriptor) int { return strings.Compare(a.ID, b.ID) })
	return result
}

func (r *Registry) Prepare(id string, scope json.RawMessage, operation, resource string, input json.RawMessage) (Plan, error) {
	a, err := r.Lookup(id)
	if err != nil {
		return Plan{}, err
	}
	d := a.Descriptor()
	var selected *Operation
	for _, op := range d.Operations {
		if op.Name == operation {
			selected = &op
			break
		}
	}
	if selected == nil {
		return Plan{}, fmt.Errorf("operation %q is not supported by %s", operation, id)
	}
	if err := a.ValidateScope(scope); err != nil {
		return Plan{}, err
	}
	if len(input) == 0 {
		input = json.RawMessage(`{}`)
	}
	plan, err := a.Prepare(scope, operation, resource, input)
	if err != nil {
		return Plan{}, err
	}
	plan.Adapter, plan.AdapterVersion, plan.Operation, plan.ReadOnly = d.ID, d.Version, operation, selected.ReadOnly
	plan.HTTP.Headers = maps.Clone(plan.HTTP.Headers)
	if len(plan.HTTP.Body) > 0 {
		hasContentType := false
		for name := range plan.HTTP.Headers {
			hasContentType = hasContentType || strings.EqualFold(name, "Content-Type")
		}
		if !hasContentType {
			if plan.HTTP.Headers == nil {
				plan.HTTP.Headers = make(map[string]string)
			}
			plan.HTTP.Headers["Content-Type"] = "application/json"
		}
	}
	if err := ValidatePlan(d, plan); err != nil {
		return Plan{}, err
	}
	return plan, nil
}

// ValidatePlan prevents adapters from supplying destinations or authentication
// behavior outside their reviewed declaration. It does not sandbox adapter code.
func ValidatePlan(d Descriptor, p Plan) error {
	u, err := url.Parse(p.HTTP.URL)
	if err != nil || u.Scheme != "https" || u.User != nil || u.Fragment != "" || u.Opaque != "" ||
		u.Host != u.Hostname() || !slices.Contains(d.AllowedHosts, u.Host) {
		return fmt.Errorf("request destination is outside the adapter's declared HTTPS hosts")
	}
	if p.HTTP.Method != http.MethodGet && p.HTTP.Method != http.MethodPost {
		return fmt.Errorf("unsupported request method")
	}
	if p.ReadOnly && p.HTTP.Method != http.MethodGet {
		return fmt.Errorf("read-only operations must use GET")
	}
	if p.HTTP.Method == http.MethodGet && len(p.HTTP.Body) != 0 {
		return fmt.Errorf("GET requests cannot have a body")
	}
	headers := make(map[string]bool)
	for name, value := range p.HTTP.Headers {
		canonical := http.CanonicalHeaderKey(name)
		if headers[canonical] {
			return fmt.Errorf("duplicate request header")
		}
		headers[canonical] = true
		switch canonical {
		case "Accept", "Content-Type", "X-Github-Api-Version":
		default:
			return fmt.Errorf("header %q is not allowed in a prepared request", name)
		}
		if strings.ContainsAny(value, "\x00\r\n") {
			return fmt.Errorf("invalid request header")
		}
	}
	if p.Authentication.Scheme != d.AuthScheme {
		return fmt.Errorf("authentication does not match the adapter declaration")
	}
	if d.AuthScheme == "aws-sigv4" {
		if p.Authentication.Service != d.SigningService || p.Authentication.Region == "" ||
			u.Host != "s3."+p.Authentication.Region+".amazonaws.com" {
			return fmt.Errorf("AWS signing region and destination do not match")
		}
	} else if p.Authentication.Region != "" || p.Authentication.Service != "" {
		return fmt.Errorf("unexpected signing parameters")
	}
	if p.Review.Title == "" || p.Review.Consequence == "" || len(p.Review.Fields) == 0 {
		return fmt.Errorf("a complete human review is required")
	}
	for _, field := range p.Review.Fields {
		if strings.TrimSpace(field.Label) == "" {
			return fmt.Errorf("review fields require labels")
		}
	}
	data, err := json.Marshal(p)
	if err != nil || len(data) > 1024*1024 {
		return fmt.Errorf("prepared plan exceeds 1 MiB or cannot be encoded")
	}
	return nil
}

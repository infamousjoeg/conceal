// Package example shows the complete contribution surface for a compiled
// adapter. example.invalid is intentionally nonfunctional. This package is not
// registered in Conceal; replace it with a reviewed service-specific adapter.
package example

import (
	"encoding/json"
	"fmt"
	"net/url"
	"slices"
	"strings"

	"github.com/infamousjoeg/conceal/pkg/conceal/adapter"
)

const List = "example.items.list"

type Adapter struct{}

type scope struct {
	Projects []string `json:"projects"`
}

func (Adapter) Descriptor() adapter.Descriptor {
	return adapter.Descriptor{
		APIVersion: adapter.APIVersion, ID: "example", Version: "1", Name: "Example service",
		AllowedHosts: []string{"api.example.invalid"}, AuthScheme: "bearer",
		Credentials: []adapter.Credential{{Name: "token", Description: "Example API token", Required: true}},
		ScopeSchema: json.RawMessage(`{"type":"object","required":["projects"],"additionalProperties":false,"properties":{"projects":{"type":"array","minItems":1,"items":{"type":"string"}}}}`),
		Operations: []adapter.Operation{{
			Name: List, Title: "List project items", Description: "Read one bounded page of item metadata.", ReadOnly: true,
			InputSchema: json.RawMessage(`{"type":"object","additionalProperties":false,"properties":{"limit":{"type":"integer","minimum":1,"maximum":25,"default":10}}}`),
		}},
	}
}

func (Adapter) ValidateScope(raw json.RawMessage) error {
	var s scope
	if err := adapter.Decode(raw, &s); err != nil {
		return err
	}
	if len(s.Projects) == 0 {
		return fmt.Errorf("explicit projects are required")
	}
	for _, project := range s.Projects {
		if project == "" || len(project) > 100 || strings.ContainsAny(project, "*/\\?#\r\n\x00") || project == "." || project == ".." {
			return fmt.Errorf("invalid project")
		}
	}
	return nil
}

func (a Adapter) Prepare(raw json.RawMessage, operation, resource string, input json.RawMessage) (adapter.Plan, error) {
	if err := a.ValidateScope(raw); err != nil {
		return adapter.Plan{}, err
	}
	var s scope
	if err := adapter.Decode(raw, &s); err != nil {
		return adapter.Plan{}, err
	}
	if operation != List || !slices.Contains(s.Projects, resource) {
		return adapter.Plan{}, fmt.Errorf("operation or project is outside the contract")
	}
	parameters := struct {
		Limit int `json:"limit"`
	}{Limit: 10}
	if len(input) == 0 {
		input = json.RawMessage(`{}`)
	}
	if err := adapter.Decode(input, &parameters); err != nil {
		return adapter.Plan{}, err
	}
	if parameters.Limit < 1 || parameters.Limit > 25 {
		return adapter.Plan{}, fmt.Errorf("limit must be 1–25")
	}
	return adapter.Plan{
		HTTP:           adapter.HTTPRequest{Method: "GET", URL: "https://api.example.invalid/projects/" + url.PathEscape(resource) + "/items?limit=" + fmt.Sprint(parameters.Limit)},
		Authentication: adapter.Authentication{Scheme: "bearer"},
		Review: adapter.Review{
			Title: "List project items", Consequence: "Returns one page of item metadata. Item names may be private.",
			Fields: []adapter.Field{{Label: "Project", Value: resource}, {Label: "Maximum results", Value: fmt.Sprint(parameters.Limit)}},
		},
	}, nil
}

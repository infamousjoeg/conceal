// Package contracts prepares and executes narrowly scoped, approved API actions.
package contracts

import (
	"encoding/json"
	"fmt"
	"slices"
	"strings"

	"github.com/infamousjoeg/conceal/internal/adapters"
	"github.com/infamousjoeg/conceal/pkg/conceal/adapter"
	"github.com/infamousjoeg/conceal/pkg/conceal/adapters/github"
)

const (
	ListIssues  = github.ListIssues
	CreateDraft = github.CreateDraft
)

// The first six fields and their JSON order preserve schema-1 request digests.
// Schema 2 uses adapter/credentials/scope instead of credential/repositories.
type Contract struct {
	SchemaVersion      int               `json:"schema_version"`
	Name               string            `json:"name"`
	Credential         string            `json:"credential,omitempty"`
	Repositories       []string          `json:"repositories,omitempty"`
	Actions            []string          `json:"actions"`
	ApprovalTTLSeconds int               `json:"approval_ttl_seconds"`
	Adapter            string            `json:"adapter,omitempty"`
	Credentials        map[string]string `json:"credentials,omitempty"`
	Scope              json.RawMessage   `json:"scope,omitempty"`
}

type HTTPRequest = adapter.HTTPRequest

func Decode(data []byte, target any) error { return adapter.Decode(data, target) }

func (c Contract) adapterID() string {
	if c.SchemaVersion == 1 {
		return "github"
	}
	return c.Adapter
}

func (c Contract) scope() json.RawMessage {
	if c.SchemaVersion == 1 {
		data, _ := json.Marshal(github.Scope{Repositories: c.Repositories})
		return data
	}
	return c.Scope
}

func (c Contract) credentialReferences() map[string]string {
	if c.SchemaVersion == 1 {
		return map[string]string{"token": c.Credential}
	}
	return c.Credentials
}

func (c Contract) Validate() error { return c.validate(adapters.NewRegistry()) }

func (c Contract) validate(registry *adapter.Registry) error {
	if c.SchemaVersion != 1 && c.SchemaVersion != 2 {
		return fmt.Errorf("unsupported contract schema version")
	}
	if strings.TrimSpace(c.Name) == "" || len(c.Actions) == 0 {
		return fmt.Errorf("contract name and explicit actions are required")
	}
	if c.ApprovalTTLSeconds < 1 || c.ApprovalTTLSeconds > 3600 {
		return fmt.Errorf("approval_ttl_seconds must be between 1 and 3600")
	}
	if c.SchemaVersion == 1 && (c.Adapter != "" || c.Credentials != nil || len(c.Scope) != 0) {
		return fmt.Errorf("schema 1 does not accept adapter, credentials, or scope")
	}
	if c.SchemaVersion == 2 && (c.Credential != "" || len(c.Repositories) != 0) {
		return fmt.Errorf("schema 2 uses credentials and scope, not credential or repositories")
	}
	a, err := registry.Lookup(c.adapterID())
	if err != nil {
		return err
	}
	d := a.Descriptor()
	if err := a.ValidateScope(c.scope()); err != nil {
		return err
	}
	for _, name := range c.Actions {
		if !slices.ContainsFunc(d.Operations, func(op adapter.Operation) bool { return op.Name == name }) {
			return fmt.Errorf("unsupported action %q", name)
		}
	}
	references := c.credentialReferences()
	for _, role := range d.Credentials {
		if role.Required && strings.TrimSpace(references[role.Name]) == "" {
			return fmt.Errorf("credential reference %q is required", role.Name)
		}
	}
	for role, reference := range references {
		if !slices.ContainsFunc(d.Credentials, func(c adapter.Credential) bool { return c.Name == role }) {
			return fmt.Errorf("unknown credential role %q", role)
		}
		if strings.TrimSpace(reference) == "" || strings.ContainsAny(reference, "\x00\r\n") {
			return fmt.Errorf("credential references must be nonempty secret names")
		}
	}
	return nil
}

func (c Contract) plan(registry *adapter.Registry, action, resource string, input json.RawMessage) (adapter.Plan, error) {
	if err := c.validate(registry); err != nil {
		return adapter.Plan{}, err
	}
	if !slices.Contains(c.Actions, action) {
		return adapter.Plan{}, fmt.Errorf("action %q is not allowed by this contract", action)
	}
	return registry.Prepare(c.adapterID(), c.scope(), action, resource, input)
}

// Build retains the schema-1 CLI compatibility surface. Version-2 requests also
// persist the complete adapter plan, including authentication and review fields.
func Build(contract Contract, action, resource string, input json.RawMessage) (HTTPRequest, error) {
	plan, err := contract.plan(adapters.NewRegistry(), action, resource, input)
	if err != nil {
		return HTTPRequest{}, err
	}
	if contract.SchemaVersion == 1 {
		plan.HTTP.Headers = nil
	}
	return plan.HTTP, nil
}

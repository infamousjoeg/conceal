// Package github is the reference adapter for repository-scoped GitHub actions.
package github

import (
	"encoding/json"
	"fmt"
	"net/url"
	"regexp"
	"slices"
	"strings"

	"github.com/infamousjoeg/conceal/pkg/conceal/adapter"
)

const (
	ListIssues  = "github.issues.list"
	CreateDraft = "github.pulls.create-draft"
)

type Adapter struct{}

type Scope struct {
	Repositories []string `json:"repositories"`
}

var repositoryName = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9_.-]*/[A-Za-z0-9_.-]+$`)

func (Adapter) Descriptor() adapter.Descriptor {
	return adapter.Descriptor{
		APIVersion: adapter.APIVersion, ID: "github", Version: "1", Name: "GitHub",
		AuthScheme: "bearer", AllowedHosts: []string{"api.github.com"},
		Credentials: []adapter.Credential{{Name: "token", Description: "GitHub token", Required: true}},
		ScopeSchema: json.RawMessage(`{"type":"object","required":["repositories"],"additionalProperties":false,"properties":{"repositories":{"type":"array","minItems":1,"items":{"type":"string"}}}}`),
		Operations: []adapter.Operation{
			{Name: ListIssues, Title: "List repository issues", ReadOnly: true,
				Description: "Read one page of GitHub issues, including pull requests.",
				InputSchema: json.RawMessage(`{"type":"object","additionalProperties":false,"properties":{"state":{"type":"string","enum":["open","closed","all"],"default":"open"},"limit":{"type":"integer","minimum":1,"maximum":100,"default":30}}}`)},
			{Name: CreateDraft, Title: "Open a draft pull request",
				Description: "Open a draft from an already pushed branch. Does not push or merge code.",
				InputSchema: json.RawMessage(`{"type":"object","required":["title","head","base"],"additionalProperties":false,"properties":{"title":{"type":"string","minLength":1},"head":{"type":"string","minLength":1},"base":{"type":"string","minLength":1},"body":{"type":"string"}}}`)},
		},
	}
}

func (Adapter) ValidateScope(raw json.RawMessage) error {
	var scope Scope
	if err := adapter.Decode(raw, &scope); err != nil {
		return fmt.Errorf("GitHub scope: %w", err)
	}
	if len(scope.Repositories) == 0 {
		return fmt.Errorf("explicit repositories are required")
	}
	for _, repo := range scope.Repositories {
		_, name, _ := strings.Cut(repo, "/")
		if !repositoryName.MatchString(repo) || name == "." || name == ".." {
			return fmt.Errorf("repository must be an explicit owner/name, without wildcards")
		}
	}
	return nil
}

func (a Adapter) Prepare(raw json.RawMessage, operation, resource string, input json.RawMessage) (adapter.Plan, error) {
	if err := a.ValidateScope(raw); err != nil {
		return adapter.Plan{}, err
	}
	var scope Scope
	if err := adapter.Decode(raw, &scope); err != nil {
		return adapter.Plan{}, err
	}
	if !slices.Contains(scope.Repositories, resource) {
		return adapter.Plan{}, fmt.Errorf("repository %q is not allowed by this contract", resource)
	}
	if len(input) == 0 {
		input = json.RawMessage(`{}`)
	}
	plan := adapter.Plan{Authentication: adapter.Authentication{Scheme: "bearer"}}
	plan.HTTP.Headers = map[string]string{"Accept": "application/vnd.github+json", "X-GitHub-Api-Version": "2022-11-28"}
	base := "https://api.github.com/repos/" + resource
	switch operation {
	case ListIssues:
		parameters := struct {
			State string `json:"state"`
			Limit int    `json:"limit"`
		}{State: "open", Limit: 30}
		if err := adapter.Decode(input, &parameters); err != nil {
			return adapter.Plan{}, fmt.Errorf("issue input: %w", err)
		}
		if parameters.State != "open" && parameters.State != "closed" && parameters.State != "all" {
			return adapter.Plan{}, fmt.Errorf("state must be open, closed, or all")
		}
		if parameters.Limit < 1 || parameters.Limit > 100 {
			return adapter.Plan{}, fmt.Errorf("limit must be between 1 and 100")
		}
		query := url.Values{"state": {parameters.State}, "per_page": {fmt.Sprint(parameters.Limit)}}
		plan.HTTP.Method, plan.HTTP.URL = "GET", base+"/issues?"+query.Encode()
		plan.Review = adapter.Review{Title: "List repository issues",
			Consequence: "Reads one page of issues and pull requests. Repository data will be returned to the caller.",
			Fields: []adapter.Field{
				{Label: "Repository", Value: resource}, {Label: "Issue state", Value: parameters.State},
				{Label: "Maximum results", Value: fmt.Sprint(parameters.Limit)},
			}}
	case CreateDraft:
		var parameters struct {
			Title string `json:"title"`
			Head  string `json:"head"`
			Base  string `json:"base"`
			Body  string `json:"body"`
		}
		if err := adapter.Decode(input, &parameters); err != nil {
			return adapter.Plan{}, fmt.Errorf("draft pull-request input: %w", err)
		}
		if parameters.Title == "" || parameters.Head == "" || parameters.Base == "" {
			return adapter.Plan{}, fmt.Errorf("title, head, and base are required")
		}
		body, err := json.Marshal(struct {
			Title               string `json:"title"`
			Head                string `json:"head"`
			Base                string `json:"base"`
			Body                string `json:"body"`
			Draft               bool   `json:"draft"`
			MaintainerCanModify bool   `json:"maintainer_can_modify"`
		}{parameters.Title, parameters.Head, parameters.Base, parameters.Body, true, false})
		if err != nil {
			return adapter.Plan{}, err
		}
		plan.HTTP.Method, plan.HTTP.URL, plan.HTTP.Body = "POST", base+"/pulls", body
		plan.Review = adapter.Review{Title: "Open a draft pull request",
			Consequence: "Creates a draft pull request on GitHub. The source branch must already be pushed. Does not merge code.",
			Fields: []adapter.Field{
				{Label: "Repository", Value: resource}, {Label: "Title", Value: parameters.Title},
				{Label: "From branch", Value: parameters.Head}, {Label: "Into branch", Value: parameters.Base},
				{Label: "Description", Value: parameters.Body}, {Label: "Draft", Value: "Yes"},
				{Label: "Allow maintainer edits", Value: "No"},
			}}
	default:
		return adapter.Plan{}, fmt.Errorf("unsupported GitHub operation")
	}
	return plan, nil
}

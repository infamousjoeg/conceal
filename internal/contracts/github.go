// Package contracts prepares and executes narrowly scoped, approved API actions.
package contracts

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/url"
	"regexp"
	"slices"
	"strings"
)

const (
	ListIssues  = "github.issues.list"
	CreateDraft = "github.pulls.create-draft"
)

// Contract is a reviewable access policy. Credentials are references, not values.
type Contract struct {
	SchemaVersion      int      `json:"schema_version"`
	Name               string   `json:"name"`
	Credential         string   `json:"credential"`
	Repositories       []string `json:"repositories"`
	Actions            []string `json:"actions"`
	ApprovalTTLSeconds int      `json:"approval_ttl_seconds"`
}

var repositoryName = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9_.-]*/[A-Za-z0-9_.-]+$`)

func Decode(data []byte, target any) error {
	if err := unambiguousObject(data); err != nil {
		return err
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		return err
	}
	if err := decoder.Decode(new(any)); err != io.EOF {
		return fmt.Errorf("expected a single JSON document")
	}
	return nil
}

// Duplicate keys can display differently in a reviewer and an executor. Reject
// them, including case variants accepted by encoding/json's struct matching.
func unambiguousObject(data []byte) error {
	if trimmed := bytes.TrimSpace(data); len(trimmed) == 0 || trimmed[0] != '{' {
		return fmt.Errorf("expected a JSON object")
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.UseNumber()
	var visit func(int) error
	visit = func(depth int) error {
		if depth > 64 {
			return fmt.Errorf("JSON nesting exceeds 64 levels")
		}
		token, err := decoder.Token()
		if err != nil {
			return err
		}
		delim, ok := token.(json.Delim)
		if !ok {
			return nil
		}
		switch delim {
		case '{':
			keys := make(map[string]bool)
			for decoder.More() {
				key, err := decoder.Token()
				if err != nil {
					return err
				}
				name, ok := key.(string)
				if !ok {
					return fmt.Errorf("expected an object key")
				}
				folded := strings.ToLower(name)
				if keys[folded] {
					return fmt.Errorf("duplicate JSON field %q", name)
				}
				keys[folded] = true
				if err := visit(depth + 1); err != nil {
					return err
				}
			}
		case '[':
			for decoder.More() {
				if err := visit(depth + 1); err != nil {
					return err
				}
			}
		default:
			return fmt.Errorf("unexpected JSON delimiter")
		}
		_, err = decoder.Token()
		return err
	}
	if err := visit(0); err != nil {
		return err
	}
	if _, err := decoder.Token(); err != io.EOF {
		return fmt.Errorf("expected a single JSON document")
	}
	return nil
}

func (c Contract) Validate() error {
	if c.SchemaVersion != 1 {
		return fmt.Errorf("unsupported contract schema version")
	}
	if strings.TrimSpace(c.Name) == "" || strings.TrimSpace(c.Credential) == "" || strings.ContainsAny(c.Credential, "\x00\r\n") {
		return fmt.Errorf("contract name and credential reference are required")
	}
	if len(c.Repositories) == 0 || len(c.Actions) == 0 {
		return fmt.Errorf("explicit repositories and actions are required")
	}
	if c.ApprovalTTLSeconds < 1 || c.ApprovalTTLSeconds > 3600 {
		return fmt.Errorf("approval_ttl_seconds must be between 1 and 3600")
	}
	for _, repo := range c.Repositories {
		_, name, _ := strings.Cut(repo, "/")
		if !repositoryName.MatchString(repo) || name == "." || name == ".." {
			return fmt.Errorf("repository must be an explicit owner/name, without wildcards")
		}
	}
	for _, action := range c.Actions {
		if action != ListIssues && action != CreateDraft {
			return fmt.Errorf("unsupported action %q", action)
		}
	}
	return nil
}

type HTTPRequest struct {
	Method string          `json:"method"`
	URL    string          `json:"url"`
	Body   json.RawMessage `json:"body,omitempty"`
}

// Build does not accept arbitrary URLs, HTTP methods, headers, or API actions.
func Build(contract Contract, action, repo string, input json.RawMessage) (HTTPRequest, error) {
	if err := contract.Validate(); err != nil {
		return HTTPRequest{}, err
	}
	if !slices.Contains(contract.Repositories, repo) {
		return HTTPRequest{}, fmt.Errorf("repository %q is not allowed by this contract", repo)
	}
	if !slices.Contains(contract.Actions, action) {
		return HTTPRequest{}, fmt.Errorf("action %q is not allowed by this contract", action)
	}
	if len(input) == 0 {
		input = json.RawMessage(`{}`)
	}
	base := "https://api.github.com/repos/" + repo
	switch action {
	case ListIssues:
		parameters := struct {
			State string `json:"state"`
			Limit int    `json:"limit"`
		}{State: "open", Limit: 30}
		if err := Decode(input, &parameters); err != nil {
			return HTTPRequest{}, fmt.Errorf("issue input: %w", err)
		}
		if parameters.State != "open" && parameters.State != "closed" && parameters.State != "all" {
			return HTTPRequest{}, fmt.Errorf("state must be open, closed, or all")
		}
		if parameters.Limit < 1 || parameters.Limit > 100 {
			return HTTPRequest{}, fmt.Errorf("limit must be between 1 and 100")
		}
		query := url.Values{"state": {parameters.State}, "per_page": {fmt.Sprint(parameters.Limit)}}
		return HTTPRequest{Method: "GET", URL: base + "/issues?" + query.Encode()}, nil
	case CreateDraft:
		var parameters struct {
			Title string `json:"title"`
			Head  string `json:"head"`
			Base  string `json:"base"`
			Body  string `json:"body"`
		}
		if err := Decode(input, &parameters); err != nil {
			return HTTPRequest{}, fmt.Errorf("draft pull-request input: %w", err)
		}
		if parameters.Title == "" || parameters.Head == "" || parameters.Base == "" {
			return HTTPRequest{}, fmt.Errorf("title, head, and base are required")
		}
		body, err := json.Marshal(struct {
			Title               string `json:"title"`
			Head                string `json:"head"`
			Base                string `json:"base"`
			Body                string `json:"body"`
			Draft               bool   `json:"draft"`
			MaintainerCanModify bool   `json:"maintainer_can_modify"`
		}{parameters.Title, parameters.Head, parameters.Base, parameters.Body, true, false})
		return HTTPRequest{Method: "POST", URL: base + "/pulls", Body: body}, err
	default:
		return HTTPRequest{}, fmt.Errorf("unsupported action")
	}
}

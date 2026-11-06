package github_test

import (
	"encoding/json"
	"testing"

	"github.com/infamousjoeg/conceal/pkg/conceal/adapter"
	"github.com/infamousjoeg/conceal/pkg/conceal/adapter/adaptertest"
	"github.com/infamousjoeg/conceal/pkg/conceal/adapters/github"
)

func TestConformance(t *testing.T) {
	scope := json.RawMessage(`{"repositories":["owner/repo"]}`)
	adaptertest.Run(t, github.Adapter{}, []adaptertest.Case{
		{Name: "issues", Scope: scope, Operation: github.ListIssues, Resource: "owner/repo", Check: func(t *testing.T, p adapter.Plan) {
			t.Helper()
			if !p.ReadOnly || p.HTTP.URL != "https://api.github.com/repos/owner/repo/issues?per_page=30&state=open" {
				t.Fatal("unexpected listing")
			}
		}},
		{Name: "draft", Scope: scope, Operation: github.CreateDraft, Resource: "owner/repo",
			Input: json.RawMessage(`{"title":"Review","head":"work","base":"main","body":"Complete description"}`),
			Check: func(t *testing.T, p adapter.Plan) {
				t.Helper()
				var body map[string]any
				if err := json.Unmarshal(p.HTTP.Body, &body); err != nil {
					t.Fatal(err)
				}
				if body["draft"] != true || body["maintainer_can_modify"] != false || p.ReadOnly || len(p.Review.Fields) != 7 || p.HTTP.Headers["Content-Type"] != "application/json" {
					t.Fatal("draft guarantees or review incomplete")
				}
			}},
		{Name: "outside repository", Scope: scope, Operation: github.ListIssues, Resource: "other/repo", Reject: true},
		{Name: "ambiguous input", Scope: scope, Operation: github.ListIssues, Resource: "owner/repo", Input: json.RawMessage(`{"limit":1,"Limit":2}`), Reject: true},
		{Name: "draft override", Scope: scope, Operation: github.CreateDraft, Resource: "owner/repo", Input: json.RawMessage(`{"title":"x","head":"x","base":"main","draft":false}`), Reject: true},
	})
}

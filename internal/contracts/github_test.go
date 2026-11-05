package contracts

import (
	"encoding/json"
	"testing"
)

func TestContractLimitsAndDraftEnforcement(t *testing.T) {
	contract := Contract{1, "test", "github/token", []string{"owner/repo"}, []string{ListIssues, CreateDraft}, 300}
	request, err := Build(contract, CreateDraft, "owner/repo", json.RawMessage(`{"title":"Fix","head":"feature","base":"main","body":"Description"}`))
	if err != nil {
		t.Fatal(err)
	}
	var body map[string]any
	if err := json.Unmarshal(request.Body, &body); err != nil {
		t.Fatal(err)
	}
	if body["draft"] != true || body["maintainer_can_modify"] != false || request.Method != "POST" || request.URL != "https://api.github.com/repos/owner/repo/pulls" {
		t.Fatalf("wrong draft request: %#v", request)
	}
	for _, test := range []struct{ action, repo, input string }{
		{ListIssues, "other/repo", `{}`},
		{"github.pulls.merge", "owner/repo", `{}`},
		{CreateDraft, "owner/repo", `{"title":"Fix","head":"feature","base":"main","draft":false}`},
		{CreateDraft, "owner/repo", `{"title":"Fix"}`},
		{ListIssues, "owner/repo", `{"url":"https://example.invalid"}`},
		{ListIssues, "owner/repo", `{"limit":101}`},
		{ListIssues, "owner/repo", `{"state":"unknown"}`},
		{ListIssues, "owner/repo", `{} {}`},
		{ListIssues, "owner/repo", `null`},
		{ListIssues, "owner/repo", `{"limit":5,"limit":100}`},
		{ListIssues, "owner/repo", `{"limit":5,"Limit":100}`},
	} {
		if _, err := Build(contract, test.action, test.repo, json.RawMessage(test.input)); err == nil {
			t.Errorf("accepted disallowed action: %#v", test)
		}
	}
	for _, repo := range []string{"*", "owner/*", "owner/.", "owner/..", "owner/repo/../../other", "owner/repo?x=1", "owner/repo%2fother"} {
		contract.Repositories = []string{repo}
		if err := contract.Validate(); err == nil {
			t.Errorf("accepted repository %q", repo)
		}
	}
}

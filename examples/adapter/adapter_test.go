package example

import (
	"encoding/json"
	"testing"

	"github.com/infamousjoeg/conceal/pkg/conceal/adapter"
	"github.com/infamousjoeg/conceal/pkg/conceal/adapter/adaptertest"
)

func TestConformance(t *testing.T) {
	scope := json.RawMessage(`{"projects":["my-project"]}`)
	adaptertest.Run(t, Adapter{}, []adaptertest.Case{
		{Name: "bounded list", Scope: scope, Operation: List, Resource: "my-project", Input: json.RawMessage(`{"limit":5}`),
			Check: func(t *testing.T, p adapter.Plan) {
				t.Helper()
				if p.HTTP.URL != "https://api.example.invalid/projects/my-project/items?limit=5" || !p.ReadOnly {
					t.Fatal("unexpected prepared request")
				}
			}},
		{Name: "wrong project", Scope: scope, Operation: List, Resource: "other-project", Reject: true},
		{Name: "too many items", Scope: scope, Operation: List, Resource: "my-project", Input: json.RawMessage(`{"limit":26}`), Reject: true},
		{Name: "unknown option", Scope: scope, Operation: List, Resource: "my-project", Input: json.RawMessage(`{"url":"https://other.invalid"}`), Reject: true},
		{Name: "ambiguous input", Scope: scope, Operation: List, Resource: "my-project", Input: json.RawMessage(`{"limit":1,"Limit":2}`), Reject: true},
	})
}

package aws_test

import (
	"encoding/json"
	"net/url"
	"testing"

	"github.com/infamousjoeg/conceal/pkg/conceal/adapter"
	"github.com/infamousjoeg/conceal/pkg/conceal/adapter/adaptertest"
	"github.com/infamousjoeg/conceal/pkg/conceal/adapters/aws"
)

func TestConformance(t *testing.T) {
	scope := json.RawMessage(`{"region":"us-east-1","buckets":[{"name":"example-backups","prefix":"reports/"}]}`)
	cases := []adaptertest.Case{{
		Name: "bounded listing", Scope: scope, Operation: aws.ListObjects, Resource: "example-backups",
		Input: json.RawMessage(`{"prefix":"reports/a +&雪/","limit":5}`),
		Check: func(t *testing.T, p adapter.Plan) {
			t.Helper()
			u, err := url.Parse(p.HTTP.URL)
			if err != nil || u.Host != "s3.us-east-1.amazonaws.com" || u.Path != "/example-backups" ||
				u.Query().Get("prefix") != "reports/a +&雪/" || u.Query().Get("max-keys") != "5" || len(u.Query()) != 4 ||
				!p.ReadOnly || p.Authentication.Region != "us-east-1" {
				t.Fatalf("unexpected plan: %#v", p)
			}
		},
	}, {
		Name: "default prefix", Scope: scope, Operation: aws.ListObjects, Resource: "example-backups",
		Check: func(t *testing.T, p adapter.Plan) {
			t.Helper()
			u, _ := url.Parse(p.HTTP.URL)
			if u.Query().Get("prefix") != "reports/" || u.Query().Get("max-keys") != "30" {
				t.Fatal("defaults broadened the scope")
			}
		},
	}}
	for _, input := range []string{
		`{"prefix":""}`, `{"prefix":"report"}`, `{"prefix":"reports-other/"}`, `{"limit":0}`, `{"limit":101}`,
		`{"limit":1,"Limit":100}`, `{"endpoint":"https://example.invalid"}`, `{"prefix":"reports/\n"}`,
		`{"continuation_token":"next"}`, `{"limit":1.5}`, `[]`,
	} {
		cases = append(cases, adaptertest.Case{Name: "reject " + input, Scope: scope, Operation: aws.ListObjects, Resource: "example-backups", Input: json.RawMessage(input), Reject: true})
	}
	cases = append(cases, adaptertest.Case{Name: "wrong bucket", Scope: scope, Operation: aws.ListObjects, Resource: "other-bucket", Reject: true})
	for _, invalid := range []string{
		`{"region":"us-east-2","buckets":[{"name":"example-backups","prefix":""}]}`,
		`{"region":"us-east-1","buckets":[{"name":"example-backups"}]}`,
		`{"region":"us-east-1","buckets":[{"name":"example-backups","prefix":null}]}`,
		`{"region":"us-east-1","buckets":[{"name":"example-backups/../other","prefix":""}]}`,
		`{"region":"us-east-1","buckets":[{"name":"192.0.2.1","prefix":""}]}`,
		`{"region":"us-east-1","buckets":[{"name":"example-backups","prefix":""}],"endpoint":"https://evil.invalid"}`,
	} {
		cases = append(cases, adaptertest.Case{Name: "scope " + invalid, Scope: json.RawMessage(invalid), Operation: aws.ListObjects, Resource: "example-backups", Reject: true})
	}
	adaptertest.Run(t, aws.Adapter{}, cases)
}

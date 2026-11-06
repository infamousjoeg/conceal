package contracts

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/infamousjoeg/conceal/pkg/conceal/adapter"
	"github.com/infamousjoeg/conceal/pkg/conceal/adapters/aws"
	"github.com/infamousjoeg/conceal/pkg/conceal/adapters/github"
)

func awsFixture(t *testing.T) (Manager, Proposal, *int, *int, *[][]byte) {
	t.Helper()
	root := t.TempDir()
	path := filepath.Join(root, "s3.json")
	c := Contract{
		SchemaVersion: 2, Name: "backups", Adapter: "aws", ApprovalTTLSeconds: 60,
		Credentials: map[string]string{"access_key_id": "aws/id", "secret_access_key": "aws/secret", "session_token": "aws/session"},
		Actions:     []string{aws.ListObjects}, Scope: json.RawMessage(`{"region":"us-west-2","buckets":[{"name":"example-backups","prefix":"reports/"}]}`),
	}
	data, _ := json.Marshal(c)
	if err := os.WriteFile(path, data, 0600); err != nil {
		t.Fatal(err)
	}
	reads, calls := new(int), new(int)
	buffers := new([][]byte)
	m := Manager{Directory: filepath.Join(root, "state"), Now: func() time.Time { return time.Date(2026, 11, 5, 12, 0, 0, 0, time.UTC) }}
	m.ReadSecret = func(ref string) ([]byte, error) {
		*reads++
		value := []byte(map[string]string{"aws/id": "AKIDEXAMPLE", "aws/secret": "secret-example", "aws/session": "session-example"}[ref])
		*buffers = append(*buffers, value)
		return value, nil
	}
	m.Transport = transportFunc(func(r *http.Request) (*http.Response, error) {
		*calls++
		if r.Method != "GET" || r.URL.Host != "s3.us-west-2.amazonaws.com" || r.URL.Query().Get("prefix") != "reports/" ||
			r.Header.Get("X-Amz-Security-Token") != "session-example" || r.Header.Get("X-Amz-Date") != "20261105T120000Z" ||
			!strings.HasPrefix(r.Header.Get("Authorization"), "AWS4-HMAC-SHA256 Credential=AKIDEXAMPLE/20261105/us-west-2/s3/aws4_request,") ||
			r.GetBody != nil || !r.Close {
			t.Fatalf("incorrect signed dispatch: %#v %v", r, r.Header)
		}
		return &http.Response{StatusCode: 200, Header: http.Header{}, Body: io.NopCloser(strings.NewReader(`<List>AKIDEXAMPLE secret-example session-example</List>`)), Request: r}, nil
	})
	p, err := m.Prepare(path, aws.ListObjects, "example-backups", nil)
	if err != nil {
		t.Fatal(err)
	}
	return m, p, reads, calls, buffers
}

func TestAWSExactApprovalAndSigning(t *testing.T) {
	m, p, reads, calls, buffers := awsFixture(t)
	view, err := m.Review(p.Request.ID)
	if err != nil || view.Status != "pending" || len(view.Credentials) != 3 || view.Proposal.Request.Plan.AdapterVersion != "1" {
		t.Fatalf("invalid review: %#v %v", view, err)
	}
	if _, err := m.Execute(context.Background(), p.Request.ID); err == nil || *reads != 0 || *calls != 0 {
		t.Fatal("unapproved action reached credentials or network")
	}
	approve(t, m, p)
	if *reads != 0 {
		t.Fatal("approval read a credential")
	}
	result, err := m.Execute(context.Background(), p.Request.ID)
	if err != nil || result.Receipt.Adapter != "aws" || result.Receipt.Resource != "example-backups" || *reads != 3 || *calls != 1 {
		t.Fatalf("execution failed: %#v %v", result, err)
	}
	for _, secret := range []string{"AKIDEXAMPLE", "secret-example", "session-example"} {
		if bytes.Contains(result.Response, []byte(secret)) {
			t.Fatal("response includes credentials")
		}
	}
	for _, buffer := range *buffers {
		if !bytes.Equal(buffer, make([]byte, len(buffer))) {
			t.Fatal("credential buffer was not cleared")
		}
	}
	if _, err := m.Execute(context.Background(), p.Request.ID); err == nil || *calls != 1 {
		t.Fatal("replayed approval")
	}
}

func TestExpiryDuringCredentialLookupDoesNotDispatch(t *testing.T) {
	m, p, _, calls, _ := awsFixture(t)
	approve(t, m, p)
	now := m.now()
	read := m.ReadSecret
	m.Now = func() time.Time { return now }
	m.ReadSecret = func(ref string) ([]byte, error) {
		now = now.Add(time.Minute)
		return read(ref)
	}
	if _, err := m.Execute(context.Background(), p.Request.ID); err == nil || *calls != 0 {
		t.Fatal("expired approval dispatched after slow credential lookup")
	}
	if err := m.unused(p.Request.ID); err != nil {
		t.Fatal("undispatched action consumed approval")
	}
}

func TestCredentialFailureClearsLoadedBuffers(t *testing.T) {
	m, p, _, calls, buffers := awsFixture(t)
	approve(t, m, p)
	read := m.ReadSecret
	m.ReadSecret = func(ref string) ([]byte, error) {
		if ref == "aws/secret" {
			return []byte("partial-secret"), fmt.Errorf("backend exposed private detail")
		}
		return read(ref)
	}
	_, err := m.Execute(context.Background(), p.Request.ID)
	if err == nil || strings.Contains(err.Error(), "private") || *calls != 0 {
		t.Fatalf("unsafe error: %v", err)
	}
	for _, buffer := range *buffers {
		if !bytes.Equal(buffer, make([]byte, len(buffer))) {
			t.Fatal("credential buffer retained")
		}
	}
}

type revisedGitHub struct{ github.Adapter }

func (a revisedGitHub) Descriptor() adapter.Descriptor {
	d := a.Adapter.Descriptor()
	d.Version = "2"
	return d
}

type customContentType struct{ github.Adapter }

func (a customContentType) Prepare(scope json.RawMessage, operation, resource string, input json.RawMessage) (adapter.Plan, error) {
	p, err := a.Adapter.Prepare(scope, operation, resource, input)
	p.HTTP.Headers["Content-Type"] = "application/vnd.github+json"
	return p, err
}

func TestDispatchPreservesDeclaredContentType(t *testing.T) {
	m, path, _, _ := fixture(t)
	m.Registry, _ = adapter.NewRegistry(customContentType{})
	m.Transport = transportFunc(func(r *http.Request) (*http.Response, error) {
		if r.Header.Get("Content-Type") != "application/vnd.github+json" {
			t.Fatal("dispatch replaced the adapter's declared content type")
		}
		return &http.Response{StatusCode: 201, Header: http.Header{}, Body: io.NopCloser(strings.NewReader(`{}`)), Request: r}, nil
	})
	p, err := m.Prepare(path, CreateDraft, "owner/repo", json.RawMessage(`{"title":"Review","head":"work","base":"main"}`))
	if err != nil {
		t.Fatal(err)
	}
	approve(t, m, p)
	if _, err := m.Execute(context.Background(), p.Request.ID); err != nil {
		t.Fatal(err)
	}
}

func TestAdapterUpgradeInvalidatesSchemaTwoApproval(t *testing.T) {
	m, path, reads, calls := fixture(t)
	c := Contract{SchemaVersion: 2, Name: "review", Adapter: "github", Credentials: map[string]string{"token": "github/token"},
		Actions: []string{ListIssues}, ApprovalTTLSeconds: 300, Scope: json.RawMessage(`{"repositories":["owner/repo"]}`)}
	data, _ := json.Marshal(c)
	if err := os.WriteFile(path, data, 0600); err != nil {
		t.Fatal(err)
	}
	p := prepare(t, m, path)
	approve(t, m, p)
	m.Registry, _ = adapter.NewRegistry(revisedGitHub{})
	if _, err := m.Review(p.Request.ID); err == nil {
		t.Fatal("review accepted a changed adapter")
	}
	if _, err := m.Execute(context.Background(), p.Request.ID); err == nil || reads.Load() != 0 || calls.Load() != 0 {
		t.Fatal("changed adapter reached credentials")
	}
}

func TestLegacyRequestEncodingPreserved(t *testing.T) {
	c := Contract{SchemaVersion: 1, Name: "review", Credential: "github/token", Repositories: []string{"owner/repo"},
		Actions: []string{ListIssues}, ApprovalTTLSeconds: 300}
	r := Request{ID: strings.Repeat("a", 32), CreatedAt: time.Date(2026, 11, 5, 12, 0, 0, 0, time.UTC),
		ContractPath: "contract.json", Contract: c, Action: ListIssues, Repository: "owner/repo", Input: json.RawMessage(`{}`),
		HTTP: HTTPRequest{Method: "GET", URL: "https://api.github.com/repos/owner/repo/issues?per_page=30&state=open"}}
	const legacy = `{"id":"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa","created_at":"2026-11-05T12:00:00Z","contract_path":"contract.json","contract":{"schema_version":1,"name":"review","credential":"github/token","repositories":["owner/repo"],"actions":["github.issues.list"],"approval_ttl_seconds":300},"action":"github.issues.list","repository":"owner/repo","input":{},"http":{"method":"GET","url":"https://api.github.com/repos/owner/repo/issues?per_page=30\u0026state=open"}}`
	data, err := json.Marshal(r)
	if err != nil || string(data) != legacy {
		t.Fatalf("legacy digest encoding changed:\n%s\n%v", data, err)
	}
}

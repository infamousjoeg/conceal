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
	"runtime"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

type transportFunc func(*http.Request) (*http.Response, error)

func (f transportFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func fixture(t *testing.T) (Manager, string, *atomic.Int32, *atomic.Int32) {
	t.Helper()
	root := t.TempDir()
	contract := Contract{1, "review", "github/token", []string{"owner/repo"}, []string{ListIssues, CreateDraft}, 300}
	path := filepath.Join(root, "contract.json")
	data, _ := json.Marshal(contract)
	if err := os.WriteFile(path, data, 0600); err != nil {
		t.Fatal(err)
	}
	reads, calls := new(atomic.Int32), new(atomic.Int32)
	m := Manager{
		Directory: filepath.Join(root, "state"),
		ReadSecret: func(id string) ([]byte, error) {
			reads.Add(1)
			if id != "github/token" {
				return nil, fmt.Errorf("wrong credential")
			}
			return []byte("fixture-token"), nil
		},
		Transport: transportFunc(func(r *http.Request) (*http.Response, error) {
			calls.Add(1)
			if r.URL.Scheme != "https" || r.URL.Host != "api.github.com" || r.Header.Get("Authorization") != "Bearer fixture-token" {
				return nil, fmt.Errorf("unexpected target or credential")
			}
			return &http.Response{StatusCode: 200, Header: http.Header{}, Body: io.NopCloser(strings.NewReader(`{"message":"fixture-token"}`)), Request: r}, nil
		}),
	}
	return m, path, reads, calls
}

func prepare(t *testing.T, m Manager, path string) Proposal {
	t.Helper()
	proposal, err := m.Prepare(path, ListIssues, "owner/repo", json.RawMessage(`{"limit":5}`))
	if err != nil {
		t.Fatal(err)
	}
	return proposal
}

func approve(t *testing.T, m Manager, proposal Proposal) {
	t.Helper()
	if _, err := m.Approve(proposal.Request.ID, proposal.Digest); err != nil {
		t.Fatal(err)
	}
}

func TestApprovedExecutionIsExactAndSingleUse(t *testing.T) {
	m, path, reads, calls := fixture(t)
	proposal := prepare(t, m, path)
	if proposal.Request.HTTP.URL != "https://api.github.com/repos/owner/repo/issues?per_page=5&state=open" {
		t.Fatalf("unexpected request: %#v", proposal.Request.HTTP)
	}
	approve(t, m, proposal)
	if reads.Load() != 0 || calls.Load() != 0 {
		t.Fatal("preparation or approval used credentials")
	}
	result, err := m.Execute(context.Background(), proposal.Request.ID)
	if err != nil {
		t.Fatal(err)
	}
	if result.Receipt.Outcome != "succeeded" || bytes.Contains(result.Response, []byte("fixture-token")) || !bytes.Contains(result.Response, []byte("[REDACTED]")) {
		t.Fatalf("unsafe or unsuccessful result: %#v", result)
	}
	if _, err := m.Execute(context.Background(), proposal.Request.ID); err == nil {
		t.Fatal("reused approval")
	}
	if _, err := m.Approve(proposal.Request.ID, proposal.Digest); err == nil {
		t.Fatal("reapproved spent request")
	}
	if reads.Load() != 1 || calls.Load() != 1 {
		t.Fatalf("reads=%d calls=%d", reads.Load(), calls.Load())
	}
	receipt, err := os.ReadFile(filepath.Join(m.Directory, proposal.Request.ID+".receipt.json"))
	if err != nil {
		t.Fatal(err)
	}
	if bytes.Contains(receipt, []byte("fixture-token")) || bytes.Contains(receipt, []byte("response")) {
		t.Fatal("receipt includes credential or response data")
	}
}

func TestDeniedRequestsNeverReadCredentials(t *testing.T) {
	for _, name := range []string{"unapproved", "wrong digest", "expired", "changed request", "changed contract", "revoked", "cancelled"} {
		t.Run(name, func(t *testing.T) {
			m, path, reads, calls := fixture(t)
			now := time.Now().UTC()
			m.Now = func() time.Time { return now }
			proposal := prepare(t, m, path)
			id := proposal.Request.ID
			ctx := context.Background()
			if name != "unapproved" && name != "wrong digest" {
				approve(t, m, proposal)
			}
			switch name {
			case "wrong digest":
				if _, err := m.Approve(id, strings.Repeat("0", 64)); err == nil {
					t.Fatal("approved wrong digest")
				}
			case "expired":
				now = now.Add(301 * time.Second)
			case "changed request":
				proposal.Request.Input = json.RawMessage(`{"limit":100}`)
				proposal.Request.HTTP, _ = Build(proposal.Request.Contract, ListIssues, "owner/repo", proposal.Request.Input)
				proposal.Digest = digest(proposal.Request)
				data, _ := json.Marshal(proposal)
				if err := os.WriteFile(filepath.Join(m.Directory, id+".request.json"), data, 0600); err != nil {
					t.Fatal(err)
				}
			case "changed contract":
				contract := proposal.Request.Contract
				contract.Credential = "another/token"
				data, _ := json.Marshal(contract)
				if err := os.WriteFile(path, data, 0600); err != nil {
					t.Fatal(err)
				}
			case "revoked":
				if _, err := m.Revoke(id); err != nil {
					t.Fatal(err)
				}
			case "cancelled":
				var cancel context.CancelFunc
				ctx, cancel = context.WithCancel(ctx)
				cancel()
			}
			if _, err := m.Execute(ctx, id); err == nil {
				t.Fatal("unsafe request executed")
			}
			if reads.Load() != 0 || calls.Load() != 0 {
				t.Fatal("denied request reached credentials or network")
			}
		})
	}
}

func TestNoRedirectOrRetryAndFailedApprovalIsSpent(t *testing.T) {
	for _, status := range []int{302, 307, 401, 500} {
		t.Run(fmt.Sprint(status), func(t *testing.T) {
			m, path, _, calls := fixture(t)
			m.Transport = transportFunc(func(r *http.Request) (*http.Response, error) {
				calls.Add(1)
				if r.URL.Host != "api.github.com" {
					t.Error("followed credential redirect")
				}
				return &http.Response{StatusCode: status, Header: http.Header{"Location": {"https://example.invalid/steal"}}, Body: io.NopCloser(strings.NewReader(`{}`)), Request: r}, nil
			})
			proposal := prepare(t, m, path)
			approve(t, m, proposal)
			result, err := m.Execute(context.Background(), proposal.Request.ID)
			if err == nil || result.Receipt.HTTPStatus != status {
				t.Fatalf("result=%#v err=%v", result, err)
			}
			if _, err := m.Execute(context.Background(), proposal.Request.ID); err == nil {
				t.Fatal("retried spent approval")
			}
			if calls.Load() != 1 {
				t.Fatalf("made %d requests", calls.Load())
			}
		})
	}
}

func TestTransportFailureIsRecordedAndNotRetried(t *testing.T) {
	m, path, _, calls := fixture(t)
	m.Transport = transportFunc(func(*http.Request) (*http.Response, error) { calls.Add(1); return nil, fmt.Errorf("fixture-token") })
	proposal := prepare(t, m, path)
	approve(t, m, proposal)
	result, err := m.Execute(context.Background(), proposal.Request.ID)
	if err == nil || strings.Contains(err.Error(), "fixture-token") || result.Receipt.Outcome != "transport_error" {
		t.Fatalf("result=%#v err=%v", result, err)
	}
	if _, err := m.Execute(context.Background(), proposal.Request.ID); err == nil {
		t.Fatal("retried uncertain outcome")
	}
	if calls.Load() != 1 {
		t.Fatal("unexpected retry")
	}
}

func TestConcurrentExecutionsConsumeOnlyOnce(t *testing.T) {
	m, path, reads, calls := fixture(t)
	proposal := prepare(t, m, path)
	approve(t, m, proposal)
	var group sync.WaitGroup
	var successes atomic.Int32
	for i := 0; i < 12; i++ {
		group.Add(1)
		go func() {
			defer group.Done()
			if _, err := m.Execute(context.Background(), proposal.Request.ID); err == nil {
				successes.Add(1)
			}
		}()
	}
	group.Wait()
	if successes.Load() != 1 || reads.Load() != 1 || calls.Load() != 1 {
		t.Fatal("approval was not consumed exactly once")
	}
}

func TestInvalidIDsAndDigestTampering(t *testing.T) {
	m, path, _, _ := fixture(t)
	for _, id := range []string{"../contract", "", `..\contract`, strings.Repeat("a", 33)} {
		if _, err := m.Inspect(id); err == nil {
			t.Fatal("accepted invalid request ID")
		}
	}
	proposal := prepare(t, m, path)
	proposal.Request.HTTP.URL = "https://example.invalid"
	data, _ := json.Marshal(proposal)
	if err := os.WriteFile(filepath.Join(m.Directory, proposal.Request.ID+".request.json"), data, 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := m.Inspect(proposal.Request.ID); err == nil {
		t.Fatal("accepted modified request")
	}
}

func TestOversizedProposalDoesNotCreateState(t *testing.T) {
	m, path, _, _ := fixture(t)
	input, _ := json.Marshal(map[string]string{
		"title": "Fixture", "head": "feature", "base": "main",
		"body": strings.Repeat("x", maxDocument/2),
	})
	if _, err := m.Prepare(path, CreateDraft, "owner/repo", input); err == nil {
		t.Fatal("stored a proposal too large to inspect")
	}
	entries, err := os.ReadDir(m.Directory)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 0 {
		t.Fatal("oversized proposal left partial state")
	}
}

func TestPrivateStateDirectory(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Windows directory privacy is controlled by ACLs")
	}
	m, path, _, _ := fixture(t)
	if err := os.Mkdir(m.Directory, 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(m.Directory, 0777); err != nil {
		t.Fatal(err)
	}
	if _, err := m.Prepare(path, ListIssues, "owner/repo", nil); err == nil {
		t.Fatal("accepted a shared approval directory")
	}
}

func TestCredentialFailuresDoNotConsumeOrDispatch(t *testing.T) {
	for _, value := range []string{"", "token\r\ninjected: header", "token\x00"} {
		m, path, _, calls := fixture(t)
		m.ReadSecret = func(string) ([]byte, error) { return []byte(value), nil }
		proposal := prepare(t, m, path)
		approve(t, m, proposal)
		if _, err := m.Execute(context.Background(), proposal.Request.ID); err == nil {
			t.Fatal("accepted invalid API token")
		}
		if calls.Load() != 0 {
			t.Fatal("invalid token reached transport")
		}
		if err := m.unused(proposal.Request.ID); err != nil {
			t.Fatal("pre-dispatch failure spent approval")
		}
	}
}

func TestResponseLimitConsumesApproval(t *testing.T) {
	m, path, _, calls := fixture(t)
	m.Transport = transportFunc(func(r *http.Request) (*http.Response, error) {
		calls.Add(1)
		return &http.Response{StatusCode: 200, Header: http.Header{}, Body: io.NopCloser(strings.NewReader(strings.Repeat("x", 4*maxDocument+1))), Request: r}, nil
	})
	proposal := prepare(t, m, path)
	approve(t, m, proposal)
	result, err := m.Execute(context.Background(), proposal.Request.ID)
	if err == nil || result.Receipt.Outcome != "response_error" || len(result.Response) != 0 {
		t.Fatal("oversized response was exposed")
	}
	if _, err := m.Execute(context.Background(), proposal.Request.ID); err == nil || calls.Load() != 1 {
		t.Fatal("response error allowed replay")
	}
}

package contracts

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"time"
)

const maxDocument = 1024 * 1024

var requestID = regexp.MustCompile(`^[a-f0-9]{32}$`)

type Request struct {
	ID           string          `json:"id"`
	CreatedAt    time.Time       `json:"created_at"`
	ContractPath string          `json:"contract_path"`
	Contract     Contract        `json:"contract"`
	Action       string          `json:"action"`
	Repository   string          `json:"repository"`
	Input        json.RawMessage `json:"input"`
	HTTP         HTTPRequest     `json:"http"`
}

type Proposal struct {
	Request Request `json:"request"`
	Digest  string  `json:"digest"`
}

type Approval struct {
	Digest     string    `json:"digest"`
	ApprovedAt time.Time `json:"approved_at"`
	ExpiresAt  time.Time `json:"expires_at"`
}

type Receipt struct {
	SchemaVersion int       `json:"schema_version"`
	ID            string    `json:"id"`
	Digest        string    `json:"digest"`
	Action        string    `json:"action"`
	Repository    string    `json:"repository"`
	Outcome       string    `json:"outcome"`
	HTTPStatus    int       `json:"http_status,omitempty"`
	Time          time.Time `json:"time"`
}

type Result struct {
	Receipt  Receipt         `json:"receipt"`
	Response json.RawMessage `json:"response,omitempty"`
}

// Manager's state directory belongs to the approving user or trusted broker.
// It must not be writable by an untrusted agent. Same-user unrestricted code is
// not an isolation boundary; use OS/sandbox separation when that is required.
type Manager struct {
	Directory  string
	ReadSecret func(string) ([]byte, error)
	Transport  http.RoundTripper
	Now        func() time.Time
}

func (m Manager) now() time.Time {
	if m.Now != nil {
		return m.Now().UTC()
	}
	return time.Now().UTC()
}

func digest(value any) string {
	data, _ := json.Marshal(value)
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}

func readDocument(path string, target any) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer func() { _ = file.Close() }()
	data, err := io.ReadAll(io.LimitReader(file, maxDocument+1))
	if err != nil {
		return err
	}
	if len(data) > maxDocument {
		return fmt.Errorf("document exceeds 1 MiB")
	}
	return Decode(data, target)
}

func LoadContract(path string) (Contract, error) {
	var contract Contract
	if err := readDocument(path, &contract); err != nil {
		return contract, err
	}
	return contract, contract.Validate()
}

func (m Manager) path(id, extension string) (string, error) {
	if !requestID.MatchString(id) {
		return "", fmt.Errorf("invalid request ID")
	}
	return filepath.Join(m.Directory, id+extension), nil
}

func (m Manager) initialize() error {
	if m.Directory == "" {
		return fmt.Errorf("approval state directory is required")
	}
	if err := os.MkdirAll(m.Directory, 0700); err != nil {
		return err
	}
	info, err := os.Lstat(m.Directory)
	if err != nil {
		return err
	}
	if !info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
		return fmt.Errorf("approval state must be a regular directory")
	}
	if runtime.GOOS != "windows" && info.Mode().Perm()&0077 != 0 {
		return fmt.Errorf("approval state directory must be private (mode 0700)")
	}
	return nil
}

func writeExclusive(path string, value any) error {
	data, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return err
	}
	if len(data)+1 > maxDocument {
		return fmt.Errorf("document exceeds 1 MiB")
	}
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	if err != nil {
		return err
	}
	_, writeErr := file.Write(append(data, '\n'))
	if writeErr == nil {
		writeErr = file.Sync()
	}
	closeErr := file.Close()
	return errors.Join(writeErr, closeErr)
}

func (m Manager) Prepare(contractPath, action, repo string, input json.RawMessage) (Proposal, error) {
	var proposal Proposal
	if err := m.initialize(); err != nil {
		return proposal, err
	}
	path, err := filepath.Abs(contractPath)
	if err != nil {
		return proposal, err
	}
	contract, err := LoadContract(path)
	if err != nil {
		return proposal, err
	}
	if len(input) == 0 {
		input = json.RawMessage(`{}`)
	}
	// Normalize JSON before hashing so approved data is stable across encoding.
	var compact bytes.Buffer
	if err := json.Compact(&compact, input); err != nil {
		return proposal, err
	}
	input = compact.Bytes()
	request, err := Build(contract, action, repo, input)
	if err != nil {
		return proposal, err
	}
	idBytes := make([]byte, 16)
	if _, err := rand.Read(idBytes); err != nil {
		return proposal, err
	}
	id := hex.EncodeToString(idBytes)
	proposal.Request = Request{id, m.now(), path, contract, action, repo, input, request}
	proposal.Digest = digest(proposal.Request)
	destination, _ := m.path(id, ".request.json")
	err = writeExclusive(destination, proposal)
	return proposal, err
}

func (m Manager) Inspect(id string) (Proposal, error) {
	var proposal Proposal
	path, err := m.path(id, ".request.json")
	if err != nil {
		return proposal, err
	}
	if err := readDocument(path, &proposal); err != nil {
		return proposal, err
	}
	if proposal.Request.ID != id || proposal.Digest != digest(proposal.Request) {
		return proposal, fmt.Errorf("request changed; prepare a new request")
	}
	contract, err := LoadContract(proposal.Request.ContractPath)
	if err != nil {
		return proposal, err
	}
	if digest(contract) != digest(proposal.Request.Contract) {
		return proposal, fmt.Errorf("contract changed; prepare a new request")
	}
	expected, err := Build(contract, proposal.Request.Action, proposal.Request.Repository, proposal.Request.Input)
	if err != nil {
		return proposal, err
	}
	if digest(expected) != digest(proposal.Request.HTTP) {
		return proposal, fmt.Errorf("request does not match its contract")
	}
	return proposal, nil
}

func (m Manager) lock(id string) (func(), error) {
	if err := m.initialize(); err != nil {
		return nil, err
	}
	path, err := m.path(id, ".lock")
	if err != nil {
		return nil, err
	}
	if err := os.Mkdir(path, 0700); err != nil {
		return nil, fmt.Errorf("request is busy or a previous operation was interrupted; prepare a new request: %w", err)
	}
	return func() { _ = os.Remove(path) }, nil
}

func (m Manager) unused(id string) error {
	path, _ := m.path(id, ".used")
	_, err := os.Lstat(path)
	if err == nil {
		return fmt.Errorf("approval already consumed or revoked")
	}
	if !os.IsNotExist(err) {
		return err
	}
	return nil
}

// Approve requires the exact digest the approving user reviewed.
func (m Manager) Approve(id, reviewedDigest string) (Approval, error) {
	var approval Approval
	unlock, err := m.lock(id)
	if err != nil {
		return approval, err
	}
	defer unlock()
	proposal, err := m.Inspect(id)
	if err != nil {
		return approval, err
	}
	if reviewedDigest != proposal.Digest {
		return approval, fmt.Errorf("reviewed digest does not match this request")
	}
	if err := m.unused(id); err != nil {
		return approval, err
	}
	now := m.now()
	if proposal.Request.CreatedAt.After(now) || now.Sub(proposal.Request.CreatedAt) > 24*time.Hour {
		return approval, fmt.Errorf("request is stale; prepare a new request")
	}
	approval = Approval{proposal.Digest, now, now.Add(time.Duration(proposal.Request.Contract.ApprovalTTLSeconds) * time.Second)}
	path, _ := m.path(id, ".approval.json")
	err = writeExclusive(path, approval)
	return approval, err
}

func (m Manager) saveReceipt(proposal Proposal, outcome string, status int) (Receipt, error) {
	receipt := Receipt{1, proposal.Request.ID, proposal.Digest, proposal.Request.Action, proposal.Request.Repository, outcome, status, m.now()}
	path, _ := m.path(proposal.Request.ID, ".receipt.json")
	return receipt, writeExclusive(path, receipt)
}

func (m Manager) Revoke(id string) (Receipt, error) {
	unlock, err := m.lock(id)
	if err != nil {
		return Receipt{}, err
	}
	defer unlock()
	proposal, err := m.Inspect(id)
	if err != nil {
		return Receipt{}, err
	}
	path, _ := m.path(id, ".used")
	if err := writeExclusive(path, struct{ Outcome string }{"revoked"}); err != nil {
		return Receipt{}, fmt.Errorf("request already consumed or revoked: %w", err)
	}
	return m.saveReceipt(proposal, "revoked", 0)
}

// Execute consumes approval before dispatching a request. No redirect or
// application-level retry is allowed; uncertain outcomes require a fresh review.
func (m Manager) Execute(ctx context.Context, id string) (Result, error) {
	var result Result
	unlock, err := m.lock(id)
	if err != nil {
		return result, err
	}
	defer unlock()
	proposal, err := m.Inspect(id)
	if err != nil {
		return result, err
	}
	if err := m.unused(id); err != nil {
		return result, err
	}
	var approval Approval
	path, _ := m.path(id, ".approval.json")
	if err := readDocument(path, &approval); err != nil {
		return result, fmt.Errorf("request requires approval: %w", err)
	}
	now := m.now()
	maxExpiry := approval.ApprovedAt.Add(time.Duration(proposal.Request.Contract.ApprovalTTLSeconds) * time.Second)
	if approval.Digest != proposal.Digest || !now.Before(approval.ExpiresAt) || approval.ApprovedAt.After(now) || approval.ExpiresAt.After(maxExpiry) {
		return result, fmt.Errorf("approval expired or does not match the exact request")
	}
	if err := ctx.Err(); err != nil {
		return result, err
	}
	if m.ReadSecret == nil {
		return result, fmt.Errorf("contract credential store unavailable")
	}
	credential, err := m.ReadSecret(proposal.Request.Contract.Credential)
	if err != nil {
		return result, fmt.Errorf("contract credential unavailable")
	}
	defer clear(credential)
	if len(credential) == 0 || strings.ContainsAny(string(credential), "\x00\r\n") {
		return result, fmt.Errorf("contract credential is not a valid API token")
	}
	call := proposal.Request.HTTP
	request, err := http.NewRequestWithContext(ctx, call.Method, call.URL, bytes.NewReader(call.Body))
	if err != nil {
		return result, err
	}
	request.Header.Set("Authorization", "Bearer "+string(credential))
	request.Header.Set("Accept", "application/vnd.github+json")
	request.Header.Set("X-GitHub-Api-Version", "2022-11-28")
	request.Header.Set("User-Agent", "Conceal")
	if len(call.Body) > 0 {
		request.Header.Set("Content-Type", "application/json")
	}
	// Disallow transport retries from rewinding an approved request body.
	request.GetBody = nil
	request.Close = true
	used, _ := m.path(id, ".used")
	if err := writeExclusive(used, struct{ Digest string }{proposal.Digest}); err != nil {
		return result, err
	}
	transport := m.Transport
	if transport == nil {
		// A fresh pool avoids automatic retries on stale reused connections.
		direct := http.DefaultTransport.(*http.Transport).Clone()
		direct.DisableKeepAlives = true
		defer direct.CloseIdleConnections()
		transport = direct
	}
	client := &http.Client{Transport: transport, Timeout: 30 * time.Second, CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}
	response, callErr := client.Do(request)
	if callErr != nil {
		result.Receipt, err = m.saveReceipt(proposal, "transport_error", 0)
		return result, errors.Join(fmt.Errorf("request failed; approval consumed; inspect the remote service before retrying"), err)
	}
	defer func() { _ = response.Body.Close() }()
	data, readErr := io.ReadAll(io.LimitReader(response.Body, 4*maxDocument+1))
	outcome := "succeeded"
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		outcome = "http_error"
	}
	if readErr != nil || len(data) > 4*maxDocument {
		outcome = "response_error"
	}
	result.Receipt, err = m.saveReceipt(proposal, outcome, response.StatusCode)
	if outcome == "response_error" {
		return result, errors.Join(fmt.Errorf("response unreadable or exceeds 4 MiB; approval consumed"), err)
	}
	encoded, _ := json.Marshal(string(credential))
	data = bytes.ReplaceAll(data, encoded[1:len(encoded)-1], []byte("[REDACTED]"))
	data = bytes.ReplaceAll(data, credential, []byte("[REDACTED]"))
	if json.Valid(data) {
		result.Response = data
	} else {
		result.Response, _ = json.Marshal(string(data))
	}
	if outcome != "succeeded" {
		return result, errors.Join(fmt.Errorf("GitHub returned HTTP %d; approval consumed", response.StatusCode), err)
	}
	return result, err
}

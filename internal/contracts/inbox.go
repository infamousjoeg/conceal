package contracts

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"github.com/infamousjoeg/conceal/pkg/conceal/adapter"
)

// ReviewDocument is a credential-free, versioned interface for approval clients.
// The client's approval must include Proposal.Digest from the displayed document.
type ReviewDocument struct {
	SchemaVersion int             `json:"schema_version"`
	Proposal      Proposal        `json:"proposal"`
	Review        adapter.Review  `json:"review"`
	Adapter       string          `json:"adapter"`
	Credentials   []adapter.Field `json:"credentials"`
	Status        string          `json:"status"`
	Approval      *Approval       `json:"approval,omitempty"`
	Receipt       *Receipt        `json:"receipt,omitempty"`
	CanApprove    bool            `json:"can_approve"`
	CanExecute    bool            `json:"can_execute"`
	CanReject     bool            `json:"can_reject"`
	ServerTime    time.Time       `json:"server_time"`
}

type InboxItem struct {
	ID        string    `json:"id"`
	Title     string    `json:"title"`
	Adapter   string    `json:"adapter"`
	Action    string    `json:"action"`
	Resource  string    `json:"resource"`
	CreatedAt time.Time `json:"created_at"`
	Status    string    `json:"status"`
	Problem   string    `json:"problem,omitempty"`
}

type Inbox struct {
	SchemaVersion int         `json:"schema_version"`
	Items         []InboxItem `json:"items"`
	Total         int         `json:"total"`
	Truncated     bool        `json:"truncated"`
}

func (m Manager) Review(id string) (ReviewDocument, error) {
	var view ReviewDocument
	if err := m.initialize(); err != nil {
		return view, err
	}
	proposal, err := m.Inspect(id)
	if err != nil {
		return view, err
	}
	plan, err := proposal.Request.Contract.plan(m.registry(), proposal.Request.Action, proposal.Request.target(), proposal.Request.Input)
	if err != nil {
		return view, err
	}
	view = ReviewDocument{
		SchemaVersion: 1, Proposal: proposal, Review: plan.Review, Adapter: plan.Adapter,
		Credentials: []adapter.Field{}, Status: "pending", ServerTime: m.now(), CanApprove: true, CanReject: true,
	}
	a, err := m.registry().Lookup(plan.Adapter)
	if err != nil {
		return view, err
	}
	references := proposal.Request.Contract.credentialReferences()
	for _, role := range a.Descriptor().Credentials {
		if name := references[role.Name]; name != "" {
			view.Credentials = append(view.Credentials, adapter.Field{Label: role.Description, Value: name})
		}
	}
	if lock, _ := m.path(id, ".lock"); pathExists(lock) {
		view.Status, view.CanApprove, view.CanReject = "busy", false, false
		return view, nil
	}
	if used, _ := m.path(id, ".used"); pathExists(used) {
		view.Status, view.CanApprove, view.CanReject = "consumed", false, false
		receiptPath, _ := m.path(id, ".receipt.json")
		if pathExists(receiptPath) {
			var receipt Receipt
			if err := readDocument(receiptPath, &receipt); err != nil {
				return view, fmt.Errorf("receipt cannot be read")
			}
			if receipt.ID != id || receipt.Digest != proposal.Digest {
				return view, fmt.Errorf("receipt does not match the request")
			}
			view.Receipt = &receipt
			if receipt.Outcome == "revoked" {
				view.Status = "rejected"
			}
		}
		return view, nil
	}
	approvalPath, _ := m.path(id, ".approval.json")
	if pathExists(approvalPath) {
		var approval Approval
		if err := readDocument(approvalPath, &approval); err != nil {
			return view, fmt.Errorf("approval cannot be read")
		}
		view.Approval, view.CanApprove = &approval, false
		maxExpiry := approval.ApprovedAt.Add(time.Duration(proposal.Request.Contract.ApprovalTTLSeconds) * time.Second)
		if approval.Digest != proposal.Digest || approval.ApprovedAt.After(view.ServerTime) || approval.ExpiresAt.After(maxExpiry) {
			return view, fmt.Errorf("approval does not match the request")
		}
		if !view.ServerTime.Before(approval.ExpiresAt) {
			view.Status = "expired"
		} else {
			view.Status, view.CanExecute = "approved", true
		}
	} else if proposal.Request.CreatedAt.After(view.ServerTime) || view.ServerTime.Sub(proposal.Request.CreatedAt) > 24*time.Hour {
		view.Status, view.CanApprove = "stale", false
	}
	return view, nil
}

func pathExists(path string) bool {
	_, err := os.Lstat(path)
	// Treat errors other than absence as present so the subsequent guarded read
	// fails closed; an unreadable approval must never look like a new proposal.
	return !os.IsNotExist(err)
}

// List returns recent proposal metadata, not credentials or remote responses.
func (m Manager) List(limit int) (Inbox, error) {
	result := Inbox{SchemaVersion: 1, Items: []InboxItem{}}
	if limit < 1 || limit > 200 {
		return result, fmt.Errorf("inbox limit must be between 1 and 200")
	}
	if err := m.initialize(); err != nil {
		return result, err
	}
	entries, err := os.ReadDir(m.Directory)
	if err != nil {
		return result, err
	}
	type candidate struct {
		id   string
		time time.Time
	}
	var candidates []candidate
	for _, entry := range entries {
		if !strings.HasSuffix(entry.Name(), ".request.json") || !entry.Type().IsRegular() {
			continue
		}
		id := strings.TrimSuffix(entry.Name(), ".request.json")
		if !requestID.MatchString(id) {
			continue
		}
		info, err := entry.Info()
		if err != nil {
			continue
		}
		candidates = append(candidates, candidate{id: id, time: info.ModTime()})
	}
	slices.SortFunc(candidates, func(a, b candidate) int {
		if order := b.time.Compare(a.time); order != 0 {
			return order
		}
		return strings.Compare(a.id, b.id)
	})
	result.Total = len(candidates)
	result.Truncated = len(candidates) > limit
	if result.Truncated {
		candidates = candidates[:limit]
	}
	for _, candidate := range candidates {
		view, err := m.Review(candidate.id)
		if err != nil {
			result.Items = append(result.Items, InboxItem{ID: candidate.id, Title: "Request cannot be reviewed", Status: "invalid", Problem: err.Error()})
			continue
		}
		result.Items = append(result.Items, InboxItem{
			ID: candidate.id, Title: view.Review.Title, Adapter: view.Adapter, Action: view.Proposal.Request.Action,
			Resource: view.Proposal.Request.target(), CreatedAt: view.Proposal.Request.CreatedAt, Status: view.Status,
		})
	}
	return result, nil
}

// StateDirectory exposes no configurable path to the frontend. It is shared by
// the CLI and desktop sidecar in the trusted operator's OS account.
func StateDirectory() (string, error) {
	root, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(root, "conceal", "actions"), nil
}

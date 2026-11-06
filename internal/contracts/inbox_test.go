package contracts

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestInboxStates(t *testing.T) {
	for _, state := range []string{"pending", "approved", "expired", "stale", "rejected", "busy", "invalid"} {
		t.Run(state, func(t *testing.T) {
			m, path, reads, calls := fixture(t)
			now := time.Now().UTC()
			m.Now = func() time.Time { return now }
			p := prepare(t, m, path)
			switch state {
			case "approved", "expired":
				approve(t, m, p)
				if state == "expired" {
					now = now.Add(time.Hour)
				}
			case "stale":
				now = now.Add(25 * time.Hour)
			case "rejected":
				if _, err := m.Revoke(p.Request.ID); err != nil {
					t.Fatal(err)
				}
			case "busy":
				unlock, err := m.lock(p.Request.ID)
				if err != nil {
					t.Fatal(err)
				}
				defer unlock()
			case "invalid":
				if err := os.WriteFile(filepath.Join(m.Directory, p.Request.ID+".request.json"), []byte(`{}`), 0600); err != nil {
					t.Fatal(err)
				}
			}
			inbox, err := m.List(100)
			if err != nil || len(inbox.Items) != 1 || inbox.Items[0].Status != state {
				t.Fatalf("%#v %v", inbox, err)
			}
			view, err := m.Review(p.Request.ID)
			if state == "invalid" {
				if err == nil {
					t.Fatal("invalid review accepted")
				}
			} else if err != nil || view.CanApprove != (state == "pending") || view.CanExecute != (state == "approved") {
				t.Fatalf("incorrect actions: %#v %v", view, err)
			}
			if reads.Load() != 0 || calls.Load() != 0 {
				t.Fatal("inbox accessed credentials or network")
			}
		})
	}
}

func TestInboxLimits(t *testing.T) {
	m, path, _, _ := fixture(t)
	prepare(t, m, path)
	prepare(t, m, path)
	list, err := m.List(1)
	if err != nil || len(list.Items) != 1 || list.Total != 2 || !list.Truncated {
		t.Fatalf("%#v %v", list, err)
	}
	for _, limit := range []int{-1, 0, 201} {
		if _, err := m.List(limit); err == nil {
			t.Fatal("invalid limit accepted")
		}
	}
}

package cmd

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestAdapterDiscoveryAndDesktopCommandsDoNotReadCredentials(t *testing.T) {
	store := &memoryStore{}
	catalog, _, err := execute(t, store, strings.NewReader(""), "action", "adapters")
	if err != nil || !strings.Contains(catalog, `"aws.s3.objects.list"`) || !strings.Contains(catalog, `"github.issues.list"`) {
		t.Fatalf("adapter discovery failed: %s %v", catalog, err)
	}
	root := t.TempDir()
	for _, id := range []string{"aws", "github"} {
		t.Run(id, func(t *testing.T) {
			example, _, err := execute(t, store, strings.NewReader(""), "action", "example", "--adapter", id)
			if err != nil {
				t.Fatal(err)
			}
			path := filepath.Join(root, id+".json")
			if err := os.WriteFile(path, []byte(example), 0600); err != nil {
				t.Fatal(err)
			}
			if _, _, err := execute(t, store, nil, "action", "validate", path); err != nil {
				t.Fatal(err)
			}
			operation, resource := "aws.s3.objects.list", "example-backups"
			if id == "github" {
				operation, resource = "github.issues.list", "OWNER/REPOSITORY"
			}
			base := []string{"action", "--state-dir", filepath.Join(root, id+"-state")}
			invoke := func(args ...string) string {
				t.Helper()
				out, _, err := execute(t, store, nil, append(append([]string{}, base...), args...)...)
				if err != nil {
					t.Fatal(err)
				}
				return out
			}
			proposed := invoke("prepare", "--contract", path, "--operation", operation, "--resource", resource)
			var result struct {
				Request struct {
					ID string `json:"id"`
				} `json:"request"`
			}
			if err := json.Unmarshal([]byte(proposed), &result); err != nil {
				t.Fatal(err)
			}
			review := invoke("review", result.Request.ID)
			inbox := invoke("inbox", "--limit", "10")
			if !strings.Contains(review, `"can_approve": true`) || !strings.Contains(inbox, `"status": "pending"`) {
				t.Fatal("missing desktop state")
			}
			invoke("revoke", result.Request.ID)
			if !strings.Contains(invoke("review", result.Request.ID), `"status": "rejected"`) {
				t.Fatal("rejection not reflected")
			}
		})
	}
	if len(store.reads) != 0 {
		t.Fatal("discovery or review retrieved a credential")
	}
}

package cmd

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/infamousjoeg/conceal/internal/contracts"
	"github.com/infamousjoeg/conceal/pkg/conceal/keychain"
)

type memoryStore struct {
	values  map[string][]byte
	reads   []string
	listErr error
}

func (s *memoryStore) List() ([]keychain.QueryResult, error) {
	if s.listErr != nil {
		return nil, s.listErr
	}
	items := []keychain.QueryResult{}
	for name := range s.values {
		items = append(items, keychain.QueryResult{Account: name})
	}
	return items, nil
}
func (s *memoryStore) ReadSecret(name string) ([]byte, error) {
	s.reads = append(s.reads, name)
	if value, ok := s.values[name]; ok {
		return bytes.Clone(value), nil
	}
	return nil, errors.New("missing fixture")
}
func (s *memoryStore) AddSecret(name string, value []byte) error {
	if s.values == nil {
		s.values = make(map[string][]byte)
	}
	if _, ok := s.values[name]; ok {
		return errors.New("duplicate fixture")
	}
	s.values[name] = bytes.Clone(value)
	return nil
}
func (s *memoryStore) UpdateSecret(name string, value []byte) error {
	if _, ok := s.values[name]; !ok {
		return errors.New("missing fixture")
	}
	s.values[name] = bytes.Clone(value)
	return nil
}
func (s *memoryStore) DeleteSecret(name string) error {
	if _, ok := s.values[name]; !ok {
		return errors.New("missing fixture")
	}
	delete(s.values, name)
	return nil
}
func execute(t *testing.T, store secretStore, input io.Reader, args ...string) (string, string, error) {
	t.Helper()
	root := newRootCommand(store)
	var out, stderr bytes.Buffer
	root.SetOut(&out)
	root.SetErr(&stderr)
	root.SetIn(input)
	root.SetArgs(args)
	err := root.Execute()
	return out.String(), stderr.String(), err
}
func statusCode(err error) int {
	if err == nil {
		return 0
	}
	var status *statusError
	if errors.As(err, &status) {
		return status.code
	}
	return 1
}

func TestSecretInputPreservesContent(t *testing.T) {
	for _, test := range []struct {
		name, input, want string
		raw               bool
	}{
		{"EOF without newline", "token", "token", false},
		{"terminal newline", "token\n", "token", false},
		{"Windows newline", "token\r\n", "token", false},
		{"multiline whitespace", "  first\nsecond \n\n", "  first\nsecond \n", false},
		{"raw", "first\r\nsecond\r\n", "first\r\nsecond\r\n", true},
		{"binary", "zero\x00byte", "zero\x00byte", false},
		{"empty", "", "", false},
	} {
		t.Run(test.name, func(t *testing.T) {
			store := &memoryStore{}
			args := []string{"set", "fixture", "--non-interactive"}
			if test.raw {
				args = append(args, "--raw")
			}
			_, stderr, err := execute(t, store, strings.NewReader(test.input), args...)
			if err != nil || stderr != "" {
				t.Fatalf("err=%v stderr=%q", err, stderr)
			}
			if string(store.values["fixture"]) != test.want {
				t.Fatal("stored content changed")
			}
			_, _, err = execute(t, store, strings.NewReader("changed"), "update", "fixture")
			if err != nil || string(store.values["fixture"]) != "changed" {
				t.Fatal("update failed")
			}
		})
	}
}

func TestInputErrorsDoNotWrite(t *testing.T) {
	for _, test := range []struct {
		args  []string
		input string
	}{
		{[]string{"set"}, "value"},
		{[]string{"set", "fixture", "extra"}, "value"},
		{[]string{"set", "bad\nname"}, "value"},
		{[]string{"set", "fixture"}, strings.Repeat("x", maxInputSize+1)},
		{[]string{"update", "missing"}, "value"},
	} {
		store := &memoryStore{}
		_, _, err := execute(t, store, strings.NewReader(test.input), test.args...)
		if err == nil || len(store.values) != 0 {
			t.Fatalf("unexpected write for %q", test.args)
		}
	}
}

func TestDiscoveryNeverReadsValues(t *testing.T) {
	store := &memoryStore{values: map[string][]byte{"z": []byte("not-output"), "a": []byte("also-not-output")}}
	out, _, err := execute(t, store, strings.NewReader(""), "list", "--json")
	if err != nil {
		t.Fatal(err)
	}
	var list struct {
		Schema  int                    `json:"schema_version"`
		Secrets []keychain.QueryResult `json:"secrets"`
	}
	if err := json.Unmarshal([]byte(out), &list); err != nil {
		t.Fatal(err)
	}
	if list.Schema != 1 || len(list.Secrets) != 2 || list.Secrets[0].Account != "a" {
		t.Fatal("unsorted or invalid discovery")
	}
	for _, test := range []struct {
		ids  []string
		code int
	}{
		{[]string{"a", "z"}, 0}, {[]string{"a", "missing"}, 2},
	} {
		out, _, err := execute(t, store, strings.NewReader(""), append([]string{"check", "--json"}, test.ids...)...)
		if statusCode(err) != test.code {
			t.Fatalf("status=%d want=%d", statusCode(err), test.code)
		}
		var result preflight
		if err := json.Unmarshal([]byte(out), &result); err != nil {
			t.Fatal(err)
		}
		if result.Ready != (test.code == 0) || len(result.Secrets) != len(test.ids) {
			t.Fatal("invalid preflight")
		}
	}
	store.listErr = errors.New("sensitive backend diagnostic")
	out, _, err = execute(t, store, strings.NewReader(""), "check", "--json", "a")
	if statusCode(err) != 3 || strings.Contains(out, "sensitive") {
		t.Fatal("unsafe backend failure")
	}
	var unavailable preflight
	if err := json.Unmarshal([]byte(out), &unavailable); err != nil {
		t.Fatal(err)
	}
	if len(unavailable.Secrets) != 0 || unavailable.Error == "" {
		t.Fatal("backend failure misreported as missing secrets")
	}
	if len(store.reads) != 0 {
		t.Fatal("discovery retrieved credentials")
	}
}

func TestProviderProtocolIsExact(t *testing.T) {
	value := []byte(" raw\x00value\r\n")
	store := &memoryStore{values: map[string][]byte{"fixture": value}}
	var out bytes.Buffer
	if err := executeProvider(store, []string{"fixture"}, &out); err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(out.Bytes(), value) {
		t.Fatal("provider altered bytes")
	}
	out.Reset()
	if err := executeProvider(store, []string{"--version"}, &out); err != nil {
		t.Fatal(err)
	}
	if out.String() != versionString()+"\n" || len(store.reads) != 1 {
		t.Fatal("version protocol read a secret")
	}
	for _, args := range [][]string{nil, {"fixture", "extra"}, {"missing"}} {
		if err := executeProvider(store, args, io.Discard); err == nil {
			t.Fatalf("accepted invalid args %q", args)
		}
	}
}

func TestProviderInstallation(t *testing.T) {
	dir := t.TempDir()
	source := filepath.Join(dir, "source")
	if err := os.WriteFile(source, []byte("first binary"), 0700); err != nil {
		t.Fatal(err)
	}
	destination := filepath.Join(dir, "providers")
	target, err := installProvider(source, destination, "conceal_summon", false)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := installProvider(source, destination, "conceal_summon", false); err == nil {
		t.Fatal("overwrote existing provider")
	}
	if err := os.WriteFile(source, []byte("replacement binary"), 0700); err != nil {
		t.Fatal(err)
	}
	if _, err := installProvider(source, destination, "conceal_summon", true); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(target)
	if err != nil || string(data) != "replacement binary" {
		t.Fatal("replacement did not install")
	}
	if _, err := installProvider(target, destination, "conceal_summon", true); err == nil {
		t.Fatal("replaced source with itself")
	}
	if err := os.Mkdir(filepath.Join(destination, "directory"), 0700); err != nil {
		t.Fatal(err)
	}
	if _, err := installProvider(source, destination, "directory", true); err == nil {
		t.Fatal("replaced a directory")
	}
	files, err := os.ReadDir(destination)
	if err != nil {
		t.Fatal(err)
	}
	for _, file := range files {
		if strings.HasPrefix(file.Name(), ".conceal-provider-") {
			t.Fatal("temporary installer left behind")
		}
	}
}

// The test binary acts as a portable child, avoiding shell-specific behavior.
func TestRunChild(t *testing.T) {
	if os.Getenv("CONCEAL_FIXTURE_CHILD") != "yes" {
		return
	}
	// os.Exit runs coverage hooks even when --clean-env removes GOCOVERDIR.
	// Keep helper coverage in a directory owned and cleaned by the parent test.
	if err := os.Setenv("GOCOVERDIR", os.Args[len(os.Args)-2]); err != nil {
		os.Exit(2)
	}
	mode := os.Args[len(os.Args)-1]
	switch mode {
	case "output":
		value := os.Getenv("CONCEAL_FIXTURE_VALUE")
		_, _ = fmt.Fprint(os.Stdout, "before:")
		for _, char := range value {
			_, _ = fmt.Fprint(os.Stdout, string(char))
		}
		_, _ = fmt.Fprint(os.Stdout, ":after\n")
		_, _ = fmt.Fprint(os.Stderr, value)
		if os.Getenv("CONCEAL_FIXTURE_PARENT") != "" {
			_, _ = fmt.Fprint(os.Stdout, "inherited\n")
		}
		input, _ := io.ReadAll(os.Stdin)
		_, _ = fmt.Fprint(os.Stdout, string(input))
	case "exit":
		os.Exit(7)
	case "wait":
		time.Sleep(10 * time.Second)
	}
	os.Exit(0)
}

func TestRunScopesSecretsAndRedactsOutput(t *testing.T) {
	t.Setenv("CONCEAL_FIXTURE_PARENT", "present")
	executable, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	for _, clean := range []bool{false, true} {
		for _, unmasked := range []bool{false, true} {
			store := &memoryStore{values: map[string][]byte{
				"child-mode": []byte("yes"), "selected": []byte("fixture-秘密-token"), "unselected": []byte("must-not-read"),
			}}
			args := []string{"run", "--env", "CONCEAL_FIXTURE_CHILD=child-mode", "--env", "CONCEAL_FIXTURE_VALUE=selected"}
			if clean {
				args = append(args, "--clean-env")
			}
			if unmasked {
				args = append(args, "--no-redact")
			}
			args = append(args, "--", executable, "-test.run=^TestRunChild$", "--", t.TempDir(), "output")
			out, stderr, err := execute(t, store, strings.NewReader("input-forwarded"), args...)
			if err != nil {
				t.Fatal(err)
			}
			expected := "[REDACTED]"
			if unmasked {
				expected = "fixture-秘密-token"
			}
			if !strings.Contains(out, "before:"+expected+":after\n") || stderr != expected || !strings.HasSuffix(out, "input-forwarded") {
				t.Fatalf("unexpected child output: %q %q", out, stderr)
			}
			if strings.Contains(out, "inherited") == clean {
				t.Fatal("incorrect parent environment inheritance")
			}
			if !reflect.DeepEqual(store.reads, []string{"child-mode", "selected"}) {
				t.Fatal("read undeclared secret")
			}
		}
	}
}

func TestRunExitAndTimeout(t *testing.T) {
	executable, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	for _, test := range []struct {
		mode, timeout string
		code          int
	}{
		{"exit", "0", 7}, {"wait", "100ms", 124},
	} {
		store := &memoryStore{values: map[string][]byte{"child-mode": []byte("yes")}}
		_, _, err := execute(t, store, strings.NewReader(""), "run", "--timeout", test.timeout, "--env", "CONCEAL_FIXTURE_CHILD=child-mode", "--", executable, "-test.run=^TestRunChild$", "--", t.TempDir(), test.mode)
		if statusCode(err) != test.code {
			t.Fatalf("%s status=%d want=%d, err=%v", test.mode, statusCode(err), test.code, err)
		}
	}
}

func TestRunInvalidBindingsDoNotReadSecrets(t *testing.T) {
	for _, args := range [][]string{
		{"run", "--", "not-a-real-program"},
		{"run", "--env", "BAD-NAME=fixture", "--", "not-a-real-program"},
		{"run", "--env", "A=fixture", "--env", "A=other", "--", "not-a-real-program"},
		{"run", "--env", "A=fixture", "not-a-real-program"},
		{"run", "--env", "A=fixture", "--timeout=-1s", "--", "not-a-real-program"},
	} {
		store := &memoryStore{}
		_, _, err := execute(t, store, strings.NewReader(""), args...)
		if err == nil || len(store.reads) != 0 {
			t.Fatalf("accepted invalid invocation %q", args)
		}
	}
	if runtime.GOOS == "windows" {
		if _, err := parseBindings([]string{"Token=a", "TOKEN=b"}); err == nil {
			t.Fatal("accepted case-variant Windows bindings")
		}
	}
}

func TestCancelledRun(t *testing.T) {
	executable, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	store := &memoryStore{values: map[string][]byte{"fixture": []byte("yes")}}
	root := newRootCommand(store)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	root.SetContext(ctx)
	root.SetOut(io.Discard)
	root.SetErr(io.Discard)
	root.SetIn(strings.NewReader(""))
	root.SetArgs([]string{"run", "--env", "CONCEAL_FIXTURE_CHILD=fixture", "--", executable, "-test.run=^TestRunChild$", "--", t.TempDir(), "wait"})
	if statusCode(root.Execute()) != 130 {
		t.Fatal("cancelled run did not propagate interruption")
	}
	if len(store.reads) != 0 {
		t.Fatal("cancelled run retrieved credentials")
	}
}

func TestActionCLIUsesSelectedStateAndNoCredentialsBeforeExecution(t *testing.T) {
	store := &memoryStore{}
	directory := t.TempDir()
	contractPath := filepath.Join(directory, "contract.json")
	example, _, err := execute(t, store, strings.NewReader(""), "action", "example")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(contractPath, []byte(example), 0600); err != nil {
		t.Fatal(err)
	}
	base := []string{"action", "--state-dir", filepath.Join(directory, "state")}
	invoke := func(args ...string) (string, error) {
		t.Helper()
		argv := append(append([]string{}, base...), args...)
		out, _, err := execute(t, store, strings.NewReader(""), argv...)
		return out, err
	}
	if _, err := invoke("validate", contractPath); err != nil {
		t.Fatal(err)
	}
	out, err := invoke("prepare", "--contract", contractPath, "--operation", contracts.ListIssues, "--repo", "OWNER/REPOSITORY")
	if err != nil {
		t.Fatal(err)
	}
	var proposal contracts.Proposal
	if err := json.Unmarshal([]byte(out), &proposal); err != nil {
		t.Fatal(err)
	}
	id := proposal.Request.ID
	inspected, err := invoke("inspect", id)
	if err != nil || inspected != out {
		t.Fatal("inspect returned a different proposal")
	}
	if _, err := invoke("execute", id); err == nil {
		t.Fatal("executed without approval")
	}
	if _, err := invoke("approve", id); err == nil {
		t.Fatal("approved without explicit digest")
	}
	if _, err := invoke("approve", id, "--digest", proposal.Digest); err != nil {
		t.Fatal(err)
	}
	if _, err := invoke("revoke", id); err != nil {
		t.Fatal(err)
	}
	if _, err := invoke("execute", id); err == nil {
		t.Fatal("executed revoked request")
	}
	if len(store.reads) != 0 {
		t.Fatal("action preflight or rejected execution retrieved a secret")
	}
}

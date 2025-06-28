package conceal

import (
	"io"
	"os"
	"strings"
	"testing"
)

func captureOutput(f func()) string {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	f()
	_ = w.Close()
	os.Stdout = old
	out, _ := io.ReadAll(r)
	return string(out)
}

func TestPrintSuccess(t *testing.T) {
	out := captureOutput(func() { PrintSuccess("hi %s", "joe") })
	if !strings.Contains(out, "✅ hi joe") {
		t.Errorf("unexpected output: %q", out)
	}
}

func TestPrintFailure(t *testing.T) {
	out := captureOutput(func() { PrintFailure("bad") })
	if !strings.Contains(out, "❌ bad") {
		t.Errorf("unexpected output: %q", out)
	}
}

func TestPrintInfo(t *testing.T) {
	out := captureOutput(func() { PrintInfo("info") })
	if !strings.Contains(out, "ℹ️ info") {
		t.Errorf("unexpected output: %q", out)
	}
}

func TestGetSecretNameArgs(t *testing.T) {
	if name := GetSecretName([]string{"foo"}); name != "foo" {
		t.Fatalf("expected foo got %s", name)
	}
}

func TestGetSecretNamePrompt(t *testing.T) {
	old := os.Stdin
	tmp, err := os.CreateTemp("", "stdin")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.Remove(tmp.Name()) }()
	if _, err := tmp.WriteString("bar\n"); err != nil {
		t.Fatal(err)
	}
	if _, err := tmp.Seek(0, 0); err != nil {
		t.Fatal(err)
	}
	os.Stdin = tmp
	name := GetSecretName(nil)
	os.Stdin = old
	if name != "bar" {
		t.Fatalf("expected bar got %s", name)
	}
}

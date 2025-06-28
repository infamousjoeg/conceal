package migrate

import "testing"

func TestFilter(t *testing.T) {
	src := map[string][]byte{"a/b": {}, "c/d": {}}
	out := Filter(src, "a/*")
	if len(out) != 1 || out["a/b"] == nil {
		t.Fatalf("unexpected result: %v", out)
	}
}

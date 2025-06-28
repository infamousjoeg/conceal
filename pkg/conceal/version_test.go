package conceal

import "testing"

func TestFullVersionName(t *testing.T) {
	expected := Version + "-" + Tag
	if FullVersionName != expected {
		t.Fatalf("expected %s got %s", expected, FullVersionName)
	}
}

// Package adaptertest provides the shared acceptance suite for adapter authors.
// It checks preparation through the same registry validation used by Conceal.
// Tests cannot prove absence of I/O; review the adapter source as trusted code.
package adaptertest

import (
	"bytes"
	"encoding/json"
	"reflect"
	"testing"

	"github.com/infamousjoeg/conceal/pkg/conceal/adapter"
)

type Case struct {
	Name      string
	Scope     json.RawMessage
	Operation string
	Resource  string
	Input     json.RawMessage
	Reject    bool
	Check     func(*testing.T, adapter.Plan)
}

// Run verifies registration, deterministic preparation, immutable inputs, host
// validation, and the supplied positive and negative policy cases.
func Run(t *testing.T, implementation adapter.Adapter, cases []Case) {
	t.Helper()
	registry, err := adapter.NewRegistry(implementation)
	if err != nil {
		t.Fatal(err)
	}
	if len(cases) == 0 {
		t.Fatal("conformance requires policy cases")
	}
	for _, test := range cases {
		t.Run(test.Name, func(t *testing.T) {
			scope, input := bytes.Clone(test.Scope), bytes.Clone(test.Input)
			plan, err := registry.Prepare(implementation.Descriptor().ID, test.Scope, test.Operation, test.Resource, test.Input)
			if !bytes.Equal(scope, test.Scope) || !bytes.Equal(input, test.Input) {
				t.Fatal("adapter mutated its inputs")
			}
			if test.Reject {
				if err == nil {
					t.Fatal("out-of-policy input was accepted")
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			again, err := registry.Prepare(implementation.Descriptor().ID, scope, test.Operation, test.Resource, input)
			if err != nil || !reflect.DeepEqual(plan, again) {
				t.Fatal("preparation is not deterministic")
			}
			if test.Check != nil {
				test.Check(t, plan)
			}
		})
	}
}

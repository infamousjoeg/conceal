// Package adapters registers the reviewed adapters shipped in this executable.
package adapters

import (
	"github.com/infamousjoeg/conceal/pkg/conceal/adapter"
	"github.com/infamousjoeg/conceal/pkg/conceal/adapters/aws"
	"github.com/infamousjoeg/conceal/pkg/conceal/adapters/github"
)

func NewRegistry() *adapter.Registry {
	registry, err := adapter.NewRegistry(github.Adapter{}, aws.Adapter{})
	if err != nil {
		panic(err) // An invalid compiled declaration is a programming error.
	}
	return registry
}

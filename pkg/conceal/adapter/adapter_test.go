package adapter_test

import (
	"encoding/json"
	"testing"

	"github.com/infamousjoeg/conceal/pkg/conceal/adapter"
	"github.com/infamousjoeg/conceal/pkg/conceal/adapters/github"
)

type altered struct {
	github.Adapter
	descriptor func(*adapter.Descriptor)
	plan       func(*adapter.Plan)
}

func (a altered) Descriptor() adapter.Descriptor {
	d := a.Adapter.Descriptor()
	if a.descriptor != nil {
		a.descriptor(&d)
	}
	return d
}

func (a altered) Prepare(s json.RawMessage, op, resource string, input json.RawMessage) (adapter.Plan, error) {
	p, err := a.Adapter.Prepare(s, op, resource, input)
	if a.plan != nil {
		a.plan(&p)
	}
	return p, err
}

func TestHostRejectsUnsafePlans(t *testing.T) {
	for name, mutate := range map[string]func(*adapter.Plan){
		"HTTP":           func(p *adapter.Plan) { p.HTTP.URL = "http://api.github.com/" },
		"another host":   func(p *adapter.Plan) { p.HTTP.URL = "https://api.github.com.evil.invalid/" },
		"userinfo":       func(p *adapter.Plan) { p.HTTP.URL = "https://secret@api.github.com/" },
		"port":           func(p *adapter.Plan) { p.HTTP.URL = "https://api.github.com:443/" },
		"fragment":       func(p *adapter.Plan) { p.HTTP.URL += "#hidden" },
		"auth header":    func(p *adapter.Plan) { p.HTTP.Headers["Authorization"] = "Bearer bad" },
		"header newline": func(p *adapter.Plan) { p.HTTP.Headers["Accept"] = "text/plain\r\nX-Bad: yes" },
		"header alias":   func(p *adapter.Plan) { p.HTTP.Headers["accept"] = "text/plain" },
		"method":         func(p *adapter.Plan) { p.HTTP.Method = "DELETE" },
		"GET body":       func(p *adapter.Plan) { p.HTTP.Body = json.RawMessage(`{}`) },
		"readonly POST":  func(p *adapter.Plan) { p.HTTP.Method = "POST" },
		"auth scheme":    func(p *adapter.Plan) { p.Authentication.Scheme = "aws-sigv4" },
		"signing":        func(p *adapter.Plan) { p.Authentication.Region = "us-east-1" },
		"no review":      func(p *adapter.Plan) { p.Review = adapter.Review{} },
	} {
		t.Run(name, func(t *testing.T) {
			registry, err := adapter.NewRegistry(altered{plan: mutate})
			if err != nil {
				t.Fatal(err)
			}
			_, err = registry.Prepare("github", json.RawMessage(`{"repositories":["owner/repo"]}`), github.ListIssues, "owner/repo", nil)
			if err == nil {
				t.Fatal("unsafe plan accepted")
			}
		})
	}
}

func TestInvalidDescriptors(t *testing.T) {
	for name, mutate := range map[string]func(*adapter.Descriptor){
		"API":            func(d *adapter.Descriptor) { d.APIVersion++ },
		"empty host":     func(d *adapter.Descriptor) { d.AllowedHosts = []string{""} },
		"wildcard":       func(d *adapter.Descriptor) { d.AllowedHosts = []string{"*.github.com"} },
		"no token":       func(d *adapter.Descriptor) { d.Credentials = nil },
		"optional token": func(d *adapter.Descriptor) { d.Credentials[0].Required = false },
		"extra role": func(d *adapter.Descriptor) {
			d.Credentials = append(d.Credentials, adapter.Credential{Name: "other", Description: "Other"})
		},
		"bearer signer":       func(d *adapter.Descriptor) { d.SigningService = "s3" },
		"operation namespace": func(d *adapter.Descriptor) { d.Operations[0].Name = "other.list" },
	} {
		t.Run(name, func(t *testing.T) {
			if _, err := adapter.NewRegistry(altered{descriptor: mutate}); err == nil {
				t.Fatal("invalid adapter accepted")
			}
		})
	}
}

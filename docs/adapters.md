# Add a service to Conceal

[Back to Conceal](../README.md) · [Contributing](../CONTRIBUTING.md) · [Desktop inbox](desktop.md)

An adapter turns a small, useful service operation into something a person can review and approve. Start with one task: list backups, open a draft, or inspect a deployment. Give it a clear resource boundary and describe the consequence in ordinary language.

**Adapters prepare data. The shared engine handles credentials, approval, signing, dispatch, and receipts.** Contributors do not need to build another approval screen.

## Choose a starting point

| Example | Use it to learn |
| --- | --- |
| [Minimal starter](../examples/adapter/adapter.go) and [tests](../examples/adapter/adapter_test.go) | The complete interface for a bounded bearer-authenticated GET |
| [GitHub reference](../pkg/conceal/adapters/github/github.go) | Read operations, a draft-creation operation, and complete review fields |
| [AWS S3 adapter](../pkg/conceal/adapters/aws/aws.go) | Multiple credential roles, prefix scope, and host-owned SigV4 signing |

Run the examples without credentials or network calls:

```sh
go test ./examples/adapter ./pkg/conceal/adapters/... ./internal/contracts
```

The starter uses `api.example.invalid` and is not installed in the production registry.

## The public SDK

Import `github.com/infamousjoeg/conceal/pkg/conceal/adapter`. SDK API version **1** has three methods:

```go
type Adapter interface {
    Descriptor() Descriptor
    ValidateScope(json.RawMessage) error
    Prepare(scope json.RawMessage, operation, resource string, input json.RawMessage) (Plan, error)
}
```

1. **Describe the adapter.** Declare a stable ID, adapter behavior version, explicit HTTPS hosts, authentication scheme, credential roles, operations, and JSON schemas. Operation names begin with the adapter ID.
2. **Validate scope.** Reject unknown fields and ambiguous/duplicate JSON with `adapter.Decode`. Require explicit resource choices. Treat the JSON schema as discovery metadata; validate every rule in Go too.
3. **Prepare deterministically.** Return the exact method, URL, body, and stable headers, plus a title, consequence, and every decision-relevant field. No network, credentials, filesystem, environment, clock, or random access.

`conceal action adapters` returns the installed descriptors as JSON. An agent can discover supported inputs without seeing a credential. `action example --adapter github` and `--adapter aws` print ready-to-customize schema-2 contracts.

### What the engine accepts

- Explicit declared HTTPS hosts; no caller-supplied endpoints, userinfo, fragments, ports, or wildcards.
- GET without a body, or POST. Read-only operations must use GET.
- Stable `Accept`, `Content-Type`, and `X-GitHub-Api-Version` headers only. Authentication headers come from the engine.
- JSON bodies without a declared content type receive `application/json` during preparation, so the default is included in the schema-2 approval.
- `bearer` authentication with a required `token` role; or `aws-sigv4` with required `access_key_id`/`secret_access_key` and optional `session_token`. Signing currently supports S3's explicitly declared regional endpoints.
- A complete review document. Include the destination/resource, all user input, and side effects. Do not hide fields behind reassuring summaries.

If another service needs a new authentication scheme, method, or header, propose a focused core change and tests. Do not run a CLI, invoke a shell, accept arbitrary HTTP requests, or tunnel a broader operation through an existing one.

## Register and test your adapter

1. Copy the starter to `pkg/conceal/adapters/<service>/` and rename its package/ID/operation.
2. Replace the placeholder destination and semantics with the service's documented API. State the least service-side permissions needed.
3. Add it to the explicit registry in [`internal/adapters/registry.go`](../internal/adapters/registry.go).
4. Use the shared [`adaptertest.Run`](../pkg/conceal/adapter/adaptertest/conformance.go) harness. Include valid requests, resource/prefix escapes, unknown and duplicate input, limit boundaries, and forbidden operations. Assert the exact request and complete review.
5. Add engine-level tests when introducing authentication or execution behavior. Use a fake transport and synthetic credentials. Prove that denial occurs before credential access and that approval permits one dispatch.
6. Add a contract/input example and a short user walkthrough. Check its review in the desktop. No frontend service-specific component should be required.

The harness checks registration, host validation, determinism, immutable inputs, and your policy cases. Tests cannot prove that trusted adapter code never performs I/O; code review remains required.

## Versioning and compatibility

SDK API version and adapter behavior version are separate. Increment `Descriptor.Version` whenever request construction, allowed destinations, defaults, authentication, scope, or review meaning changes. Schema-2 proposals bind the entire prepared plan and adapter version. An upgrade that changes either invalidates the old request; prepare and review it again.

Schema-1 GitHub contracts remain supported with their original request serialization. New integrations should use schema 2:

```json
{
  "schema_version": 2,
  "name": "example-review",
  "actions": ["example.items.list"],
  "approval_ttl_seconds": 300,
  "adapter": "example",
  "credentials": { "token": "example/token" },
  "scope": { "projects": ["my-project"] }
}
```

The engine binds credential **names**, not secret values, so credentials can rotate without putting values into a proposal.

## Contribution and trust model

This release uses **reviewed adapters compiled into Conceal**. There is no runtime plugin loader, marketplace, dynamic binary download, or adapter subprocess protocol. The SDK is public so contributors can build and test adapters in their own Go projects before proposing inclusion.

An interface is not a sandbox. Installed Go code is trusted code and could access the process's privileges. Review dependencies and source, keep adapters narrow, and never describe a third-party adapter as safe merely because it implements the interface. A future isolated extension protocol would be a separate security design.

Open a contribution with the user task, API documentation, exact permissions, sample review, tests, and known limits. See [Contributing](../CONTRIBUTING.md) for the normal review process.

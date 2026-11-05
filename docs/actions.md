# Action contracts and exact approvals

Conceal's action adapter lets an agent request a specific operation without receiving the credential used to perform it. A contract limits the repositories and operations. An approval authorizes one exact request for a short period. Conceal checks both before retrieving the credential and dispatching the API request.

Explore the behavior in the [playground](playground.html). It is a simulation, requires no credentials, and makes no network requests.

## What ships in this candidate

| Operation | Allowed inputs | GitHub request |
| --- | --- | --- |
| `github.issues.list` | `state`: `open`, `closed`, or `all`; `limit`: 1–100 | One `GET /repos/{owner}/{repo}/issues` |
| `github.pulls.create-draft` | Required `title`, `head`, `base`; optional `body` | One `POST /repos/{owner}/{repo}/pulls`, with `draft: true` and `maintainer_can_modify: false` |

The issue endpoint returns GitHub issue records, including pull requests as GitHub defines them. This candidate retrieves one page; it does not follow pagination links. Draft creation does not create branches or commits: `head` must already exist. Fork heads supported by GitHub can be specified, but the destination repository remains the one approved.

Only `https://api.github.com` is supported. There are no arbitrary URL, header, HTTP method, merge, delete, or credential-read operations. Redirects are rejected. GitHub Enterprise hosts and additional service adapters are outside this candidate.

## The trust boundary

Conceal enforces the contract inside a trusted execution context. It does not establish that context for you. The approving operator must control:

- the executable and its environment;
- the contract file;
- the approval state directory;
- access to the native credential store.

The agent should submit proposals through a trusted wrapper, task runner, or operator. That boundary must allow only the intended `prepare`, `inspect`, and `execute` operations and fix the contract path, state directory, and environment. The agent must not be able to invoke `approve`, replace the executable, modify state files, select a weaker contract, or use `summon show` or `run` to retrieve the same credential.

A local CLI running under one unrestricted user cannot distinguish a human from an agent running as that same user. The `--digest` flag is an explicit approval input, not human authentication. Use an OS account boundary, a sandbox with an allowlisted host interface, or another trusted execution layer when the agent itself is untrusted. This candidate does not include a sandbox or a network approval service.

On macOS and Linux, newly created state directories use mode `0700`, state files use `0600`, and an existing directory with group/other access is rejected. Windows uses the user's private configuration directory and inherited ACLs; protect a custom location with appropriate ACLs. Symlinks as the state directory are rejected. Protect parent directories as well.

## Create a contract

Start from [`examples/github-contract.json`](../examples/github-contract.json), or:

```sh
conceal action example
```

The fields are:

| Field | Meaning |
| --- | --- |
| `schema_version` | Must be `1` |
| `name` | Human-readable contract name |
| `credential` | Existing Conceal secret name, never a token value |
| `repositories` | Explicit, case-sensitive `owner/repository` entries; no wildcards |
| `actions` | Explicit supported operation names |
| `approval_ttl_seconds` | Approval lifetime, from 1 to 3600 seconds |

Use a GitHub fine-grained token with the smallest repository scope and permissions the selected actions require. Issue reads typically need Issues read permission for private repositories; draft creation needs Pull requests write permission. Organization policies and token type may impose additional requirements. See GitHub's [issues](https://docs.github.com/en/rest/issues/issues#list-repository-issues) and [pull requests](https://docs.github.com/en/rest/pulls/pulls#create-a-pull-request) endpoint documentation.

Store the token through the hidden terminal prompt:

```sh
conceal set github/token
conceal action validate examples/github-contract.json
```

Contract validation and request preparation do not retrieve the token. The token must be available when executing. Approval binds its secret reference, not its current value, so authorized credential rotation does not require rewriting the contract.

## Prepare, review, approve, execute

Edit the contract's repository entry before running:

```sh
conceal action prepare --contract examples/github-contract.json --operation github.issues.list --repo OWNER/REPOSITORY --input examples/list-issues.json
```

For a draft:

```sh
conceal action prepare --contract examples/github-contract.json --operation github.pulls.create-draft --repo OWNER/REPOSITORY --input examples/draft-pull-request.json
```

`--input -` reads JSON from stdin. Omitted input means `{}`, suitable for issue reads with the default open state and limit of 30. Unknown fields, duplicate fields (including case variants), non-object input, and unsupported values are rejected. Persisted JSON documents are limited to 1 MiB; large payloads may reach this limit sooner because the proposal includes both source inputs and the generated HTTP body.

Preparation returns JSON containing a random request ID, creation time, absolute contract path, contract snapshot, inputs, the generated HTTP request, and a SHA-256 digest. No API call has occurred.

From the trusted approving context:

```sh
conceal action inspect REQUEST_ID
conceal action approve REQUEST_ID --digest REVIEWED_SHA256
conceal action execute REQUEST_ID
```

Review the entire `inspect` output, especially repository, title/body, branch names, and requested operation. Pass the digest obtained from that trusted inspection. Do not approve an agent-supplied digest without checking the corresponding request.

The digest covers the request ID, timestamp, contract path and contents, action, repository, inputs, and generated HTTP request. This binds the approval to one concrete proposal. Inspection reloads the contract and rebuilds the HTTP request; any semantic contract or request change requires a new proposal and approval. JSON formatting alone does not change a decoded contract.

Unapproved proposals older than 24 hours cannot be approved. Approval lifetime starts when `approve` succeeds, using the host clock. An existing approval cannot be renewed or replaced. To revoke:

```sh
conceal action revoke REQUEST_ID
```

All commands use the same state directory. By default:

- macOS: `~/Library/Application Support/conceal/actions`
- Windows: `%AppData%\conceal\actions`
- Linux: `$XDG_CONFIG_HOME/conceal/actions`, falling back to `~/.config/conceal/actions`

`conceal action --state-dir PATH ...` selects a trusted custom directory. The location must be fixed by the trusted wrapper, not chosen by an untrusted caller.

## Single use and uncertain outcomes

Conceal locks a request while approving, executing, or revoking it. Execution checks the approval and expiration before reading the credential. It writes the consumed marker before network dispatch. A competing execution cannot spend the same approval.

Credential lookup failures before dispatch do not spend approval. After the marker is written, the approval remains consumed even if GitHub rejects the request, the connection fails, the response exceeds 4 MiB, or the process crashes. There are no application-level retries and no redirect following. Request timeout is 30 seconds.

This gives single-use authorization, not a distributed exactly-once transaction. A remote operation can succeed while the response is lost. Check GitHub before preparing another draft-creation request. A new proposal receives a new ID and can produce another remote operation if approved.

A crash can leave a lock directory and/or a consumed marker without a receipt. Do not delete state to retry the same request. Inspect the remote service, then prepare a new request if appropriate. Restoring state from backups can resurrect old approvals; treat restoration as security-sensitive and discard outstanding approvals.

## Results and receipts

Successful execution returns JSON:

```json
{
  "receipt": {
    "schema_version": 1,
    "id": "request-id",
    "digest": "sha256-of-reviewed-request",
    "action": "github.issues.list",
    "repository": "owner/repository",
    "outcome": "succeeded",
    "http_status": 200,
    "time": "2026-11-05T12:00:00Z"
  },
  "response": []
}
```

Outcomes include `succeeded`, `http_error`, `transport_error`, `response_error`, and `revoked`. Errors produce a nonzero exit status. When a receipt exists, the CLI prints it before reporting an execution error to stderr. Failed checks before consumption do not create an execution receipt.

Receipts persist action identity, digest, repository, outcome, status, and time. They do not persist the token or remote response. Exact token values and their JSON-escaped form are masked in returned responses as a defensive measure; arbitrary transformations are not detected. Responses can contain private repository data and must be handled accordingly.

Receipts are local records, not cryptographically signed audit evidence. The approving user can modify their own files. The request files also contain titles, bodies, and other potentially private proposal data; keep state private and apply a retention policy. Conceal does not currently prune state.

## Relationship to `run` and Summon

`conceal run` and Summon inject secrets into trusted processes. Those processes receive the credential and choose what to do with it.

The action adapter keeps credential use inside Conceal and exposes a limited operation, with an exact approval before dispatch. You can use both approaches in the same installation. Neither substitutes for sandboxing an agent that already has unrestricted access to the host.

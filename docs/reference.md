# Command and platform reference

[Back to Conceal](../README.md) · [Getting started](getting-started.md) · [Reviewed GitHub actions](actions.md)

## Platforms

| Platform | Native store | Release architectures | Requirements |
| --- | --- | --- | --- |
| macOS | Keychain, service `summon` | Apple Silicon and Intel | macOS 13+ for Go 1.27 release binaries |
| Windows | Generic Credentials, target `summon/<name>` | x64 and ARM64 | Windows 10+ or Server 2016+ |
| Linux | Kernel user keyring, description `summon:<name>` | x64, ARM64, i386 | User keyrings enabled and permitted by the session/container |

Linux keyrings are kernel memory, not an encrypted disk vault. Secrets do not survive reboot; session and keyring lifetime policies can remove them earlier. Containers may deny `keyctl` entirely. `conceal check --json <name>` distinguishes an unavailable backend from a missing secret.

Clipboard access requires a desktop session. On Linux, the clipboard library uses `xclip` or `xsel` and an X11-compatible display. Headless operations (`set`, `update`, `list`, `run`, and `action`) do not require clipboard support.

The [desktop inbox](desktop.md) has separate WebView2/WebKitGTK requirements and native installer artifacts. The table above describes CLI archives.

## Store and manage secrets

```sh
conceal set github/token
conceal update github/token
conceal list
conceal get github/token
conceal unset github/token
```

`set` and `update` hide input when run in a terminal. Both also accept stdin, including input ending at EOF without a newline. They remove one final LF or CRLF by default; `--raw` preserves every byte. Internal newlines and surrounding spaces are preserved. Input is limited to 1 MiB, and native stores impose smaller platform-specific limits. Do not type real secret values into command arguments or shell history.

`--non-interactive` disables prompts; supply the secret name and pipe its value from a trusted source. `get` copies a value for 15 seconds, clearing it only if the clipboard still contains that value. Clipboard managers may retain their own history.

## Run a trusted process

```sh
conceal run --env GITHUB_TOKEN=github/token -- gh issue list
conceal run --clean-env --timeout 2m --env API_TOKEN=service/token -- my-tool
```

Only named secrets are loaded. `--` is required before the executable; no shell is inserted. The parent environment is inherited unless `--clean-env` is supplied. Exact secret values are masked in stdout and stderr, including values split across output chunks; `--no-redact` disables masking.

The child receives the real values and must be trusted. Masking does not prevent encoding, network exfiltration, file writes, or reads from other credential stores. Timeouts terminate the direct child, not a process tree. Child exit codes are preserved; timeout is 124 and interruption is 130. `--clean-env` may remove variables a program needs, and Windows may supply essential system variables.

## Discover names and availability

```sh
conceal list --json
conceal check --json github/token service/token
```

JSON includes `schema_version: 1`. `check` exits 0 when every name is present, 2 when a secret is missing, and 3 when the backend is unavailable. Presence does not prove a credential is valid or authorized at the remote service. These commands return names and availability only. Windows enumeration loads native credential records internally; Conceal clears their value buffers and exposes names only.

## Summon provider

For a complete `secrets.yml` example, see [getting started](getting-started.md#use-your-existing-summon-workflow).

```sh
conceal summon install
conceal summon install --dir ./providers
```

The default directory respects `SUMMON_PROVIDER_PATH`; otherwise it uses `/usr/local/lib/summon` on macOS/Linux and `%ProgramW6432%\Cyberark Conjur\Summon\Providers` on Windows. The standard location may require elevated filesystem access; `--dir` selects a writable location. Existing providers are preserved unless `--force` is supplied. Re-run the installer after upgrading Conceal.

The installed provider is `conceal_summon` (`conceal_summon.exe` on Windows), a native copy of the executable that implements Summon's single-argument provider protocol. Use its path with `summon --provider <path> <command>`. It writes exact value bytes, without an added newline; `--version` reports its version.

`conceal summon show <name>` writes the exact value to stdout for trusted consumers. It can expose credentials in terminal logs.

## Discover and review service actions

```sh
conceal action adapters
conceal action example --adapter github
conceal action example --adapter aws
conceal action inbox --limit 100
conceal action review REQUEST_ID
```

`adapters` returns a schema-versioned catalog of installed operations, input/scope schemas, declared hosts, and credential roles. `inbox` returns up to 200 recent request summaries (100 by default). `review` returns the complete validated proposal, structured human review, credential names, approval state, and allowed transitions. These commands do not retrieve credentials or contact services.

Use `--resource` for schema-2 preparation. `--repo` remains a mutually exclusive alias for existing GitHub workflows. Calling `action example` without an adapter retains the schema-1 GitHub example. See [GitHub actions](actions.md), [S3 listings](aws.md), and the [adapter SDK](adapters.md).

## Development and release checks

```sh
go test ./...
go test -race ./...
go vet ./...
```

Unit tests use synthetic stores and do not touch your credential store or clipboard. To run the native lifecycle test, explicitly set `CONCEAL_INTEGRATION=1`; it creates random test entries and removes them. An opted-in test fails if the backend is unavailable. The race detector needs a supported C compiler.

CI tests Go 1.26 and 1.27 on macOS, Windows, and Linux, runs native credential lifecycles, lints each platform, scans dependencies, and builds all seven release architectures. The packaging job runs on macOS to support CGO. Release workflows currently produce downloadable verification artifacts only; they do not publish releases or update package registries.

The [desktop workflow](https://github.com/infamousjoeg/conceal/actions/workflows/desktop.yml) builds native installers on all three systems, checks the frontend and Rust bridge, and exercises the Windows WebView2 app with an isolated test profile. See the [desktop guide](desktop.md#build-or-contribute) for local commands.

The Go packages remain available under `pkg/conceal`. Prefer `keychain.List()` (which reports backend failures) and `keychain.ReadSecret()` (caller-owned bytes to clear after use); `ListSecrets()` is retained for compatibility. `clipboard.Secret()` reports delivery/cleanup errors.

See [CONTRIBUTING.md](../CONTRIBUTING.md) for the development workflow.

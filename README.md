# Conceal

Conceal manages local secrets and gives applications and AI agents controlled ways to use them. It uses macOS Keychain, Windows Credential Manager, or the Linux kernel user keyring. It runs on its own and also works as a native [Summon](https://cyberark.github.io/summon) provider.

The 4.3 release candidate adds **action contracts and exact approvals**: an agent proposes an operation; an operator reviews the exact request; Conceal uses the credential to execute that approved request once. The agent receives the result, not the API token.

[![Go Test](https://github.com/infamousjoeg/conceal/actions/workflows/go-test.yml/badge.svg)](https://github.com/infamousjoeg/conceal/actions/workflows/go-test.yml)
[![Code Quality](https://github.com/infamousjoeg/conceal/actions/workflows/golangci-lint.yml/badge.svg)](https://github.com/infamousjoeg/conceal/actions/workflows/golangci-lint.yml)
[![Security Checks](https://github.com/infamousjoeg/conceal/actions/workflows/security.yml/badge.svg)](https://github.com/infamousjoeg/conceal/actions/workflows/security.yml)

**Explore:** [Action playground](docs/playground.html) · [Action contract guide](docs/actions.md) · [Example contract](examples/github-contract.json)

The playground is a self-contained page. Download/open `docs/playground.html`, or serve a checkout with `python -m http.server 8000` and visit `http://localhost:8000/docs/playground.html`. It uses simulated data and makes no API calls.

## Platforms

| Platform | Native store | Release architectures | Requirements |
| --- | --- | --- | --- |
| macOS | Keychain, service `summon` | Apple Silicon and Intel | macOS 13+ for Go 1.27 release binaries |
| Windows | Generic Credentials, target `summon/<name>` | x64 and ARM64 | Windows 10+ or Server 2016+ |
| Linux | Kernel user keyring, description `summon:<name>` | x64, ARM64, i386 | User keyrings enabled and permitted by the session/container |

Linux keyrings are kernel memory, not an encrypted disk vault. Secrets do not survive reboot; session and keyring lifetime policies can remove them earlier. Containers may deny `keyctl` entirely. `conceal check --json <name>` distinguishes an unavailable backend from a missing secret.

Clipboard access requires a desktop session. On Linux, the clipboard library uses `xclip` or `xsel` and an X11-compatible display. Headless operations (`set`, `update`, `list`, `run`, and `action`) do not require clipboard support.

## Installation

Stable releases remain available through [GitHub Releases](https://github.com/infamousjoeg/conceal/releases), Homebrew, and Chocolatey:

```sh
brew tap infamousjoeg/tap
brew install conceal
```

```powershell
choco install conceal
```

The commands above install the published stable version, which may not include release-candidate features. For the candidate, build this branch or download the `release-candidate-packages` artifact from its successful [packaging workflow](https://github.com/infamousjoeg/conceal/actions/workflows/test-release.yml).

Archives are named `conceal_<OS>_<architecture>.tar.gz` (Windows uses `.zip`). Verify against `checksums.txt`, extract, place the executable on `PATH`, and run `conceal version`. Candidate macOS archives are not Developer ID signed or notarized.

### Build from source

Use Go 1.26 or newer; release builds use Go 1.27.3. macOS builds require Xcode Command Line Tools and CGO for the Security framework. Windows and Linux binaries do not require CGO.

```sh
git clone https://github.com/infamousjoeg/conceal.git
cd conceal
go build -trimpath -o bin/ .
```

Run `bin/conceal version` (`bin\conceal.exe version` on Windows).

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

## Applications and agents

### Execute an approved action

The first adapter supports GitHub issue reads and **draft-only** pull-request creation, for repositories explicitly named in a contract. It does not expose arbitrary URLs, shell commands, merges, or raw credential retrieval.

```sh
conceal action validate examples/github-contract.json
conceal action prepare --contract examples/github-contract.json --operation github.issues.list --repo OWNER/REPOSITORY --input examples/list-issues.json
conceal action inspect REQUEST_ID
conceal action approve REQUEST_ID --digest REVIEWED_SHA256
conceal action execute REQUEST_ID
```

Replace `OWNER/REPOSITORY` in the example contract before preparing a request. Review the complete `inspect` output from the trusted approving context, then supply that digest to `approve`. Changed requests, changed contracts, expired approvals, and consumed approvals are rejected. See the [guide](docs/actions.md) for deployment boundaries, request schemas, receipts, and failure handling.

The policy, approval state, and native store must remain outside the untrusted agent's control. Conceal is not an operating-system sandbox: an unrestricted process under the same user can approve requests or read that user's secrets through other tools. `approve` is a privileged operator command, not proof of human presence.

### Run a trusted process

```sh
conceal run --env GITHUB_TOKEN=github/token -- gh issue list
conceal run --clean-env --timeout 2m --env API_TOKEN=service/token -- my-tool
```

Only named secrets are loaded. `--` is required before the executable; no shell is inserted. The parent environment is inherited unless `--clean-env` is supplied. Exact secret values are masked in stdout and stderr, including values split across output chunks; `--no-redact` disables masking.

The child receives the real values and must be trusted. Masking does not prevent encoding, network exfiltration, file writes, or reads from other credential stores. Timeouts terminate the direct child, not a process tree. Child exit codes are preserved; timeout is 124 and interruption is 130. `--clean-env` may remove variables a program needs, and Windows may supply essential system variables.

### Discover requirements without returning values

```sh
conceal list --json
conceal check --json github/token service/token
```

JSON includes `schema_version: 1`. `check` exits 0 when every name is present, 2 when a secret is missing, and 3 when the backend is unavailable. Presence does not prove a credential is valid or authorized at the remote service. These commands return names and availability only. Windows enumeration loads native credential records internally; Conceal clears their value buffers and exposes names only.

## Summon integration

Summon is optional. Install the native provider executable with:

```sh
conceal summon install
conceal summon install --dir ./providers
```

The default directory respects `SUMMON_PROVIDER_PATH`; otherwise it uses `/usr/local/lib/summon` on macOS/Linux and `%ProgramW6432%\Cyberark Conjur\Summon\Providers` on Windows. The standard location may require elevated filesystem access; `--dir` selects a writable location. Existing providers are preserved unless `--force` is supplied. Re-run the installer after upgrading Conceal.

The installed provider is `conceal_summon` (`conceal_summon.exe` on Windows), a native copy of the executable that implements Summon's single-argument provider protocol. Use its path with `summon --provider <path> <command>`.

`conceal summon show <name>` writes the exact value to stdout for trusted consumers. It can expose credentials in terminal logs; use the action adapter or `run` when appropriate.

## Development and release checks

```sh
go test ./...
go test -race ./...
go vet ./...
```

Unit tests use synthetic stores and do not touch your credential store or clipboard. To run the native lifecycle test, explicitly set `CONCEAL_INTEGRATION=1`; it creates random test entries and removes them. An opted-in test fails if the backend is unavailable. The race detector needs a supported C compiler.

CI tests Go 1.26 and 1.27 on macOS, Windows, and Linux, runs native credential lifecycles, lints each platform, scans dependencies, and builds all seven release architectures. The packaging job runs on macOS to support CGO. Release workflows currently produce downloadable verification artifacts only; they do not publish releases or update package registries.

The Go packages remain available under `pkg/conceal`. Prefer `keychain.List()` (which reports backend failures) and `keychain.ReadSecret()` (caller-owned bytes to clear after use); `ListSecrets()` is retained for compatibility. `clipboard.Secret()` reports delivery/cleanup errors.

## Maintainer and license

Maintained by [Joe Garcia](https://github.com/infamousjoeg). Contributions are welcome. See [LICENSE](LICENSE).

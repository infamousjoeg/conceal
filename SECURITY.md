# Security Policy

## Supported Versions

Use the latest [published stable release](https://github.com/infamousjoeg/conceal/releases) for stable deployments. The 4.3 branch is a release candidate: please report findings against it, but do not treat candidate status as a production security certification. Include the exact output of `conceal version` in reports.

## Security Features

### Choose the appropriate workflow

| Workflow | Who receives the key? | What you must trust |
| --- | --- | --- |
| `conceal run` or the Summon provider | The launched program | That program and its environment |
| `conceal action` / desktop inbox | Conceal's shared execution engine; the caller receives the result | The operator's execution context, installed adapter code, contract, approval state, and credential store |
| `conceal get` | The clipboard and applications that can read it | The desktop session and clipboard consumers |
| `conceal summon show` | Standard output | Every consumer of that output |

Conceal uses macOS Keychain, Windows Credential Manager, or the Linux kernel user keyring. OS access controls apply. Linux values are held in kernel memory and do not survive reboot; they are not an encrypted disk vault.

Conceal does not implement Touch ID, Face ID, or Windows Hello approval. The `action approve` command does not authenticate human presence. A process with unrestricted access to the same account can invoke approval or retrieve keys through another route.

### Action authorization

The shared action engine checks an explicit resource/operation contract and a short-lived approval bound to the exact request before retrieving credentials. It consumes approval before dispatch, rejects redirects, and does not retry automatically. Schema-2 requests also bind the full adapter plan and behavior version. The AWS adapter uses host-owned SigV4 signing; it does not discover credentials or call the AWS CLI.

Adapters are trusted code compiled into the application. The public SDK, endpoint declarations, and conformance tests are not a sandbox. Installed adapter code and its dependencies require review. There is no dynamic plugin loader.

The desktop sends the digest of the displayed review to the same engine. Its Rust bridge accepts fixed commands and invokes a bundled executable; the webview has no general shell permission. Practice mode uses synthetic in-memory data. Neither a checkbox nor an application window proves human presence or prevents unrestricted same-account code from approving through the CLI. See [desktop operator context](docs/desktop.md#keep-the-operator-context-trusted).

The candidate requires an operator or separately configured trusted boundary. It does not include an agent connector, sandbox, or network approval service. See the [action threat model](docs/actions.md#the-trust-boundary), [uncertain outcomes](docs/actions.md#single-use-and-uncertain-outcomes), and [receipt limitations](docs/actions.md#results-and-receipts) before deployment.

### Output, memory, and clipboard limits

`run` masks exact key values in stdout/stderr by default; `--no-redact` disables that masking. Encoding, transformation, files, and network traffic are outside this protection. The launched program receives the real key. The action engine also masks exact credential values and their JSON string escapes in returned responses, but those responses may contain other private data.

Conceal clears owned secret buffers where possible. This does not guarantee erasure of every copy in Go, OS APIs, child processes, or crash dumps.

`get` clears the clipboard after 15 seconds only if it still contains the copied value. Other applications and clipboard history may retain copies. Hidden input prevents terminal echo; never put real keys in command arguments or shell history.

## Automated Security

CI runs `golangci-lint`, `govulncheck`, npm audit, RustSec cargo audit, dependency review on pull requests, and platform tests. The [security workflow](https://github.com/infamousjoeg/conceal/blob/release/v4.3.0-rc.1/.github/workflows/security.yml) defines scan triggers. Automated checks reduce known risks; passing them does not prove an absence of vulnerabilities.

Candidate archives include checksums for integrity verification. They are not Developer ID signed or notarized. Current release workflows build verification artifacts without publishing releases or updating registries.

### Desktop dependency warnings

The Tauri Linux dependency tree requires GTK 3 / GLib 0.18.5. Conceal includes a [verified local backport](https://github.com/infamousjoeg/conceal/tree/release/v4.3.0-rc.1/desktop/src-tauri/vendor) of the upstream fix for [RUSTSEC-2024-0429](https://rustsec.org/advisories/RUSTSEC-2024-0429). The fix makes the `VariantStrIter` output pointer mutable, removing undefined behavior that can cause a null-pointer dereference. CI compares every vendored file against the checksum-verified official archive, permits only the upstream two-line change, and runs a regression test with release optimization. GLib 0.20 is not compatible with this GTK dependency.

Version-only scanners may still report the upstream GLib advisory because the patched package retains version 0.18.5. RustSec also reports [RUSTSEC-2024-0370](https://rustsec.org/advisories/RUSTSEC-2024-0370): the build-time `proc-macro-error` 1.0.4 dependency used by the GTK/GLib macros is unmaintained. Cargo audit warnings remain visible; there is no cargo advisory ignore list. Windows and macOS do not use the GTK runtime. Remove the backport when a compatible fixed dependency becomes available.

## Reporting a Vulnerability

### Responsible Disclosure

We appreciate security researchers who report vulnerabilities responsibly. Please follow our responsible disclosure process:

1. **Contact**: Email `joe dot garcia at cyberark dot com` directly
2. **Include**:
   - Detailed description of the vulnerability
   - Steps to reproduce the issue
   - Potential impact assessment
   - Your contact information for follow-up

### What to Expect

| Timeline | Action |
|----------|---------|
| 24 hours | Acknowledgment of your report |
| 48 hours | Initial assessment and classification |
| 7 days   | Detailed investigation and response plan |
| 30 days  | Fix implementation and testing |
| Release  | Public disclosure and credit (if desired) |

### Security Response

**Critical Vulnerabilities:**
- Immediate investigation
- Emergency patch within 48-72 hours
- Public security advisory

**High/Medium Vulnerabilities:**
- Investigation within 7 days
- Patch in next scheduled release
- Security note in release notes

**Low Vulnerabilities:**
- Investigation within 30 days
- Fix in future release
- Standard development process

## Security Best Practices

### For Users

1. **Keep Updated**: Always use the latest version
2. **Secure Environment**: Use Conceal only on trusted systems
3. **Access Control**: Limit who has access to your credential stores
4. **Regular Audits**: Periodically review stored secrets with `conceal list`
5. **Recovery**: Keep an appropriate recovery method for the original credential; Linux keyring values disappear at reboot. Do not restore outstanding action approvals from backups.
6. **Secret Input**: Use the hidden `set` or `update` prompt, or stdin from a trusted source. See [secret input behavior](docs/reference.md#store-and-manage-secrets).

### For Developers

1. **Code Review**: All code changes require security-focused review
2. **Dependency Management**: Keep dependencies updated and vetted
3. **Testing**: Include security test cases in your contributions
4. **Documentation**: Document any security-relevant changes
5. **Build Tags**: Use appropriate build tags to limit platform exposure

## Threat Model

### Assets
- User secrets stored in OS credential stores
- Application source code and binaries
- User systems and credential stores

### Scope

Conceal helps avoid keys in project files, limit which keys a trusted process receives, and authorize specific service requests. It does not protect against a compromised OS, unrestricted same-account code, malicious trusted applications or adapters, or privileged memory inspection. Local receipts are editable by their owner and are not signed audit evidence.

### Assumptions
- User's operating system and credential store are secure
- User has legitimate access to their credential store
- Development and build environments are secure
- Go toolchain and standard library are secure

## Security Contact

For security-related questions or concerns that don't constitute vulnerabilities:

- **Email**: joe dot garcia at cyberark dot com  
- **Subject**: [Conceal Security] Your Question
- **Response Time**: 7-14 business days

For urgent security matters, please include "URGENT" in the subject line.

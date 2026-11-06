# Review requests in the desktop inbox

[Back to Conceal](../README.md) · [GitHub walkthrough](actions.md) · [S3 walkthrough](aws.md)

Conceal's desktop app puts the decision in one place: what will happen, which resource it affects, every input, and which stored credentials will be used.

**Try practice mode first.** It contains a draft pull request and an S3 listing. Approve one, run it, and try rejecting the other. Practice uses in-memory examples: it does not read credentials, write approval files, or contact a service.

## Get the candidate

Download a successful run's artifact from [Desktop candidate checks](https://github.com/infamousjoeg/conceal/actions/workflows/desktop.yml). Select the artifact for your operating system and verify the included `checksums.txt`. GitHub requires a signed-in account to download workflow artifacts.

| Platform | Candidate package | Requirements |
| --- | --- | --- |
| Windows | NSIS setup `.exe` | Windows 10+; WebView2 runtime (the installer can install it) |
| macOS | `.dmg` | macOS 13+; match the artifact's CPU architecture |
| Linux | `.deb` and `.AppImage` | Built on Ubuntu 22.04; GTK 3/WebKitGTK 4.1 and an available kernel user keyring |

These are unsigned candidate builds, without Apple notarization or Windows publisher signing. They may trigger operating system warnings. Candidate installers are verification artifacts, not a published stable release. The desktop includes its own matching Conceal engine; install the [CLI candidate](getting-started.md#install-the-release-candidate) separately to prepare requests.

For the Linux x64 AppImage, this launch mode works without FUSE:

```sh
chmod u+x Conceal_4.3.0-rc.1_amd64.AppImage
APPIMAGE_EXTRACT_AND_RUN=1 ./Conceal_4.3.0-rc.1_amd64.AppImage
```

See the [documented Linux dependency warnings](../SECURITY.md#desktop-dependency-warnings) when evaluating the candidate.

## Use a real request

1. Create a contract using the [GitHub](actions.md#create-a-contract) or [S3](aws.md) guide. Keep it in your trusted operator environment.
2. Prepare a request with `conceal action prepare`. Preparation reads no credentials and sends no API request.
3. Open Conceal and choose **Refresh**. The CLI and desktop share the default approval directory in your operating system's user configuration folder.
4. Review the target, all fields, consequence, credential names, and approval lifetime. **Inspect the complete request** shows the HTTP request, contract, and digest.
5. Check the review confirmation and choose **Approve once**, or choose **Reject request**.
6. After approval, choose **Run approved action** when ready. Approval alone does not run anything. The service response appears after execution.

A changed request, contract, or schema-2 adapter plan fails validation. The app submits the digest from the document you actually reviewed; it does not silently approve a newer version.

The inbox shows up to 100 recent requests. Refresh explicitly after preparing or changing a request. Invalid requests remain listed but cannot be approved. The CLI's `action inspect REQUEST_ID` can inspect older requests.

## Understand the outcome

| State | What to do |
| --- | --- |
| Needs review | Read every field, then approve once or reject |
| Approved once | Run before the displayed deadline, or reject |
| Approval expired / Request expired | Prepare a fresh request; approval cannot be renewed |
| Approval used | Inspect the receipt and service result; the approval cannot run again |
| Rejected | No dispatch occurred; prepare a new request if needed |
| In progress | Refresh; an interrupted operation may leave a lock |
| Cannot review | Resolve the changed or invalid contract/request and prepare a new proposal |

If a request was dispatched but the response failed, its approval is still consumed. Check the remote service before preparing another request. Local receipts are useful history, but they are not signed audit evidence.

## Keep the operator context trusted

The desktop improves review; it does not authenticate human presence or sandbox an agent. Unrestricted code running as your account can use the CLI or credential store directly. Keep the approval directory, contracts, app, and credential store outside the untrusted agent's writable environment. See [the trust boundary](actions.md#the-trust-boundary).

The webview has no general shell API. Its Rust backend accepts only fixed inbox/review/approve/reject/execute commands and invokes the bundled Go engine. Review fields and responses are rendered as text, with invisible control/direction characters shown explicitly.

For a separate trusted operator profile or isolated test, set `CONCEAL_DESKTOP_STATE_DIR` **before launching the app**, and pass the same directory to the CLI's `--state-dir`. This is a trusted launcher setting, not a frontend option. The directory must meet the same private-directory requirements as the CLI. Close that profile when finished; the app does not import approvals from agent-controlled folders.

## Build or contribute

Install Go 1.27.3, Node.js 22.12+, Rust 1.92+, and the [Tauri platform prerequisites](https://v2.tauri.app/start/prerequisites/). macOS also needs Xcode Command Line Tools and CGO. Windows needs the MSVC C++ build tools and WebView2.

From the checkout:

```sh
cd desktop
npm ci
npm run desktop:dev
```

For a production bundle:

```sh
npm run desktop:build
```

The build script compiles the Go sidecar for the native Rust target. Output is in `desktop/src-tauri/target/release/bundle/`. Set `TAURI_ENV_TARGET_TRIPLE` only when deliberately cross-compiling with the matching SDK/toolchain, and pass the same target to Tauri.

Checks:

```sh
npm run build
npx playwright install chromium
npm test
npm run sidecar
cargo test --manifest-path src-tauri/Cargo.toml --locked
cargo clippy --manifest-path src-tauri/Cargo.toml --locked -- -D warnings
```

On Windows, `PLAYWRIGHT_CHANNEL=msedge` selects an installed Edge for browser tests; set it with PowerShell's `$env:PLAYWRIGHT_CHANNEL = 'msedge'`. Browser tests exercise practice, review controls, stale-digest rejection, text rendering, and narrow/dark layouts. Native Rust tests check command arguments. The Go tests cover authorization and dispatch.

After a Windows production build, `npm run test:native` launches the actual WebView2 app in an isolated temporary profile. It prepares a synthetic request, approves it through the UI, verifies denial with a nonexistent credential, and rejects it. No service is contacted. The app closes afterward; its temporary profile and screenshots are retained for inspection.

`npm run dev` opens a browser-only practice version. It never connects to an approval engine. Application icons are generated from `desktop/src-tauri/icons/source.svg` with `npx tauri icon`; keep the desktop PNG/ICO/ICNS outputs when updating the source.

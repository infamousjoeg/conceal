import { execFileSync } from "node:child_process";
import { mkdirSync } from "node:fs";
import { fileURLToPath } from "node:url";
import path from "node:path";

const root = fileURLToPath(new URL("../../", import.meta.url));
const host = execFileSync("rustc", ["-vV"], { encoding: "utf8" }).match(/^host: (.+)$/m)?.[1];
const target = process.env.TAURI_ENV_TARGET_TRIPLE || host;
const targets = {
  "x86_64-pc-windows-msvc": ["windows", "amd64"],
  "aarch64-pc-windows-msvc": ["windows", "arm64"],
  "x86_64-apple-darwin": ["darwin", "amd64"],
  "aarch64-apple-darwin": ["darwin", "arm64"],
  "x86_64-unknown-linux-gnu": ["linux", "amd64"],
  "aarch64-unknown-linux-gnu": ["linux", "arm64"],
};
if (!target || !targets[target]) throw new Error(`Unsupported desktop target: ${target}`);
const [GOOS, GOARCH] = targets[target];
// macOS uses Security.framework for the Keychain; cross builds require its SDK.
const CGO_ENABLED = GOOS === "darwin" ? "1" : "0";
const output = path.join(root, "desktop/src-tauri/binaries", `conceal-engine-${target}${GOOS === "windows" ? ".exe" : ""}`);
mkdirSync(path.dirname(output), { recursive: true });
execFileSync("go", ["build", "-trimpath", "-ldflags", "-s -w -X github.com/infamousjoeg/conceal/pkg/conceal.Version=4.3.0-rc.1 -X github.com/infamousjoeg/conceal/pkg/conceal.Tag=", "-o", output, "."], {
  cwd: root, stdio: "inherit",
  env: { ...process.env, GOOS, GOARCH, CGO_ENABLED, MACOSX_DEPLOYMENT_TARGET: "13.0" },
});
console.log(`Built ${path.basename(output)}`);

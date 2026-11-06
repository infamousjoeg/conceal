// Exercise the packaged Windows webview, Rust bridge, and bundled Go engine.
// Uses an isolated profile and a nonexistent credential; no service can run.
import { chromium, expect } from "@playwright/test";
import { execFileSync, spawn } from "node:child_process";
import { mkdtemp, mkdir, readFile, writeFile } from "node:fs/promises";
import { createServer } from "node:net";
import { tmpdir } from "node:os";
import path from "node:path";
import { fileURLToPath } from "node:url";

if (process.platform !== "win32") throw new Error("This smoke test requires Windows WebView2.");
const desktop = fileURLToPath(new URL("../", import.meta.url));
const binary = path.join(desktop, "src-tauri/target/release/conceal-desktop.exe");
const engine = path.join(desktop, "src-tauri/target/release/conceal-engine.exe");
const directory = await mkdtemp(path.join(tmpdir(), "conceal-desktop-smoke-"));
const state = path.join(directory, "state");
const contract = path.join(directory, "contract.json");
await writeFile(contract, JSON.stringify({
  schema_version: 2, name: "native-smoke", adapter: "github",
  credentials: { token: `conceal-smoke-missing/${path.basename(directory)}` },
  actions: ["github.issues.list"], approval_ttl_seconds: 300, scope: { repositories: ["example/review-test"] },
}));
const cli = (...args) => JSON.parse(execFileSync(engine, ["action", "--state-dir", state, ...args], { encoding: "utf8", windowsHide: true }));
const proposal = cli("prepare", "--contract", contract, "--operation", "github.issues.list", "--resource", "example/review-test");
const server = createServer();
await new Promise(resolve => server.listen(0, "127.0.0.1", resolve));
const port = server.address().port;
await new Promise(resolve => server.close(resolve));
const processHandle = spawn(binary, [], {
  windowsHide: true, stdio: "ignore",
  env: { ...process.env, CONCEAL_DESKTOP_STATE_DIR: state, WEBVIEW2_USER_DATA_FOLDER: path.join(directory, "webview"),
    WEBVIEW2_ADDITIONAL_BROWSER_ARGUMENTS: `--remote-debugging-port=${port}` },
});
let browser;
try {
  const endpoint = `http://127.0.0.1:${port}`;
  const deadline = Date.now() + 45_000;
  while (true) {
    try { if ((await fetch(`${endpoint}/json/version`)).ok) break; } catch {}
    if (Date.now() > deadline || processHandle.exitCode !== null) throw new Error("Native webview did not start.");
    await new Promise(resolve => setTimeout(resolve, 250));
  }
  browser = await chromium.connectOverCDP(endpoint);
  const page = browser.contexts()[0].pages()[0];
  await expect(page.getByText("This computer", { exact: true })).toBeVisible();
  await expect(page.getByRole("heading", { name: "List repository issues" })).toBeVisible();
  await expect(page.getByText("example/review-test", { exact: true }).last()).toBeVisible();
  await expect(page.getByRole("button", { name: "Approve once" })).toBeDisabled();
  const shots = process.env.CONCEAL_SCREENSHOT_DIR || directory;
  await mkdir(shots, { recursive: true });
  await page.screenshot({ path: path.join(shots, "native-inbox.png"), fullPage: true });
  await page.getByRole("checkbox").check();
  await page.getByRole("button", { name: "Approve once" }).click();
  await expect(page.getByRole("button", { name: "Run approved action" })).toBeEnabled();
  const approval = JSON.parse(await readFile(path.join(state, `${proposal.request.id}.approval.json`), "utf8"));
  if (approval.digest !== proposal.digest) throw new Error("Displayed approval did not reach the engine intact.");
  await page.getByRole("button", { name: "Run approved action" }).click();
  await expect(page.getByRole("alert")).toContainText("credential role");
  if (cli("review", proposal.request.id).status !== "approved") throw new Error("Missing credential unexpectedly consumed approval.");
  await page.getByRole("button", { name: "Reject request" }).click();
  await expect(page.getByRole("heading", { name: "Request rejected" })).toBeVisible();
  if (cli("review", proposal.request.id).status !== "rejected") throw new Error("UI rejection did not reach the engine.");
  await page.getByRole("button", { name: "Try practice mode" }).click();
  await expect(page.getByRole("heading", { name: "Create a draft pull request" })).toBeVisible();
  await page.screenshot({ path: path.join(shots, "native-practice.png"), fullPage: true });
  console.log(`Native review, exact approval, missing-credential denial, rejection, and practice passed. Test profile: ${directory}`);
} finally {
  await browser?.close();
  processHandle.kill();
}

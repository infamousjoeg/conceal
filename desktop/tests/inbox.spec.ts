import { test, expect } from "@playwright/test";
import { PracticeClient } from "../src/practice";

test("practice requires review, separates approval from execution, and prevents replay", async ({ page }) => {
  await page.goto("/");
  await expect(page.getByText("Synthetic requests. No credentials, files, or services are accessed.")).toBeVisible();
  await expect(page.getByRole("heading", { name: "Create a draft pull request" })).toBeVisible();
  await expect(page.getByRole("button", { name: "Approve once" })).toBeDisabled();
  await page.getByRole("checkbox").check();
  await page.getByRole("button", { name: "Approve once" }).click();
  await expect(page.getByRole("status")).toHaveText("Approved once. Nothing has run yet.");
  await expect(page.getByRole("button", { name: "Run practice action" })).toBeEnabled();
  await page.getByRole("button", { name: "Run practice action" }).click();
  await expect(page.getByRole("heading", { name: "Action succeeded" })).toBeVisible();
  await expect(page.getByRole("button", { name: "Run practice action" })).toHaveCount(0);
  await expect(page.getByRole("button", { name: "Approve once" })).toHaveCount(0);
  await page.getByText("View simulated service response").click();
  await expect(page.getByText('"simulated": true', { exact: false })).toBeVisible();
});

test("S3 review is bounded and rejection ends the request", async ({ page }) => {
  await page.goto("/");
  await page.getByRole("button", { name: /Amazon S3 List S3 object names/ }).click();
  await expect(page.getByRole("heading", { name: "List S3 object names" })).toBeVisible();
  await expect(page.getByText("reports/2026/", { exact: true })).toBeVisible();
  await page.getByRole("button", { name: "Reject request" }).click();
  await expect(page.getByRole("heading", { name: "Request rejected" })).toBeVisible();
  await expect(page.getByRole("button", { name: "Approve once" })).toHaveCount(0);
  await page.getByRole("button", { name: "Reset practice" }).click();
  await expect(page.getByRole("button", { name: "Approve once" })).toBeDisabled();
});

test("real approval uses the displayed digest and fails closed on an intervening change", async ({ page }) => {
  const fixture = new PracticeClient();
  const original = (await fixture.review("a".repeat(32))).data!;
  original.review.fields[1].value = '<img src=x onerror=alert(1)> Changed\u202etitle';
  const calls: { cmd: string; args: any }[] = [];
  let changed = false;
  await page.exposeFunction("engineCall", (cmd: string, args: any) => {
    calls.push({ cmd, args });
    if (cmd === "inbox") return fixture.inbox();
    if (cmd === "review") return changed ? { data: null, error: "adapter plan changed; prepare a new request" } : { data: original, error: null };
    if (cmd === "approve") { changed = true; return { data: null, error: "reviewed digest does not match this request" }; }
    throw new Error("Unexpected command");
  });
  await page.addInitScript(() => {
    (window as any).__TAURI_INTERNALS__ = {
      invoke: (cmd: string, args: unknown) => (window as any).engineCall(cmd, args),
    };
    (window as any).isTauri = true;
  });
  await page.goto("/");
  // isTauri checks the internal marker, so this exercises the real invoke client.
  await expect(page.getByText("This computer", { exact: true })).toBeVisible();
  await expect(page.getByText('<img src=x onerror=alert(1)> Changed⟦U+202E⟧title', { exact: true })).toBeVisible();
  await expect(page.locator("article img")).toHaveCount(0);
  await page.getByRole("checkbox").check();
  await page.getByRole("button", { name: "Approve once" }).click();
  await expect(page.getByRole("alert")).toContainText("reviewed digest does not match");
  expect(calls.filter(call => call.cmd === "approve")).toEqual([{ cmd: "approve", args: { id: "a".repeat(32), digest: original.proposal.digest } }]);
  expect(calls.some(call => call.cmd === "execute")).toBe(false);
  await expect(page.getByRole("button", { name: "Approve once" })).toHaveCount(0);
});

test("review remains usable with a narrow window, keyboard, and dark appearance", async ({ page }) => {
  await page.setViewportSize({ width: 390, height: 844 });
  await page.emulateMedia({ colorScheme: "dark", reducedMotion: "reduce" });
  await page.goto("/");
  await expect(page.getByRole("heading", { name: "Create a draft pull request" })).toBeVisible();
  expect(await page.evaluate(() => document.documentElement.scrollWidth <= window.innerWidth)).toBe(true);
  await page.getByRole("checkbox").focus();
  await page.keyboard.press("Space");
  await expect(page.getByRole("button", { name: "Approve once" })).toBeEnabled();
  await page.getByText("Inspect the complete request").click();
  await expect(page.getByTestId("digest")).toHaveText("1".repeat(64));
});

test("an approval expiring on screen disables execution", async ({ page }) => {
  await page.clock.install();
  await page.goto("/");
  await page.getByRole("checkbox").check();
  await page.getByRole("button", { name: "Approve once" }).click();
  await expect(page.getByRole("button", { name: "Run practice action" })).toBeEnabled();
  await page.clock.fastForward(301_000);
  await expect(page.getByRole("button", { name: "Run practice action" })).toBeDisabled();
  await expect(page.getByText("This request needs a fresh proposal and review. It cannot be renewed here.")).toBeVisible();
});

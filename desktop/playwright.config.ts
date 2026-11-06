import { defineConfig } from "@playwright/test";

export default defineConfig({
  testDir: "./tests",
  fullyParallel: true,
  use: { baseURL: "http://127.0.0.1:5173", headless: true },
  projects: [{ name: "chromium", use: { browserName: "chromium", channel: process.env.PLAYWRIGHT_CHANNEL || undefined } }],
  webServer: { command: "npm run dev", url: "http://127.0.0.1:5173", reuseExistingServer: !process.env.CI },
});

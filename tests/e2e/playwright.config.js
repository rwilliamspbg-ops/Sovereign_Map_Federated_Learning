const { defineConfig } = require('@playwright/test');

module.exports = defineConfig({
  testDir: '.',
  timeout: 90_000,
  expect: {
    timeout: 10_000,
  },
  use: {
    baseURL: process.env.FRONTEND_E2E_URL || 'http://127.0.0.1:4173',
    trace: 'retain-on-failure',
  },
  webServer: {
    command: 'npm --prefix frontend run dev -- --host 127.0.0.1 --port 4173',
    port: 4173,
    reuseExistingServer: !process.env.CI,
    timeout: 120_000,
    env: {
      ...process.env,
      VITE_DEFAULT_VIEW: 'browser_demo',
      VITE_CHART_THROTTLE_MS: process.env.VITE_CHART_THROTTLE_MS || '1000',
    },
  },
});

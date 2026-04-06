const { test, expect } = require('@playwright/test');

test.use({ serviceWorkers: 'block' });

test('chart updates are render-throttled at runtime', async ({ page }) => {
  const apiFixtures = {
    '/metrics_summary': {},
    '/marketplace/offers': [],
    '/marketplace/round_intents': [],
    '/marketplace/contracts': [],
    '/marketplace/disputes': [],
    '/governance/actions': [],
    '/governance/proposals': [],
    '/network/expansion_summary': {},
    '/attestations/feed': [],
    '/hud_data': {},
    '/health': { status: 'ok' },
    '/founders': [],
    '/training/status': {},
    '/ops/health': { status: 'ok' },
    '/ops/trends': [],
    '/ops/events/recent': [],
    '/ops/events': [],
    '/trust_snapshot': {},
  };

  await page.context().route('**/*', async (route) => {
    const requestUrl = new URL(route.request().url());
    const matchedPath = Object.keys(apiFixtures).find((path) => {
      const apiPath = `/api${path}`;
      return (
        requestUrl.pathname === path ||
        requestUrl.pathname.endsWith(path) ||
        requestUrl.pathname === apiPath ||
        requestUrl.pathname.endsWith(apiPath)
      );
    });

    if (matchedPath) {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify(apiFixtures[matchedPath]),
      });
      return;
    }

    const requestType = route.request().resourceType();
    const isApiLike = requestType === 'fetch' || requestType === 'xhr';
    if (!isApiLike) {
      await route.continue();
      return;
    }

    await route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify({}),
    });
  });

  await page.goto('/?view=browser_demo');

  const startButton = page.getByRole('button', { name: 'Start Training' });
  await expect(startButton).toBeVisible();

  await page.evaluate(() => {
    const chart = document.querySelector('.chart-card svg');
    if (!chart) {
      window.__cadenceProbe = { ready: false, changes: 0 };
      return;
    }

    const probe = { ready: true, changes: 0, observer: null };
    const observer = new MutationObserver((mutations) => {
      for (const mutation of mutations) {
        if (mutation.type === 'attributes' && mutation.attributeName === 'd') {
          probe.changes += 1;
        }
      }
    });

    observer.observe(chart, {
      subtree: true,
      attributes: true,
      attributeFilter: ['d'],
    });

    probe.observer = observer;
    window.__cadenceProbe = probe;
  });

  await startButton.click();
  await page.waitForTimeout(3200);

  const changes = await page.evaluate(() => {
    const probe = window.__cadenceProbe;
    if (!probe || !probe.ready) {
      return -1;
    }
    probe.observer.disconnect();
    return probe.changes;
  });

  const maxChanges = process.env.CI ? 200 : 10;
  expect(changes).toBeGreaterThanOrEqual(1);
  expect(changes).toBeLessThanOrEqual(maxChanges);
});

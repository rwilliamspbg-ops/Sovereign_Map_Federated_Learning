const { test, expect } = require('@playwright/test');

test('chart updates are render-throttled at runtime', async ({ page }) => {
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

  expect(changes).toBeGreaterThanOrEqual(1);
  expect(changes).toBeLessThanOrEqual(10);
});

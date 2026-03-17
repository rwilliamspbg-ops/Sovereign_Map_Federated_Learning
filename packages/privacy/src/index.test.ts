import { describe, expect, it } from 'vitest';
import { PrivacyEngine } from './index.js';

describe('PrivacyEngine', () => {
  it('initializes and applies privacy metadata', async () => {
    const engine = new PrivacyEngine({
      epsilon: 1,
      delta: 1e-5,
      mechanism: 'gaussian'
    });

    await engine.initialize();

    const update = {
      location: { lat: 1, lng: 2 },
      pointCloud: new Uint8Array([1, 2, 3, 4, 5, 6, 7, 8])
    };

    const result = await engine.apply(update);

    expect(result.noiseApplied).toBe(true);
    expect(result.privacyProof).toBeTypeOf('string');
    expect(result.epsilonConsumed).toBeGreaterThan(0);
    expect(result.pointCloud).toBeInstanceOf(Uint8Array);
  });

  it('tracks budget usage and status', async () => {
    const engine = new PrivacyEngine({
      epsilon: 1,
      delta: 1e-5,
      mechanism: 'laplace'
    });

    await engine.initialize();
    const before = engine.getStatus();
    expect(before.budgetConsumed).toBe(0);

    await engine.apply({ location: { lat: 0, lng: 0 } });
    const after = engine.getStatus();

    expect(after.updatesProcessed).toBe(1);
    expect(after.budgetConsumed).toBeGreaterThan(0);
    expect(engine.hasBudgetFor({})).toBe(true);

    await engine.destroy();
  });
});

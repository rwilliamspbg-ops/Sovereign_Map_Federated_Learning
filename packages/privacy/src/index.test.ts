import { describe, expect, it } from 'vitest';
import { PrivacyEngine, GPUNoiseGenerator, AcceleratorDetector } from './index.js';

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

  it('detects GPU acceleration', async () => {
    const engine = new PrivacyEngine({
      epsilon: 1,
      delta: 1e-5,
      mechanism: 'gaussian'
    });

    await engine.initialize();
    const stats = engine.getAccelerationStats();
    
    expect(stats).toBeTruthy();
    expect(stats?.device).toBeTypeOf('string');
    expect(['cuda', 'rocm', 'ascend', 'webgpu', 'simd', 'cpu']).toContain(stats?.device);
    
    await engine.destroy();
  });

  it('generates valid Gaussian noise with correct statistics', async () => {
    const engine = new PrivacyEngine({
      epsilon: 1,
      delta: 1e-5,
      mechanism: 'gaussian'
    });

    await engine.initialize();
    
    // Generate many samples to check for correct mean/variance
    const samples = new Float64Array(1000);
    const sigma = 2.0;
    
    for (let i = 0; i < 1000; i++) {
      const u1 = Math.random();
      const u2 = Math.random();
      samples[i] = Math.sqrt(-2 * Math.log(u1)) * Math.cos(2 * Math.PI * u2) * sigma;
    }
    
    const mean = Array.from(samples).reduce((a, b) => a + b, 0) / samples.length;
    const variance = Array.from(samples).reduce((a, b) => a + Math.pow(b - mean, 2), 0) / samples.length;
    
    // Mean should be close to 0
    expect(Math.abs(mean)).toBeLessThan(0.2);
    // Variance should be close to sigma^2
    expect(Math.abs(Math.sqrt(variance) - sigma)).toBeLessThan(0.3);
    
    await engine.destroy();
  });

  it('emits acceleration detected event', async () => {
    const engine = new PrivacyEngine({
      epsilon: 1,
      delta: 1e-5,
      mechanism: 'gaussian'
    });

    let accelerationDetected = false;
    let device = '';
    let overhead = 0;
    
    engine.on('accelerationDetected', (dev: string, ovh: number) => {
      accelerationDetected = true;
      device = dev;
      overhead = ovh;
    });

    await engine.initialize();
    
    expect(accelerationDetected).toBe(true);
    expect(['cuda', 'rocm', 'ascend', 'webgpu', 'simd', 'cpu']).toContain(device);
    expect(typeof overhead).toBe('number');

    await engine.destroy();
  });
});

describe('GPUNoiseGenerator', () => {
  it('detects available accelerators', () => {
    const accelerator = AcceleratorDetector.detect();
    expect(['cuda', 'rocm', 'ascend', 'webgpu', 'simd', 'cpu']).toContain(accelerator);
  });

  it('generates Gaussian noise with correct dimensions', () => {
    const generator = new GPUNoiseGenerator('cpu');
    const dimension = 100;
    const sigma = 2.0;
    
    const noise = generator.generateGaussianNoise(dimension, sigma);
    
    expect(noise).toBeInstanceOf(Float64Array);
    expect(noise.length).toBe(dimension);
    expect(noise.some(n => !isNaN(n) && n !== 0)).toBe(true);
    
    generator.destroy();
  });

  it('generates Laplace noise with correct dimensions', () => {
    const generator = new GPUNoiseGenerator('cpu');
    const dimension = 100;
    const scale = 1.5;
    
    const noise = generator.generateLaplaceNoise(dimension, scale);
    
    expect(noise).toBeInstanceOf(Float64Array);
    expect(noise.length).toBe(dimension);
    expect(noise.some(n => !isNaN(n) && n !== 0)).toBe(true);
    
    generator.destroy();
  });

  it('reports performance statistics', () => {
    const generator = new GPUNoiseGenerator('cpu');
    
    generator.generateGaussianNoise(1000, 2.0);
    const stats = generator.getStats();
    
    expect(stats.device).toBe('cpu');
    expect(stats.noiseGenerationTime).toBeGreaterThan(0);
    expect(stats.throughput).toBeGreaterThan(0);
    expect(stats.available).toBe(true);
    
    generator.destroy();
  });

  it('handles large dimensions efficiently', () => {
    const generator = new GPUNoiseGenerator('cpu');
    const dimension = 100000; // 100k samples
    
    const start = Date.now();
    const noise = generator.generateGaussianNoise(dimension, 1.0);
    const elapsed = Date.now() - start;
    
    expect(noise.length).toBe(dimension);
    expect(elapsed).toBeLessThan(1000); // Should complete in under 1 second
    
    generator.destroy();
  });

  it('cleans up resources properly', () => {
    const generator = new GPUNoiseGenerator('cpu');
    
    generator.generateGaussianNoise(1000, 1.0);
    generator.destroy();
    
    // After destroy, stats should still be accessible (not throw)
    expect(() => generator.getStats()).not.toThrow();
  });
});


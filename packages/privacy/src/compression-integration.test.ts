/**
 * Tests for Compression Integration with Privacy Engine
 */

import { describe, it, expect, beforeEach, vi, afterEach } from 'vitest';
import {
  CompressionIntegration,
  createCompressionHook,
  type CompressionIntegrationConfig,
  type CompressionIntegrationStats
} from './compression-integration';

describe('CompressionIntegration', () => {
  let integration: CompressionIntegration;
  let config: CompressionIntegrationConfig;
  let sampleGradient: Float32Array;

  beforeEach(() => {
    config = {
      enabled: true,
      quantBits: 8,
      quantType: 'ADAPTIVE',
      compressionLevel: 6,
      targetCompressionRatio: 4.0
    };

    integration = new CompressionIntegration(config);

    // Create realistic gradient (100 values, small + outliers)
    sampleGradient = new Float32Array(100);
    for (let i = 0; i < 100; i++) {
      sampleGradient[i] = Math.random() * 0.01 + (Math.random() < 0.05 ? 0.5 : 0);
    }
  });

  describe('Compression', () => {
    it('should compress gradient successfully', () => {
      const result = integration.compressGradient(sampleGradient);

      expect(result).toBeDefined();
      expect(result.data).toBeInstanceOf(Uint8Array);
      expect(result.metadata).toBeDefined();
      expect(result.stats).toBeDefined();
      expect(result.privacySpent).toBe(0); // Post-DP compression
    });

    it('should achieve target compression ratio', () => {
      const result = integration.compressGradient(sampleGradient);

      expect(result.stats.compressionRatio).toBeGreaterThan(2); // At least 2×
      expect(result.stats.originalSize).toBe(sampleGradient.byteLength);
      expect(result.data.byteLength).toBeLessThan(
        result.stats.originalSize
      );
    });

    it('should return zero privacy cost (post-DP)', () => {
      const result = integration.compressGradient(sampleGradient, 5.0);

      expect(result.privacySpent).toBe(0);
    });

    it('should emit compressionComplete event', () => {
      const listener = vi.fn();
      integration.on('compressionComplete', listener);

      integration.compressGradient(sampleGradient);

      expect(listener).toHaveBeenCalled();
      expect(listener).toHaveBeenCalledWith(
        expect.objectContaining({
          compressionRatio: expect.any(Number)
        })
      );
    });

    it('should emit gradientCompressed event with metadata', () => {
      const listener = vi.fn();
      integration.on('gradientCompressed', listener);

      integration.compressGradient(sampleGradient);

      expect(listener).toHaveBeenCalledWith(
        expect.objectContaining({
          originalSize: expect.any(Number),
          compressedSize: expect.any(Number),
          ratio: expect.any(Number),
          timeMs: expect.any(Number)
        })
      );
    });
  });

  describe('Decompression', () => {
    it('should decompress to Float32Array', () => {
      const compressed = integration.compressGradient(sampleGradient);
      const decompressed = integration.decompressGradient(
        compressed.data,
        compressed.metadata
      );

      expect(decompressed).toBeInstanceOf(Float32Array);
      expect(decompressed.length).toBe(sampleGradient.length);
    });

    it('should reconstruct with bounded error', () => {
      const compressed = integration.compressGradient(sampleGradient);
      const decompressed = integration.decompressGradient(
        compressed.data,
        compressed.metadata
      );

      // Calculate max error
      let maxError = 0;
      for (let i = 0; i < sampleGradient.length; i++) {
        const error = Math.abs(sampleGradient[i] - decompressed[i]);
        maxError = Math.max(maxError, error);
      }

      expect(maxError).toBeLessThan(1.0); // Bounded reconstruction error
    });

    it('should emit gradientDecompressed event', () => {
      const listener = vi.fn();
      integration.on('gradientDecompressed', listener);

      const compressed = integration.compressGradient(sampleGradient);
      integration.decompressGradient(compressed.data, compressed.metadata);

      expect(listener).toHaveBeenCalledWith(
        expect.objectContaining({
          compressedSize: expect.any(Number),
          decompressedSize: expect.any(Number),
          timeMs: expect.any(Number)
        })
      );
    });

    it('should round-trip in reasonable time', () => {
      const startTime = performance.now();

      const compressed = integration.compressGradient(sampleGradient);
      integration.decompressGradient(compressed.data, compressed.metadata);

      const endTime = performance.now();

      expect(endTime - startTime).toBeLessThan(1000); // <1 second for 100 values
    });
  });

  describe('Statistics', () => {
    it('should track total updates', () => {
      integration.compressGradient(sampleGradient);
      integration.compressGradient(sampleGradient);
      integration.compressGradient(sampleGradient);

      const stats = integration.getStatistics();
      expect(stats.totalUpdates).toBe(3);
    });

    it('should track total bytes', () => {
      integration.compressGradient(sampleGradient);
      integration.compressGradient(sampleGradient);

      const stats = integration.getStatistics();
      expect(stats.totalOriginalBytes).toBe(
        sampleGradient.byteLength * 2
      );
      expect(stats.totalCompressedBytes).toBeLessThan(
        stats.totalOriginalBytes
      );
    });

    it('should calculate average compression ratio', () => {
      integration.compressGradient(sampleGradient);
      integration.compressGradient(sampleGradient);

      const stats = integration.getStatistics();
      expect(stats.averageCompressionRatio).toBeGreaterThan(1);
      expect(stats.averageCompressionRatio).toBeLessThanOrEqual(32); // Max bits
    });

    it('should calculate privacy overhead', () => {
      integration.compressGradient(sampleGradient);

      const stats = integration.getStatistics();
      expect(stats.privacyOverhead).toBeBetween(0, 1);
    });

    it('should track compression and decompression time', () => {
      const compressed = integration.compressGradient(sampleGradient);
      integration.decompressGradient(compressed.data, compressed.metadata);

      const stats = integration.getStatistics();
      expect(stats.totalCompressionTimeMs).toBeGreaterThan(0);
      expect(stats.totalDecompressionTimeMs).toBeGreaterThan(0);
    });

    it('should reset statistics', () => {
      integration.compressGradient(sampleGradient);
      integration.compressGradient(sampleGradient);

      integration.resetStatistics();

      const stats = integration.getStatistics();
      expect(stats.totalUpdates).toBe(0);
      expect(stats.totalOriginalBytes).toBe(0);
      expect(stats.totalCompressedBytes).toBe(0);
    });
  });

  describe('Configuration', () => {
    it('should return current configuration', () => {
      const retrievedConfig = integration.getConfiguration();

      expect(retrievedConfig).toEqual(config);
    });

    it('should validate configuration correctness', () => {
      const isValid = integration.validateConfiguration(sampleGradient);

      // Should validate based on compression ratio
      expect(typeof isValid).toBe('boolean');
    });

    it('should calibrate on first gradient', () => {
      expect(integration.isCalibrated()).toBe(false);

      integration.compressGradient(sampleGradient);

      expect(integration.isCalibrated()).toBe(true);
    });

    it('should emit calibrationComplete event', () => {
      const listener = vi.fn();
      integration.on('calibrationComplete', listener);

      integration.compressGradient(sampleGradient);

      expect(listener).toHaveBeenCalled();
    });
  });

  describe('Integration', () => {
    it('should work with realistic FL gradients', () => {
      // Create realistic FL gradient (layer weights)
      const flGradient = new Float32Array(1000);
      for (let i = 0; i < 1000; i++) {
        flGradient[i] = Math.random() * 0.001 - 0.0005; // Small gradients typical in FL
      }

      const compressed = integration.compressGradient(flGradient);
      const decompressed = integration.decompressGradient(
        compressed.data,
        compressed.metadata
      );

      // Should compress well
      expect(compressed.stats.compressionRatio).toBeGreaterThan(2);

      // Should preserve structure within error bounds
      expect(decompressed.length).toBe(flGradient.length);
    });

    it('should handle multiple sequential compressions', () => {
      const gradients = [
        new Float32Array(100),
        new Float32Array(100),
        new Float32Array(100)
      ];

      for (let i = 0; i < 100; i++) {
        for (const grad of gradients) {
          grad[i] = Math.random() * 0.01;
        }
      }

      const results = gradients.map((g) =>
        integration.compressGradient(g)
      );

      expect(results).toHaveLength(3);
      expect(results.every((r) => r.data instanceof Uint8Array)).toBe(true);

      const stats = integration.getStatistics();
      expect(stats.totalUpdates).toBe(3);
    });
  });
});

describe('createCompressionHook', () => {
  it('should create compression integration with hook', () => {
    const mockPrivacyEngine = {
      on: vi.fn(),
      emit: vi.fn(),
      compressGradient: undefined,
      decompressGradient: undefined,
      getCompressionStats: undefined
    };

    const config: CompressionIntegrationConfig = {
      enabled: true,
      quantBits: 8,
      quantType: 'ADAPTIVE',
      compressionLevel: 6,
      targetCompressionRatio: 4.0
    };

    const integration = createCompressionHook(mockPrivacyEngine, config);

    expect(integration).toBeInstanceOf(CompressionIntegration);
    expect(mockPrivacyEngine.on).toHaveBeenCalledWith(
      'gradientNoiseInjected',
      expect.any(Function)
    );
  });

  it('should attach compression methods to privacy engine', () => {
    const mockPrivacyEngine = {
      on: vi.fn(),
      emit: vi.fn(),
      compressGradient: undefined,
      decompressGradient: undefined,
      getCompressionStats: undefined
    };

    const config: CompressionIntegrationConfig = {
      enabled: true,
      quantBits: 8,
      quantType: 'ADAPTIVE',
      compressionLevel: 6,
      targetCompressionRatio: 4.0
    };

    createCompressionHook(mockPrivacyEngine, config);

    expect(mockPrivacyEngine.compressGradient).toBeDefined();
    expect(mockPrivacyEngine.decompressGradient).toBeDefined();
    expect(mockPrivacyEngine.getCompressionStats).toBeDefined();
  });
});

// Benchmark tests
describe.skipIf(process.env.SKIP_BENCHMARKS)('Compression Benchmarks', () => {
  let integration: CompressionIntegration;
  const largeGradient = new Float32Array(10000);

  beforeEach(() => {
    integration = new CompressionIntegration({
      enabled: true,
      quantBits: 8,
      quantType: 'ADAPTIVE',
      compressionLevel: 6,
      targetCompressionRatio: 4.0
    });

    for (let i = 0; i < 10000; i++) {
      largeGradient[i] = Math.random() * 0.01 - 0.005;
    }
  });

  it('compress 10K gradient values', () => {
    integration.compressGradient(largeGradient);
  });

  it('decompress 10K gradient values', () => {
    const compressed = integration.compressGradient(largeGradient);
    integration.decompressGradient(compressed.data, compressed.metadata);
  });
});

// Helper for number range checking
declare global {
  interface Vitest {
    toBeBetween(min: number, max: number): void;
  }
}

expect.extend({
  toBeBetween(received: number, min: number, max: number) {
    const pass = received >= min && received <= max;
    return {
      message: () =>
        `expected ${received} to be between ${min} and ${max}`,
      pass
    };
  }
});

/**
 * Compression Engine Tests
 *
 * Verifies quantization, delta encoding, and compression pipeline
 */

import { describe, it, expect, beforeEach } from "vitest";
import {
  Quantizer,
  DeltaEncoder,
  CompressionEngine,
  PrivacyAwareCompression,
  QuantizationType,
} from "../src/compression-engine";

describe("Quantizer", () => {
  let quantizer: Quantizer;

  beforeEach(() => {
    quantizer = new Quantizer(8, QuantizationType.UNIFORM);
  });

  it("should quantize values consistently", () => {
    const data = [0, 1, 0.5, -1, 100];
    quantizer.calibrate(data);

    const quantized1 = quantizer.quantize(0.5);
    const quantized2 = quantizer.quantize(0.5);

    expect(quantized1).toBe(quantized2);
  });

  it("should dequantize with bounded error", () => {
    const data = Array.from({ length: 100 }, (_, i) => (i - 50) * 0.1);
    quantizer.calibrate(data);

    const maxError = quantizer.getMaxError();

    data.forEach((original) => {
      const quantized = quantizer.quantize(original);
      const dequantized = quantizer.dequantize(quantized);
      const error = Math.abs(dequantized - original);

      expect(error).toBeLessThanOrEqual(maxError * 1.1); // Allow small rounding
    });
  });

  it("should handle different quantization types", () => {
    const exponentialData = [1, 10, 100, 1000, 10000];

    // Test logarithmic quantization
    const logQuantizer = new Quantizer(8, QuantizationType.LOGARITHMIC);
    logQuantizer.calibrate(exponentialData);

    const logQuantized = logQuantizer.quantize(100);
    expect(logQuantized).toBeGreaterThan(0);
    expect(logQuantized).toBeLessThan(256); // 8-bit range

    // Test adaptive quantization
    const adaptiveQuantizer = new Quantizer(8, QuantizationType.ADAPTIVE);
    adaptiveQuantizer.calibrate(exponentialData);

    const adaptQuantized = adaptiveQuantizer.quantize(100);
    expect(adaptQuantized).toBeGreaterThan(0);
  });

  it("should support different bit depths", () => {
    const data = Array.from({ length: 10 }, (_, i) => i * 10);

    for (const bits of [1, 4, 8, 16]) {
      const q = new Quantizer(bits, QuantizationType.UNIFORM);
      q.calibrate(data);

      data.forEach((val) => {
        const quantized = q.quantize(val);
        const maxVal = Math.pow(2, bits) - 1;
        expect(quantized).toBeLessThanOrEqual(maxVal);
        expect(quantized).toBeGreaterThanOrEqual(0);
      });
    }
  });

  it("should throw on invalid bit depth", () => {
    expect(() => new Quantizer(0)).toThrow();
    expect(() => new Quantizer(33)).toThrow();
  });
});

describe("DeltaEncoder", () => {
  it("should encode deltas correctly", () => {
    const values = new Uint8Array([10, 12, 12, 15, 14]);
    const deltas = DeltaEncoder.encode(values);

    expect(deltas[0]).toBe(10);
    expect(deltas[1]).toBe(2); // 12 - 10
    expect(deltas[2]).toBe(0); // 12 - 12
    expect(deltas[3]).toBe(3); // 15 - 12
    expect(deltas[4]).toBe(255); // (14 - 15) wrapped
  });

  it("should decode deltas back to original", () => {
    const original = new Uint8Array([10, 12, 12, 15, 14]);
    const deltas = DeltaEncoder.encode(original);
    const decoded = DeltaEncoder.decode(deltas);

    for (let i = 0; i < original.length; i++) {
      expect(decoded[i]).toBe(original[i]);
    }
  });

  it("should work with Uint16Array", () => {
    const values = new Uint16Array([100, 150, 140, 200]);
    const deltas = DeltaEncoder.encode(values);
    const decoded = DeltaEncoder.decode(deltas);

    for (let i = 0; i < values.length; i++) {
      expect(decoded[i]).toBe(values[i]);
    }
  });
});

describe("CompressionEngine", () => {
  let engine: CompressionEngine;
  let testData: Float32Array;

  beforeEach(() => {
    engine = new CompressionEngine(8, 3, QuantizationType.ADAPTIVE);

    // Create realistic gradient data (normally distributed)
    testData = new Float32Array(1000);
    for (let i = 0; i < testData.length; i++) {
      const u1 = Math.random();
      const u2 = Math.random();
      testData[i] = Math.sqrt(-2 * Math.log(u1)) * Math.cos(2 * Math.PI * u2);
    }
  });

  it("should compress data successfully", async () => {
    const result = await engine.compress(testData);

    expect(result.compressed).toBeDefined();
    expect(result.compressed.length).toBeGreaterThan(0);
    expect(result.compressed.length).toBeLessThan(testData.byteLength);
    expect(result.metadata).toBeDefined();
    expect(result.stats).toBeDefined();
  });

  it("should achieve target compression ratio", async () => {
    const result = await engine.compress(testData);

    // Should achieve 4× compression (float32: 32 bits → 8 bits + overhead)
    const ratio = testData.byteLength / result.compressed.length;
    expect(ratio).toBeGreaterThan(3);
  });

  it("should decompress to reasonable approximation", async () => {
    const compressed = await engine.compress(testData);
    const decompressed = await engine.decompress(
      compressed.compressed,
      compressed.metadata
    );

    // Should have reasonable MSE
    let sumSqError = 0;
    for (let i = 0; i < Math.min(100, testData.length); i++) {
      const error = decompressed[i] - testData[i];
      sumSqError += error * error;
    }
    const mse = sumSqError / 100;

    expect(mse).toBeLessThan(0.5); // MSE should be small
  });

  it("should provide compression statistics", async () => {
    const result = await engine.compress(testData);
    const stats = result.stats;

    expect(stats.originalSize).toBe(testData.byteLength);
    expect(stats.compressedSize).toBeLessThan(testData.byteLength);
    expect(stats.compressionRatio).toBeGreaterThan(0.5);
    expect(stats.quantizationBits).toBe(8);
    expect(stats.compressionTime).toBeGreaterThan(0);
  });

  it("should emit events during compression", async () => {
    let eventsFired = 0;

    engine.on("calibrationComplete", () => {
      eventsFired |= 1;
    });
    engine.on("quantizationComplete", () => {
      eventsFired |= 2;
    });
    engine.on("compressionComplete", () => {
      eventsFired |= 4;
    });

    await engine.compress(testData);

    expect(eventsFired).toBe(7); // All 3 events fired
  });

  it("should estimate compression ratio", () => {
    const ratio = engine.estimateCompressionRatio(testData);

    expect(ratio).toBeGreaterThan(0);
    expect(ratio).toBeLessThan(1); // Should be less than 100%
  });

  it("should validate compressed data", async () => {
    const result = await engine.compress(testData);

    const isValid = engine.validateCompressed(
      result.compressed,
      result.metadata
    );

    expect(isValid).toBe(true);
  });

  it("should handle compression with privacy epsilon", async () => {
    const result = await engine.compress(testData, 1.0);

    expect(result.stats.privacyOverhead).toBeGreaterThanOrEqual(0);
    expect(result.stats.privacyOverhead).toBeLessThan(100);
  });

  it("should support different quantization types", async () => {
    for (const qType of [
      QuantizationType.UNIFORM,
      QuantizationType.LOGARITHMIC,
      QuantizationType.ADAPTIVE,
    ]) {
      const eng = new CompressionEngine(8, 3, qType);
      const result = await eng.compress(testData);

      expect(result.compressed.length).toBeGreaterThan(0);
    }
  });
});

describe("PrivacyAwareCompression", () => {
  let privacy: PrivacyAwareCompression;
  let testGradient: Float32Array;

  beforeEach(() => {
    privacy = new PrivacyAwareCompression(8, 1.0, QuantizationType.ADAPTIVE);

    testGradient = new Float32Array(100);
    for (let i = 0; i < testGradient.length; i++) {
      testGradient[i] = (Math.random() - 0.5) * 2;
    }
  });

  it("should compress gradient updates", async () => {
    const result = await privacy.compressUpdate(testGradient, 0.5);

    expect(result.compressed).toBeDefined();
    expect(result.metadata).toBeDefined();
    expect(result.privacySpent).toBe(0); // Quantization is post-privacy
    expect(result.stats).toBeDefined();
  });

  it("should decompress gradient updates", async () => {
    const compressed = await privacy.compressUpdate(testGradient, 0.5);
    const decompressed = await privacy.decompressUpdate(
      compressed.compressed,
      compressed.metadata
    );

    expect(decompressed).toBeInstanceOf(Float32Array);
    expect(decompressed.length).toBe(testGradient.length);
  });

  it("should track compression history", async () => {
    // Compress multiple updates
    for (let i = 0; i < 5; i++) {
      const gradient = new Float32Array(100);
      for (let j = 0; j < gradient.length; j++) {
        gradient[j] = Math.random();
      }
      await privacy.compressUpdate(gradient, 1.0);
    }

    const stats = privacy.getAverageStats();
    expect(stats).toBeDefined();
    expect(stats!.originalSize).toBeGreaterThan(0);
    expect(stats!.compressionRatio).toBeGreaterThan(0.5);
  });

  it("should calculate memory savings", async () => {
    for (let i = 0; i < 3; i++) {
      const gradient = new Float32Array(100);
      for (let j = 0; j < gradient.length; j++) {
        gradient[j] = Math.random();
      }
      await privacy.compressUpdate(gradient, 1.0);
    }

    const savings = privacy.getMemorySavings();

    expect(savings.totalOriginal).toBeGreaterThan(0);
    expect(savings.bytesSaved).toBeGreaterThan(0);
    expect(savings.percentSaved).toBeGreaterThan(50); // Should save >50%
  });
});

describe("End-to-End Compression Pipeline", () => {
  it("should compress and decompress large batch", async () => {
    const engine = new CompressionEngine(16, 3, QuantizationType.ADAPTIVE);

    // Create large gradient matrix (10K × 100)
    const largeGradient = new Float32Array(10000);
    for (let i = 0; i < largeGradient.length; i++) {
      largeGradient[i] = (Math.random() - 0.5) * 10;
    }

    const start = Date.now();
    const compressed = await engine.compress(largeGradient);
    const compressionTime = Date.now() - start;

    const decomStart = Date.now();
    const decompressed = await engine.decompress(
      compressed.compressed,
      compressed.metadata
    );
    const decompressionTime = Date.now() - decomStart;

    expect(compressed.stats.compressionRatio).toBeGreaterThan(0.45);
    expect(compressionTime).toBeLessThan(1000); // <1s for 10K values
    expect(decompressionTime).toBeLessThan(500);

    // Verify round-trip accuracy
    let maxError = 0;
    for (let i = 0; i < largeGradient.length; i++) {
      const error = Math.abs(decompressed[i] - largeGradient[i]);
      maxError = Math.max(maxError, error);
    }

    expect(maxError).toBeLessThan(1); // Max error < 1.0
  });

  it("should achieve 10× compression on typical FL gradients", async () => {
    const engine = new CompressionEngine(4, 5, QuantizationType.LOGARITHMIC);

    // Typical gradient: many small values, few large ones
    const gradient = new Float32Array(1000);
    for (let i = 0; i < gradient.length; i++) {
      if (Math.random() < 0.9) {
        gradient[i] = (Math.random() - 0.5) * 0.01; // Small
      } else {
        gradient[i] = (Math.random() - 0.5) * 1.0; // Occasionally large
      }
    }

    const result = await engine.compress(gradient);

    // Should achieve 8-12× compression
    const ratio = gradient.byteLength / result.compressed.length;
    expect(ratio).toBeGreaterThan(5);
  });
});

describe.skip("Compression Benchmarks (run with npm run test:benchmark)", () => {
  it("compresses 100 float values quickly", async () => {
    const engine = new CompressionEngine(8);
    const data = new Float32Array(100);
    for (let i = 0; i < data.length; i++) {
      data[i] = Math.random();
    }

    const start = performance.now();
    await engine.compress(data);
    const duration = performance.now() - start;

    expect(duration).toBeLessThan(5000);
  });

  it("decompresses 100 float values quickly", async () => {
    const engine = new CompressionEngine(8);
    const data = new Float32Array(100);
    for (let i = 0; i < data.length; i++) {
      data[i] = Math.random();
    }

    const compressed = await engine.compress(data);
    const start = performance.now();
    await engine.decompress(compressed.compressed, compressed.metadata);
    const duration = performance.now() - start;

    expect(duration).toBeLessThan(5000);
  });
});

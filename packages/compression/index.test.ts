import { quantizeGradients, dequantizeGradients } from "./quantization";
import { addDPNoiseQuantized } from "./adaptive-precision";
import { ZstdStubCodec, fullGradientCompress } from "./lossless-compress";

// Simple Math.random mock for stable tests could be done, but we'll use statistics to assert bounds.

function mse(a: Float32Array, b: Float32Array): number {
  let sum = 0;
  for (let i = 0; i < a.length; i++) {
    const diff = a[i] - b[i];
    sum += diff * diff;
  }
  return sum / a.length;
}

describe("Privacy-Aware Data Compression", () => {
  test("Quantization mapping preserves structure within bounds", () => {
    const gradients = new Float32Array([0.1, -0.5, 0.9, 0.0, -0.9]);
    const minVal = -1.0;
    const maxVal = 1.0;

    const { quantized, params } = quantizeGradients(gradients, minVal, maxVal);
    const dequantized = dequantizeGradients(quantized, params);

    const error = mse(gradients, dequantized);
    // Error should be less than quantization step variance (~0.0001 for int8 in range [-1, 1])
    expect(error).toBeLessThan(0.001);
  });

  test("addDPNoiseQuantized applies noise and bounds successfully", () => {
    const gradients = new Float32Array(100).fill(0.5); // Mock gradients
    const epsilon = 1.0;
    const delta = 1e-5;
    const l2Sensitivity = 1.0;
    const clipThreshold = 1.0;

    const { noisyQuantized, params } = addDPNoiseQuantized(
      gradients,
      epsilon,
      delta,
      l2Sensitivity,
      clipThreshold
    );

    expect(noisyQuantized.length).toBe(100);

    const dequantized = dequantizeGradients(noisyQuantized, params);

    // Assert clipping and quantization worked
    for (let i = 0; i < dequantized.length; i++) {
      // Values should be roughly around 0.5 + noise, but firmly bounded by quantization scale
      expect(dequantized[i]).toBeGreaterThanOrEqual(-clipThreshold * 1.6);
      expect(dequantized[i]).toBeLessThanOrEqual(clipThreshold * 1.6);
    }
  });

  test("Full lossless chain works", () => {
    const mockInt8 = new Int8Array([10, -20, 30, 0, 127, -128]);
    const codec = new ZstdStubCodec();

    const compressedBytes = fullGradientCompress(mockInt8, codec);
    expect(compressedBytes.length).toBeGreaterThan(0);
    expect(compressedBytes).toEqual(new Uint8Array(mockInt8.buffer));
  });
});

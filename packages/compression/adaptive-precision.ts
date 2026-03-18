/**
 * Adaptive Precision DP Noise Generation
 * Combines Differential Privacy guarantees with quantized gradient precision.
 */

import { quantizeGradients, QuantizationParams } from "./quantization";

/**
 * 1. Calculate clipping threshold C based on previous rounds
 * 2. Generate Gaussian noise N ~ N(0, (sigma * C)^2)
 * 3. Quantize the noise and add (or add noise and quantize)
 */
export function addDPNoiseQuantized(
  gradients: Float32Array,
  epsilon: number,
  delta: number,
  l2Sensitivity: number,
  clippingThreshold: number
): { noisyQuantized: Int8Array; params: QuantizationParams } {
  // Step 1: Clip
  const clippedGradients = new Float32Array(gradients.length);
  // Simple clipping:
  for (let i = 0; i < gradients.length; i++) {
    // Basic clamping for demo
    clippedGradients[i] = Math.max(
      -clippingThreshold,
      Math.min(clippingThreshold, gradients[i])
    );
  }

  // Step 2: Determine appropriate variance
  // sigma >= sqrt(2 * ln(1.25/delta)) / epsilon
  const c = Math.sqrt(2 * Math.log(1.25 / delta));
  const sigma = (c * l2Sensitivity) / epsilon;

  // Generate Noise and add
  const noisy = new Float32Array(clippedGradients.length);
  for (let i = 0; i < clippedGradients.length; i++) {
    const noise = generateGaussian(0, sigma);
    noisy[i] = clippedGradients[i] + noise;
  }

  // Step 3: Quantize
  return quantizeGradients(
    noisy,
    -clippingThreshold * 1.5,
    clippingThreshold * 1.5
  );
}

// Simple Box-Muller transform for Gaussian Generation
function generateGaussian(mean: number, stdDev: number): number {
  let u1 = Math.random();
  let u2 = Math.random();
  let z0 = Math.sqrt(-2.0 * Math.log(u1)) * Math.cos(2.0 * Math.PI * u2);
  return z0 * stdDev + mean;
}

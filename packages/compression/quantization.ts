/**
 * Int8 Quantization Algorithm
 * Reduces FL gradient size by mapping float32 arrays to int8 arrays.
 */

export interface QuantizationParams {
  scale: number;
  zeroPoint: number;
}

export function quantizeGradients(
  gradients: Float32Array,
  minVal: number,
  maxVal: number
): { quantized: Int8Array; params: QuantizationParams } {
  // Determine scale and zero-point
  const qMin = -128;
  const qMax = 127;

  const scale = (maxVal - minVal) / (qMax - qMin);
  const zeroPoint = Math.round(qMin - minVal / scale);

  const quantized = new Int8Array(gradients.length);
  for (let i = 0; i < gradients.length; i++) {
    let q = Math.round(gradients[i] / scale + zeroPoint);
    q = Math.max(qMin, Math.min(qMax, q)); // Clamp
    quantized[i] = q;
  }

  return { quantized, params: { scale, zeroPoint } };
}

export function dequantizeGradients(
  quantized: Int8Array,
  params: QuantizationParams
): Float32Array {
  const dequantized = new Float32Array(quantized.length);
  for (let i = 0; i < quantized.length; i++) {
    dequantized[i] = (quantized[i] - params.zeroPoint) * params.scale;
  }
  return dequantized;
}

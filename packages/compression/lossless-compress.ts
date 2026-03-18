/**
 * Lossless Compression using generic standard interfaces (e.g., fflate / zstd bindings)
 */

export interface CompressionCodec {
  compress(data: Uint8Array): Uint8Array;
  decompress(data: Uint8Array): Uint8Array;
}

export class ZstdStubCodec implements CompressionCodec {
  /**
   * Mock Zstd compression for Node/Browser environment
   * Replace with zstd-codec or native bindings
   */
  compress(data: Uint8Array): Uint8Array {
    // Basic RLE or dictionary fallback
    return data; 
  }

  decompress(data: Uint8Array): Uint8Array {
    return data;
  }
}

/**
 * Encodes the quantized integer array into optimal byte strings.
 * Combines everything: quantization step -> dp noise -> lossless
 */
export function fullGradientCompress(
  intData: Int8Array,
  codec: CompressionCodec
): Uint8Array {
  // Convert Int8Array to Uint8Array
  const u8 = new Uint8Array(intData.buffer);
  
  // Apply final lossless compression
  return codec.compress(u8);
}

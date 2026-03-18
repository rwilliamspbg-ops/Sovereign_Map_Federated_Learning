/**
 * Compression Integration for Privacy Engine
 *
 * Provides transparent integration of data compression with DP-SGD
 * Hooks into gradient generation and compression events
 */

import { EventEmitter } from "eventemitter3";
import {
  CompressionEngine,
  PrivacyAwareCompression,
  CompressionStats,
  QuantizationType,
} from "@sovereignmap/compression";

/**
 * Configuration for privacy-aware compression
 */
export interface CompressionIntegrationConfig {
  enabled: boolean;
  quantBits: number;
  quantType: QuantizationType;
  compressionLevel: number;
  targetCompressionRatio: number;
}

/**
 * Statistics for compression in privacy context
 */
export interface CompressionIntegrationStats {
  totalUpdates: number;
  totalOriginalBytes: number;
  totalCompressedBytes: number;
  averageCompressionRatio: number;
  averageMaxError: number;
  totalCompressionTimeMs: number;
  totalDecompressionTimeMs: number;
  privacyOverhead: number;
}

/**
 * Privacy-aware compression integration layer
 *
 * Combines compression engine with privacy tracking:
 * - Post-DP-SGD compression (no epsilon cost)
 * - Automatic calibration on first gradient
 * - Event-driven integration with PrivacyEngine
 *
 * @example
 * ```typescript
 * const integration = new CompressionIntegration({
 *   enabled: true,
 *   quantBits: 8,
 *   quantType: 'ADAPTIVE',
 *   compressionLevel: 6,
 *   targetCompressionRatio: 4.0
 * });
 *
 * const compressed = integration.compressGradient(noisyGradient);
 * const decompressed = integration.decompressGradient(compressed.data, compressed.metadata);
 *
 * const stats = integration.getStatistics();
 * console.log(`Average compression: ${stats.averageCompressionRatio}×`);
 * ```
 */
export class CompressionIntegration extends EventEmitter {
  private engine: CompressionEngine;
  private privacyAware: PrivacyAwareCompression;
  private stats: CompressionIntegrationStats = {
    totalUpdates: 0,
    totalOriginalBytes: 0,
    totalCompressedBytes: 0,
    averageCompressionRatio: 0,
    averageMaxError: 0,
    totalCompressionTimeMs: 0,
    totalDecompressionTimeMs: 0,
    privacyOverhead: 0,
  };
  private config: CompressionIntegrationConfig;
  private calibrationComplete: boolean = false;

  constructor(config: CompressionIntegrationConfig) {
    super();
    this.config = config;

    // Initialize engines
    this.engine = new CompressionEngine(
      config.quantBits,
      config.compressionLevel,
      config.quantType
    );

    this.privacyAware = new PrivacyAwareCompression(
      config.quantBits,
      10.0, // Not consumed, post-DP
      config.quantType
    );

    // Event handling
    this.engine.on("calibrationComplete", (calibStats: unknown) => {
      this.calibrationComplete = true;
      this.emit("calibrationComplete", calibStats);
    });

    this.engine.on("compressionComplete", (compStats: unknown) => {
      this.emit("compressionComplete", compStats);
    });
  }

  /**
   * Compress a gradient update
   *
   * @param gradient Float32Array of gradient values
   * @param epsilonRemaining Current privacy budget (informational only, not consumed)
   * @returns Compressed data with metadata and statistics
   */
  compressGradient(
    gradient: Float32Array,
    epsilonRemaining?: number
  ): {
    data: Uint8Array;
    metadata: any;
    stats: CompressionStats;
    privacySpent: number; // Always 0 - post-DP compression
  } {
    const startTime = performance.now();

    // Compress using privacy-aware engine
    const result = this.privacyAware.compressUpdate(
      gradient,
      epsilonRemaining || 10.0
    );

    // Normalize ratio as X:1 (original/compressed) for integration consumers.
    const compressionRatio =
      result.compressed.byteLength > 0
        ? gradient.byteLength / result.compressed.byteLength
        : 0;
    const normalizedStats: CompressionStats = {
      ...result.stats,
      compressionRatio,
    };

    const endTime = performance.now();
    const compressionTimeMs = endTime - startTime;

    if (!this.calibrationComplete) {
      this.calibrationComplete = true;
      this.emit("calibrationComplete", {
        sampleSize: Math.max(100, Math.floor(gradient.length * 0.1)),
      });
    }

    // Update statistics
    this.stats.totalUpdates++;
    this.stats.totalOriginalBytes += gradient.byteLength;
    this.stats.totalCompressedBytes += result.compressed.byteLength;
    this.stats.totalCompressionTimeMs += compressionTimeMs;

    // Recalculate average metrics
    this.updateAverageMetrics(normalizedStats);

    this.emit("compressionComplete", normalizedStats);

    // Emit event
    this.emit("gradientCompressed", {
      originalSize: gradient.byteLength,
      compressedSize: result.compressed.byteLength,
      ratio: normalizedStats.compressionRatio,
      timeMs: compressionTimeMs,
    });

    return {
      data: result.compressed,
      metadata: result.metadata,
      stats: normalizedStats,
      privacySpent: 0, // Post-DP compression has no privacy cost
    };
  }

  /**
   * Decompress a gradient update
   *
   * @param compressed Compressed data buffer
   * @param metadata Compression metadata from compression operation
   * @returns Decompressed Float32Array
   */
  decompressGradient(compressed: Uint8Array, metadata: any): Float32Array {
    const startTime = performance.now();

    const decompressed = this.privacyAware.decompressUpdate(
      Buffer.from(compressed),
      metadata
    );

    const endTime = performance.now();
    const decompressionTimeMs = endTime - startTime;

    // Update statistics
    this.stats.totalDecompressionTimeMs += decompressionTimeMs;

    // Emit event
    this.emit("gradientDecompressed", {
      compressedSize: compressed.byteLength,
      decompressedSize: decompressed.byteLength,
      timeMs: decompressionTimeMs,
    });

    return decompressed as Float32Array;
  }

  /**
   * Validate compression configuration matches target
   *
   * @param gradient Sample gradient for validation
   * @returns true if compression meets target ratio and error bounds
   */
  validateConfiguration(gradient: Float32Array): boolean {
    if (!this.calibrationComplete) {
      // Need to calibrate first
      this.compressGradient(gradient);
    }

    const { stats } = this.privacyAware.compressUpdate(
      gradient,
      10.0 // Not consumed
    );

    const normalizedRatio =
      stats.compressedSize > 0 ? stats.originalSize / stats.compressedSize : 0;

    return (
      normalizedRatio >= this.config.targetCompressionRatio * 0.9 && // Allow 10% tolerance
      stats.quantizationError < 1.0 // Max error per value
    );
  }

  /**
   * Get current compression statistics
   *
   * @returns Aggregated statistics across all operations
   */
  getStatistics(): CompressionIntegrationStats {
    return { ...this.stats };
  }

  /**
   * Reset statistics
   */
  resetStatistics(): void {
    this.stats = {
      totalUpdates: 0,
      totalOriginalBytes: 0,
      totalCompressedBytes: 0,
      averageCompressionRatio: 0,
      averageMaxError: 0,
      totalCompressionTimeMs: 0,
      totalDecompressionTimeMs: 0,
      privacyOverhead: 0,
    };
  }

  /**
   * Check if calibration is complete
   */
  isCalibrated(): boolean {
    return this.calibrationComplete;
  }

  /**
   * Get compression configuration
   */
  getConfiguration(): CompressionIntegrationConfig {
    return { ...this.config };
  }

  /**
   * Update privacy overhead calculation
   *
   * @param compressionStats Statistics from compression operation
   * @private
   */
  private updateAverageMetrics(compressionStats: CompressionStats): void {
    if (this.stats.totalUpdates === 0) return;

    // Average compression ratio
    this.stats.averageCompressionRatio =
      this.stats.totalOriginalBytes / this.stats.totalCompressedBytes;

    // Average max error
    this.stats.averageMaxError =
      (this.stats.averageMaxError * (this.stats.totalUpdates - 1) +
        compressionStats.quantizationError) /
      this.stats.totalUpdates;

    // Privacy overhead: quantization error relative to DP noise
    // Assumption: DP noise std dev >> quantization error
    // Overhead = max_error / expected_noise_std (typically 0.1-1.0)
    const expectedNoiseStd = 0.1; // Typical for DP-SGD with strong privacy
    this.stats.privacyOverhead =
      compressionStats.quantizationError / expectedNoiseStd;
  }
}

/**
 * Hook for integrating compression into PrivacyEngine
 *
 * @param privacyEngine PrivacyEngine instance
 * @param config Compression configuration
 * @returns Compression integration instance
 */
export function createCompressionHook(
  privacyEngine: any, // PrivacyEngine type
  config: CompressionIntegrationConfig
): CompressionIntegration {
  const integration = new CompressionIntegration(config);

  // Subscribe to privacy engine events
  privacyEngine.on("gradientNoiseInjected", (event: any) => {
    const compressed = integration.compressGradient(
      event.noisyGradient,
      event.epsilonUsed
    );

    // Forward compressed gradient
    privacyEngine.emit("gradientCompressed", {
      gradient: compressed.data,
      metadata: compressed.metadata,
      stats: compressed.stats,
      originalGradient: event.noisyGradient,
    });
  });

  // Export integration methods
  privacyEngine.compressGradient = (gradient: Float32Array) =>
    integration.compressGradient(gradient);

  privacyEngine.decompressGradient = (data: Uint8Array, metadata: any) =>
    integration.decompressGradient(data, metadata);

  privacyEngine.getCompressionStats = () => integration.getStatistics();

  return integration;
}

/**
 * Compression result type
 */
export interface CompressionResult {
  data: Uint8Array;
  metadata: any;
  stats: CompressionStats;
  privacySpent: number;
}

export default CompressionIntegration;

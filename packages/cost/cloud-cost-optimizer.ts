/**
 * Real-time Cloud Cost Optimization Tracker
 * Measures dollars saved based on the 10x differential privacy compression algorithm.
 */

// Prices per GB based on standard cloud egress patterns
const AWS_EGRESS_COST_PER_GB = 0.09;
const GCP_EGRESS_COST_PER_GB = 0.085;

export interface CostReport {
  originalDataSizeGB: number;
  compressedDataSizeGB: number;
  bandwidthSavedGB: number;
  dollarsSavedAWS: number;
  dollarsSavedGCP: number;
}

export class CloudCostTracker {
  private totalFloat32BytesProcessed = 0;
  private totalCompressedBytesEgressed = 0;

  /**
   * Called automatically during normal FL rounds.
   */
  public recordTrafficTransfer(originalBytes: number, compressedBytes: number) {
    this.totalFloat32BytesProcessed += originalBytes;
    this.totalCompressedBytesEgressed += compressedBytes;
  }

  /**
   * Generates an executive ROI summary of the bandwidth features
   */
  public generateROISummary(): CostReport {
    const originalGB = this.totalFloat32BytesProcessed / 1024 ** 3;
    const compressedGB = this.totalCompressedBytesEgressed / 1024 ** 3;
    const savedGB = originalGB - compressedGB;

    return {
      originalDataSizeGB: parseFloat(originalGB.toFixed(4)),
      compressedDataSizeGB: parseFloat(compressedGB.toFixed(4)),
      bandwidthSavedGB: parseFloat(savedGB.toFixed(4)),
      dollarsSavedAWS: parseFloat(
        (savedGB * AWS_EGRESS_COST_PER_GB).toFixed(2)
      ),
      dollarsSavedGCP: parseFloat(
        (savedGB * GCP_EGRESS_COST_PER_GB).toFixed(2)
      ),
    };
  }

  public reportToOpsLogger(): void {
    const report = this.generateROISummary();
    console.log(
      `[CostOptimizer] Bandwidth Saved: ${report.bandwidthSavedGB} GB`
    );
    console.log(
      `[CostOptimizer] Estimated Egress Savings: $${report.dollarsSavedAWS} (AWS) / $${report.dollarsSavedGCP} (GCP)`
    );
  }
}

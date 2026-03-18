/**
 * Enhanced Prometheus Metrics Exporter for Sovereign Map
 *
 * Exports comprehensive metrics for GPU, privacy, byzantium, and network monitoring
 * Integrates with existing PrivacyEngine and consensus modules
 */

import { EventEmitter } from "eventemitter3";

/**
 * Metric types
 */
export interface MetricPoint {
  name: string;
  value: number;
  timestamp: number;
  labels: Record<string, string>;
}

/**
 * Gauge metric (can go up and down)
 */
export class GaugeMetric {
  private value: number = 0;

  set(value: number) {
    this.value = value;
  }

  inc(delta: number = 1) {
    this.value += delta;
  }

  dec(delta: number = 1) {
    this.value -= delta;
  }

  get(): number {
    return this.value;
  }
}

/**
 * Counter metric (only increases)
 */
export class CounterMetric {
  private value: number = 0;

  inc(delta: number = 1) {
    this.value += delta;
  }

  get(): number {
    return this.value;
  }
}

/**
 * Histogram metric (tracks distribution)
 */
export class HistogramMetric {
  private samples: number[] = [];
  private readonly buckets: number[];

  constructor(buckets: number[] = [1, 2, 5, 10, 20, 50, 100, 200, 500, 1000]) {
    this.buckets = buckets;
  }

  observe(value: number) {
    this.samples.push(value);
    // Keep only last 1000 samples for memory efficiency
    if (this.samples.length > 1000) {
      this.samples.shift();
    }
  }

  getPercentile(p: number): number {
    if (this.samples.length === 0) return 0;
    const sorted = [...this.samples].sort((a, b) => a - b);
    const index = Math.ceil((p / 100) * sorted.length) - 1;
    return sorted[Math.max(0, index)];
  }

  getStats() {
    if (this.samples.length === 0) {
      return { count: 0, sum: 0, mean: 0, p50: 0, p99: 0, p999: 0 };
    }

    const sorted = [...this.samples].sort((a, b) => a - b);
    const sum = this.samples.reduce((a, b) => a + b, 0);

    return {
      count: this.samples.length,
      sum,
      mean: sum / this.samples.length,
      p50: sorted[Math.floor(sorted.length * 0.5)],
      p99: sorted[Math.floor(sorted.length * 0.99)],
      p999: sorted[Math.floor(sorted.length * 0.999)],
    };
  }
}

/**
 * GPU Metrics
 */
export class GPUMetrics {
  // Per-node metrics
  gpuUtilization = new GaugeMetric(); // 0-100%
  gpuMemoryUsed = new GaugeMetric(); // bytes
  gpuMemoryTotal = new GaugeMetric(); // bytes
  gpuPower = new GaugeMetric(); // watts
  gpuTemperature = new GaugeMetric(); // celsius

  // Per-operation metrics
  noiseGenerationLatency = new HistogramMetric([0.1, 0.5, 1, 2, 5, 10, 20, 50]);
  gaussianNoiseLatency = new HistogramMetric([0.01, 0.05, 0.1, 0.5, 1, 2, 5]);
  laplaceNoiseLatency = new HistogramMetric([0.01, 0.05, 0.1, 0.5, 1, 2, 5]);

  // Throughput
  noiseSamplesPerSecond = new GaugeMetric();
  operationsPerSecond = new GaugeMetric();

  // Acceleration detection
  gpuDetectedCount = new CounterMetric();
  gpuDetectionFailures = new CounterMetric();

  // Device stats
  activeGPUDevices = new GaugeMetric();
  fallbackToCPU = new CounterMetric();
}

/**
 * Privacy Metrics
 */
export class PrivacyMetrics {
  // Budget tracking
  epsilonConsumedTotal = new CounterMetric();
  epsilonConsumedPerRound = new GaugeMetric();
  epsilonRemaining = new GaugeMetric();
  deltaRemaining = new GaugeMetric();

  // Noise injection
  noiseInjectedCount = new CounterMetric();
  noiseMagnitude = new HistogramMetric([0.01, 0.05, 0.1, 0.5, 1, 2, 5, 10]);

  // Mechanisms
  gaussianNoiseCount = new CounterMetric();
  laplaceNoiseCount = new CounterMetric();

  // Privacy overhead
  privacyOverheadPercent = new GaugeMetric();
  cpuVsGPUOverhead = new GaugeMetric(); // Negative = GPU faster

  // Per-update metrics
  updatePrivacyTime = new HistogramMetric([1, 5, 10, 20, 50, 100, 200, 500]);
}

/**
 * Consensus & Byzantine Metrics
 */
export class ConsensusMetrics {
  // Round metrics
  roundsCompleted = new CounterMetric();
  roundDuration = new HistogramMetric([100, 500, 1000, 2000, 5000, 10000]);
  participationRate = new GaugeMetric(); // 0-100%

  // Byzantine detection
  byzantineNodesDetected = new CounterMetric();
  byzantineDetectionEvents = new CounterMetric();
  byzantineNodesList = new GaugeMetric(); // Count of detected nodes

  // Consensus achievement
  consensusReached = new CounterMetric();
  consensusFailed = new CounterMetric();
  consensusTime = new HistogramMetric([100, 500, 1000, 2000, 5000]);

  // Quorum status
  quorumSize = new GaugeMetric();
  nodesOnline = new GaugeMetric();
  nodesOffline = new GaugeMetric();
}

/**
 * Network Partition Metrics
 */
export class NetworkPartitionMetrics {
  // Detection
  partitionsDetected = new CounterMetric();
  partitionDetectionLatency = new HistogramMetric([1000, 5000, 10000, 15000]);

  // Recovery
  partitionRecoveryAttempts = new CounterMetric();
  partitionRecoverySuccess = new CounterMetric();
  partitionRecoveryFailure = new CounterMetric();
  recoveryTime = new HistogramMetric([5000, 10000, 20000, 30000, 50000]);

  // Current state
  currentPartitionSize = new GaugeMetric();
  partitionIsolatedNodes = new GaugeMetric();

  // Attestation validation
  attestationValidations = new CounterMetric();
  attestationValidationFailures = new CounterMetric();
}

/**
 * System Metrics
 */
export class SystemMetrics {
  // Memory
  processMpemory = new GaugeMetric(); // bytes
  systemMemoryUsage = new GaugeMetric(); // bytes

  // CPU
  cpuUsage = new GaugeMetric(); // 0-100%
  cpuCores = new GaugeMetric(); // count

  // Network
  networkBytesIn = new CounterMetric();
  networkBytesOut = new CounterMetric();
  networkLatency = new HistogramMetric([1, 5, 10, 20, 50, 100, 200]);

  // Uptime
  uptimeSeconds = new CounterMetric();
  restarts = new CounterMetric();
}

/**
 * Complete Metrics Registry
 */
export class MetricsRegistry {
  gpu = new GPUMetrics();
  privacy = new PrivacyMetrics();
  consensus = new ConsensusMetrics();
  partition = new NetworkPartitionMetrics();
  system = new SystemMetrics();

  /**
   * Export all metrics in Prometheus text format
   */
  exportPrometheus(): string {
    let output = "";

    // GPU metrics
    output += this.exportGaugeMetric(
      "gpu_utilization_percent",
      this.gpu.gpuUtilization.get(),
      { device: "primary" }
    );
    output += this.exportGaugeMetric(
      "gpu_memory_used_bytes",
      this.gpu.gpuMemoryUsed.get(),
      { device: "primary" }
    );
    output += this.exportGaugeMetric(
      "gpu_memory_total_bytes",
      this.gpu.gpuMemoryTotal.get(),
      { device: "primary" }
    );
    output += this.exportHistogramMetric(
      "gpu_noise_latency_seconds",
      this.gpu.noiseGenerationLatency
    );
    output += this.exportGaugeMetric(
      "gpu_active_devices",
      this.gpu.activeGPUDevices.get()
    );

    // Privacy metrics
    output += this.exportGaugeMetric(
      "privacy_epsilon_remaining",
      this.privacy.epsilonRemaining.get()
    );
    output += this.exportGaugeMetric(
      "privacy_overhead_percent",
      this.privacy.privacyOverheadPercent.get()
    );
    output += this.exportConstMetric(
      "privacy_noise_injected_total",
      this.privacy.noiseInjectedCount.get()
    );
    output += this.exportHistogramMetric(
      "privacy_noise_magnitude",
      this.privacy.noiseMagnitude
    );

    // Consensus metrics
    output += this.exportGaugeMetric(
      "consensus_participation_rate",
      this.consensus.participationRate.get()
    );
    output += this.exportGaugeMetric(
      "consensus_nodes_online",
      this.consensus.nodesOnline.get()
    );
    output += this.exportHistogramMetric(
      "consensus_round_duration_seconds",
      this.consensus.roundDuration
    );
    output += this.exportConstMetric(
      "consensus_rounds_completed",
      this.consensus.roundsCompleted.get()
    );

    // Byzantine metrics
    output += this.exportConstMetric(
      "byzantine_nodes_detected_total",
      this.consensus.byzantineNodesDetected.get()
    );
    output += this.exportGaugeMetric(
      "byzantine_nodes_currently_detected",
      this.consensus.byzantineNodesList.get()
    );

    // Network metrics
    output += this.exportConstMetric(
      "network_partitions_detected_total",
      this.partition.partitionsDetected.get()
    );
    output += this.exportHistogramMetric(
      "network_partition_recovery_seconds",
      this.partition.recoveryTime
    );
    output += this.exportConstMetric(
      "network_partition_recovery_success",
      this.partition.partitionRecoverySuccess.get()
    );

    // System metrics
    output += this.exportGaugeMetric(
      "system_memory_bytes",
      this.system.processMpemory.get()
    );
    output += this.exportGaugeMetric(
      "system_cpu_usage_percent",
      this.system.cpuUsage.get()
    );
    output += this.exportHistogramMetric(
      "system_network_latency_seconds",
      this.system.networkLatency
    );

    return output;
  }

  private exportGaugeMetric(
    name: string,
    value: number,
    labels?: Record<string, string>
  ): string {
    const labelStr = labels
      ? ` {${Object.entries(labels)
          .map(([k, v]) => `${k}="${v}"`)
          .join(",")}}`
      : "";
    return `# TYPE ${name} gauge\n${name}${labelStr} ${value}\n\n`;
  }

  private exportConstMetric(name: string, value: number): string {
    return `# TYPE ${name} counter\n${name} ${value}\n\n`;
  }

  private exportHistogramMetric(name: string, metric: HistogramMetric): string {
    const stats = metric.getStats();
    let output = `# TYPE ${name} histogram\n`;
    output += `${name}_count ${stats.count}\n`;
    output += `${name}_sum ${stats.sum}\n`;
    output += `${name}_bucket{le="0.5"} ${
      stats.count * (stats.p50 <= 0.5 ? 1 : 0)
    }\n`;
    output += `${name}_bucket{le="1"} ${
      stats.count * (stats.p50 <= 1 ? 1 : 0)
    }\n`;
    output += `${name}_bucket{le="5"} ${
      stats.count * (stats.p50 <= 5 ? 1 : 0)
    }\n`;
    output += `${name}_bucket{le="+Inf"} ${stats.count}\n\n`;
    return output;
  }

  /**
   * Export SLA compliance status
   */
  exportSLAStatus(): Record<string, boolean> {
    return {
      privacy_overhead_sla: this.privacy.privacyOverheadPercent.get() < 12,
      gpu_detection_sla: this.gpu.gpuDetectionFailures.get() === 0,
      consensus_participation_sla:
        this.consensus.participationRate.get() > 66.67,
      network_latency_sla: this.system.networkLatency.getPercentile(99) < 100,
      byzantine_tolerance_sla:
        this.consensus.byzantineNodesDetected.get() < 334, // <33%
    };
  }
}

export default MetricsRegistry;

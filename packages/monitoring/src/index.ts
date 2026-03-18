/**
 * Monitoring Integration Module
 *
 * Integrates metrics collection across Privacy, Consensus, Network modules
 * Emits metrics to Prometheus server for real-time monitoring
 */

import { EventEmitter } from "eventemitter3";
import { MetricsRegistry } from "./metrics-registry";
import { PrometheusServer } from "./prometheus-server";

/**
 * Integration hooks for PrivacyEngine metrics
 */
export class PrivacyMetricsCollector {
  private metrics: MetricsRegistry;

  constructor(metrics: MetricsRegistry) {
    this.metrics = metrics;
  }

  /**
   * Hook into PrivacyEngine events
   * @param energyEngine PrivacyEngine instance
   */
  integrateWithEngine(privacyEngine: any) {
    // Track noise injection
    privacyEngine.on("noiseGenerated", (data: any) => {
      const { epsilon, delta, mechanism, magnitude, overhead, latency } = data;

      // Privacy budget tracking
      this.metrics.recordGauge("privacy_epsilon_remaining", epsilon);
      this.metrics.recordGauge("privacy_delta_remaining", delta);

      // Noise statistics
      this.metrics.recordCounter("privacy_noise_injected_total", 1, {
        mechanism, // gaussian or laplace
      });
      this.metrics.recordHistogram(
        "privacy_noise_magnitude_distribution",
        magnitude
      );

      // Overhead tracking
      this.metrics.recordGauge("privacy_overhead_percent", overhead);
      this.metrics.recordHistogram("privacy_update_latency_seconds", latency);

      // Mechanism-specific counters
      if (mechanism === "gaussian") {
        this.metrics.recordCounter("privacy_gaussian_noise_count_total", 1);
      } else if (mechanism === "laplace") {
        this.metrics.recordCounter("privacy_laplace_noise_count_total", 1);
      }
    });

    // Track privacy failures
    privacyEngine.on("noiseInjectionFailed", (error: any) => {
      this.metrics.recordCounter("privacy_noise_injection_failures", 1, {
        errorType: error.type || "unknown",
      });
    });

    // Track acceleration status
    privacyEngine.on("accelerationDetected", (device: string) => {
      this.metrics.recordCounter("gpu_detection_attempts", 1);
      this.metrics.recordGauge("gpu_active_devices", 1, { device });
    });

    privacyEngine.on("accelerationFallback", (device: string) => {
      this.metrics.recordCounter("gpu_detection_failures", 1);
      this.metrics.recordCounter("gpu_cpu_fallback_total", 1, { device });
    });
  }

  /**
   * Record GPU stats from PrivacyEngine
   * @param stats GPU acceleration statistics
   */
  recordGPUStats(stats: any) {
    const {
      utilization,
      memoryUsed,
      memoryTotal,
      temperature,
      latency,
      throughput,
    } = stats;

    this.metrics.recordGauge("gpu_utilization_percent", utilization);
    this.metrics.recordGauge("gpu_memory_used_bytes", memoryUsed);
    this.metrics.recordGauge("gpu_memory_total_bytes", memoryTotal);
    this.metrics.recordGauge("gpu_temperature_celsius", temperature);
    this.metrics.recordHistogram("gpu_noise_latency_seconds", latency);
    this.metrics.recordGauge("gpu_samples_per_second", throughput);
  }
}

/**
 * Integration hooks for ConsensusModule metrics
 */
export class ConsensusMetricsCollector {
  private metrics: MetricsRegistry;

  constructor(metrics: MetricsRegistry) {
    this.metrics = metrics;
  }

  /**
   * Hook into ConsensusModule events
   * @param consensusModule ConsensusModule instance
   */
  integrateWithConsensus(consensusModule: any) {
    // Track round completion
    consensusModule.on("roundCompleted", (data: any) => {
      const { roundNumber, duration, participants, participation } = data;

      this.metrics.recordCounter("consensus_rounds_completed", 1);
      this.metrics.recordHistogram(
        "consensus_round_duration_seconds",
        duration
      );
      this.metrics.recordGauge("consensus_participation_rate", participation);
      this.metrics.recordGauge("consensus_nodes_online", participants.online);
      this.metrics.recordGauge("consensus_nodes_offline", participants.offline);
    });

    // Track round failures
    consensusModule.on("roundFailed", (error: any) => {
      this.metrics.recordCounter("consensus_failed", 1, {
        reason: error.reason || "unknown",
      });
    });

    // Track Byzantine detection
    consensusModule.on("byzantineNodeDetected", (data: any) => {
      const { nodeId, confidence } = data;

      this.metrics.recordCounter("byzantine_nodes_detected_total", 1);
      this.metrics.recordGauge("byzantine_nodes_currently_detected", 1, {
        nodeId,
      });
      this.metrics.recordHistogram(
        "byzantine_detection_confidence",
        confidence
      );
    });

    // Track Byzantine node recovery (removal from Byzantine list)
    consensusModule.on("byzantineNodeRecovered", (nodeId: string) => {
      this.metrics.recordGauge("byzantine_nodes_currently_detected", 0, {
        nodeId,
      });
    });

    // Track consensus timeouts
    consensusModule.on("roundTimeout", () => {
      this.metrics.recordCounter("consensus_round_timeout_total", 1);
    });

    // Track quorum changes
    consensusModule.on("quorumChanged", (data: any) => {
      const { quorumSize, totalNodes } = data;
      this.metrics.recordGauge("consensus_quorum_size", quorumSize);
    });
  }

  /**
   * Record consensus state snapshot
   * @param state Current consensus state
   */
  recordConsensusState(state: any) {
    const {
      roundsCompleted,
      participation,
      nodesOnline,
      nodesOffline,
      byzantineCount,
    } = state;

    this.metrics.recordGauge("consensus_rounds_completed", roundsCompleted);
    this.metrics.recordGauge("consensus_participation_rate", participation);
    this.metrics.recordGauge("consensus_nodes_online", nodesOnline);
    this.metrics.recordGauge("consensus_nodes_offline", nodesOffline);
    this.metrics.recordGauge(
      "byzantine_nodes_currently_detected",
      byzantineCount
    );
  }
}

/**
 * Integration hooks for PartitionRecoveryManager metrics
 */
export class NetworkMetricsCollector {
  private metrics: MetricsRegistry;

  constructor(metrics: MetricsRegistry) {
    this.metrics = metrics;
  }

  /**
   * Hook into PartitionRecoveryManager events
   * @param recoveryManager PartitionRecoveryManager instance
   */
  integrateWithPartitionRecovery(recoveryManager: any) {
    // Track partition detection
    recoveryManager.on("partitionDetected", (data: any) => {
      const { isolatedNodes, detectionMethod } = data;

      this.metrics.recordCounter("network_partitions_detected_total", 1);
      this.metrics.recordGauge(
        "network_partition_isolated_nodes",
        isolatedNodes.length,
        {
          detectionMethod,
        }
      );
    });

    // Track recovery attempts
    recoveryManager.on("recoveryStarted", (data: any) => {
      const { timestamp } = data;

      this.metrics.recordCounter("network_partition_recovery_attempts", 1);
      this.metrics.recordGauge(
        "network_partition_recovery_start_time",
        timestamp
      );
    });

    // Track successful recovery
    recoveryManager.on("recoverySucceeded", (data: any) => {
      const { duration, nodesRecovered } = data;

      this.metrics.recordCounter("network_partition_recovery_success", 1);
      this.metrics.recordHistogram(
        "network_partition_recovery_seconds",
        duration
      );
      this.metrics.recordGauge("network_partition_isolated_nodes", 0);
    });

    // Track recovery failures
    recoveryManager.on("recoveryFailed", (error: any) => {
      this.metrics.recordCounter("network_partition_recovery_failure", 1, {
        reason: error.reason || "unknown",
      });
    });

    // Track attestation validation
    recoveryManager.on("attestationValidated", () => {
      this.metrics.recordCounter("attestation_validations_total", 1);
    });

    recoveryManager.on("attestationValidationFailed", (error: any) => {
      this.metrics.recordCounter("attestation_validation_failures", 1, {
        reason: error.reason || "unknown",
      });
    });

    // Track state synchronization
    recoveryManager.on("synchronizationCompleted", (data: any) => {
      const { duration, bytesTransferred } = data;

      this.metrics.recordHistogram("network_sync_duration_seconds", duration);
      this.metrics.recordGauge(
        "network_sync_bytes_transferred",
        bytesTransferred
      );
    });
  }

  /**
   * Record network health snapshot
   * @param health Current network health
   */
  recordNetworkHealth(health: any) {
    const {
      partitionsDetected,
      isolatedNodes,
      recoveryTime,
      attestationPassRate,
    } = health;

    this.metrics.recordGauge(
      "network_partitions_detected_total",
      partitionsDetected
    );
    this.metrics.recordGauge("network_partition_isolated_nodes", isolatedNodes);
    this.metrics.recordGauge(
      "network_partition_recovery_seconds",
      recoveryTime
    );
    this.metrics.recordGauge(
      "attestation_validation_success_rate",
      attestationPassRate
    );
  }
}

/**
 * System metrics collector
 */
export class SystemMetricsCollector {
  private metrics: MetricsRegistry;
  private intervalId: NodeJS.Timeout | null = null;

  constructor(metrics: MetricsRegistry) {
    this.metrics = metrics;
  }

  /**
   * Start collecting system metrics at regular interval
   * @param intervalMs Collection interval in milliseconds (default: 10s)
   */
  startCollecting(intervalMs: number = 10000) {
    if (this.intervalId) return;

    this.intervalId = setInterval(() => {
      this.collectSystemMetrics();
    }, intervalMs);
  }

  /**
   * Stop collecting system metrics
   */
  stopCollecting() {
    if (this.intervalId) {
      clearInterval(this.intervalId);
      this.intervalId = null;
    }
  }

  /**
   * Collect current system metrics
   */
  private collectSystemMetrics() {
    // Memory metrics
    const memUsage = process.memoryUsage();
    this.metrics.recordGauge("system_memory_bytes", memUsage.heapUsed);
    this.metrics.recordGauge("system_memory_total_bytes", memUsage.heapTotal);

    // CPU metrics (mock, requires os module for accuracy)
    const cpuUsage = process.cpuUsage();
    const cpuPercent = (cpuUsage.user + cpuUsage.system) / 1000000; // Convert to percentage
    this.metrics.recordGauge(
      "system_cpu_usage_percent",
      Math.min(cpuPercent, 100)
    );

    // Uptime
    this.metrics.recordGauge("system_uptime_seconds", process.uptime());
  }

  /**
   * Record network latency measurement
   * @param latencyMs Latency in milliseconds
   */
  recordNetworkLatency(latencyMs: number) {
    this.metrics.recordHistogram(
      "system_network_latency_seconds",
      latencyMs / 1000
    );
  }

  /**
   * Record process restart
   */
  recordRestart() {
    this.metrics.recordCounter("system_restarts", 1);
  }
}

/**
 * Unified monitoring orchestrator
 */
export class MonitoringOrchestrator {
  private metrics: MetricsRegistry;
  private server: PrometheusServer;
  private privacyCollector: PrivacyMetricsCollector;
  private consensusCollector: ConsensusMetricsCollector;
  private networkCollector: NetworkMetricsCollector;
  private systemCollector: SystemMetricsCollector;

  constructor(port: number = 9090) {
    this.metrics = new MetricsRegistry();
    this.server = new PrometheusServer(this.metrics, port);
    this.privacyCollector = new PrivacyMetricsCollector(this.metrics);
    this.consensusCollector = new ConsensusMetricsCollector(this.metrics);
    this.networkCollector = new NetworkMetricsCollector(this.metrics);
    this.systemCollector = new SystemMetricsCollector(this.metrics);
  }

  /**
   * Initialize all monitoring components
   */
  async initialize() {
    await this.server.start();
    this.systemCollector.startCollecting();
  }

  /**
   * Shutdown all monitoring components
   */
  async shutdown() {
    this.systemCollector.stopCollecting();
    await this.server.stop();
  }

  /**
   * Integrate with PrivacyEngine
   */
  integratePrivacy(privacyEngine: any) {
    this.privacyCollector.integrateWithEngine(privacyEngine);
  }

  /**
   * Integrate with ConsensusModule
   */
  integrateConsensus(consensusModule: any) {
    this.consensusCollector.integrateWithConsensus(consensusModule);
  }

  /**
   * Integrate with PartitionRecoveryManager
   */
  integrateNetwork(recoveryManager: any) {
    this.networkCollector.integrateWithPartitionRecovery(recoveryManager);
  }

  /**
   * Get metrics registry for direct access
   */
  getMetrics(): MetricsRegistry {
    return this.metrics;
  }

  /**
   * Get Prometheus server for endpoint configuration
   */
  getServer(): PrometheusServer {
    return this.server;
  }
}

/**
 * Example usage:
 *
 * const monitoring = new MonitoringOrchestrator(9090);
 * await monitoring.initialize();
 *
 * monitoring.integratePrivacy(privacyEngine);
 * monitoring.integrateConsensus(consensusModule);
 * monitoring.integrateNetwork(partitionRecoveryManager);
 *
 * // Metrics now flow to:
 * // - http://localhost:9090/metrics (Prometheus scraping)
 * // - http://localhost:9090/health (Kubernetes probe)
 * // - http://localhost:9090/sla (SLA compliance)
 */

export {
  MetricsRegistry,
  PrometheusServer,
  PrivacyMetricsCollector,
  ConsensusMetricsCollector,
  NetworkMetricsCollector,
  SystemMetricsCollector,
};

export default MonitoringOrchestrator;

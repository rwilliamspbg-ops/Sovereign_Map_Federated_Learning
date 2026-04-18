/**
 * Integration Tests for Monitoring System
 *
 * Tests integration of MonitoringOrchestrator with PrivacyEngine,
 * ConsensusModule, and PartitionRecoveryManager
 */

import { describe, it, expect, beforeEach, afterEach, vi } from "vitest";
import { EventEmitter } from "eventemitter3";
import {
  MonitoringOrchestrator,
  PrivacyMetricsCollector,
  ConsensusMetricsCollector,
  NetworkMetricsCollector,
  SystemMetricsCollector,
  MetricsRegistry,
  PrometheusServer,
} from "../src/index";

// Mock module implementations
class MockPrivacyEngine extends EventEmitter {
  generateNoise() {
    this.emit("noiseGenerated", {
      epsilon: 0.95,
      delta: 1e-5,
      mechanism: "gaussian",
      magnitude: 0.42,
      overhead: 8.5,
      latency: 0.025,
    });
  }

  detectAcceleration() {
    this.emit("accelerationDetected", "cuda");
  }

  triggerFallback() {
    this.emit("accelerationFallback", "cuda");
  }

  triggerNoiseFailure() {
    this.emit("noiseInjectionFailed", { type: "memoryExhaustion" });
  }
}

class MockConsensusModule extends EventEmitter {
  completeRound() {
    this.emit("roundCompleted", {
      roundNumber: 42,
      duration: 5.2,
      participants: { online: 95, offline: 5 },
      participation: 95.0,
    });
  }

  detectByzantine() {
    this.emit("byzantineNodeDetected", {
      nodeId: "node-42",
      confidence: 0.98,
    });
  }

  triggerTimeout() {
    this.emit("roundTimeout");
  }

  triggerRoundFailure() {
    this.emit("roundFailed", { reason: "quorumLoss" });
  }
}

class MockPartitionRecoveryManager extends EventEmitter {
  detectPartition() {
    this.emit("partitionDetected", {
      isolatedNodes: ["node-1", "node-2", "node-3"],
      detectionMethod: "heartbeat",
    });
  }

  startRecovery() {
    this.emit("recoveryStarted", {
      timestamp: Date.now(),
    });
  }

  succeedRecovery() {
    this.emit("recoverySucceeded", {
      duration: 35.5,
      nodesRecovered: 3,
    });
  }

  failRecovery() {
    this.emit("recoveryFailed", {
      reason: "byzantineDispute",
    });
  }

  validateAttestation() {
    this.emit("attestationValidated");
  }

  failAttestation() {
    this.emit("attestationValidationFailed", {
      reason: "invalidSignature",
    });
  }
}

describe("MonitoringOrchestrator Integration Tests", () => {
  let orchestrator: MonitoringOrchestrator;
  let privacyEngine: MockPrivacyEngine;
  let consensusModule: MockConsensusModule;
  let partitionRecovery: MockPartitionRecoveryManager;

  beforeEach(async () => {
    orchestrator = new MonitoringOrchestrator(19090); // Use different port for tests
    privacyEngine = new MockPrivacyEngine();
    consensusModule = new MockConsensusModule();
    partitionRecovery = new MockPartitionRecoveryManager();

    await orchestrator.initialize();
  });

  afterEach(async () => {
    await orchestrator.shutdown();
  });

  describe("PrivacyMetricsCollector Integration", () => {
    it("should collect noise generation metrics", (done) => {
      orchestrator.integratePrivacy(privacyEngine);
      const metrics = orchestrator.getMetrics();

      privacyEngine.generateNoise();

      // Give event loop time to process
      setTimeout(() => {
        expect(metrics.exportPrometheus()).toContain(
          "privacy_epsilon_remaining"
        );
        expect(metrics.exportPrometheus()).toContain(
          "privacy_overhead_percent"
        );
        expect(metrics.exportPrometheus()).toContain(
          "privacy_noise_injected_total"
        );
        done();
      }, 50);
    });

    it("should detect GPU acceleration", (done) => {
      orchestrator.integratePrivacy(privacyEngine);
      const metrics = orchestrator.getMetrics();

      privacyEngine.detectAcceleration();

      setTimeout(() => {
        expect(metrics.exportPrometheus()).toContain("gpu_detection_attempts");
        expect(metrics.exportPrometheus()).toContain("gpu_active_devices");
        done();
      }, 50);
    });

    it("should track GPU acceleration fallback", (done) => {
      orchestrator.integratePrivacy(privacyEngine);
      const metrics = orchestrator.getMetrics();

      privacyEngine.triggerFallback();

      setTimeout(() => {
        expect(metrics.exportPrometheus()).toContain("gpu_detection_failures");
        expect(metrics.exportPrometheus()).toContain("gpu_cpu_fallback_total");
        done();
      }, 50);
    });

    it("should track noise injection failures", (done) => {
      orchestrator.integratePrivacy(privacyEngine);
      const metrics = orchestrator.getMetrics();

      privacyEngine.triggerNoiseFailure();

      setTimeout(() => {
        expect(metrics.exportPrometheus()).toContain(
          "privacy_noise_injection_failures"
        );
        done();
      }, 50);
    });

    it("should measure Gaussian vs Laplace mechanisms", (done) => {
      orchestrator.integratePrivacy(privacyEngine);
      const metrics = orchestrator.getMetrics();

      privacyEngine.generateNoise();

      setTimeout(() => {
        expect(metrics.exportPrometheus()).toContain(
          "privacy_gaussian_noise_count_total"
        );
        done();
      }, 50);
    });
  });

  describe("ConsensusMetricsCollector Integration", () => {
    it("should collect round completion metrics", (done) => {
      orchestrator.integrateConsensus(consensusModule);
      const metrics = orchestrator.getMetrics();

      consensusModule.completeRound();

      setTimeout(() => {
        expect(metrics.exportPrometheus()).toContain(
          "consensus_rounds_completed"
        );
        expect(metrics.exportPrometheus()).toContain(
          "consensus_round_duration_seconds"
        );
        expect(metrics.exportPrometheus()).toContain(
          "consensus_participation_rate"
        );
        done();
      }, 50);
    });

    it("should detect Byzantine nodes", (done) => {
      orchestrator.integrateConsensus(consensusModule);
      const metrics = orchestrator.getMetrics();

      consensusModule.detectByzantine();

      setTimeout(() => {
        expect(metrics.exportPrometheus()).toContain(
          "byzantine_nodes_detected_total"
        );
        expect(metrics.exportPrometheus()).toContain(
          "byzantine_nodes_currently_detected"
        );
        done();
      }, 50);
    });

    it("should track consensus timeouts", (done) => {
      orchestrator.integrateConsensus(consensusModule);
      const metrics = orchestrator.getMetrics();

      consensusModule.triggerTimeout();

      setTimeout(() => {
        expect(metrics.exportPrometheus()).toContain(
          "consensus_round_timeout_total"
        );
        done();
      }, 50);
    });

    it("should track round failures", (done) => {
      orchestrator.integrateConsensus(consensusModule);
      const metrics = orchestrator.getMetrics();

      consensusModule.triggerRoundFailure();

      setTimeout(() => {
        expect(metrics.exportPrometheus()).toContain("consensus_failed");
        done();
      }, 50);
    });

    it("should track online/offline nodes", (done) => {
      orchestrator.integrateConsensus(consensusModule);
      const metrics = orchestrator.getMetrics();

      consensusModule.completeRound();

      setTimeout(() => {
        expect(metrics.exportPrometheus()).toContain("consensus_nodes_online");
        expect(metrics.exportPrometheus()).toContain("consensus_nodes_offline");
        done();
      }, 50);
    });
  });

  describe("NetworkMetricsCollector Integration", () => {
    it("should detect network partitions", (done) => {
      orchestrator.integrateNetwork(partitionRecovery);
      const metrics = orchestrator.getMetrics();

      partitionRecovery.detectPartition();

      setTimeout(() => {
        expect(metrics.exportPrometheus()).toContain(
          "network_partitions_detected_total"
        );
        expect(metrics.exportPrometheus()).toContain(
          "network_partition_isolated_nodes"
        );
        done();
      }, 50);
    });

    it("should track recovery attempts", (done) => {
      orchestrator.integrateNetwork(partitionRecovery);
      const metrics = orchestrator.getMetrics();

      partitionRecovery.startRecovery();

      setTimeout(() => {
        expect(metrics.exportPrometheus()).toContain(
          "network_partition_recovery_attempts"
        );
        done();
      }, 50);
    });

    it("should track successful recovery", (done) => {
      orchestrator.integrateNetwork(partitionRecovery);
      const metrics = orchestrator.getMetrics();

      partitionRecovery.succeedRecovery();

      setTimeout(() => {
        expect(metrics.exportPrometheus()).toContain(
          "network_partition_recovery_success"
        );
        expect(metrics.exportPrometheus()).toContain(
          "network_partition_recovery_seconds"
        );
        done();
      }, 50);
    });

    it("should track recovery failures", (done) => {
      orchestrator.integrateNetwork(partitionRecovery);
      const metrics = orchestrator.getMetrics();

      partitionRecovery.failRecovery();

      setTimeout(() => {
        expect(metrics.exportPrometheus()).toContain(
          "network_partition_recovery_failure"
        );
        done();
      }, 50);
    });

    it("should track attestation validation", (done) => {
      orchestrator.integrateNetwork(partitionRecovery);
      const metrics = orchestrator.getMetrics();

      partitionRecovery.validateAttestation();

      setTimeout(() => {
        expect(metrics.exportPrometheus()).toContain(
          "attestation_validations_total"
        );
        done();
      }, 50);
    });

    it("should track attestation validation failures", (done) => {
      orchestrator.integrateNetwork(partitionRecovery);
      const metrics = orchestrator.getMetrics();

      partitionRecovery.failAttestation();

      setTimeout(() => {
        expect(metrics.exportPrometheus()).toContain(
          "attestation_validation_failures"
        );
        done();
      }, 50);
    });
  });

  describe("SystemMetricsCollector Integration", () => {
    it("should collect system metrics periodically", (done) => {
      const metrics = orchestrator.getMetrics();

      setTimeout(() => {
        const prometheusOutput = metrics.exportPrometheus();
        expect(prometheusOutput).toContain("system_memory_bytes");
        expect(prometheusOutput).toContain("system_cpu_usage_percent");
        expect(prometheusOutput).toContain("system_uptime_seconds");
        done();
      }, 500); // Wait for first collection cycle
    });

    it("should record network latency measurements", (done) => {
      const metrics = orchestrator.getMetrics();
      const systemCollector = new SystemMetricsCollector(metrics);

      systemCollector.recordNetworkLatency(45.5);

      setTimeout(() => {
        expect(metrics.exportPrometheus()).toContain(
          "system_network_latency_seconds"
        );
        done();
      }, 50);
    });

    it("should track process restarts", (done) => {
      const metrics = orchestrator.getMetrics();
      const systemCollector = new SystemMetricsCollector(metrics);

      systemCollector.recordRestart();

      setTimeout(() => {
        expect(metrics.exportPrometheus()).toContain("system_restarts");
        done();
      }, 50);
    });
  });

  describe("End-to-End Integration", () => {
    it("should handle simultaneous events from multiple modules", (done) => {
      orchestrator.integratePrivacy(privacyEngine);
      orchestrator.integrateConsensus(consensusModule);
      orchestrator.integrateNetwork(partitionRecovery);

      const metrics = orchestrator.getMetrics();

      // Emit events from all modules simultaneously
      privacyEngine.generateNoise();
      consensusModule.completeRound();
      partitionRecovery.detectPartition();

      setTimeout(() => {
        const prometheusOutput = metrics.exportPrometheus();

        // Verify all subsystems reporting metrics
        expect(prometheusOutput).toContain("privacy_overhead_percent");
        expect(prometheusOutput).toContain("consensus_participation_rate");
        expect(prometheusOutput).toContain("network_partitions_detected_total");
        expect(prometheusOutput).toContain("system_uptime_seconds");

        done();
      }, 100);
    });

    it("should support SLA compliance checks", (done) => {
      orchestrator.integratePrivacy(privacyEngine);
      orchestrator.integrateConsensus(consensusModule);

      // Emit healthy metrics
      privacyEngine.generateNoise();
      consensusModule.completeRound();

      setTimeout(() => {
        const metrics = orchestrator.getMetrics();
        const slaStatus = metrics.exportSLAStatus();

        expect(slaStatus).toHaveProperty("overallStatus");
        expect(slaStatus).toHaveProperty("checks");
        expect(slaStatus.checks.length).toBeGreaterThan(0);

        done();
      }, 100);
    });

    it("should export metrics in Prometheus text format", (done) => {
      orchestrator.integratePrivacy(privacyEngine);
      privacyEngine.generateNoise();

      setTimeout(() => {
        const metrics = orchestrator.getMetrics();
        const prometheusOutput = metrics.exportPrometheus();

        // Verify Prometheus text format
        expect(prometheusOutput).toMatch(/^# HELP/m); // Has help section
        expect(prometheusOutput).toMatch(/^# TYPE/m); // Has type section
        expect(prometheusOutput).toMatch(/\{.*\}\s+[\d.]+/); // Has metric values

        done();
      }, 50);
    });
  });

  describe("Error Handling", () => {
    it("should handle collector gracefully during shutdown", async () => {
      orchestrator.integratePrivacy(privacyEngine);
      orchestrator.integrateConsensus(consensusModule);

      // Should shutdown without errors
      await expect(orchestrator.shutdown()).resolves.toBeUndefined();
    });

    it("should support multiple integrations on same orchestrator", async () => {
      orchestrator.integratePrivacy(privacyEngine);
      orchestrator.integrateConsensus(consensusModule);
      orchestrator.integrateNetwork(partitionRecovery);

      // All three should be integrated without conflicts
      const metrics = orchestrator.getMetrics();
      expect(metrics).toBeDefined();
    });

    it("should handle rapid event sequences", (done) => {
      orchestrator.integratePrivacy(privacyEngine);
      const metrics = orchestrator.getMetrics();

      // Hammer the system with rapid events
      for (let i = 0; i < 100; i++) {
        privacyEngine.generateNoise();
      }

      setTimeout(() => {
        // Should not crash and should report metrics
        expect(metrics.exportPrometheus()).toContain(
          "privacy_noise_injected_total"
        );
        done();
      }, 200);
    });
  });

  describe("Performance Characteristics", () => {
    it("should have minimal metric collection overhead", (done) => {
      const startTime = Date.now();

      orchestrator.integratePrivacy(privacyEngine);
      orchestrator.integrateConsensus(consensusModule);
      orchestrator.integrateNetwork(partitionRecovery);

      // Emit 1000 events
      for (let i = 0; i < 1000; i++) {
        privacyEngine.generateNoise();
        consensusModule.completeRound();
        partitionRecovery.validateAttestation();
      }

      const elapsed = Date.now() - startTime;

      // 1000 events across 3 modules should complete in < 500ms
        return expect(elapsed).toBeLessThan(500);
    });

    it("should maintain metric accuracy under load", (done) => {
      orchestrator.integratePrivacy(privacyEngine);
      const metrics = orchestrator.getMetrics();

      // Emit 50 known events
      for (let i = 0; i < 50; i++) {
        privacyEngine.generateNoise();
      }

      setTimeout(() => {
        const prometheusOutput = metrics.exportPrometheus();

        // Should have incremented counter by 50
        const noiseMatch = prometheusOutput.match(
          /privacy_noise_injected_total\s+(\d+)/
        );
        expect(noiseMatch).toBeTruthy();
        expect(parseInt(noiseMatch![1])).toBe(50);

        done();
      }, 100);
    });
  });
});

/**
 * Manual testing guide:
 *
 * # 1. Start monitoring orchestrator
 * const orch = new MonitoringOrchestrator(9090);
 * await orch.initialize();
 *
 * # 2. Verify endpoints
 * curl http://localhost:9090/metrics   # Should show metric lines
 * curl http://localhost:9090/health    # Should return {"status": "healthy"}
 * curl http://localhost:9090/sla       # Should show SLA status
 *
 * # 3. Integrate with modules
 * orch.integratePrivacy(privacyEngine);
 * orch.integrateConsensus(consensusModule);
 * orch.integrateNetwork(partitionManager);
 *
 * # 4. Verify metrics flowing
 * setInterval(() => {
 *   console.log(orch.getMetrics().exportPrometheus().split('\n').filter(l => !l.startsWith('#')).length);
 * }, 1000);
 *
 * # 5. Configure Prometheus scraping
 * # In prometheus.yml:
 * # scrape_configs:
 * #   - job_name: 'sovereign-map'
 * #     static_configs:
 * #       - targets: ['localhost:9090']
 */

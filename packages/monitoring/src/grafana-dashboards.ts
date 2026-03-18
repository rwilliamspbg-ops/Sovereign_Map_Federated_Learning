/**
 * Grafana Dashboard Definitions
 *
 * Exportable JSON dashboards for common monitoring scenarios
 */

export const GPU_DASHBOARD = {
  dashboard: {
    title: "GPU Acceleration Monitor",
    description: "Real-time GPU performance and utilization monitoring",
    tags: ["gpu", "performance", "sovereign-map"],
    timezone: "UTC",
    panels: [
      {
        id: 1,
        title: "GPU Utilization (%)",
        type: "graph",
        targets: [
          {
            expr: 'gpu_utilization_percent{device="primary"}',
          },
        ],
        yaxis: { format: "percent", max: 100, min: 0 },
      },
      {
        id: 2,
        title: "GPU Memory Usage (GB)",
        type: "graph",
        targets: [
          {
            expr: "(gpu_memory_used_bytes / 1024 / 1024 / 1024)",
          },
        ],
      },
      {
        id: 3,
        title: "Noise Generation Latency (p99, ms)",
        type: "graph",
        targets: [
          {
            expr: "gpu_noise_latency_seconds * 1000",
          },
        ],
      },
      {
        id: 4,
        title: "Active GPU Devices",
        type: "stat",
        targets: [
          {
            expr: "gpu_active_devices",
          },
        ],
      },
      {
        id: 5,
        title: "GPU Failure Rate (%)",
        type: "stat",
        targets: [
          {
            expr: "(gpu detected failures / gpu detection attempts) * 100",
          },
        ],
        thresholds: {
          mode: "absolute",
          steps: [
            { color: "green", value: null },
            { color: "red", value: 1 },
          ],
        },
      },
    ],
  },
};

export const PRIVACY_DASHBOARD = {
  dashboard: {
    title: "Privacy Budget Monitor",
    description:
      "Epsilon/delta consumption, noise injection, overhead tracking",
    tags: ["privacy", "sgp-001", "sovereign-map"],
    timezone: "UTC",
    panels: [
      {
        id: 1,
        title: "Epsilon Remaining",
        type: "stat",
        targets: [
          {
            expr: "privacy_epsilon_remaining",
          },
        ],
        thresholds: {
          steps: [
            { color: "red", value: 0.1 },
            { color: "yellow", value: 0.5 },
            { color: "green", value: null },
          ],
        },
      },
      {
        id: 2,
        title: "Privacy Overhead (%)",
        type: "stat",
        targets: [
          {
            expr: "privacy_overhead_percent",
          },
        ],
        thresholds: {
          steps: [
            { color: "green", value: null },
            { color: "yellow", value: 20 },
            { color: "red", value: 100 },
          ],
        },
      },
      {
        id: 3,
        title: "Noise Injection Rate (per second)",
        type: "graph",
        targets: [
          {
            expr: "rate(privacy_noise_injected_total[1m])",
          },
        ],
      },
      {
        id: 4,
        title: "Noise Magnitude Distribution",
        type: "heatmap",
        targets: [
          {
            expr: "privacy_noise_magnitude",
          },
        ],
      },
      {
        id: 5,
        title: "Gaussian vs Laplace Mix",
        type: "piechart",
        targets: [
          {
            expr: "privacy_gaussian_noise_count_total",
          },
          {
            expr: "privacy_laplace_noise_count_total",
          },
        ],
      },
      {
        id: 6,
        title: "CPU vs GPU Overhead",
        type: "stat",
        targets: [
          {
            expr: "privacy_cpu_vs_gpu_overhead",
          },
        ],
      },
    ],
  },
};

export const CONSENSUS_DASHBOARD = {
  dashboard: {
    title: "Consensus & Byzantine Detection",
    description: "Round metrics, Byzantine node detection, consensus quality",
    tags: ["consensus", "byzantine", "sovereign-map"],
    timezone: "UTC",
    panels: [
      {
        id: 1,
        title: "Round Completion Time (p99, seconds)",
        type: "graph",
        targets: [
          {
            expr: "consensus_round_duration_seconds",
          },
        ],
      },
      {
        id: 2,
        title: "Participation Rate (%)",
        type: "gauge",
        targets: [
          {
            expr: "consensus_participation_rate",
          },
        ],
        thresholds: {
          steps: [
            { color: "red", value: 0 },
            { color: "yellow", value: 66.7 },
            { color: "green", value: 75 },
          ],
        },
      },
      {
        id: 3,
        title: "Nodes Online / Offline",
        type: "stat",
        targets: [
          {
            expr: "consensus_nodes_online",
          },
          {
            expr: "consensus_nodes_offline",
          },
        ],
      },
      {
        id: 4,
        title: "Byzantine Nodes Detected (Current)",
        type: "stat",
        targets: [
          {
            expr: "byzantine_nodes_currently_detected",
          },
        ],
        thresholds: {
          steps: [
            { color: "green", value: null },
            { color: "red", value: 1 },
          ],
        },
      },
      {
        id: 5,
        title: "Byzantine Detection Rate (per hour)",
        type: "graph",
        targets: [
          {
            expr: "rate(byzantine_nodes_detected_total[1h])",
          },
        ],
      },
      {
        id: 6,
        title: "Rounds Completed (Total)",
        type: "stat",
        targets: [
          {
            expr: "consensus_rounds_completed",
          },
        ],
      },
      {
        id: 7,
        title: "Consensus Success Rate (%)",
        type: "stat",
        targets: [
          {
            expr: "(consensus_rounds_completed / (consensus_rounds_completed + consensus_failed)) * 100",
          },
        ],
      },
    ],
  },
};

export const NETWORK_DASHBOARD = {
  dashboard: {
    title: "Network Partition & Recovery",
    description: "Partition detection, recovery time, attestation validation",
    tags: ["network", "partition", "recovery", "sovereign-map"],
    timezone: "UTC",
    panels: [
      {
        id: 1,
        title: "Partitions Detected (Total)",
        type: "stat",
        targets: [
          {
            expr: "network_partitions_detected_total",
          },
        ],
      },
      {
        id: 2,
        title: "Current Partition Size",
        type: "stat",
        targets: [
          {
            expr: "network_partition_isolated_nodes",
          },
        ],
        thresholds: {
          steps: [
            { color: "green", value: null },
            { color: "yellow", value: 10 },
            { color: "red", value: 100 },
          ],
        },
      },
      {
        id: 3,
        title: "Recovery Time (p99, seconds)",
        type: "graph",
        targets: [
          {
            expr: "network_partition_recovery_seconds",
          },
        ],
      },
      {
        id: 4,
        title: "Recovery Success Rate (%)",
        type: "stat",
        targets: [
          {
            expr: "(network_partition_recovery_success / (network_partition_recovery_success + network_partition_recovery_failure)) * 100",
          },
        ],
        thresholds: {
          steps: [
            { color: "red", value: 0 },
            { color: "yellow", value: 90 },
            { color: "green", value: 99 },
          ],
        },
      },
      {
        id: 5,
        title: "Attestation Validation Failures",
        type: "graph",
        targets: [
          {
            expr: "rate(attestation_validation_failures[1m])",
          },
        ],
      },
      {
        id: 6,
        title: "Partition Timeline",
        type: "timeline",
        targets: [
          {
            expr: "network_partitions_detected_total",
          },
        ],
      },
    ],
  },
};

export const SYSTEM_DASHBOARD = {
  dashboard: {
    title: "System Health",
    description: "Memory, CPU, network latency, uptime",
    tags: ["system", "infrastructure", "sovereign-map"],
    timezone: "UTC",
    panels: [
      {
        id: 1,
        title: "Memory Usage (GB)",
        type: "graph",
        targets: [
          {
            expr: "system_memory_bytes / 1024 / 1024 / 1024",
          },
        ],
      },
      {
        id: 2,
        title: "CPU Usage (%)",
        type: "gauge",
        targets: [
          {
            expr: "system_cpu_usage_percent",
          },
        ],
        thresholds: {
          steps: [
            { color: "green", value: null },
            { color: "yellow", value: 70 },
            { color: "red", value: 90 },
          ],
        },
      },
      {
        id: 3,
        title: "Network Latency (p99, ms)",
        type: "graph",
        targets: [
          {
            expr: "system_network_latency_seconds * 1000",
          },
        ],
      },
      {
        id: 4,
        title: "Process Uptime (days)",
        type: "stat",
        targets: [
          {
            expr: "system_uptime_seconds / 86400",
          },
        ],
      },
      {
        id: 5,
        title: "Process Restarts",
        type: "stat",
        targets: [
          {
            expr: "system_restarts",
          },
        ],
      },
    ],
  },
};

export const SLA_DASHBOARD = {
  dashboard: {
    title: "SLA Compliance Monitor",
    description: "Real-time SLA status and breach alerts",
    tags: ["sla", "compliance", "sovereign-map"],
    timezone: "UTC",
    panels: [
      {
        id: 1,
        title: "Overall SLA Status",
        type: "stat",
        targets: [
          {
            expr: "sla_overall_status",
          },
        ],
        thresholds: {
          steps: [
            { color: "red", value: 0 },
            { color: "green", value: 1 },
          ],
        },
      },
      {
        id: 2,
        title: "Privacy Overhead SLA (<12%)",
        type: "stat",
        targets: [
          {
            expr: "privacy_overhead_percent < 12",
          },
        ],
      },
      {
        id: 3,
        title: "GPU Detection SLA (>99%)",
        type: "stat",
        targets: [
          {
            expr: "gpu_detection_success_rate > 0.99",
          },
        ],
      },
      {
        id: 4,
        title: "Consensus Participation SLA (>66.7%)",
        type: "stat",
        targets: [
          {
            expr: "consensus_participation_rate > 66.7",
          },
        ],
      },
      {
        id: 5,
        title: "Network Latency SLA (<100ms p99)",
        type: "stat",
        targets: [
          {
            expr: "system_network_latency_p99_ms < 100",
          },
        ],
      },
      {
        id: 6,
        title: "Byzantine Tolerance SLA (<33%)",
        type: "stat",
        targets: [
          {
            expr: "consensus_byzantine_node_ratio < 0.33",
          },
        ],
      },
    ],
  },
};

/**
 * Export all dashboards
 */
export const DASHBOARDS = {
  gpu: GPU_DASHBOARD,
  privacy: PRIVACY_DASHBOARD,
  consensus: CONSENSUS_DASHBOARD,
  network: NETWORK_DASHBOARD,
  system: SYSTEM_DASHBOARD,
  sla: SLA_DASHBOARD,
};

export default DASHBOARDS;

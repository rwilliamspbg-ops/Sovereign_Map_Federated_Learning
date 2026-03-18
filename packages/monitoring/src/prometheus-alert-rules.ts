/**
 * Prometheus Alerting Rules for Sovereign Map FL
 * 
 * Add to prometheus.yml:
 * rule_files:
 *   - '/path/to/prometheus-alert-rules.yml'
 */

export const PROMETHEUS_ALERT_RULES = `
groups:
  - name: sovereign_map_alerts
    interval: 30s
    rules:
      # Privacy Alerts
      - alert: HighPrivacyOverhead
        expr: privacy_overhead_percent > 20
        for: 5m
        labels:
          severity: warning
          category: privacy
        annotations:
          summary: "High privacy overhead detected ({{ $value | humanize }}%)"
          description: "Privacy overhead has exceeded 20% threshold on {{ $labels.instance }}"
          runbook: "docs/PRIVACY_RUNBOOK.md"

      - alert: PrivacyBudgetExhausted
        expr: privacy_epsilon_remaining < 0.1
        for: 1m
        labels:
          severity: critical
          category: privacy
        annotations:
          summary: "Privacy epsilon budget exhausted"
          description: "Remaining epsilon budget ({{ $value }}) on {{ $labels.instance }} is critically low"
          action: "Stop new updates or deploy with larger epsilon budget"

      - alert: NoiseInjectionFailure
        expr: rate(privacy_noise_injection_failures[5m]) > 0.01
        for: 5m
        labels:
          severity: warning
          category: privacy
        annotations:
          summary: "Noise injection failures detected"
          description: "Failure rate: {{ $value | humanizePercentage }} on {{ $labels.instance }}"

      # GPU Alerts
      - alert: GPUDetectionFailure
        expr: rate(gpu_detection_failures[5m]) > 0.1
        for: 5m
        labels:
          severity: warning
          category: gpu
        annotations:
          summary: "GPU detection failure rate elevated"
          description: "{{ $value | humanizePercentage }} failures on {{ $labels.instance }}"
          action: "Verify GPU drivers and CUDA/ROCm installation"

      - alert: GPUMemoryExhaustion
        expr: (gpu_memory_used_bytes / gpu_memory_total_bytes) * 100 > 90
        for: 2m
        labels:
          severity: critical
          category: gpu
        annotations:
          summary: "GPU memory usage critical ({{ $value | humanize }}%)"
          description: "GPU memory on {{ $labels.instance }} is {{ $value | humanize }}% full"
          action: "Reduce batch size or enable GPU memory swapping"

      - alert: GPUHighTemperature
        expr: gpu_temperature_celsius > 85
        for: 3m
        labels:
          severity: warning
          category: gpu
        annotations:
          summary: "GPU temperature elevated ({{ $value | humanize }}°C)"
          description: "GPU temperature on {{ $labels.instance }} has exceeded 85°C"
          action: "Check cooling system, reduce workload, or thermal throttle settings"

      - alert: GPUNoiseLatency
        expr: gpu_noise_latency_seconds > 0.1
        for: 5m
        labels:
          severity: warning
          category: gpu
        annotations:
          summary: "Noise generation latency high ({{ $value | humanize}}s)"
          description: "GPU noise latency on {{ $labels.instance }} exceeds 100ms target"

      # Consensus Alerts
      - alert: LowConsensusParticipation
        expr: consensus_participation_rate < 66.7
        for: 5m
        labels:
          severity: critical
          category: consensus
        annotations:
          summary: "Consensus participation below 66.7%"
          description: "Participation rate: {{ $value | humanize }}% on {{ $labels.instance }}"
          action: "Investigate node connectivity and Byzantine detection"

      - alert: ByzantineNodesDetected
        expr: byzantine_nodes_currently_detected >= 1
        for: 1m
        labels:
          severity: critical
          category: byzantine
        annotations:
          summary: "Byzantine nodes detected (count: {{ $value | humanize }})"
          description: "{{ $value | humanize }} Byzantine nodes on {{ $labels.instance }}"
          action: "Review validation logs and consider node quarantine"

      - alert: HighByzantineRatio
        expr: (byzantine_nodes_detected_total / consensus_nodes_online) * 100 > 25
        for: 5m
        labels:
          severity: critical
          category: byzantine
        annotations:
          summary: "Byzantine ratio exceeds 25% threshold"
          description: "Byzantine ratio: {{ $value | humanize }}% on {{ $labels.instance }}"
          action: "URGENT: Investigate potential coordinated attack"

      - alert: ConsensusRoundTimeout
        expr: rate(consensus_round_timeout_total[5m]) > 0.01
        for: 5m
        labels:
          severity: warning
          category: consensus
        annotations:
          summary: "Consensus round timeouts detected"
          description: "Timeout rate: {{ $value | humanizePercentage }} on {{ $labels.instance }}"
          action: "Check network latency and node responsiveness"

      # Network Partition Alerts
      - alert: NetworkPartitionDetected
        expr: network_partitions_detected_total > 0
        for: 1m
        labels:
          severity: warning
          category: network
        annotations:
          summary: "Network partition detected"
          description: "Partition detected on {{ $labels.instance }}"
          action: "Review partition recovery manager logs"

      - alert: PartitionPersists
        expr: network_partition_isolated_nodes > 0
        for: 5m
        labels:
          severity: critical
          category: network
        annotations:
          summary: "Network partition persists ({{ $value | humanize }} nodes isolated)"
          description: "Partition on {{ $labels.instance }} has lasted >5 minutes"
          action: "URGENT: Check network connectivity, may indicate ISP outage"

      - alert: PartitionRecoveryFailure
        expr: rate(network_partition_recovery_failure[5m]) > 0.1
        for: 5m
        labels:
          severity: critical
          category: network
        annotations:
          summary: "Partition recovery failures detected"
          description: "Failure rate: {{ $value | humanizePercentage }} on {{ $labels.instance }}"
          action: "Review attestation chain validator, check clock sync"

      - alert: AttestationValidationFailure
        expr: rate(attestation_validation_failures[5m]) > 0.01
        for: 5m
        labels:
          severity: warning
          category: network
        annotations:
          summary: "Attestation validation failures elevated"
          description: "Failure rate: {{ $value | humanizePercentage }} on {{ $labels.instance }}"
          action: "Verify attestation chain integrity and Byzantine node isolation"

      # System Alerts
      - alert: HighCPUUsage
        expr: system_cpu_usage_percent > 90
        for: 5m
        labels:
          severity: warning
          category: system
        annotations:
          summary: "CPU usage critical ({{ $value | humanize }}%)"
          description: "CPU on {{ $labels.instance }} is {{ $value | humanize }}% utilized"
          action: "Review process usage, consider horizontal scaling"

      - alert: HighMemoryUsage
        expr: (system_memory_bytes / 1024 / 1024 / 1024) > 90
        for: 5m
        labels:
          severity: warning
          category: system
        annotations:
          summary: "System memory usage high ({{ $value | humanize }}GB)"
          description: "Memory on {{ $labels.instance }} is {{ $value | humanize }}GB"
          action: "Review memory leaks, consider memory limits"

      - alert: HighNetworkLatency
        expr: system_network_latency_seconds * 1000 > 100
        for: 5m
        labels:
          severity: warning
          category: system
        annotations:
          summary: "Network latency high (p99: {{ $value | humanize }}ms)"
          description: "Network latency on {{ $labels.instance }} exceeds 100ms"
          action: "Investigate network infrastructure, ISP latency"

      - alert: ProcessDown
        expr: up{job="sovereign-map"} == 0
        for: 1m
        labels:
          severity: critical
          category: system
        annotations:
          summary: "Sovereign Map process is down"
          description: "Process on {{ $labels.instance }} is not responding"
          action: "URGENT: Restart process, check logs for crash"

      - alert: HighRestartFrequency
        expr: rate(system_restarts[1h]) > 0.01
        for: 1m
        labels:
          severity: warning
          category: system
        annotations:
          summary: "Process restarting frequently"
          description: "Restart rate: {{ $value | humanize }} per hour on {{ $labels.instance }}"
          action: "Review logs for crash causes"

      # SLA Alerts
      - alert: SLAViolation
        expr: sla_overall_status == 0
        for: 1m
        labels:
          severity: critical
          category: sla
        annotations:
          summary: "SLA violation detected"
          description: "One or more SLA metrics failed on {{ $labels.instance }}"
          action: "Review SLA dashboard for details"

      - alert: PrivacyOverheadSLABreach
        expr: privacy_overhead_percent > 12
        for: 5m
        labels:
          severity: critical
          category: sla
        annotations:
          summary: "Privacy overhead SLA breached (>12%)"
          description: "Current overhead: {{ $value | humanize }}%"
          action: "Enable GPU acceleration or optimize CPU implementation"

      - alert: GPUDetectionSLABreach
        expr: gpu_detection_success_rate < 0.99
        for: 5m
        labels:
          severity: warning
          category: sla
        annotations:
          summary: "GPU detection SLA breached (<99%)"
          description: "Success rate: {{ $value | humanizePercentage }}"
          action: "Investigate GPU detection failures"

      - alert: NetworkLatencySLABreach
        expr: system_network_latency_milliseconds_p99 > 100
        for: 5m
        labels:
          severity: warning
          category: sla
        annotations:
          summary: "Network latency SLA breached (>100ms p99)"
          description: "Latency: {{ $value | humanize }}ms"
          action: "Investigate network infrastructure"

      - alert: ByzantineSLABreach
        expr: consensus_byzantine_node_ratio > 0.33
        for: 5m
        labels:
          severity: critical
          category: sla
        annotations:
          summary: "Byzantine tolerance SLA breached (>33%)"
          description: "Ratio: {{ $value | humanizePercentage }}"
          action: "CRITICAL: Investigate potential coordinated attack"

      # Aggregate Alerts
      - alert: MultipleSystemsDown
        expr: count(up{job="sovereign-map"} == 0) >= 3
        for: 2m
        labels:
          severity: critical
          category: system
        annotations:
          summary: "Multiple Sovereign Map instances down"
          description: "{{ $value | humanize }} instances not responding"
          action: "URGENT: Investigate infrastructure, potential regional outage"

      - alert: WidespreadByzantineActivity
        expr: count(byzantine_nodes_currently_detected >= 1) >= 2
        for: 5m
        labels:
          severity: critical
          category: security
        annotations:
          summary: "Byzantine activity detected across multiple instances"
          description: "{{ $value | humanize }} instances reporting Byzantine nodes"
          action: "CRITICAL: Coordinate with security team, review attack logs"
`;

/**
 * Export as structured object for programmatic use
 */
export const ALERT_RULES = {
  privacy: {
    highOverhead: {
      name: "HighPrivacyOverhead",
      expr: "privacy_overhead_percent > 20",
      severity: "warning",
      threshold: 20
    },
    budgetExhausted: {
      name: "PrivacyBudgetExhausted",
      expr: "privacy_epsilon_remaining < 0.1",
      severity: "critical",
      threshold: 0.1
    }
  },
  gpu: {
    detectionFailure: {
      name: "GPUDetectionFailure",
      expr: "rate(gpu_detection_failures[5m]) > 0.1",
      severity: "warning"
    },
    memoryExhaustion: {
      name: "GPUMemoryExhaustion",
      expr: "(gpu_memory_used_bytes / gpu_memory_total_bytes) * 100 > 90",
      severity: "critical",
      threshold: 90
    }
  },
  consensus: {
    lowParticipation: {
      name: "LowConsensusParticipation",
      expr: "consensus_participation_rate < 66.7",
      severity: "critical",
      threshold: 66.7
    },
    byzantineDetected: {
      name: "ByzantineNodesDetected",
      expr: "byzantine_nodes_currently_detected >= 1",
      severity: "critical"
    }
  },
  network: {
    partitionDetected: {
      name: "NetworkPartitionDetected",
      expr: "network_partitions_detected_total > 0",
      severity: "warning"
    },
    partitionPersists: {
      name: "PartitionPersists",
      expr: "network_partition_isolated_nodes > 0",
      severity: "critical",
      duration: "5m"
    }
  },
  system: {
    highCPU: {
      name: "HighCPUUsage",
      expr: "system_cpu_usage_percent > 90",
      severity: "warning",
      threshold: 90
    },
    processDown: {
      name: "ProcessDown",
      expr: 'up{job="sovereign-map"} == 0',
      severity: "critical"
    }
  },
  sla: {
    privacyBreach: {
      name: "PrivacyOverheadSLABreach",
      expr: "privacy_overhead_percent > 12",
      severity: "critical",
      slaTarget: 12
    },
    byzantineBreach: {
      name: "ByzantineSLABreach",
      expr: "consensus_byzantine_node_ratio > 0.33",
      severity: "critical",
      slaTarget: 0.33
    }
  }
};

export default PROMETHEUS_ALERT_RULES;

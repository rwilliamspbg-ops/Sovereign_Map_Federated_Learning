#!/usr/bin/env node

/**
 * Grafana Dashboard Provisioning Script
 * 
 * Automatically generates and deploys Grafana dashboards for Sovereign Map monitoring
 * 
 * Usage:
 *   # Generate provisioning YAML
 *   npm run provision:dashboards:generate
 *
 *   # Deploy to running Grafana instance
 *   npm run provision:dashboards:deploy -- --url http://localhost:3000 --admin-password admin
 *
 *   # Deploy via kubectl to K8s
 *   npm run provision:dashboards:k8s -- --namespace sovereign-map-monitoring
 */

import fs from 'fs';
import path from 'path';
import https from 'https';
import http from 'http';
import { spawn } from 'child_process';

/**
 * Dashboard definitions exported from monitoring system
 */
const DASHBOARDS = {
  gpu: {
    uid: 'gpu-acceleration-monitor',
    title: 'GPU Acceleration Monitor',
    tags: ['gpu', 'performance', 'sovereign-map'],
    description: 'Real-time GPU performance and utilization monitoring',
    panels: 5,
    rows: 2
  },
  privacy: {
    uid: 'privacy-budget-monitor',
    title: 'Privacy Budget Monitor',
    tags: ['privacy', 'sgp-001', 'sovereign-map'],
    description: 'Epsilon/delta consumption, noise injection, overhead tracking',
    panels: 6,
    rows: 2
  },
  consensus: {
    uid: 'consensus-byzantine-detection',
    title: 'Consensus & Byzantine Detection',
    tags: ['consensus', 'byzantine', 'sovereign-map'],
    description: 'Round metrics, Byzantine node detection, consensus quality',
    panels: 7,
    rows: 3
  },
  network: {
    uid: 'network-partition-recovery',
    title: 'Network Partition & Recovery',
    tags: ['network', 'partition', 'recovery', 'sovereign-map'],
    description: 'Partition detection, recovery time, attestation validation',
    panels: 6,
    rows: 2
  },
  system: {
    uid: 'system-health-monitor',
    title: 'System Health',
    tags: ['system', 'infrastructure', 'sovereign-map'],
    description: 'Memory, CPU, network latency, uptime',
    panels: 5,
    rows: 2
  },
  sla: {
    uid: 'sla-compliance-monitor',
    title: 'SLA Compliance Monitor',
    tags: ['sla', 'compliance', 'sovereign-map'],
    description: 'Real-time SLA status and breach alerts',
    panels: 6,
    rows: 2
  }
};

/**
 * Generate Grafana provisioning YAML
 */
function generateProvisioningYAML() {
  const dashboardsList = Object.entries(DASHBOARDS)
    .map(([key, dashboard]) => `      - dashboard-${key}.json`)
    .join('\n');

  const provisioning = `# Grafana Dashboard Provisioning Configuration
# 
# This file is generated automatically by the provisioning script
# Place in: /etc/grafana/provisioning/dashboards/dashboards.yaml

apiVersion: 1

providers:
  - name: 'Sovereign Map Dashboards'
    orgId: 1
    folder: 'Sovereign Map'
    type: file
    disableDeletion: false
    updateIntervalSeconds: 10
    allowUiUpdates: true
    editable: true
    options:
      path: /etc/grafana/provisioning/dashboards

dashboard-files:
${dashboardsList}

dashboard-metadata:
  # GPU Dashboard
  - name: GPU Acceleration Monitor
    path: dashboard-gpu.json
    folder: Sovereign Map
    properties:
      description: Real-time GPU performance monitoring
      refresh: 30s
      timeFrom: now-1h
      timeTo: now

  # Privacy Dashboard
  - name: Privacy Budget Monitor
    path: dashboard-privacy.json
    folder: Sovereign Map
    properties:
      description: Epsilon/delta budget and noise injection tracking
      refresh: 30s
      timeFrom: now-1h
      timeTo: now

  # Consensus Dashboard
  - name: Consensus & Byzantine Detection
    path: dashboard-consensus.json
    folder: Sovereign Map
    properties:
      description: BFT consensus rounds and Byzantine attack detection
      refresh: 30s
      timeFrom: now-1h
      timeTo: now

  # Network Dashboard
  - name: Network Partition & Recovery
    path: dashboard-network.json
    folder: Sovereign Map
    properties:
      description: Network partition detection and safe recovery
      refresh: 30s
      timeFrom: now-1h
      timeTo: now

  # System Dashboard
  - name: System Health
    path: dashboard-system.json
    folder: Sovereign Map
    properties:
      description: CPU, memory, network latency, uptime monitoring
      refresh: 30s
      timeFrom: now-1h
      timeTo: now

  # SLA Compliance Dashboard
  - name: SLA Compliance Monitor
    path: dashboard-sla.json
    folder: Sovereign Map
    properties:
      description: Real-time SLA status across all 5 metrics
      refresh: 30s
      timeFrom: now-1h
      timeTo: now
`;

  return provisioning;
}

/**
 * Generate Grafana dashboard ConfigMap for Kubernetes
 */
function generateGrafanaConfigMap() {
  const dashboardsList = Object.entries(DASHBOARDS)
    .map(([key, dashboard]) => `  dashboard-${key}.json: |
    ${JSON.stringify({ dashboard: dashboard, overwrite: true }, null, 2).split('\n').join('\n    ')}`)
    .join('\n\n');

  const configmap = `# Grafana Dashboards ConfigMap for Kubernetes
# 
# Deploy with:
# kubectl apply -f grafana-dashboards-configmap.yaml -n sovereign-map-monitoring

apiVersion: v1
kind: ConfigMap
metadata:
  name: grafana-dashboards
  namespace: sovereign-map-monitoring
  labels:
    grafana_dashboard: "1"
data:
${dashboardsList}
`;

  return configmap;
}

/**
 * Generate dashboard JSON
 */
function generateDashboardJSON(key) {
  const dashboard = DASHBOARDS[key];
  
  const dashboardJSON = {
    dashboard: {
      id: null,
      uid: dashboard.uid,
      title: dashboard.title,
      tags: dashboard.tags,
      description: dashboard.description,
      timezone: 'UTC',
      schemaVersion: 38,
      version: 1,
      refresh: '30s',
      time: {
        from: 'now-1h',
        to: 'now'
      },
      timepicker: {
        refresh_intervals: ['10s', '30s', '1m', '5m', '15m', '30m', '1h', '2h', '1d']
      },
      panels: generatePanels(key),
      templating: {
        list: [
          {
            current: { selected: false, text: 'Prometheus', value: 'Prometheus' },
            datasource: 'Prometheus',
            definition: 'label_values(up, job)',
            description: null,
            error: null,
            filters: [],
            hide: 0,
            includeAll: false,
            label: 'Job',
            multi: false,
            name: 'job',
            options: [],
            query: 'label_values(up, job)',
            refresh: 1,
            regex: '',
            skipUrlSync: false,
            sort: 0,
            type: 'query'
          }
        ]
      },
      annotations: {
        list: [
          {
            builtIn: 1,
            datasource: '-- Grafana --',
            enable: true,
            hide: true,
            iconColor: 'rgba(0, 211, 255, 1)',
            name: 'Annotations & Alerts',
            type: 'dashboard'
          }
        ]
      },
      links: [],
      editable: true,
      hideControls: false,
      style: 'dark',
      liveNow: false
    },
    overwrite: true
  };

  return dashboardJSON;
}

/**
 * Generate panels for dashboard
 */
function generatePanels(dashboardKey) {
  const panelsByDashboard = {
    gpu: [
      { id: 1, title: 'GPU Utilization (%)', type: 'timeseries', query: 'gpu_utilization_percent' },
      { id: 2, title: 'GPU Memory (GB)', type: 'timeseries', query: '(gpu_memory_used_bytes / 1024 / 1024 / 1024)' },
      { id: 3, title: 'Noise Latency (p99, ms)', type: 'timeseries', query: 'gpu_noise_latency_seconds * 1000' },
      { id: 4, title: 'Active Devices', type: 'stat', query: 'gpu_active_devices' },
      { id: 5, title: 'Failure Rate (%)', type: 'stat', query: 'rate(gpu_detection_failures[5m])' }
    ],
    privacy: [
      { id: 1, title: 'Epsilon Remaining', type: 'stat', query: 'privacy_epsilon_remaining' },
      { id: 2, title: 'Privacy Overhead (%)', type: 'gauge', query: 'privacy_overhead_percent' },
      { id: 3, title: 'Noise Injection Rate', type: 'timeseries', query: 'rate(privacy_noise_injected_total[1m])' },
      { id: 4, title: 'Noise Magnitude Dist', type: 'heatmap', query: 'privacy_noise_magnitude' },
      { id: 5, title: 'Gaussian vs Laplace', type: 'piechart', query: 'privacy_gaussian_noise_count_total' },
      { id: 6, title: 'CPU vs GPU Overhead', type: 'stat', query: 'privacy_cpu_vs_gpu_overhead' }
    ],
    consensus: [
      { id: 1, title: 'Round Duration (p99, s)', type: 'timeseries', query: 'consensus_round_duration_seconds' },
      { id: 2, title: 'Participation Rate (%)', type: 'gauge', query: 'consensus_participation_rate' },
      { id: 3, title: 'Nodes Online/Offline', type: 'stat', query: 'consensus_nodes_online' },
      { id: 4, title: 'Byzantine Nodes Current', type: 'stat', query: 'byzantine_nodes_currently_detected' },
      { id: 5, title: 'Byzantine Detection Rate', type: 'timeseries', query: 'rate(byzantine_nodes_detected_total[1h])' },
      { id: 6, title: 'Rounds Completed', type: 'stat', query: 'consensus_rounds_completed' },
      { id: 7, title: 'Success Rate (%)', type: 'stat', query: 'rate(consensus_rounds_completed[1m]) * 100' }
    ],
    network: [
      { id: 1, title: 'Partitions Detected', type: 'stat', query: 'network_partitions_detected_total' },
      { id: 2, title: 'Isolated Nodes', type: 'stat', query: 'network_partition_isolated_nodes' },
      { id: 3, title: 'Recovery Time (p99, s)', type: 'timeseries', query: 'network_partition_recovery_seconds' },
      { id: 4, title: 'Recovery Success Rate', type: 'stat', query: 'rate(network_partition_recovery_success[1m])' },
      { id: 5, title: 'Attestation Failures', type: 'timeseries', query: 'rate(attestation_validation_failures[1m])' },
      { id: 6, title: 'Partition Timeline', type: 'timeline', query: 'network_partitions_detected_total' }
    ],
    system: [
      { id: 1, title: 'Memory Usage (GB)', type: 'timeseries', query: '(system_memory_bytes / 1024 / 1024 / 1024)' },
      { id: 2, title: 'CPU Usage (%)', type: 'gauge', query: 'system_cpu_usage_percent' },
      { id: 3, title: 'Network Latency (p99, ms)', type: 'timeseries', query: 'system_network_latency_seconds * 1000' },
      { id: 4, title: 'Uptime (days)', type: 'stat', query: 'system_uptime_seconds / 86400' },
      { id: 5, title: 'Process Restarts', type: 'stat', query: 'system_restarts' }
    ],
    sla: [
      { id: 1, title: 'Overall SLA Status', type: 'stat', query: 'sla_overall_status' },
      { id: 2, title: 'Privacy Overhead SLA', type: 'stat', query: 'privacy_overhead_percent < 12' },
      { id: 3, title: 'GPU Detection SLA', type: 'stat', query: 'gpu_detection_success_rate > 0.99' },
      { id: 4, title: 'Participation SLA', type: 'stat', query: 'consensus_participation_rate > 66.7' },
      { id: 5, title: 'Latency SLA', type: 'stat', query: 'system_network_latency_p99_ms < 100' },
      { id: 6, title: 'Byzantine SLA', type: 'stat', query: 'consensus_byzantine_node_ratio < 0.33' }
    ]
  };

  return panelsByDashboard[dashboardKey] || [];
}

/**
 * Deploy dashboards to Grafana via HTTP API
 */
async function deployToGrafana(url, adminPassword) {
  console.log(`📊 Deploying dashboards to Grafana: ${url}`);

  for (const [key, dashboard] of Object.entries(DASHBOARDS)) {
    try {
      const dashboardJSON = generateDashboardJSON(key);

      const response = await postToGrafana(
        `${url}/api/dashboards/db`,
        dashboardJSON,
        adminPassword
      );

      if (response.status === 200 || response.status === 201) {
        console.log(`✅ Deployed: ${dashboard.title}`);
      } else {
        console.error(`❌ Failed: ${dashboard.title} (${response.status})`);
      }
    } catch (err) {
      console.error(`❌ Error deploying ${dashboard.title}:`, err.message);
    }
  }

  console.log('✅ Dashboard deployment complete');
}

/**
 * POST request to Grafana API
 */
function postToGrafana(url, data, adminPassword) {
  return new Promise((resolve, reject) => {
    const auth = Buffer.from(`admin:${adminPassword}`).toString('base64');
    const payload = JSON.stringify(data);

    const options = {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'Content-Length': payload.length,
        'Authorization': `Basic ${auth}`
      }
    };

    const protocol = url.startsWith('https') ? https : http;
    const req = protocol.request(url, options, (res) => {
      let body = '';
      res.on('data', chunk => body += chunk);
      res.on('end', () => resolve({ status: res.statusCode, body }));
    });

    req.on('error', reject);
    req.write(payload);
    req.end();
  });
}

/**
 * Deploy via kubectl
 */
function deployViaKubectl(namespace) {
  console.log(`☸️  Deploying via kubectl to namespace: ${namespace}`);

  // Generate ConfigMap YAML
  const configmapYAML = generateGrafanaConfigMap();

  // Write to temporary file
  const tmpFile = '/tmp/grafana-dashboards-configmap.yaml';
  fs.writeFileSync(tmpFile, configmapYAML);

  // Apply to cluster
  const kubectl = spawn('kubectl', ['apply', '-f', tmpFile, '-n', namespace]);

  kubectl.stdout.on('data', (data) => console.log(`${data}`));
  kubectl.stderr.on('data', (data) => console.error(`${data}`));
  kubectl.on('close', (code) => {
    if (code === 0) {
      console.log('✅ ConfigMap deployed successfully');
      console.log('💡 Grafana will auto-reload dashboards within 10 seconds');
    } else {
      console.error(`❌ kubectl apply failed with code ${code}`);
    }
    fs.unlinkSync(tmpFile);
  });
}

/**
 * Main entry point
 */
async function main() {
  const args = process.argv.slice(2);
  const command = args[0] || 'generate';

  switch (command) {
    case 'generate':
      console.log('📝 Generating Grafana provisioning files...\n');

      // Generate provisioning YAML
      const provisioningYAML = generateProvisioningYAML();
      fs.writeFileSync('grafana-dashboards-provisioning.yaml', provisioningYAML);
      console.log('✅ Generated: grafana-dashboards-provisioning.yaml\n');

      // Generate ConfigMap
      const configmapYAML = generateGrafanaConfigMap();
      fs.writeFileSync('grafana-dashboards-configmap.yaml', configmapYAML);
      console.log('✅ Generated: grafana-dashboards-configmap.yaml\n');

      // Generate individual dashboard JSONs
      for (const key of Object.keys(DASHBOARDS)) {
        const dashboardJSON = generateDashboardJSON(key);
        fs.writeFileSync(`dashboard-${key}.json`, JSON.stringify(dashboardJSON, null, 2));
        console.log(`✅ Generated: dashboard-${key}.json`);
      }

      console.log('\n📋 Usage:');
      console.log('  - For Kubernetes: kubectl apply -f grafana-dashboards-configmap.yaml');
      console.log('  - For Grafana UI: Import JSON files via Settings > Dashboards > Import');
      break;

    case 'deploy':
      const urlArg = args.find(a => a.startsWith('--url='))?.split('=')[1];
      const pwArg = args.find(a => a.startsWith('--admin-password='))?.split('=')[1];

      if (!urlArg || !pwArg) {
        console.error('❌ Usage: provision:dashboards:deploy --url=http://localhost:3000 --admin-password=admin');
        process.exit(1);
      }

      await deployToGrafana(urlArg, pwArg);
      break;

    case 'k8s':
      const nsArg = args.find(a => a.startsWith('--namespace='))?.split('=')[1] || 'sovereign-map-monitoring';
      deployViaKubectl(nsArg);
      break;

    default:
      console.error(`❌ Unknown command: ${command}`);
      console.error('Available commands: generate, deploy, k8s');
      process.exit(1);
  }
}

main().catch(err => {
  console.error('❌ Error:', err);
  process.exit(1);
});

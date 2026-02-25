#!/usr/bin/env python
"""
GRAFANA MONITORING DASHBOARD for Byzantine Fault Tolerance
Real-time BFT metrics, amplification detection, and production monitoring
"""

import json
import time

# ============================================================================
# PROMETHEUS METRICS EXPORTER
# ============================================================================

class BFTMetricsExporter:
    """Export BFT metrics in Prometheus format"""
    
    def __init__(self, port=8000):
        self.port = port
        self.metrics = {
            'bft_convergence_accuracy': {},
            'bft_byzantine_level': {},
            'bft_amplification_factor': {},
            'bft_recovery_time': {},
            'bft_throughput': {},
            'bft_memory_usage': {},
            'bft_consensus_rounds': {},
            'bft_attack_detected': {}
        }
    
    def update_metric(self, metric_name, value, labels=None):
        """Update a metric value"""
        if metric_name not in self.metrics:
            self.metrics[metric_name] = {}
        
        key = str(labels) if labels else 'default'
        self.metrics[metric_name][key] = {
            'value': value,
            'timestamp': time.time()
        }
    
    def export_prometheus_format(self):
        """Export metrics in Prometheus format"""
        lines = []
        
        for metric_name, values in self.metrics.items():
            # Add HELP and TYPE
            lines.append(f"# HELP {metric_name} Byzantine Fault Tolerance metric")
            lines.append(f"# TYPE {metric_name} gauge")
            
            # Add metric values
            for label_key, data in values.items():
                if label_key == 'default':
                    lines.append(f"{metric_name} {data['value']}")
                else:
                    lines.append(f"{metric_name}{label_key} {data['value']}")
        
        return "\n".join(lines)
    
    def get_http_response(self):
        """Format for HTTP response"""
        return self.export_prometheus_format()

# ============================================================================
# GRAFANA DASHBOARD JSON
# ============================================================================

def create_bft_dashboard():
    """Create comprehensive BFT monitoring dashboard"""
    
    dashboard = {
        "dashboard": {
            "title": "Byzantine Fault Tolerance Monitoring - 100K Nodes",
            "tags": ["BFT", "Federated Learning", "Production"],
            "timezone": "browser",
            "panels": [
                # Panel 1: Convergence Accuracy
                {
                    "id": 1,
                    "title": "Convergence Accuracy by Byzantine Level",
                    "type": "graph",
                    "gridPos": {"h": 8, "w": 12, "x": 0, "y": 0},
                    "targets": [
                        {
                            "expr": "bft_convergence_accuracy",
                            "legendFormat": "Byzantine {{byzantine_level}}%",
                            "refId": "A"
                        }
                    ],
                    "yaxes": [
                        {"format": "percent", "label": "Accuracy"}
                    ],
                    "alert": {
                        "conditions": [
                            {"evaluator": {"params": [70], "type": "lt"}}
                        ],
                        "message": "Convergence accuracy below 70%"
                    }
                },
                
                # Panel 2: Byzantine Detection
                {
                    "id": 2,
                    "title": "Byzantine Attack Detection",
                    "type": "graph",
                    "gridPos": {"h": 8, "w": 12, "x": 12, "y": 0},
                    "targets": [
                        {
                            "expr": "bft_attack_detected",
                            "legendFormat": "{{attack_type}}",
                            "refId": "A"
                        }
                    ],
                    "alert": {
                        "conditions": [
                            {"evaluator": {"params": [1], "type": "gt"}}
                        ],
                        "message": "Byzantine attack detected!"
                    }
                },
                
                # Panel 3: Amplification Factor
                {
                    "id": 3,
                    "title": "Byzantine Amplification Factor",
                    "type": "graph",
                    "gridPos": {"h": 8, "w": 12, "x": 0, "y": 8},
                    "targets": [
                        {
                            "expr": "bft_amplification_factor",
                            "legendFormat": "Amplification {{strategy}}",
                            "refId": "A"
                        }
                    ],
                    "thresholds": "2.0, 3.0",
                    "alert": {
                        "conditions": [
                            {"evaluator": {"params": [2.5], "type": "gt"}}
                        ],
                        "message": "High amplification factor detected"
                    }
                },
                
                # Panel 4: Recovery Time
                {
                    "id": 4,
                    "title": "Byzantine Recovery Time (Rounds)",
                    "type": "graph",
                    "gridPos": {"h": 8, "w": 12, "x": 12, "y": 8},
                    "targets": [
                        {
                            "expr": "bft_recovery_time",
                            "legendFormat": "Recovery {{byzantine_level}}%",
                            "refId": "A"
                        }
                    ],
                    "alert": {
                        "conditions": [
                            {"evaluator": {"params": [10], "type": "gt"}}
                        ],
                        "message": "Recovery time exceeds 10 rounds"
                    }
                },
                
                # Panel 5: Throughput
                {
                    "id": 5,
                    "title": "System Throughput (Updates/sec)",
                    "type": "graph",
                    "gridPos": {"h": 8, "w": 12, "x": 0, "y": 16},
                    "targets": [
                        {
                            "expr": "bft_throughput",
                            "legendFormat": "Throughput {{node_count}}",
                            "refId": "A"
                        }
                    ],
                    "yaxes": [
                        {"format": "short", "label": "Updates/sec"}
                    ],
                    "alert": {
                        "conditions": [
                            {"evaluator": {"params": [50000], "type": "lt"}}
                        ],
                        "message": "Throughput below expected threshold"
                    }
                },
                
                # Panel 6: Memory Usage
                {
                    "id": 6,
                    "title": "Memory Usage Trend",
                    "type": "graph",
                    "gridPos": {"h": 8, "w": 12, "x": 12, "y": 16},
                    "targets": [
                        {
                            "expr": "bft_memory_usage",
                            "legendFormat": "Memory {{component}}",
                            "refId": "A"
                        }
                    ],
                    "yaxes": [
                        {"format": "bytes", "label": "Memory"}
                    ],
                    "alert": {
                        "conditions": [
                            {"evaluator": {"params": [1000000000], "type": "gt"}}
                        ],
                        "message": "Memory usage exceeds 1GB"
                    }
                },
                
                # Panel 7: Byzantine Tolerance Heatmap
                {
                    "id": 7,
                    "title": "Byzantine Tolerance Heatmap (Accuracy %)",
                    "type": "heatmap",
                    "gridPos": {"h": 8, "w": 24, "x": 0, "y": 24},
                    "targets": [
                        {
                            "expr": "bft_convergence_accuracy",
                            "format": "heatmap",
                            "legendFormat": "{{byzantine_level}}% Byzantine",
                            "refId": "A"
                        }
                    ],
                    "dataFormat": "timeseries"
                },
                
                # Panel 8: System Status (SingleStat)
                {
                    "id": 8,
                    "title": "Current System Status",
                    "type": "singlestat",
                    "gridPos": {"h": 4, "w": 6, "x": 0, "y": 32},
                    "targets": [
                        {
                            "expr": "bft_convergence_accuracy{byzantine_level=\"0\"}",
                            "refId": "A"
                        }
                    ],
                    "thresholds": "60,80",
                    "colors": ["#FF0000", "#FFFF00", "#00FF00"],
                    "sparkline": {"show": True}
                },
                
                # Panel 9: Active Byzantine Attacks
                {
                    "id": 9,
                    "title": "Active Attack Types",
                    "type": "singlestat",
                    "gridPos": {"h": 4, "w": 6, "x": 6, "y": 32},
                    "targets": [
                        {
                            "expr": "sum(bft_attack_detected)",
                            "refId": "A"
                        }
                    ],
                    "thresholds": "0,1"
                },
                
                # Panel 10: Nodes Status
                {
                    "id": 10,
                    "title": "Active Nodes",
                    "type": "singlestat",
                    "gridPos": {"h": 4, "w": 6, "x": 12, "y": 32},
                    "targets": [
                        {
                            "expr": "bft_active_nodes",
                            "refId": "A"
                        }
                    ],
                    "format": "short"
                },
                
                # Panel 11: Alert Status
                {
                    "id": 11,
                    "title": "Alert Status",
                    "type": "singlestat",
                    "gridPos": {"h": 4, "w": 6, "x": 18, "y": 32},
                    "targets": [
                        {
                            "expr": "sum(ALERTS{severity=\"critical\"})",
                            "refId": "A"
                        }
                    ],
                    "colorBackground": True,
                    "thresholds": "0,1"
                }
            ],
            
            "alerts": [
                {
                    "name": "Low Convergence Accuracy",
                    "condition": "bft_convergence_accuracy < 70",
                    "message": "Convergence accuracy dropped below 70% - Byzantine attack or network issue"
                },
                {
                    "name": "High Byzantine Load",
                    "condition": "bft_byzantine_level > 45",
                    "message": "Byzantine node percentage approaching critical threshold"
                },
                {
                    "name": "High Amplification",
                    "condition": "bft_amplification_factor > 2.5",
                    "message": "Byzantine amplification factor too high - coordinated attack detected"
                },
                {
                    "name": "Slow Recovery",
                    "condition": "bft_recovery_time > 10",
                    "message": "System recovery time exceeds threshold"
                },
                {
                    "name": "Low Throughput",
                    "condition": "bft_throughput < 50000",
                    "message": "System throughput below expected level"
                }
            ]
        }
    }
    
    return dashboard

# ============================================================================
# DOCKER COMPOSE FOR MONITORING STACK
# ============================================================================

def create_monitoring_docker_compose():
    """Create docker-compose for Prometheus + Grafana stack"""
    
    compose = """
version: '3.8'

services:
  prometheus:
    image: prom/prometheus:latest
    ports:
      - "9090:9090"
    volumes:
      - ./prometheus.yml:/etc/prometheus/prometheus.yml
      - prometheus_data:/prometheus
    command:
      - '--config.file=/etc/prometheus/prometheus.yml'
      - '--storage.tsdb.path=/prometheus'
    networks:
      - monitoring
    environment:
      - TZ=UTC

  grafana:
    image: grafana/grafana:latest
    ports:
      - "3000:3000"
    volumes:
      - grafana_data:/var/lib/grafana
      - ./dashboards:/etc/grafana/provisioning/dashboards
      - ./datasources:/etc/grafana/provisioning/datasources
    environment:
      - GF_SECURITY_ADMIN_USER=admin
      - GF_SECURITY_ADMIN_PASSWORD=admin
      - GF_INSTALL_PLUGINS=grafana-piechart-panel
    networks:
      - monitoring
    depends_on:
      - prometheus

  alertmanager:
    image: prom/alertmanager:latest
    ports:
      - "9093:9093"
    volumes:
      - ./alertmanager.yml:/etc/alertmanager/config.yml
      - alertmanager_data:/alertmanager
    command:
      - '--config.file=/etc/alertmanager/config.yml'
      - '--storage.path=/alertmanager'
    networks:
      - monitoring

  # BFT Metrics Collector (your app)
  bft_metrics:
    build: .
    ports:
      - "8000:8000"
    environment:
      - PROMETHEUS_PORT=8000
      - NODE_COUNT=100000
    networks:
      - monitoring
    depends_on:
      - prometheus

volumes:
  prometheus_data:
  grafana_data:
  alertmanager_data:

networks:
  monitoring:
    driver: bridge
"""
    
    return compose

# ============================================================================
# PROMETHEUS CONFIGURATION
# ============================================================================

def create_prometheus_config():
    """Create Prometheus configuration"""
    
    config = """
# Prometheus configuration for BFT monitoring

global:
  scrape_interval: 15s
  evaluation_interval: 15s
  external_labels:
    cluster: 'federated-learning'
    environment: 'production'

# Alertmanager configuration
alerting:
  alertmanagers:
    - static_configs:
        - targets:
            - localhost:9093

# Rule files
rule_files:
  - 'bft_rules.yml'

# Scrape configs
scrape_configs:
  # Prometheus self-monitoring
  - job_name: 'prometheus'
    static_configs:
      - targets: ['localhost:9090']

  # BFT Metrics Exporter
  - job_name: 'bft_metrics'
    static_configs:
      - targets: ['localhost:8000']
    scrape_interval: 10s
    metrics_path: '/metrics'

  # Grafana
  - job_name: 'grafana'
    static_configs:
      - targets: ['localhost:3000']

  # Node Exporters (for system metrics)
  - job_name: 'node_exporter'
    static_configs:
      - targets: ['localhost:9100']
"""
    
    return config

# ============================================================================
# ALERTING RULES
# ============================================================================

def create_alert_rules():
    """Create Prometheus alert rules for BFT"""
    
    rules = """
# Byzantine Fault Tolerance Alert Rules

groups:
  - name: bft_alerts
    interval: 30s
    rules:
      # Low Convergence
      - alert: LowConvergenceAccuracy
        expr: bft_convergence_accuracy < 70
        for: 5m
        annotations:
          summary: "Low convergence accuracy detected"
          description: "Byzantine nodes may be launching attack"

      # High Byzantine Load
      - alert: HighByzantineLoad
        expr: bft_byzantine_level > 45
        for: 2m
        annotations:
          summary: "High Byzantine node percentage"
          description: "Approaching critical Byzantine threshold (50%)"

      # High Amplification
      - alert: HighAmplificationFactor
        expr: bft_amplification_factor > 2.5
        for: 3m
        annotations:
          summary: "High Byzantine amplification detected"
          description: "Coordinated Byzantine attack likely"

      # System Recovery Failure
      - alert: RecoveryFailure
        expr: bft_recovery_time > 15
        for: 5m
        annotations:
          summary: "System recovery time critical"
          description: "System unable to recover from Byzantine attack"

      # Throughput Degradation
      - alert: ThroughputDegradation
        expr: bft_throughput < 50000
        for: 10m
        annotations:
          summary: "System throughput degraded"
          description: "Byzantine attack or network congestion"

      # Memory Leak
      - alert: MemoryLeak
        expr: increase(bft_memory_usage[1h]) > 100000000
        for: 5m
        annotations:
          summary: "Memory usage increasing rapidly"
          description: "Potential memory leak detected"

      # Node Failure
      - alert: NodeFailure
        expr: bft_active_nodes < (bft_total_nodes * 0.95)
        for: 2m
        annotations:
          summary: "Significant node failure"
          description: "More than 5% of nodes offline"
"""
    
    return rules

# ============================================================================
# MAIN
# ============================================================================

if __name__ == "__main__":
    print("\n" + "="*100)
    print("  GRAFANA BFT MONITORING DASHBOARD CONFIGURATION")
    print("="*100 + "\n")
    
    # Generate dashboard
    dashboard = create_bft_dashboard()
    
    # Save dashboard JSON
    with open('grafana_bft_dashboard.json', 'w') as f:
        json.dump(dashboard, f, indent=2)
    print("  [OK] Generated grafana_bft_dashboard.json")
    
    # Generate docker-compose
    compose = create_monitoring_docker_compose()
    with open('docker-compose.monitoring.yml', 'w') as f:
        f.write(compose)
    print("  [OK] Generated docker-compose.monitoring.yml")
    
    # Generate prometheus config
    prom_config = create_prometheus_config()
    with open('prometheus.yml', 'w') as f:
        f.write(prom_config)
    print("  [OK] Generated prometheus.yml")
    
    # Generate alert rules
    rules = create_alert_rules()
    with open('bft_rules.yml', 'w') as f:
        f.write(rules)
    print("  [OK] Generated bft_rules.yml")
    
    print("\n  DEPLOYMENT INSTRUCTIONS:\n")
    print("  1. Start monitoring stack:")
    print("     docker-compose -f docker-compose.monitoring.yml up -d")
    print()
    print("  2. Access dashboards:")
    print("     Grafana:       http://localhost:3000 (admin/admin)")
    print("     Prometheus:   http://localhost:9090")
    print("     Alertmanager: http://localhost:9093")
    print()
    print("  3. Import BFT dashboard:")
    print("     Grafana → Dashboards → Import → Select grafana_bft_dashboard.json")
    print()
    print("  4. Configure alerts:")
    print("     Update alertmanager.yml with email/Slack webhooks")
    print()
    print("="*100 + "\n")

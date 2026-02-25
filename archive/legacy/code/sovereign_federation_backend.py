"""
Sovereign Federation Backend: Flask API for Metrics Export & Real-time Sync
Integrates FL system with spatial visualization
"""

from flask import Flask, jsonify, request
from flask_cors import CORS
from flask_socketio import SocketIO, emit, join_room
import json
import numpy as np
from datetime import datetime
import threading
import time
from typing import Dict, List
import os
import sys

# Import FL integration modules
sys.path.insert(0, os.path.dirname(__file__))
from fl_metrics_translator import FLMetricsTranslator, FLMetric
from spatial_threat_analyzer import SpatialThreatAnalyzer, BFTMetrics
import asyncio


class SovereignFederationBackend:
    """Backend for Sovereign Federation system"""
    
    def __init__(self, port: int = 8000):
        self.app = Flask(__name__)
        self.app.config['SECRET_KEY'] = os.getenv('SECRET_KEY', 'sovereign-federation-key')
        CORS(self.app)
        self.socketio = SocketIO(self.app, cors_allowed_origins="*")
        self.port = port
        
        # Initialize components
        self.metrics_translator = FLMetricsTranslator(num_nodes=100000)
        self.threat_analyzer = SpatialThreatAnalyzer()
        
        # State
        self.connected_clients = 0
        self.last_threat_analysis = None
        self.metrics_buffer: List[FLMetric] = []
        
        # Setup routes
        self._setup_routes()
        self._setup_socketio()
        self._start_background_tasks()
    
    def _setup_routes(self):
        """Setup Flask routes"""
        
        @self.app.route('/health', methods=['GET'])
        def health():
            return jsonify({
                'status': 'healthy',
                'timestamp': datetime.now().isoformat(),
                'connected_clients': self.connected_clients,
                'total_nodes': len(self.metrics_translator.metrics_cache)
            })
        
        @self.app.route('/metrics', methods=['GET'])
        def get_metrics():
            """Get current spatial coordinates for all nodes"""
            coordinates = self.metrics_translator.get_all_coordinates()
            return jsonify({
                'timestamp': datetime.now().isoformat(),
                'node_count': len(coordinates),
                'nodes': [
                    {
                        'id': c.node_id,
                        'position': {'x': c.x, 'y': c.y, 'z': c.z},
                        'visual': {
                            'color': c.color,
                            'size': c.size,
                            'intensity': c.intensity
                        },
                        'metrics': {
                            'accuracy': c.accuracy,
                            'byzantine_level': c.byzantine_level,
                            'convergence': c.convergence,
                            'amplification_factor': c.amplification_factor
                        },
                        'status': c.status.value
                    }
                    for c in coordinates
                ]
            })
        
        @self.app.route('/metrics/stats', methods=['GET'])
        def get_stats():
            """Get aggregated statistics"""
            stats = self.metrics_translator.get_aggregated_stats()
            return jsonify({
                'timestamp': datetime.now().isoformat(),
                'statistics': stats,
                'threat_analysis': self.last_threat_analysis
            })
        
        @self.app.route('/metrics/update', methods=['POST'])
        def update_metrics():
            """Update metrics for nodes (from FL system)"""
            data = request.json
            
            if 'nodes' in data:
                for node_data in data['nodes']:
                    metric = FLMetric(
                        node_id=node_data['node_id'],
                        accuracy=node_data.get('accuracy', 80),
                        byzantine_level=node_data.get('byzantine_level', 0),
                        convergence=node_data.get('convergence', 80),
                        throughput=node_data.get('throughput', 80000),
                        recovery_time=node_data.get('recovery_time', 4),
                        amplification_factor=node_data.get('amplification_factor', 1.0),
                        timestamp=datetime.now(),
                        active=node_data.get('active', True)
                    )
                    self.metrics_translator.update_metric(metric)
                    self.metrics_buffer.append(metric)
            
            # Broadcast to connected clients
            self.socketio.emit('metrics_updated', {
                'node_count': len(self.metrics_translator.metrics_cache),
                'timestamp': datetime.now().isoformat()
            }, broadcast=True)
            
            return jsonify({'status': 'success', 'nodes_updated': len(data.get('nodes', []))})
        
        @self.app.route('/threat/analyze', methods=['POST'])
        def analyze_threat():
            """Trigger threat analysis"""
            data = request.json
            
            metrics = BFTMetrics(
                byzantine_percentage=data.get('byzantine_percentage', 0),
                amplification_factor=data.get('amplification_factor', 1.0),
                recovery_time_rounds=data.get('recovery_time_rounds', 4),
                convergence_rate=data.get('convergence_rate', 80),
                active_node_count=data.get('active_node_count', 95000),
                total_node_count=100000,
                attack_patterns=data.get('attack_patterns', []),
                throughput=data.get('throughput', 80000),
                network_partition_detected=data.get('network_partition', False),
                cascading_failure_detected=data.get('cascading_failure', False)
            )
            
            # Run analysis
            loop = asyncio.new_event_loop()
            asyncio.set_event_loop(loop)
            analysis = loop.run_until_complete(
                self.threat_analyzer.analyze_threats(metrics)
            )
            loop.close()
            
            self.threat_analyzer.store_analysis(analysis)
            self.last_threat_analysis = {
                'threat_level': analysis.threat_level.value,
                'severity_score': analysis.severity_score,
                'confidence': analysis.confidence,
                'risk_factors': analysis.risk_factors,
                'immediate_actions': analysis.immediate_actions,
                'recommended_defense': analysis.recommended_defense
            }
            
            # Broadcast to clients
            self.socketio.emit('threat_analysis', self.last_threat_analysis, broadcast=True)
            
            # Get defense protocol
            protocol = self.threat_analyzer.get_defense_protocol(analysis)
            
            return jsonify({
                'analysis': self.last_threat_analysis,
                'protocol': protocol,
                'timestamp': datetime.now().isoformat()
            })
        
        @self.app.route('/prometheus/metrics', methods=['GET'])
        def prometheus_metrics():
            """Export metrics in Prometheus format"""
            prom_format = self.metrics_translator.export_to_prometheus_format()
            return prom_format, 200, {'Content-Type': 'text/plain; charset=utf-8'}
        
        @self.app.route('/api/defense/activate', methods=['POST'])
        def activate_defense():
            """Activate Byzantine defense protocol"""
            data = request.json
            protocol_name = data.get('protocol', 'hierarchical_with_trim')
            
            response = {
                'protocol_activated': protocol_name,
                'timestamp': datetime.now().isoformat(),
                'parameters': {
                    'aggregation_method': 'hierarchical',
                    'trim_percentage': 0.20,
                    'timeout_per_round': 10,
                    'node_isolation_enabled': True
                }
            }
            
            self.socketio.emit('defense_activated', response, broadcast=True)
            return jsonify(response)
    
    def _setup_socketio(self):
        """Setup WebSocket events for real-time sync"""
        
        @self.socketio.on('connect')
        def handle_connect():
            self.connected_clients += 1
            print(f"[INFO] Client connected. Total: {self.connected_clients}")
            emit('connected', {
                'timestamp': datetime.now().isoformat(),
                'total_nodes': len(self.metrics_translator.metrics_cache)
            })
        
        @self.socketio.on('disconnect')
        def handle_disconnect():
            self.connected_clients -= 1
            print(f"[INFO] Client disconnected. Total: {self.connected_clients}")
        
        @self.socketio.on('request_metrics')
        def handle_request_metrics():
            """Client requests current metrics"""
            coordinates = self.metrics_translator.get_all_coordinates()
            emit('metrics_snapshot', {
                'timestamp': datetime.now().isoformat(),
                'nodes': [
                    {
                        'id': c.node_id,
                        'position': {'x': c.x, 'y': c.y, 'z': c.z},
                        'color': c.color,
                        'size': c.size,
                        'intensity': c.intensity,
                        'status': c.status.value
                    }
                    for c in coordinates[:1000]  # Send first 1000 for performance
                ]
            })
        
        @self.socketio.on('trigger_threat_analysis')
        def handle_threat_analysis(data):
            """Trigger threat analysis and broadcast result"""
            metrics = BFTMetrics(
                byzantine_percentage=data.get('byzantine_percentage', 0),
                amplification_factor=data.get('amplification_factor', 1.0),
                recovery_time_rounds=data.get('recovery_time_rounds', 4),
                convergence_rate=data.get('convergence_rate', 80),
                active_node_count=data.get('active_node_count', 95000),
                total_node_count=100000,
                attack_patterns=data.get('attack_patterns', []),
                throughput=data.get('throughput', 80000)
            )
            
            loop = asyncio.new_event_loop()
            asyncio.set_event_loop(loop)
            analysis = loop.run_until_complete(
                self.threat_analyzer.analyze_threats(metrics)
            )
            loop.close()
            
            emit('threat_analysis_result', {
                'threat_level': analysis.threat_level.value,
                'severity_score': analysis.severity_score,
                'risk_factors': analysis.risk_factors,
                'immediate_actions': analysis.immediate_actions
            }, broadcast=True)
    
    def _start_background_tasks(self):
        """Start background monitoring tasks"""
        
        def monitor_system():
            """Background task to monitor system health"""
            while True:
                time.sleep(5)  # Monitor every 5 seconds
                
                stats = self.metrics_translator.get_aggregated_stats()
                
                # Emit to all connected clients
                self.socketio.emit('system_stats', {
                    'timestamp': datetime.now().isoformat(),
                    'total_nodes': stats.get('total_nodes', 0),
                    'avg_accuracy': stats.get('accuracy', {}).get('mean', 0),
                    'avg_convergence': stats.get('convergence', {}).get('mean', 0),
                    'node_status': stats.get('node_status', {})
                }, broadcast=True)
        
        monitor_thread = threading.Thread(target=monitor_system, daemon=True)
        monitor_thread.start()
    
    def simulate_fl_metrics(self, num_updates: int = 100):
        """Simulate FL metrics for testing"""
        print(f"[INFO] Simulating {num_updates} node updates...")
        
        for i in range(num_updates):
            node_id = np.random.randint(0, 100000)
            metric = FLMetric(
                node_id=node_id,
                accuracy=90 + np.random.randn() * 5,
                byzantine_level=np.random.uniform(0, 50),
                convergence=80 + np.random.randn() * 10,
                throughput=80000 + np.random.randn() * 5000,
                recovery_time=4 + np.random.randn(),
                amplification_factor=1.5 + np.random.uniform(0, 2),
                timestamp=datetime.now()
            )
            self.metrics_translator.update_metric(metric)
        
        print(f"[INFO] Simulated metrics for {num_updates} nodes")
    
    def run(self):
        """Start the backend server"""
        print(f"[INFO] Starting Sovereign Federation Backend on port {self.port}")
        print(f"[INFO] Endpoints:")
        print(f"      - Health: http://localhost:{self.port}/health")
        print(f"      - Metrics: http://localhost:{self.port}/metrics")
        print(f"      - Prometheus: http://localhost:{self.port}/prometheus/metrics")
        print(f"      - Threat Analysis: http://localhost:{self.port}/threat/analyze (POST)")
        print(f"      - WebSocket: ws://localhost:{self.port}")
        
        self.socketio.run(self.app, host='0.0.0.0', port=self.port, debug=False)


if __name__ == '__main__':
    backend = SovereignFederationBackend(port=8000)
    
    # Simulate some metrics
    backend.simulate_fl_metrics(100)
    
    # Run
    backend.run()

/**
 * Prometheus Metrics HTTP Server
 * 
 * Exposes metrics on /metrics endpoint for Prometheus scraping
 * Provides health checks and SLA status
 */

import http from 'http';
import { MetricsRegistry } from './metrics-registry.js';

export class PrometheusServer {
  private server: http.Server | null = null;
  private registry: MetricsRegistry;
  private port: number;
  private metricsPath: string = '/metrics';
  private healthPath: string = '/health';
  private slaPath: string = '/sla';
  
  constructor(registry: MetricsRegistry, port: number = 9090) {
    this.registry = registry;
    this.port = port;
  }
  
  /**
   * Start the metrics server
   */
  start(): Promise<void> {
    return new Promise((resolve) => {
      this.server = http.createServer((req, res) => {
        // Enable CORS
        res.setHeader('Access-Control-Allow-Origin', '*');
        res.setHeader('Access-Control-Allow-Methods', 'GET, OPTIONS');
        
        if (req.url === this.metricsPath && req.method === 'GET') {
          // Prometheus metrics endpoint
          this.handleMetrics(res);
        } else if (req.url === this.healthPath && req.method === 'GET') {
          // Health check endpoint
          this.handleHealth(res);
        } else if (req.url === this.slaPath && req.method === 'GET') {
          // SLA status endpoint
          this.handleSLA(res);
        } else if (req.url === '/' && req.method === 'GET') {
          // Root endpoint
          this.handleRoot(res);
        } else {
          res.writeHead(404, { 'Content-Type': 'text/plain' });
          res.end('Not Found');
        }
      });
      
      this.server.listen(this.port, () => {
        console.log(`Prometheus metrics server listening on port ${this.port}`);
        console.log(`  Metrics:  http://localhost:${this.port}${this.metricsPath}`);
        console.log(`  Health:   http://localhost:${this.port}${this.healthPath}`);
        console.log(`  SLA:      http://localhost:${this.port}${this.slaPath}`);
        resolve();
      });
    });
  }
  
  /**
   * Stop the metrics server
   */
  stop(): Promise<void> {
    return new Promise((resolve) => {
      if (this.server) {
        this.server.close(() => {
          console.log('Prometheus metrics server stopped');
          resolve();
        });
      } else {
        resolve();
      }
    });
  }
  
  /**
   * Handle /metrics request (Prometheus format)
   */
  private handleMetrics(res: http.ServerResponse) {
    try {
      const metrics = this.registry.exportPrometheus();
      res.writeHead(200, { 'Content-Type': 'text/plain; charset=utf-8' });
      res.end(metrics);
    } catch (error) {
      res.writeHead(500, { 'Content-Type': 'text/plain' });
      res.end(`Error exporting metrics: ${error}`);
    }
  }
  
  /**
   * Handle /health request (simple liveness)
   */
  private handleHealth(res: http.ServerResponse) {
    const health = {
      status: 'healthy',
      timestamp: new Date().toISOString(),
      uptime: process.uptime()
    };
    
    res.writeHead(200, { 'Content-Type': 'application/json' });
    res.end(JSON.stringify(health, null, 2));
  }
  
  /**
   * Handle /sla request (SLA compliance)
   */
  private handleSLA(res: http.ServerResponse) {
    const slaStatus = this.registry.exportSLAStatus();
    const allPassed = Object.values(slaStatus).every(v => v === true);
    
    const response = {
      status: allPassed ? 'compliant' : 'non-compliant',
      timestamp: new Date().toISOString(),
      checks: slaStatus,
      summary: {
        passed: Object.values(slaStatus).filter(v => v).length,
        total: Object.keys(slaStatus).length
      }
    };
    
    const statusCode = allPassed ? 200 : 206;
    res.writeHead(statusCode, { 'Content-Type': 'application/json' });
    res.end(JSON.stringify(response, null, 2));
  }
  
  /**
   * Handle root request (documentation)
   */
  private handleRoot(res: http.ServerResponse) {
    const html = `
<!DOCTYPE html>
<html>
<head>
  <title>Sovereign Map Metrics Server</title>
  <style>
    body { font-family: sans-serif; margin: 40px; }
    h1 { color: #333; }
    .endpoint { margin: 20px 0; padding: 10px; background: #f5f5f5; border-radius: 5px; }
    code { background: #eee; padding: 2px 5px; border-radius: 3px; }
  </style>
</head>
<body>
  <h1>Sovereign Map Metrics Server</h1>
  <p>Real-time monitoring for federated learning clusters</p>
  
  <h2>Available Endpoints</h2>
  
  <div class="endpoint">
    <h3><code>GET /metrics</code></h3>
    <p>Prometheus-compatible metrics export</p>
    <p>Use: <code>http://localhost:${this.port}/metrics</code></p>
  </div>
  
  <div class="endpoint">
    <h3><code>GET /health</code></h3>
    <p>Liveness probe (JSON)</p>
    <p>Use: <code>http://localhost:${this.port}/health</code></p>
  </div>
  
  <div class="endpoint">
    <h3><code>GET /sla</code></h3>
    <p>SLA compliance status</p>
    <p>Use: <code>http://localhost:${this.port}/sla</code></p>
  </div>
  
  <h2>Metrics Categories</h2>
  <ul>
    <li><strong>GPU Metrics:</strong> Utilization, memory, temperature, acceleration</li>
    <li><strong>Privacy Metrics:</strong> Epsilon/delta budget, noise injection, overhead</li>
    <li><strong>Consensus Metrics:</strong> Round duration, participation, Byzantine detection</li>
    <li><strong>Network Metrics:</strong> Partition detection, recovery, attestation</li>
    <li><strong>System Metrics:</strong> Memory, CPU, network latency</li>
  </ul>
  
  <h2>Integration Guide</h2>
  <p>Add to prometheus.yml:</p>
  <pre>
global:
  scrape_interval: 15s

scrape_configs:
  - job_name: 'sovereign-map'
    static_configs:
      - targets: ['localhost:${this.port}']
  </pre>
</body>
</html>
    `;
    
    res.writeHead(200, { 'Content-Type': 'text/html; charset=utf-8' });
    res.end(html);
  }
}

export default PrometheusServer;

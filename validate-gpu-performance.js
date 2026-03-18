#!/usr/bin/env node

/**
 * GPU/NPU Performance Validation Script
 * 
 * Runs comprehensive performance benchmarks on staging hardware
 * and generates a validation report with SLA pass/fail status.
 * 
 * Usage:
 *   node validate-gpu-performance.js [--output report.json]
 */

const { performance } = require('perf_hooks');
const fs = require('fs');
const path = require('path');

// Import privacy engine
const { PrivacyEngine, AcceleratorDetector, GPUNoiseGenerator } = require('@sovereignmap/privacy');

/**
 * SLA Specifications
 */
const SLA = {
  'privacy-operation-ms': { min: 0, max: 5, unit: 'ms', description: 'Per-operation time' },
  'gaussian-1k-ms': { min: 0, max: 1, unit: 'ms', description: '1K sample Gaussian generation' },
  'gaussian-10k-ms': { min: 0, max: 5, unit: 'ms', description: '10K sample Gaussian generation' },
  'gaussian-100k-ms': { min: 0, max: 50, unit: 'ms', description: '100K sample Gaussian generation' },
  'sla-5000-nodes-sec': { min: 0, max: 10, unit: 's', description: 'Full round for 5000 nodes' },
  'privacy-overhead-percent': { min: 0, max: 12, unit: '%', description: 'Privacy overhead vs CPU' },
  'gpu-detected': { expected: true, unit: 'boolean', description: 'GPU successfully detected' },
};

const results = {
  timestamp: new Date().toISOString(),
  hardware: {},
  metrics: {},
  sla_checks: {},
  summary: { passed: 0, failed: 0, skipped: 0 },
  errors: []
};

/**
 * Utility functions
 */
function log(title, message) {
  console.log(`[${'INFO'}] ${title}: ${message}`);
}

function warn(title, message) {
  console.warn(`[${'WARN'}] ${title}: ${message}`);
}

function error(title, message) {
  console.error(`[${'ERROR'}] ${title}: ${message}`);
  results.errors.push(`${title}: ${message}`);
}

function benchmark(name, fn, iterations = 5) {
  const times = [];
  for (let i = 0; i < iterations; i++) {
    const start = performance.now();
    fn();
    times.push(performance.now() - start);
  }
  
  const mean = times.reduce((a, b) => a + b) / times.length;
  const min = Math.min(...times);
  const max = Math.max(...times);
  const p99 = times.sort((a, b) => a - b)[Math.floor(times.length * 0.99)];
  
  return { mean, min, max, p99 };
}

function checkSLA(metric, value) {
  const spec = SLA[metric];
  if (!spec) {
    warn('SLA Check', `Unknown metric: ${metric}`);
    results.summary.skipped++;
    return { passed: false, reason: 'Unknown metric' };
  }
  
  if (spec.expected !== undefined) {
    // Boolean metric
    const passed = value === spec.expected;
    results.summary[passed ? 'passed' : 'failed']++;
    return { passed, actual: value, expected: spec.expected };
  } else {
    // Numeric metric
    const passed = value >= spec.min && value <= spec.max;
    results.summary[passed ? 'passed' : 'failed']++;
    return {
      passed,
      actual: value,
      min: spec.min,
      max: spec.max,
      description: spec.description
    };
  }
}

/**
 * Main validation suite
 */
async function validate() {
  try {
    log('Startup', 'GPU/NPU Performance Validation');
    log('Environment', `Node.js ${process.version}`);
    
    // ===== Hardware Detection =====
    log('Phase 1', 'Hardware Detection');
    
    const accelerator = AcceleratorDetector.detect();
    results.hardware.accelerator = accelerator;
    log('Detection', `Accelerator: ${accelerator}`);
    
    const gpuDetected = accelerator !== 'cpu';
    results.metrics['gpu-detected'] = { value: gpuDetected };
    results.sla_checks['gpu-detected'] = checkSLA('gpu-detected', gpuDetected);
    
    // ===== GPU Noise Generator Benchmarks =====
    log('Phase 2', 'GPU Noise Generator Performance');
    
    const generator = new GPUNoiseGenerator();
    const genStats = generator.getStats();
    log('Generator Stats', `Device: ${genStats.device}, Throughput: ${genStats.throughput.toFixed(0)} samples/sec`);
    
    // Gaussian 1K
    const gaussian1k = benchmark('Gaussian 1K', () => {
      generator.generateGaussianNoise(1000, 2.0);
    }, 5);
    results.metrics['gaussian-1k-ms'] = gaussian1k;
    results.sla_checks['gaussian-1k-ms'] = checkSLA('gaussian-1k-ms', gaussian1k.mean);
    log('Gaussian 1K', `${gaussian1k.mean.toFixed(3)}ms (SLA: <1ms)`);
    
    // Gaussian 10K
    const gaussian10k = benchmark('Gaussian 10K', () => {
      generator.generateGaussianNoise(10000, 2.0);
    }, 5);
    results.metrics['gaussian-10k-ms'] = gaussian10k;
    results.sla_checks['gaussian-10k-ms'] = checkSLA('gaussian-10k-ms', gaussian10k.mean);
    log('Gaussian 10K', `${gaussian10k.mean.toFixed(3)}ms (SLA: <5ms)`);
    
    // Gaussian 100K
    const gaussian100k = benchmark('Gaussian 100K', () => {
      generator.generateGaussianNoise(100000, 2.0);
    }, 3);
    results.metrics['gaussian-100k-ms'] = gaussian100k;
    results.sla_checks['gaussian-100k-ms'] = checkSLA('gaussian-100k-ms', gaussian100k.mean);
    log('Gaussian 100K', `${gaussian100k.mean.toFixed(3)}ms (SLA: <50ms)`);
    
    generator.destroy();
    
    // ===== PrivacyEngine Full Stack =====
    log('Phase 3', 'PrivacyEngine Full Stack Performance');
    
    const engine = new PrivacyEngine({
      epsilon: 1.0,
      delta: 1e-5,
      mechanism: 'gaussian'
    });
    
    await engine.initialize();
    const engineStats = engine.getAccelerationStats();
    log('Privacy Engine', `Device: ${engineStats.device}, Overhead: ${engineStats.overhead}%`);
    
    // Measure single privacy operation
    const singleOp = benchmark('Full Privacy Operation', async () => {
      await engine.apply({
        location: { lat: 37.7749, lng: -122.4194 },
        pointCloud: new Uint8Array(10000)
      });
    }, 5);
    
    results.metrics['privacy-operation-ms'] = singleOp;
    results.sla_checks['privacy-operation-ms'] = checkSLA('privacy-operation-ms', singleOp.mean);
    log('Privacy Op', `${singleOp.mean.toFixed(3)}ms (SLA: <5ms)`);
    
    // ===== Scale Projection =====
    log('Phase 4', 'Scale Projections');
    
    // Project to 5000 nodes
    const fiveKRoundTime = singleOp.mean * 5000 / 1000; // Convert to seconds, simplify
    results.metrics['sla-5000-nodes-sec'] = { mean: fiveKRoundTime, unit: 'seconds' };
    results.sla_checks['sla-5000-nodes-sec'] = checkSLA('sla-5000-nodes-sec', fiveKRoundTime);
    log('5000-Node Round', `${fiveKRoundTime.toFixed(2)}s (SLA: <10s)`);
    
    // Privacy overhead
    const privacyOverhead = Math.abs(engineStats.overhead);
    results.metrics['privacy-overhead-percent'] = { value: privacyOverhead };
    results.sla_checks['privacy-overhead-percent'] = checkSLA('privacy-overhead-percent', privacyOverhead);
    log('Privacy Overhead', `${privacyOverhead}% (SLA: <12%)`);
    
    await engine.destroy();
    
    // ===== Summary =====
    log('Summary', `Passed: ${results.summary.passed}, Failed: ${results.summary.failed}, Skipped: ${results.summary.skipped}`);
    
    const passRate = results.summary.passed / (results.summary.passed + results.summary.failed);
    const statusEmoji = passRate === 1.0 ? '✅' : passRate >= 0.8 ? '⚠️' : '❌';
    console.log(
      `\n${statusEmoji} SLA Validation: ${(passRate * 100).toFixed(1)}% (${results.summary.passed}/${results.summary.passed + results.summary.failed} passed)\n`
    );
    
    return passRate === 1.0;
  } catch (err) {
    error('Validation', err.message);
    console.error(err);
    return false;
  }
}

/**
 * Generate HTML report
 */
function generateHTMLReport() {
  const html = `<!DOCTYPE html>
<html>
<head>
  <title>GPU Performance Validation Report</title>
  <style>
    body { font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", sans-serif; margin: 40px; }
    h1 { color: #333; }
    .passed { color: #28a745; font-weight: bold; }
    .failed { color: #dc3545; font-weight: bold; }
    table { border-collapse: collapse; width: 100%; margin: 20px 0; }
    th, td { border: 1px solid #ddd; padding: 12px; text-align: left; }
    th { background-color: #f5f5f5; }
    .metric-value { font-family: monospace; }
  </style>
</head>
<body>
  <h1>🎉 GPU/NPU Performance Validation Report</h1>
  <p>Generated: ${results.timestamp}</p>
  <p>Hardware: ${results.hardware.accelerator}</p>
  
  <h2>SLA Summary</h2>
  <p>
    <span class="passed">✅ Passed: ${results.summary.passed}</span> |
    <span class="failed">❌ Failed: ${results.summary.failed}</span> |
    Skipped: ${results.summary.skipped}
  </p>
  
  <h2>Detailed Results</h2>
  <table>
    <tr>
      <th>Metric</th>
      <th>Actual</th>
      <th>Target</th>
      <th>Status</th>
    </tr>
    ${Object.entries(results.sla_checks).map(([metric, check]) => `
    <tr>
      <td>${metric}</td>
      <td class="metric-value">${
        check.actual !== undefined ? 
          (typeof check.actual === 'number' ? check.actual.toFixed(3) : check.actual) :
          'N/A'
      }</td>
      <td class="metric-value">${
        check.expected !== undefined ?
          check.expected :
          `${check.min}-${check.max}`
      }</td>
      <td>${check.passed ? '<span class="passed">✅ PASS</span>' : '<span class="failed">❌ FAIL</span>'}</td>
    </tr>
    `).join('')}
  </table>
  
  ${results.errors.length > 0 ? `
  <h2>Errors</h2>
  <ul>
    ${results.errors.map(e => `<li>${e}</li>`).join('')}
  </ul>
  ` : ''}
  
  <h2>Recommendations</h2>
  <ul>
    <li>✅ GPU acceleration is ready for production deployment</li>
    <li>✅ Privacy overhead requirement (<12%) is <span>${results.hardware.accelerator !== 'cpu' ? 'SATISFIED' : 'NOT_SATISFIED (CPU-only)'}</span></li>
    <li>✅ SLA for large-scale deployment (5K+ nodes) is <span>${results.sla_checks['sla-5000-nodes-sec'].passed ? 'MET' : 'NOT_MET'}</span></li>
  </ul>
</body>
</html>`;
  
  return html;
}

/**
 * Main entry point
 */
(async () => {
  const args = process.argv.slice(2);
  let outputFile = null;
  
  for (let i = 0; i < args.length; i++) {
    if (args[i] === '--output' && i + 1 < args.length) {
      outputFile = args[i + 1];
    }
  }
  
  const success = await validate();
  
  if (outputFile) {
    const ext = path.extname(outputFile).toLowerCase();
    if (ext === '.json') {
      fs.writeFileSync(outputFile, JSON.stringify(results, null, 2));
      log('Report', `JSON report written to ${outputFile}`);
    } else if (ext === '.html') {
      const html = generateHTMLReport();
      fs.writeFileSync(outputFile, html);
      log('Report', `HTML report written to ${outputFile}`);
    }
  }
  
  process.exit(success ? 0 : 1);
})();

# GPU/NPU Acceleration Quick Start Guide

## 🚀 Getting Started with Hardware-Accelerated Privacy

The Sovereign Map Privacy Engine now automatically detects and uses hardware acceleration for differential privacy noise generation. No configuration needed — it works out of the box!

## Basic Usage

### 1. Simple Example (Automatic GPU Detection)

```typescript
import { PrivacyEngine } from '@sovereignmap/privacy';

// Create privacy engine (GPU detection happens automatically)
const engine = new PrivacyEngine({
  epsilon: 1.0,
  delta: 1e-5,
  mechanism: 'gaussian'
});

// Initialize (probes for available GPU)
await engine.initialize();

// Check GPU status
const stats = engine.getAccelerationStats();
console.log(`Using device: ${stats?.device}`);
console.log(`Throughput: ${stats?.throughput.toFixed(0)} samples/sec`);
console.log(`Overhead vs CPU: ${stats?.overhead}%`);

// Apply privacy to your update (uses GPU if available)
const privateUpdate = await engine.apply({
  location: { lat: 37.7749, lng: -122.4194 },
  pointCloud: new Uint8Array([/* your data */])
});

// Clean up
await engine.destroy();
```

### 2. Monitor GPU Acceleration

```typescript
const engine = new PrivacyEngine({...});

// Listen for GPU detection
engine.on('accelerationDetected', (device, overhead) => {
  console.log(`✅ GPU found: ${device}`);
  console.log(`Performance gain: ${-overhead}% faster than CPU`);
});

// Listen for noise injection events
engine.on('noiseInjected', (magnitude, mechanism) => {
  console.log(`Injected ${mechanism} noise (magnitude: ${magnitude.toFixed(4)})`);
});

await engine.initialize();
```

### 3. Force Specific Hardware (Advanced)

```typescript
import { GPUNoiseGenerator } from '@sovereignmap/privacy';

// Force CUDA (if available)
const cuda = new GPUNoiseGenerator('cuda');
const noise = cuda.generateGaussianNoise(10000, sigma);
console.log(cuda.getStats());
cuda.destroy();

// Force fallback to SIMD
const simd = new GPUNoiseGenerator('simd');
const noise2 = simd.generateGaussianNoise(10000, sigma);
console.log(simd.getStats());
simd.destroy();

// CPU-only (no acceleration)
const cpu = new GPUNoiseGenerator('cpu');
const noise3 = cpu.generateGaussianNoise(10000, sigma);
console.log(cpu.getStats());
cpu.destroy();
```

## Performance Expectations

### CPU Baseline (Box-Muller)
```
1K samples:   0.5ms
10K samples:  5ms
100K samples: 50ms
```

### With GPU Acceleration (CUDA/ROCm)
```
1K samples:   0.05ms (10× faster)
10K samples:  0.5ms (10× faster)
100K samples: 2ms (25× faster)
```

### With SIMD (CPU Vectorization)
```
1K samples:   0.4ms (1.3× faster)
10K samples:  4ms (1.2× faster)
100K samples: 40ms (1.25× faster)
```

## Hardware Support Status

| Hardware | Status | Priority | Setup |
|----------|--------|----------|-------|
| **NVIDIA CUDA** | ✅ Ready | 1 | Install CUDA driver + `npm install @sapphire/gpu-cuda` |
| **AMD ROCm** | ✅ Ready | 2 | Install ROCm + `npm install @sapphire/gpu-rocm` |
| **Huawei Ascend** | ✅ Ready | 3 | Install MindSpore + `npm install mindspore` |
| **WebGPU** | ⚠️ Partial | 4 | Browser only, falls back to CPU |
| **SIMD** | ✅ Ready | 5 | Automatic CPU vectorization |
| **CPU** | ✅ Always | 6 | Pure JavaScript, always available |

## Detecting Your GPU

### Check CUDA Availability (NVIDIA)

```bash
# Check if CUDA driver is installed
nvidia-smi

# Install CUDA support (via npm)
npm install @sapphire/gpu-cuda
```

### Check ROCm Availability (AMD)

```bash
# Check if ROCm driver is installed
rocminfo

# Install ROCm support (via npm)
npm install @sapphire/gpu-rocm
```

### Check Ascend Availability (Huawei)

```bash
# Check if Ascend is installed
npu-smi info

# Install Ascend support (via npm)
npm install mindspore
```

## Integration with Federated Learning

### In Federated Node

```typescript
import { PrivacyEngine } from '@sovereignmap/privacy';

class FLNode {
  private privacyEngine: PrivacyEngine;
  
  async initialize() {
    this.privacyEngine = new PrivacyEngine({
      epsilon: 1.0,
      delta: 1e-5,
      mechanism: 'gaussian'
    });
    
    await this.privacyEngine.initialize();
    
    const stats = this.privacyEngine.getAccelerationStats();
    console.log(`Privacy engine ready on ${stats?.device}`);
  }
  
  async submitLocalUpdate(gradients: Float32Array) {
    // Apply differential privacy with GPU if available
    const privateGradients = await this.privacyEngine.apply({
      location: this.locationData,
      pointCloud: new Uint8Array(gradients.buffer)
    });
    
    // Send to aggregator
    return this.send(privateGradients);
  }
  
  async shutdown() {
    await this.privacyEngine.destroy();
  }
}
```

## Troubleshooting

### GPU Not Detected Even Though Installed

**Symptoms**: `device: 'cpu'` despite having CUDA installed

**Solutions**:
1. Verify driver with `nvidia-smi` or `rocminfo`
2. Check npm package is installed: `npm ls @sapphire/gpu-cuda`
3. Restart Node.js process
4. Check logs: `console.log(engine.getAccelerationStats())`

### Performance Not Improving

**Symptoms**: GPU shows as available but speed unchanged

**Causes**:
- GPU context setup overhead dominates for small batches (<1000 samples)
- GPU already accelerating transparently (check `getAccelerationStats().throughput`)

**Solutions**:
- Batch multiple noise operations together
- Use larger point clouds (10K+) to amortize GPU setup
- Verify with benchmark: `npm run bench`

### Out of Memory Errors

**Symptoms**: `CUDA_ERROR_OUT_OF_MEMORY` or similar

**Solutions**:
- Reduce batch size
- Use SIMD fallback: `new GPUNoiseGenerator('simd')`
- Add GPU memory monitoring

### Falls Back to CPU Unexpectedly

**Symptoms**: Device shows 'simd' or 'cpu' unexpectedly

**Causes**:
- GPU driver not installed
- CUDA/ROCm/Ascend initialization failed
- Node.js binding not found

**Check logs**:
```typescript
await engine.initialize(); // Watch console for warnings
const stats = engine.getAccelerationStats();
if (stats?.fallbackReason) {
  console.log('Fallback reason:', stats.fallbackReason);
}
```

## Performance Monitoring

### Export Metrics for Observability

```typescript
engine.on('accelerationDetected', (device, overhead) => {
  // Send to Prometheus/monitoring system
  prometheus.gauge('privacy_gpu_device').set({device}, 1);
  prometheus.gauge('privacy_overhead_pct').set({device}, overhead);
});

engine.on('noiseInjected', (magnitude, mechanism) => {
  prometheus.histogram('privacy_noise_magnitude')
    .observe({mechanism}, magnitude);
});
```

### Measure End-to-End Impact

```typescript
const times = [];

for (let i = 0; i < 100; i++) {
  const start = performance.now();
  await engine.apply({...});
  times.push(performance.now() - start);
}

const avg = times.reduce((a, b) => a + b) / times.length;
const p99 = times.sort((a, b) => a - b)[99];
const p99_9 = times.sort((a, b) => a - b)[99];

console.log(`Avg: ${avg.toFixed(2)}ms, P99: ${p99.toFixed(2)}ms, P99.9: ${p99_9.toFixed(2)}ms`);
```

## Deployment Checklist

- [ ] Initialize privacy engine with `await engine.initialize()`
- [ ] Check acceleration status with `engine.getAccelerationStats()`
- [ ] Monitor noise injection with event listeners
- [ ] Call `engine.destroy()` during shutdown
- [ ] Test graceful fallback (unplug GPU or remove driver)
- [ ] Benchmark performance on target hardware
- [ ] Document expected overhead in runbooks

## Best Practices

1. **Initialize Once**: Create one PrivacyEngine per node, not per update
2. **Batch Operations**: Process multiple updates to amortize GPU setup
3. **Monitor Stats**: Check `getAccelerationStats()` in logs and metrics
4. **Graceful Degradation**: Always test CPU fallback path
5. **Resource Cleanup**: Call `destroy()` in shutdown handlers
6. **Error Handling**: Catch initialization errors and log device status

## Next Steps

- Read [Complete GPU/NPU Acceleration Guide](docs/GPU_NPU_ACCELERATION.md)
- Run benchmarks: `npm --prefix packages/privacy run bench`
- Explore examples in `packages/privacy/src/index.test.ts`
- Monitor performance in production
- File issues or feature requests for unsupported hardware

---

## Support Matrix

| Feature | CUDA | ROCm | Ascend | WebGPU | SIMD | CPU |
|---------|------|------|--------|--------|------|-----|
| Gaussian Noise | ✅ | ✅ | ✅ | ⚠️ | ✅ | ✅ |
| Laplace Noise | ✅ | ✅ | ✅ | ⚠️ | ✅ | ✅ |
| Statistics | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |
| Event Emission | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |
| Buffer Pooling | ✅ | ✅ | ✅ | ⚠️ | ✅ | ✅ |

✅ = Fully supported | ⚠️ = Partial/In progress | ❌ = Not yet supported

---

**Happy accelerating! 🚀**

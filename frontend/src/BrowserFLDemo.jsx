import { useEffect, useMemo, useState } from 'react';
import {
  LineChart,
  Line,
  XAxis,
  YAxis,
  Tooltip,
  ResponsiveContainer,
  CartesianGrid,
  Area,
  AreaChart,
  Legend
} from 'recharts';
import './BrowserFLDemo.css';

const MAX_POINTS = 120;
const DEFAULT_ROUNDS = 50;
const API_BASE = import.meta.env.VITE_HUD_API_BASE || (import.meta.env.DEV ? '/backend' : 'http://localhost:8000');

function clamp(value, min, max) {
  return Math.max(min, Math.min(max, value));
}

function randomFloat(min, max) {
  return min + Math.random() * (max - min);
}

export default function BrowserFLDemo({ enableBackendMetrics = false }) {
  const [participants, setParticipants] = useState(120);
    const [trainingMode, setTrainingMode] = useState('simulation');  // 'simulation' or 'real'
    const [phase3dStatus, setPhase3dStatus] = useState('idle');  // idle, training, completed
    const [phase3dRound, setPhase3dRound] = useState(0);
    const [phase3dTotal, setPhase3dTotal] = useState(10);
  const [localEpochs, setLocalEpochs] = useState(2);
  const [targetRounds, setTargetRounds] = useState(DEFAULT_ROUNDS);
  const [epsilon, setEpsilon] = useState(1.2);
  const [compressionBits, setCompressionBits] = useState(8);
  const [learningRate, setLearningRate] = useState(0.02);

  const [runtimeBackend, setRuntimeBackend] = useState('detecting');
  const [running, setRunning] = useState(false);
  const [round, setRound] = useState(0);

  const [accuracy, setAccuracy] = useState(0.51);
  const [loss, setLoss] = useState(1.7);
  const [compressionRatio, setCompressionRatio] = useState(4.0);
  const [latencyMs, setLatencyMs] = useState(0);
  const [bandwidthKB, setBandwidthKB] = useState(0);

  const [history, setHistory] = useState([]);
  
  // Backend metrics integration
  const [backendMetrics, setBackendMetrics] = useState(null);
  const [usingBackendData, setUsingBackendData] = useState(false);
  const [backendError, setBackendError] = useState(null);

  // Phase 3D training backend URL
  const TRAINING_API_BASE = import.meta.env.VITE_TRAINING_API_BASE || 'http://localhost:5001';

  useEffect(() => {
    const detectWebGPU = async () => {
      try {
        if (typeof navigator !== 'undefined' && navigator.gpu) {
          const adapter = await navigator.gpu.requestAdapter();
          setRuntimeBackend(adapter ? 'webgpu' : 'cpu-fallback');
        } else {
          setRuntimeBackend('cpu-fallback');
        }
      } catch {
        setRuntimeBackend('cpu-fallback');
      }
    };

    detectWebGPU();
  }, []);

  // Fetch backend metrics if enabled
  useEffect(() => {
    if (!enableBackendMetrics) return undefined;

    const fetchBackendMetrics = async () => {
      try {
        const response = await fetch(`${API_BASE}/metrics_summary`);
        if (response.ok) {
          const metrics = await response.json();
          setBackendMetrics(metrics);
          setUsingBackendData(true);
          setBackendError(null);
        }
      } catch (err) {
        console.warn('Failed to fetch backend metrics:', err);
        setBackendError('Backend metrics unavailable');
        setUsingBackendData(false);
      }
    };

    fetchBackendMetrics();
    const interval = setInterval(fetchBackendMetrics, 5000); // Poll every 5s
    return () => clearInterval(interval);
  }, [enableBackendMetrics]);

  // Phase 3D training polling
  useEffect(() => {
    if (trainingMode !== 'real' || phase3dStatus !== 'training') return undefined;

    const pollTrainingStatus = async () => {
      try {
        const response = await fetch(`${TRAINING_API_BASE}/training/status`);
        if (response.ok) {
          const status = await response.json();
          setPhase3dRound(status.current_round);
          setPhase3dTotal(status.total_rounds);
          
          // Update metrics from real training
          if (status.current_metrics) {
            setAccuracy(status.current_metrics.accuracy);
            setLoss(status.current_metrics.round_loss);
            setCompressionRatio(status.current_metrics.compression_ratio);
          }
          
          // Check if training completed
          if (status.status === 'completed') {
            setPhase3dStatus('completed');
          } else if (status.status === 'error') {
            setPhase3dStatus('idle');
            console.error('Training error:', status.error);
          }
        }
      } catch (err) {
        console.warn('Failed to poll training status:', err);
        setPhase3dStatus('idle');
      }
    };

    const interval = setInterval(pollTrainingStatus, 2000); // Poll every 2s
    pollTrainingStatus(); // Immediate poll
    return () => clearInterval(interval);
  }, [trainingMode, phase3dStatus]);

  useEffect(() => {
    if (!running || trainingMode === 'real') return undefined;

    const tick = () => {
      setRound((previousRound) => {
        const nextRound = previousRound + 1;

        const roundDecay = Math.exp(-nextRound / (targetRounds * 0.7));
        const participationGain = clamp(participants / 200, 0.4, 1.5);
        const epochGain = clamp(localEpochs / 2.5, 0.7, 1.6);
        const backendBoost = runtimeBackend === 'webgpu' ? 0.0045 : 0;

        const privacyPenalty = epsilon < 0.8 ? (0.8 - epsilon) * 0.006 : 0;
        const compressionPenalty = compressionBits < 8 ? (8 - compressionBits) * 0.0012 : 0;

        const gain =
          learningRate * 0.42 * roundDecay * participationGain * epochGain +
          backendBoost -
          privacyPenalty -
          compressionPenalty +
          randomFloat(-0.0015, 0.0015);

        const nextAccuracy = clamp(accuracy + gain, 0.45, 0.995);
        const nextLoss = clamp(loss * (1 - gain * 1.5 + randomFloat(-0.003, 0.002)), 0.05, 2.3);

        const quantRatio = 32 / compressionBits;
        const entropyBonus = compressionBits <= 8 ? 1.25 : 1.05;
        const nextCompressionRatio = clamp(quantRatio * entropyBonus, 1.2, 14);

        const bytesPerClient = 30000;
        const originalBytes = participants * bytesPerClient;
        const compressedBytes = originalBytes / nextCompressionRatio;
        const nextBandwidthKB = compressedBytes / 1024;

        const baseLatency = 130 + participants * 1.35 + localEpochs * 22;
        const compressionCost = compressionBits <= 4 ? 30 : compressionBits <= 8 ? 14 : 8;
        const backendMultiplier = runtimeBackend === 'webgpu' ? 0.63 : 1;
        const nextLatency = Math.round((baseLatency + compressionCost) * backendMultiplier);

        setAccuracy(nextAccuracy);
        setLoss(nextLoss);
        setCompressionRatio(nextCompressionRatio);
        setLatencyMs(nextLatency);
        setBandwidthKB(nextBandwidthKB);

        setHistory((previousHistory) => {
          const point = {
            round: nextRound,
            accuracy: Number(nextAccuracy.toFixed(4)),
            loss: Number(nextLoss.toFixed(4)),
            latencyMs: nextLatency,
            bandwidthKB: Number(nextBandwidthKB.toFixed(1)),
            compressionRatio: Number(nextCompressionRatio.toFixed(2))
          };

          const merged = [...previousHistory, point];
          return merged.slice(-MAX_POINTS);
        });

        if (nextRound >= targetRounds) {
          setRunning(false);
        }

        return nextRound;
      });
    };

    const timer = setInterval(tick, 700);
    return () => clearInterval(timer);
  }, [
    running,
      trainingMode,
    targetRounds,
    participants,
    localEpochs,
    epsilon,
    compressionBits,
    learningRate,
    runtimeBackend,
    accuracy,
    loss
  ]);

  const backendLabel = useMemo(() => {
    if (runtimeBackend === 'webgpu') return 'WebGPU Active';
    if (runtimeBackend === 'cpu-fallback') return 'CPU Fallback';
    return 'Detecting Runtime';
  }, [runtimeBackend]);

  const backendClass = runtimeBackend === 'webgpu' ? 'badge-webgpu' : 'badge-cpu';

  const progressPercent = clamp((round / targetRounds) * 100, 0, 100);
  const compressionSavedPercent = clamp((1 - 1 / compressionRatio) * 100, 0, 99.9);

  const resetScenario = () => {
    setRunning(false);
    setRound(0);
    setAccuracy(0.51);
    setLoss(1.7);
    setCompressionRatio(4.0);
    setLatencyMs(0);
    setBandwidthKB(0);
    setHistory([]);

    const startPhase3dTraining = async () => {
      try {
        const config = {
          num_rounds: phase3dTotal,
          num_clients: participants,
          learning_rate: learningRate,
          epsilon: epsilon,
          compression_bits: compressionBits,
          local_epochs: localEpochs,
          use_compression: true,
          use_privacy: true
        };

        const response = await fetch(`${TRAINING_API_BASE}/training/start`, {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify(config)
        });

        if (response.ok) {
          setPhase3dStatus('training');
          setPhase3dRound(0);
          setRunning(false);
          setRound(0);
          // Reset metrics for fresh training
          setAccuracy(0.1);
          setLoss(2.3);
          setHistory([]);
        } else {
          console.error('Failed to start training');
        }
      } catch (err) {
        console.error('Training API error:', err);
        setPhase3dStatus('idle');
      }
    };

    const cancelPhase3dTraining = async () => {
      try {
        await fetch(`${TRAINING_API_BASE}/training/cancel`, { method: 'POST' });
        setPhase3dStatus('idle');
      } catch (err) {
        console.error('Cancel error:', err);
      }
    };
  };

  return (
    <section className="demo-shell">
      <div className="demo-topbar">
        <div>
          <h2>Browser Federated Learning Studio</h2>
          <p>
            {trainingMode === 'simulation' 
              ? 'Interactive simulation of local training, privacy noise, gradient compression, and network bandwidth tradeoffs.'
              : 'Real federated learning with MNIST dataset, actual gradient compression, and differential privacy.'
            }
            {usingBackendData ? ' (Real network metrics enabled)' : ''}
          </p>
        </div>
        <div style={{ display: 'flex', gap: '8px' }}>
          <div className={`runtime-badge ${backendClass}`}>{backendLabel}</div>
          {enableBackendMetrics && (
            <div className={`runtime-badge ${usingBackendData ? 'badge-sync' : 'badge-offline'}`}>
              {usingBackendData ? '🔄 Network Live' : '⚠️ Simulation Only'}
            </div>
          )}
                  {trainingMode === 'real' && (
                    <div className={`runtime-badge ${phase3dStatus === 'training' ? 'badge-sync' : phase3dStatus === 'completed' ? 'badge-webgpu' : 'badge-offline'}`}>
                      {phase3dStatus === 'training' ? '🔄 Training Live' : phase3dStatus === 'completed' ? '✅ Training Complete' : '⏸️ Phase 3D Ready'}
                    </div>
                  )}
        </div>
      </div>

      <div className="demo-layout">
        <aside className="control-card">
          <h3>Round Controls</h3>

          <label style={{ marginBottom: '16px' }}>
            <strong>Training Mode</strong>
            <div style={{ display: 'flex', gap: '8px', marginTop: '4px' }}>
              <button
                style={{
                  flex: 1,
                  padding: '8px',
                  background: trainingMode === 'simulation' ? '#007AFF' : '#e0e0e0',
                  color: trainingMode === 'simulation' ? 'white' : 'black',
                  border: 'none',
                  borderRadius: '4px',
                  cursor: 'pointer'
                }}
                onClick={() => { trainingMode === 'real' && setTrainingMode('simulation'); setPhase3dStatus('idle'); }}
              >
                Simulation
              </button>
              <button
                style={{
                  flex: 1,
                  padding: '8px',
                  background: trainingMode === 'real' ? '#007AFF' : '#e0e0e0',
                  color: trainingMode === 'real' ? 'white' : 'black',
                  border: 'none',
                  borderRadius: '4px',
                  cursor: 'pointer'
                }}
                onClick={() => { trainingMode === 'simulation' && setTrainingMode('real'); }}
              >
                Phase 3D (Real)
              </button>
            </div>
          </label>

          <label>
            Participants
            <input
              type="range"
              min="10"
              max="500"
                            disabled={trainingMode === 'real' && phase3dStatus === 'training'}
              value={participants}
              onChange={(event) => setParticipants(Number(event.target.value))}
            />
            <span>{participants}</span>
          </label>

          <label>
            Local Epochs
            <input
              type="range"
              min="1"
              max="10"
                            disabled={trainingMode === 'real' && phase3dStatus === 'training'}
              value={localEpochs}
              onChange={(event) => setLocalEpochs(Number(event.target.value))}
            />
            <span>{localEpochs}</span>
          </label>

          <label>
            Privacy Epsilon
            <input
              type="range"
              min="0.2"
              max="3"
              step="0.1"
                            disabled={trainingMode === 'real' && phase3dStatus === 'training'}
              value={epsilon}
              onChange={(event) => setEpsilon(Number(event.target.value))}
            />
            <span>{epsilon.toFixed(1)}</span>
          </label>

          <label>
            Compression Bits
            <input
              type="range"
              min="4"
              max="16"
              step="1"
                            disabled={trainingMode === 'real' && phase3dStatus === 'training'}
              value={compressionBits}
              onChange={(event) => setCompressionBits(Number(event.target.value))}
            />
            <span>{compressionBits}-bit</span>
          </label>

          <label>
            Learning Rate
            <input
              type="range"
              min="0.005"
              max="0.08"
              step="0.001"
                            disabled={trainingMode === 'real' && phase3dStatus === 'training'}
              value={learningRate}
              onChange={(event) => setLearningRate(Number(event.target.value))}
            />
            <span>{learningRate.toFixed(3)}</span>
          </label>

          <label>
            Target Rounds
            <input
              type="number"
              min="5"
              max="400"
              disabled={trainingMode === 'real' && phase3dStatus === 'training'}
              value={trainingMode === 'real' ? phase3dTotal : targetRounds}
              onChange={(event) => {
                const val = clamp(Number(event.target.value) || 1, 5, 400);
                if (trainingMode === 'simulation') {
                  setTargetRounds(val);
                } else {
                  setPhase3dTotal(val);
                }
              }}
            />
          </label>

          <div className="control-actions">
                        {trainingMode === 'simulation' ? (
            <button onClick={() => setRunning((value) => !value)}>
              {running ? 'Pause Training' : 'Start Training'}
            </button>
            ) : (
              <>
                <button 
                  onClick={phase3dStatus === 'training' ? cancelPhase3dTraining : startPhase3dTraining}
                  style={{ background: phase3dStatus === 'training' ? '#ff6b6b' : '#007AFF' }}
                >
                  {phase3dStatus === 'training' ? 'Cancel Training' : 'Start Real Training'}
                </button>
              </>
            )}
            <button className="ghost" onClick={resetScenario}>
              Reset
            </button>
          </div>
        </aside>

        <div className="metrics-column">
          <div className="kpi-grid">
            <article>
              <h4>Round</h4>
                            <p>{trainingMode === 'real' ? `${phase3dRound}/${phase3dTotal}` : round}</p>
              <p>{round}</p>
            </article>
            <article>
              <h4>Accuracy</h4>
              <p>{(accuracy * 100).toFixed(2)}%</p>
            </article>
            <article>
              <h4>Loss</h4>
              <p>{loss.toFixed(3)}</p>
            </article>
            <article>
              <h4>Compression</h4>
              <p>{compressionRatio.toFixed(2)}x</p>
            </article>
            <article>
              <h4>Bandwidth</h4>
              <p>{bandwidthKB.toFixed(1)} KB</p>
            </article>
            <article>
              <h4>Round Latency</h4>
              <p>{latencyMs} ms</p>
                        {trainingMode === 'real' && (
                          <article>
                            <h4>Training Mode</h4>
                            <p>{phase3dStatus.toUpperCase()}</p>
                          </article>
                        )}
            </article>
          </div>

          <div className="progress-wrap">
            <div className="progress-labels">
              <span>Progress</span>
                            <span>{(trainingMode === 'real' ? (phase3dRound / phase3dTotal) * 100 : progressPercent).toFixed(0)}%</span>
              <span>{progressPercent.toFixed(0)}%</span>
            </div>
            <div className="progress-track">
              <div className="progress-fill" style={{ width: `${progressPercent}%` }} />
            </div>
            <small>
              Compression saves approximately {compressionSavedPercent.toFixed(1)}% of gradient
              transfer volume in this scenario.
            </small>
          </div>

          <div className="chart-card">
            <h3>Convergence</h3>
            <ResponsiveContainer width="100%" height={260}>
              <LineChart data={history}>
                <CartesianGrid strokeDasharray="3 3" stroke="#d8deeb" />
                <XAxis dataKey="round" />
                <YAxis yAxisId="left" domain={[0.4, 1]} />
                <YAxis yAxisId="right" orientation="right" domain={[0, 2.2]} />
                <Tooltip />
                <Legend />
                <Line yAxisId="left" type="monotone" dataKey="accuracy" stroke="#1678f2" dot={false} />
                <Line yAxisId="right" type="monotone" dataKey="loss" stroke="#d9480f" dot={false} />
              </LineChart>
            </ResponsiveContainer>
          </div>

          <div className="chart-card">
            <h3>Communication Cost</h3>
            <ResponsiveContainer width="100%" height={220}>
              <AreaChart data={history}>
                <CartesianGrid strokeDasharray="3 3" stroke="#d8deeb" />
                <XAxis dataKey="round" />
                <YAxis />
                <Tooltip />
                <Area type="monotone" dataKey="bandwidthKB" stroke="#0b7285" fill="#99e9f2" />
              </AreaChart>
            </ResponsiveContainer>
          </div>
        </div>
      </div>
    </section>
  );
}

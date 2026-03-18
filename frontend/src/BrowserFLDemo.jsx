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
const TRAINING_API_BASE = import.meta.env.VITE_TRAINING_API_BASE || 'http://localhost:5001';

function clamp(value, min, max) {
  return Math.max(min, Math.min(max, value));
}

function randomFloat(min, max) {
  return min + Math.random() * (max - min);
}

export default function BrowserFLDemo({ enableBackendMetrics = false, onMetricsUpdate = null }) {
  const [trainingMode, setTrainingMode] = useState('simulation');

  const [participants, setParticipants] = useState(120);
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
  const [realHistory, setRealHistory] = useState([]);

  const [usingBackendData, setUsingBackendData] = useState(false);

  const [phase3dStatus, setPhase3dStatus] = useState('idle');
  const [phase3dRound, setPhase3dRound] = useState(0);
  const [phase3dTotal, setPhase3dTotal] = useState(10);

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

  useEffect(() => {
    if (!enableBackendMetrics) return undefined;

    const fetchBackendMetrics = async () => {
      try {
        const response = await fetch(`${API_BASE}/metrics_summary`);
        setUsingBackendData(response.ok);
      } catch {
        setUsingBackendData(false);
      }
    };

    fetchBackendMetrics();
    const interval = setInterval(fetchBackendMetrics, 5000);
    return () => clearInterval(interval);
  }, [enableBackendMetrics]);

  useEffect(() => {
    if (trainingMode !== 'real' || phase3dStatus !== 'training') return undefined;

    const pollTrainingStatus = async () => {
      try {
        const response = await fetch(`${TRAINING_API_BASE}/training/status`);
        if (!response.ok) return;

        const status = await response.json();
        setPhase3dRound(status.current_round || 0);
        setPhase3dTotal(status.total_rounds || phase3dTotal);

        if (status.current_metrics) {
          const metric = status.current_metrics;
          setAccuracy(metric.accuracy ?? accuracy);
          setLoss(metric.round_loss ?? loss);
          setCompressionRatio(metric.compression_ratio ?? compressionRatio);

          setRealHistory((previous) => {
            const point = {
              round: status.current_round || previous.length + 1,
              accuracy: Number((metric.accuracy ?? accuracy).toFixed(4)),
              loss: Number((metric.round_loss ?? loss).toFixed(4)),
              compressionRatio: Number((metric.compression_ratio ?? compressionRatio).toFixed(2)),
              latencyMs,
              bandwidthKB
            };
            const merged = [...previous, point].slice(-MAX_POINTS);
            onMetricsUpdate?.(merged);
            return merged;
          });
        }

        if (status.status === 'completed') {
          setPhase3dStatus('completed');
        } else if (status.status === 'error') {
          setPhase3dStatus('idle');
        }
      } catch {
        setPhase3dStatus('idle');
      }
    };

    pollTrainingStatus();
    const interval = setInterval(pollTrainingStatus, 2000);
    return () => clearInterval(interval);
  }, [
    trainingMode,
    phase3dStatus,
    phase3dTotal,
    accuracy,
    loss,
    compressionRatio,
    latencyMs,
    bandwidthKB,
    onMetricsUpdate
  ]);

  useEffect(() => {
    if (!running || trainingMode !== 'simulation') return undefined;

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

          const merged = [...previousHistory, point].slice(-MAX_POINTS);
          onMetricsUpdate?.(merged);
          return merged;
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
    loss,
    onMetricsUpdate
  ]);

  const backendLabel = useMemo(() => {
    if (runtimeBackend === 'webgpu') return 'WebGPU Active';
    if (runtimeBackend === 'cpu-fallback') return 'CPU Fallback';
    return 'Detecting Runtime';
  }, [runtimeBackend]);

  const backendClass = runtimeBackend === 'webgpu' ? 'badge-webgpu' : 'badge-cpu';

  const activeRound = trainingMode === 'real' ? phase3dRound : round;
  const activeTarget = trainingMode === 'real' ? phase3dTotal : targetRounds;
  const activeHistory = trainingMode === 'real' ? realHistory : history;

  const progressPercent = clamp((activeRound / Math.max(activeTarget, 1)) * 100, 0, 100);
  const compressionSavedPercent = clamp((1 - 1 / compressionRatio) * 100, 0, 99.9);

  const resetScenario = () => {
    setRunning(false);
    setRound(0);
    setPhase3dRound(0);
    setAccuracy(0.51);
    setLoss(1.7);
    setCompressionRatio(4.0);
    setLatencyMs(0);
    setBandwidthKB(0);
    setHistory([]);
    setRealHistory([]);
    onMetricsUpdate?.([]);
  };

  const startPhase3dTraining = async () => {
    try {
      const config = {
        num_rounds: phase3dTotal,
        num_clients: participants,
        local_epochs: localEpochs,
        learning_rate: learningRate,
        epsilon,
        compression_bits: compressionBits,
        use_compression: true,
        use_privacy: true,
        dataset: 'cifar10',
        device: 'cuda',
        multi_gpu: true
      };

      const response = await fetch(`${TRAINING_API_BASE}/training/start`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(config)
      });

      if (!response.ok) {
        setPhase3dStatus('idle');
        return;
      }

      setPhase3dStatus('training');
      setRunning(false);
      setRound(0);
      setPhase3dRound(0);
      setRealHistory([]);
      onMetricsUpdate?.([]);
    } catch {
      setPhase3dStatus('idle');
    }
  };

  const cancelPhase3dTraining = async () => {
    try {
      await fetch(`${TRAINING_API_BASE}/training/cancel`, { method: 'POST' });
      setPhase3dStatus('idle');
    } catch {
      setPhase3dStatus('idle');
    }
  };

  return (
    <section className="demo-shell">
      <div className="demo-topbar">
        <div>
          <h2>Browser Federated Learning Studio</h2>
          <p>
            {trainingMode === 'simulation'
              ? 'Interactive simulation of local training, privacy noise, gradient compression, and network bandwidth tradeoffs.'
              : 'Real federated learning with CIFAR-10, gradient compression, and CUDA multi-GPU support.'}
            {usingBackendData ? ' (Real network metrics enabled)' : ''}
          </p>
        </div>
        <div style={{ display: 'flex', gap: '8px' }}>
          <div className={`runtime-badge ${backendClass}`}>{backendLabel}</div>
          {enableBackendMetrics && (
            <div className={`runtime-badge ${usingBackendData ? 'badge-sync' : 'badge-offline'}`}>
              {usingBackendData ? 'Network Live' : 'Simulation Only'}
            </div>
          )}
          {trainingMode === 'real' && (
            <div className={`runtime-badge ${phase3dStatus === 'training' ? 'badge-sync' : phase3dStatus === 'completed' ? 'badge-webgpu' : 'badge-offline'}`}>
              {phase3dStatus === 'training' ? 'Training Live' : phase3dStatus === 'completed' ? 'Training Complete' : 'Phase 3D Ready'}
            </div>
          )}
        </div>
      </div>

      <div className="demo-layout">
        <aside className="control-card">
          <h3>Round Controls</h3>

          <label>
            Training Mode
            <select value={trainingMode} onChange={(event) => setTrainingMode(event.target.value)}>
              <option value="simulation">Simulation</option>
              <option value="real">Phase 3D (Real)</option>
            </select>
          </label>

          <label>
            Participants
            <input
              type="range"
              min="10"
              max="500"
              value={participants}
              onChange={(event) => setParticipants(Number(event.target.value))}
              disabled={trainingMode === 'real' && phase3dStatus === 'training'}
            />
            <span>{participants}</span>
          </label>

          <label>
            Local Epochs
            <input
              type="range"
              min="1"
              max="10"
              value={localEpochs}
              onChange={(event) => setLocalEpochs(Number(event.target.value))}
              disabled={trainingMode === 'real' && phase3dStatus === 'training'}
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
              value={epsilon}
              onChange={(event) => setEpsilon(Number(event.target.value))}
              disabled={trainingMode === 'real' && phase3dStatus === 'training'}
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
              value={compressionBits}
              onChange={(event) => setCompressionBits(Number(event.target.value))}
              disabled={trainingMode === 'real' && phase3dStatus === 'training'}
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
              value={learningRate}
              onChange={(event) => setLearningRate(Number(event.target.value))}
              disabled={trainingMode === 'real' && phase3dStatus === 'training'}
            />
            <span>{learningRate.toFixed(3)}</span>
          </label>

          <label>
            Target Rounds
            <input
              type="number"
              min="5"
              max="400"
              value={trainingMode === 'real' ? phase3dTotal : targetRounds}
              onChange={(event) => {
                const value = clamp(Number(event.target.value) || 1, 5, 400);
                if (trainingMode === 'real') {
                  setPhase3dTotal(value);
                } else {
                  setTargetRounds(value);
                }
              }}
              disabled={trainingMode === 'real' && phase3dStatus === 'training'}
            />
          </label>

          <div className="control-actions">
            {trainingMode === 'simulation' ? (
              <button onClick={() => setRunning((value) => !value)}>{running ? 'Pause Training' : 'Start Training'}</button>
            ) : (
              <button onClick={phase3dStatus === 'training' ? cancelPhase3dTraining : startPhase3dTraining}>
                {phase3dStatus === 'training' ? 'Cancel Training' : 'Start Real Training'}
              </button>
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
              <p>{activeRound}</p>
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
            </article>
          </div>

          <div className="progress-wrap">
            <div className="progress-labels">
              <span>Progress</span>
              <span>{progressPercent.toFixed(0)}%</span>
            </div>
            <div className="progress-track">
              <div className="progress-fill" style={{ width: `${progressPercent}%` }} />
            </div>
            <div className="progress-meta">
              <span>
                {activeRound}/{activeTarget} rounds
              </span>
              <span>{compressionSavedPercent.toFixed(1)}% payload saved</span>
            </div>
          </div>

          <div className="chart-grid">
            <article className="chart-card">
              <h3>Accuracy and Loss</h3>
              <ResponsiveContainer width="100%" height={260}>
                <LineChart data={activeHistory}>
                  <CartesianGrid strokeDasharray="3 3" />
                  <XAxis dataKey="round" />
                  <YAxis yAxisId="left" domain={[0, 1]} />
                  <YAxis yAxisId="right" orientation="right" domain={[0, 2.5]} />
                  <Tooltip />
                  <Legend />
                  <Line yAxisId="left" type="monotone" dataKey="accuracy" stroke="#1d4ed8" dot={false} />
                  <Line yAxisId="right" type="monotone" dataKey="loss" stroke="#f97316" dot={false} />
                </LineChart>
              </ResponsiveContainer>
            </article>

            <article className="chart-card">
              <h3>Bandwidth and Compression</h3>
              <ResponsiveContainer width="100%" height={260}>
                <AreaChart data={activeHistory}>
                  <CartesianGrid strokeDasharray="3 3" />
                  <XAxis dataKey="round" />
                  <YAxis />
                  <Tooltip />
                  <Legend />
                  <Area type="monotone" dataKey="bandwidthKB" stroke="#0ea5e9" fill="#bae6fd" />
                  <Line type="monotone" dataKey="compressionRatio" stroke="#16a34a" dot={false} />
                </AreaChart>
              </ResponsiveContainer>
            </article>
          </div>
        </div>
      </div>
    </section>
  );
}

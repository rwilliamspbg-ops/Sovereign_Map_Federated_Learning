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

function clamp(value, min, max) {
  return Math.max(min, Math.min(max, value));
}

function randomFloat(min, max) {
  return min + Math.random() * (max - min);
}

export default function BrowserFLDemo() {
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
    if (!running) return undefined;

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
  };

  return (
    <section className="demo-shell">
      <div className="demo-topbar">
        <div>
          <h2>Browser Federated Learning Studio</h2>
          <p>
            Interactive simulation of local training, privacy noise, gradient compression, and
            network bandwidth tradeoffs.
          </p>
        </div>
        <div className={`runtime-badge ${backendClass}`}>{backendLabel}</div>
      </div>

      <div className="demo-layout">
        <aside className="control-card">
          <h3>Round Controls</h3>

          <label>
            Participants
            <input
              type="range"
              min="10"
              max="500"
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
              value={targetRounds}
              onChange={(event) => setTargetRounds(clamp(Number(event.target.value) || 1, 5, 400))}
            />
          </label>

          <div className="control-actions">
            <button onClick={() => setRunning((value) => !value)}>
              {running ? 'Pause Training' : 'Start Training'}
            </button>
            <button className="ghost" onClick={resetScenario}>
              Reset
            </button>
          </div>
        </aside>

        <div className="metrics-column">
          <div className="kpi-grid">
            <article>
              <h4>Round</h4>
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

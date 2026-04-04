import { useState, useEffect, useRef } from 'react';
import HUD from './HUD';
import BrowserFLDemo from './BrowserFLDemo';
import './App.css';

const resolveApiBase = () => {
  if (import.meta.env.VITE_HUD_API_BASE) {
    return import.meta.env.VITE_HUD_API_BASE;
  }

  // Always use the frontend reverse-proxy path unless explicitly overridden.
  // This avoids direct host:8000 reachability/CORS issues for browser clients.
  return '/backend';
};

const API_BASE = resolveApiBase();
const TRUST_API_BASE = API_BASE;
const DEFAULT_VIEW = String(import.meta.env.VITE_DEFAULT_VIEW || 'hud').toLowerCase();

const resolveViewMode = () => {
  if (typeof window === 'undefined') {
    return DEFAULT_VIEW;
  }
  const requestedView = new URLSearchParams(window.location.search).get('view');
  if (requestedView) {
    return String(requestedView).toLowerCase();
  }
  return DEFAULT_VIEW;
};

const opsEventKey = (evt) => {
  if (evt && Number.isFinite(Number(evt.id))) {
    return `id:${Number(evt.id)}`;
  }
  return `sig:${evt?.ts || 0}:${evt?.kind || ''}:${evt?.message || ''}`;
};

const mergeOpsEvents = (existing, incoming, limit = 220) => {
  const merged = [];
  const seen = new Set();

  [...(existing || []), ...(incoming || [])].forEach((evt) => {
    if (!evt || evt.kind === 'heartbeat') {
      return;
    }
    const key = opsEventKey(evt);
    if (seen.has(key)) {
      return;
    }
    seen.add(key);
    merged.push(evt);
  });

  return merged.slice(-limit);
};

const ema = (previous, next, alpha = 0.35) => {
  if (!Number.isFinite(next)) return previous;
  if (!Number.isFinite(previous)) return next;
  return (alpha * next) + ((1 - alpha) * previous);
};

function App() {
  const viewMode = resolveViewMode();
  const showBrowserDemo = viewMode === 'browser_demo' || viewMode === 'admin';

  const [hudData, setHudData] = useState(null);
  const [health, setHealth] = useState(null);
  const [metricsSummary, setMetricsSummary] = useState(null);
  const [trustStatus, setTrustStatus] = useState(null);
  const [policyHistory, setPolicyHistory] = useState([]);
  const [founders, setFounders] = useState([]);
  const [voiceQuery, setVoiceQuery] = useState('');
  const [voiceResponse, setVoiceResponse] = useState('');
  const [webMetrics, setWebMetrics] = useState({
    pageLoadMs: null,
    domInteractiveMs: null,
    jsHeapUsedMB: null,
    jsHeapTotalMB: null,
    viewport: null,
    cores: null,
    deviceMemoryGB: null,
    downlinkMbps: null,
    rttMs: null,
    connectionType: null,
    effectiveType: null
  });
  const [policyDraft, setPolicyDraft] = useState({
    require_proof: false,
    min_confidence_bps: 7000,
    reject_on_verification_failure: false,
    allow_consensus_proof: true,
    allow_zk_proof: true,
    allow_tee_proof: true
  });
  const [policyInitialized, setPolicyInitialized] = useState(false);
  const [policyToken, setPolicyToken] = useState('');
  const [policyRole, setPolicyRole] = useState('admin');
  const [policyMessage, setPolicyMessage] = useState('');
  const [trainingStatus, setTrainingStatus] = useState({ status: 'idle', active: false, round: 0 });
  const [opsHealth, setOpsHealth] = useState(null);
  const [opsTrends, setOpsTrends] = useState(null);
  const [opsEvents, setOpsEvents] = useState([]);
  const [opsStreamStatus, setOpsStreamStatus] = useState('disconnected');
  const [opsStreamLastError, setOpsStreamLastError] = useState('');
  const [policyTransport, setPolicyTransport] = useState({
    endpoint: '',
    status: 'idle',
    detail: ''
  });
  const [enclaveActionMessage, setEnclaveActionMessage] = useState('');
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState(null);
  const smoothingRef = useRef({
    trainingAccuracy: null,
    trainingLoss: null,
    opsLatencyMs: null,
    opsErrorRate: null,
    opsIngressMbps: null,
    opsSaturationPct: null
  });

  if (showBrowserDemo) {
    return <BrowserFLDemo enableBackendMetrics />;
  }

  const avgFlDuration = Number(metricsSummary?.avg_fl_duration);
  const totalStake = Number(metricsSummary?.total_stake);
  const cxlUtilization = Number(metricsSummary?.cxl_utilization);

  useEffect(() => {
    fetchData();
    const interval = setInterval(fetchData, 5000);
    return () => clearInterval(interval);
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  useEffect(() => {
    let stream;
    let cancelled = false;

    const connect = async () => {
      try {
        setOpsStreamStatus('connecting');

        // Prime timeline with recent events before opening stream.
        const recent = await fetch(`${API_BASE}/ops/events/recent?limit=80`);
        if (recent.ok) {
          const recentPayload = await recent.json();
          const seedEvents = Array.isArray(recentPayload?.events) ? recentPayload.events : [];
          setOpsEvents((prev) => mergeOpsEvents(prev, seedEvents));
        }

        if (cancelled) {
          return;
        }

        stream = new EventSource(`${API_BASE}/ops/events`);
        stream.onopen = () => {
          setOpsStreamStatus('connected');
          setOpsStreamLastError('');
        };
        stream.onerror = () => {
          setOpsStreamStatus('degraded');
          setOpsStreamLastError(`SSE transport error @ ${new Date().toLocaleTimeString()}`);
        };
        stream.onmessage = (evt) => {
          try {
            const payload = JSON.parse(evt.data);
            if (payload?.kind === 'heartbeat') {
              return;
            }
            setOpsEvents((prev) => mergeOpsEvents(prev, [payload]));
          } catch (parseErr) {
            console.warn('Failed to parse ops event', parseErr);
          }
        };
      } catch (streamErr) {
        console.warn('Ops event stream unavailable:', streamErr);
        setOpsStreamStatus('disconnected');
        setOpsStreamLastError(streamErr?.message || 'stream bootstrap failed');
      }
    };

    connect();

    return () => {
      cancelled = true;
      if (stream) {
        stream.close();
      }
    };
  }, []);

  useEffect(() => {
    const updateWebMetrics = () => {
      if (typeof window === 'undefined') {
        return;
      }

      const navigationEntry = performance.getEntriesByType('navigation')?.[0];
      const domInteractiveMs = navigationEntry
        ? Math.round(navigationEntry.domInteractive)
        : Math.max(0, (performance.timing?.domInteractive || 0) - (performance.timing?.navigationStart || 0));
      const pageLoadMs = navigationEntry
        ? Math.round(navigationEntry.loadEventEnd || navigationEntry.domComplete || 0)
        : Math.max(0, (performance.timing?.loadEventEnd || 0) - (performance.timing?.navigationStart || 0));

      const heap = performance.memory || null;
      const connection = navigator.connection || navigator.mozConnection || navigator.webkitConnection || null;

      setWebMetrics({
        pageLoadMs: pageLoadMs > 0 ? pageLoadMs : null,
        domInteractiveMs: domInteractiveMs > 0 ? domInteractiveMs : null,
        jsHeapUsedMB: heap ? Number((heap.usedJSHeapSize / (1024 * 1024)).toFixed(2)) : null,
        jsHeapTotalMB: heap ? Number((heap.totalJSHeapSize / (1024 * 1024)).toFixed(2)) : null,
        viewport: `${window.innerWidth}x${window.innerHeight}`,
        cores: navigator.hardwareConcurrency || null,
        deviceMemoryGB: navigator.deviceMemory || null,
        downlinkMbps: connection?.downlink || null,
        rttMs: connection?.rtt || null,
        connectionType: connection?.type || null,
        effectiveType: connection?.effectiveType || null
      });
    };

    updateWebMetrics();
    const interval = setInterval(updateWebMetrics, 5000);
    window.addEventListener('resize', updateWebMetrics);

    return () => {
      clearInterval(interval);
      window.removeEventListener('resize', updateWebMetrics);
    };
  }, []);

  const fetchTrustSnapshot = async () => {
    const trustRes = await fetch(`${TRUST_API_BASE}/trust_snapshot`);
    if (!trustRes.ok) {
      throw new Error(`Trust snapshot failed with ${trustRes.status}`);
    }

    const snapshot = await trustRes.json();
    setTrustStatus(snapshot.trust_status || null);
    setPolicyHistory(snapshot.policy_history || []);

    if (!policyInitialized && snapshot.trust_status?.verification_policy) {
      setPolicyDraft({ ...snapshot.trust_status.verification_policy });
      setPolicyInitialized(true);
    }
  };

  const fetchData = async () => {
    setLoading(true);
    try {
      const [hudFetch, healthFetch, metricsFetch, foundersFetch, trainingFetch, opsHealthFetch, opsTrendsFetch] = await Promise.allSettled([
        fetch(`${API_BASE}/hud_data`),
        fetch(`${API_BASE}/health`),
        fetch(`${API_BASE}/metrics_summary`),
        fetch(`${API_BASE}/founders`),
        fetch(`${API_BASE}/training/status`),
        fetch(`${API_BASE}/ops/health`),
        fetch(`${API_BASE}/ops/trends?limit=180`)
      ]);

      const toResponse = (settled) => (settled?.status === 'fulfilled' ? settled.value : null);
      const safeJson = async (response, fallback) => {
        if (!response || !response.ok) {
          return fallback;
        }
        try {
          return await response.json();
        } catch {
          return fallback;
        }
      };

      const [hud, healthData, metrics, foundersData, trainingData, opsHealthData, opsTrendsData] = await Promise.all([
        safeJson(toResponse(hudFetch), hudData || {}),
        safeJson(toResponse(healthFetch), health || {}),
        safeJson(toResponse(metricsFetch), metricsSummary || {}),
        safeJson(toResponse(foundersFetch), founders || []),
        safeJson(toResponse(trainingFetch), { status: 'idle', active: false, round: 0 }),
        safeJson(toResponse(opsHealthFetch), opsHealth || null),
        safeJson(toResponse(opsTrendsFetch), opsTrends || null)
      ]);

      const nextTraining = { ...(trainingData || { status: 'idle', active: false, round: 0 }) };
      if (nextTraining?.current_metrics) {
        const nextAcc = Number(nextTraining.current_metrics.accuracy);
        const nextLoss = Number(nextTraining.current_metrics.loss);
        const smoothAcc = ema(smoothingRef.current.trainingAccuracy, nextAcc, 0.35);
        const smoothLoss = ema(smoothingRef.current.trainingLoss, nextLoss, 0.30);
        smoothingRef.current.trainingAccuracy = smoothAcc;
        smoothingRef.current.trainingLoss = smoothLoss;

        nextTraining.current_metrics = {
          ...nextTraining.current_metrics,
          accuracy: Number.isFinite(smoothAcc) ? Number(smoothAcc.toFixed(4)) : nextTraining.current_metrics.accuracy,
          loss: Number.isFinite(smoothLoss) ? Number(smoothLoss.toFixed(4)) : nextTraining.current_metrics.loss
        };
      }

      let nextOpsHealth = opsHealthData;
      if (opsHealthData?.telemetry) {
        const telemetry = { ...opsHealthData.telemetry };
        const latency = Number(telemetry.api_latency_ms);
        const errorRate = Number(telemetry.api_error_rate);
        const ingress = Number(telemetry.ingress_mbps);
        const saturation = Number(telemetry.global_saturation_pct);

        const smoothLatency = ema(smoothingRef.current.opsLatencyMs, latency, 0.30);
        const smoothErrorRate = ema(smoothingRef.current.opsErrorRate, errorRate, 0.35);
        const smoothIngress = ema(smoothingRef.current.opsIngressMbps, ingress, 0.30);
        const smoothSaturation = ema(smoothingRef.current.opsSaturationPct, saturation, 0.25);

        smoothingRef.current.opsLatencyMs = smoothLatency;
        smoothingRef.current.opsErrorRate = smoothErrorRate;
        smoothingRef.current.opsIngressMbps = smoothIngress;
        smoothingRef.current.opsSaturationPct = smoothSaturation;

        nextOpsHealth = {
          ...opsHealthData,
          telemetry: {
            ...telemetry,
            api_latency_ms: Number.isFinite(smoothLatency) ? Math.round(smoothLatency) : telemetry.api_latency_ms,
            api_error_rate: Number.isFinite(smoothErrorRate) ? Number(smoothErrorRate.toFixed(3)) : telemetry.api_error_rate,
            ingress_mbps: Number.isFinite(smoothIngress) ? Math.round(smoothIngress) : telemetry.ingress_mbps,
            global_saturation_pct: Number.isFinite(smoothSaturation) ? Math.round(smoothSaturation) : telemetry.global_saturation_pct
          }
        };
      }

      setHudData(hud);
      setHealth(healthData);
      setMetricsSummary(metrics);
      setFounders(foundersData);
      setTrainingStatus(nextTraining);
      setOpsHealth(nextOpsHealth);
      setOpsTrends(opsTrendsData);

      try {
        await fetchTrustSnapshot();
      } catch (trustErr) {
        console.warn('Trust status unavailable:', trustErr);
      }

      setError(null);
    } catch (err) {
      console.error('Fetch error:', err);
      setError('Failed to sync with Sovereign Node');
    } finally {
      setLoading(false);
    }
  };

  const triggerFLRound = async () => {
    try {
      const response = await fetch(`${API_BASE}/trigger_fl`, { method: 'POST' });
      if (response.ok) {
        fetchData();
      }
    } catch (err) {
      console.error('Trigger FL round error:', err);
      setError('Failed to trigger FL round');
    }
  };

  const startTraining = async (rounds = 0) => {
    try {
      const targetRounds = Number(rounds);
      const payload = Number.isFinite(targetRounds) && targetRounds > 0
        ? { rounds: Math.floor(targetRounds) }
        : {};
      const response = await fetch(`${API_BASE}/training/start`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(payload)
      });
      if (!response.ok) {
        throw new Error(`Training start failed with ${response.status}`);
      }
      await fetchData();
    } catch (err) {
      console.error('Start training error:', err);
      setError('Failed to start real training loop');
    }
  };

  const stopTraining = async () => {
    try {
      const response = await fetch(`${API_BASE}/training/stop`, { method: 'POST' });
      if (!response.ok) {
        throw new Error(`Training stop failed with ${response.status}`);
      }
      await fetchData();
    } catch (err) {
      console.error('Stop training error:', err);
      setError('Failed to stop training loop');
    }
  };

  const createEnclave = async () => {
    try {
      const response = await fetch(`${API_BASE}/create_enclave`, { method: 'POST' });
      const payload = await response.json().catch(() => ({}));
      setEnclaveActionMessage(payload?.message || `Enclave status: ${payload?.enclave_status || 'unknown'}`);
      fetchData();

      if (response.status === 202) {
        setTimeout(fetchData, 2500);
      }
    } catch (err) {
      console.error('Enclave creation error:', err);
      setError('Enclave creation failed');
      setEnclaveActionMessage('TEE provisioning request failed');
    }
  };

  const triggerSimulation = async (simulationType) => {
    try {
      const response = await fetch(`${API_BASE}/simulate/${simulationType}`, {
        method: 'POST'
      });
      if (!response.ok) {
        throw new Error(`Simulation endpoint failed with ${response.status}`);
      }
      fetchData();
    } catch (err) {
      console.error('Simulation trigger error:', err);
      setError(`Failed to trigger simulation: ${simulationType}`);
    }
  };

    const submitVoiceQuery = async () => {
    if (!voiceQuery.trim()) {
      return;
    }

    try {
      const response = await fetch(`${API_BASE}/chat`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ query: voiceQuery })
      });
      if (response.ok) {
        const result = await response.json();
        setVoiceResponse(result.response);
      } else {
        setVoiceResponse('Unable to retrieve system status');
      }
    } catch (err) {
      console.error('Voice query error:', err);
      setVoiceResponse('Connection error communicating with node');
    }
  };

  const updatePolicyField = (field, value) => {
    setPolicyDraft((current) => ({ ...current, [field]: value }));
  };

  const submitVerificationPolicy = async () => {
    try {
      setPolicyMessage('Submitting verification policy...');
      const headers = {
        'Content-Type': 'application/json',
        'X-API-Role': policyRole || 'admin'
      };

      if (policyToken.trim()) {
        headers.Authorization = `Bearer ${policyToken.trim()}`;
      }

      const requestBody = JSON.stringify({
        ...policyDraft,
        min_confidence_bps: Number(policyDraft.min_confidence_bps) || 0
      });

      const candidateEndpoints = [
        `${TRUST_API_BASE}/verification_policy`,
        '/backend/verification_policy',
        'http://localhost:8000/verification_policy'
      ];

      let response = null;
      for (const endpoint of candidateEndpoints) {
        try {
          setPolicyTransport({ endpoint, status: 'pending', detail: 'submitting policy update' });
          const candidate = await fetch(endpoint, {
            method: 'POST',
            headers,
            body: requestBody
          });
          response = candidate;
          setPolicyTransport({
            endpoint,
            status: candidate.ok ? 'ok' : 'error',
            detail: `HTTP ${candidate.status}`
          });
          if (candidate.ok) {
            break;
          }
        } catch {
          setPolicyTransport({ endpoint, status: 'error', detail: 'network/transport failure' });
          // Try the next candidate endpoint.
        }
      }

      if (!response) {
        setPolicyTransport({ endpoint: candidateEndpoints[0], status: 'error', detail: 'no endpoint reachable' });
        throw new Error('no policy endpoint reachable');
      }

      if (!response.ok) {
        const failure = await response.text();
        setPolicyMessage(`Policy update failed (${response.status}): ${failure || 'unknown error'}`);
        return;
      }

      const result = await response.json();
      setTrustStatus((current) => ({
        ...current,
        trust_mode: 'p2p-reputation+governed-proof-verification',
        verification_policy: result.verification_policy,
        fl_verification: result.fl_verification
      }));
      setPolicyHistory(result.policy_history || []);
      setPolicyDraft({ ...result.verification_policy });
      setPolicyMessage('Verification policy updated');
    } catch (err) {
      console.error('Verification policy update error:', err);
      try {
        const backendProbe = await fetch(`${API_BASE}/health`);
        if (backendProbe.ok) {
          setPolicyMessage('Verification policy update failed: endpoint mismatch or auth policy rejected request');
          return;
        }
      } catch {
        // Keep unreachable message below.
      }
      setPolicyTransport({ endpoint: TRUST_API_BASE, status: 'error', detail: 'backend unreachable' });
      setPolicyMessage('Verification policy update failed: backend unreachable');
    }
  };

  return (
    <div className="hud-container">
      <header className="hud-header">
        <div>
          <h1>Sovereign Map Federated Interface</h1>
          <p>
            Use the live network operations HUD as the primary command deck for training, marketplace,
            wallet, and API telemetry.
          </p>
        </div>

        <div className="status-bar">
          {error ? <span className="status-error">{error}</span> : <span className="status-ok">Network Reachable</span>}
          {loading && <span className="status-loading"> Syncing...</span>}
        </div>
      </header>

      <HUD
        hudData={hudData}
        health={health}
        metricsSummary={metricsSummary}
        trustStatus={trustStatus}
        policyHistory={policyHistory}
        founders={founders}
        voiceQuery={voiceQuery}
        voiceResponse={voiceResponse}
        policyDraft={policyDraft}
        policyToken={policyToken}
        policyRole={policyRole}
        policyMessage={policyMessage}
        enclaveActionMessage={enclaveActionMessage}
        trainingStatus={trainingStatus}
        opsHealth={opsHealth}
        opsTrends={opsTrends}
        opsEvents={opsEvents}
        opsStreamStatus={opsStreamStatus}
        connectionDiagnostics={{
          apiBase: API_BASE,
          eventsEndpoint: `${API_BASE}/ops/events`,
          streamLastError: opsStreamLastError,
          policyEndpoint: policyTransport.endpoint,
          policyStatus: policyTransport.status,
          policyDetail: policyTransport.detail
        }}
        webMetrics={webMetrics}
        loading={loading}
        error={error}
        onTriggerFLRound={triggerFLRound}
        onStartTraining={startTraining}
        onStopTraining={stopTraining}
        onCreateEnclave={createEnclave}
        onTriggerSimulation={triggerSimulation}
        onSubmitVoiceQuery={submitVoiceQuery}
        onPolicyChange={updatePolicyField}
        onPolicyTokenChange={setPolicyToken}
        onPolicyRoleChange={setPolicyRole}
        onSubmitVerificationPolicy={submitVerificationPolicy}
        setVoiceQuery={setVoiceQuery}
      />

      {metricsSummary && (
        <div className="metrics-footer">
          <div className="summary-item">
            <strong>Total FL Rounds:</strong> {metricsSummary.fl_rounds_total}
          </div>
          <div className="summary-item">
            <strong>Avg FL Duration:</strong> {Number.isFinite(avgFlDuration) ? `${avgFlDuration.toFixed(2)}s` : 'N/A'}
          </div>
          <div className="summary-item">
            <strong>Total Stake:</strong> {Number.isFinite(totalStake) ? totalStake.toFixed(2) : 'N/A'}
          </div>
          <div className="summary-item">
            <strong>CXL Utilization:</strong> {Number.isFinite(cxlUtilization) ? `${(cxlUtilization * 100).toFixed(1)}%` : 'N/A'}
          </div>
        </div>
      )}
    </div>
  );
}

export default App;

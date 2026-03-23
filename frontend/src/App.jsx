import { useState, useEffect, useMemo } from 'react';
import HUD from './HUD';
import './App.css';
import PrivacyUtilityDashboard from './PrivacyUtilityDashboard';

const resolveApiBase = () => {
  if (import.meta.env.VITE_HUD_API_BASE) {
    return import.meta.env.VITE_HUD_API_BASE;
  }

  if (import.meta.env.DEV) {
    return '/backend';
  }

  if (typeof window !== 'undefined') {
    const protocol = window.location.protocol === 'https:' ? 'https:' : 'http:';
    const host = window.location.hostname;
    const apiPort = import.meta.env.VITE_HUD_API_PORT || '8000';
    return `${protocol}//${host}:${apiPort}`;
  }

  return 'http://localhost:8000';
};

const API_BASE = resolveApiBase();
const TRUST_API_BASE = API_BASE;

function App() {
  const [mode, setMode] = useState('network-hud');

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
  const [opsEvents, setOpsEvents] = useState([]);
  const [opsStreamStatus, setOpsStreamStatus] = useState('disconnected');
  const [enclaveActionMessage, setEnclaveActionMessage] = useState('');
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState(null);

  const avgFlDuration = Number(metricsSummary?.avg_fl_duration);
  const totalStake = Number(metricsSummary?.total_stake);
  const cxlUtilization = Number(metricsSummary?.cxl_utilization);

  const privacyTrainingMetrics = useMemo(() => {
    const rounds = Array.isArray(metricsSummary?.convergence?.rounds) ? metricsSummary.convergence.rounds : [];
    const accuracies = Array.isArray(metricsSummary?.convergence?.accuracies) ? metricsSummary.convergence.accuracies : [];
    const losses = Array.isArray(metricsSummary?.convergence?.losses) ? metricsSummary.convergence.losses : [];

    if (rounds.length === 0 && accuracies.length === 0 && losses.length === 0) {
      return [];
    }

    const count = Math.max(rounds.length, accuracies.length, losses.length);
    return Array.from({ length: count }, (_, idx) => ({
      round: Number.isFinite(Number(rounds[idx])) ? Number(rounds[idx]) : idx,
      accuracy: Number.isFinite(Number(accuracies[idx])) ? Number(accuracies[idx]) : 0,
      loss: Number.isFinite(Number(losses[idx])) ? Number(losses[idx]) : 0,
      privacy_overhead: 0,
      compression_ratio: Number.isFinite(cxlUtilization) && cxlUtilization > 0 ? Number((1 / cxlUtilization).toFixed(2)) : 1
    }));
  }, [metricsSummary, cxlUtilization]);

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
          setOpsEvents(Array.isArray(recentPayload?.events) ? recentPayload.events : []);
        }

        if (cancelled) {
          return;
        }

        stream = new EventSource(`${API_BASE}/ops/events`);
        stream.onopen = () => setOpsStreamStatus('connected');
        stream.onerror = () => setOpsStreamStatus('degraded');
        stream.onmessage = (evt) => {
          try {
            const payload = JSON.parse(evt.data);
            if (payload?.kind === 'heartbeat') {
              return;
            }
            setOpsEvents((prev) => {
              const next = [...prev, payload];
              return next.slice(-220);
            });
          } catch (parseErr) {
            console.warn('Failed to parse ops event', parseErr);
          }
        };
      } catch (streamErr) {
        console.warn('Ops event stream unavailable:', streamErr);
        setOpsStreamStatus('disconnected');
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
      const [hudFetch, healthFetch, metricsFetch, foundersFetch, trainingFetch, opsHealthFetch] = await Promise.allSettled([
        fetch(`${API_BASE}/hud_data`),
        fetch(`${API_BASE}/health`),
        fetch(`${API_BASE}/metrics_summary`),
        fetch(`${API_BASE}/founders`),
        fetch(`${API_BASE}/training/status`),
        fetch(`${API_BASE}/ops/health`)
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

      const [hud, healthData, metrics, foundersData, trainingData, opsHealthData] = await Promise.all([
        safeJson(toResponse(hudFetch), hudData || {}),
        safeJson(toResponse(healthFetch), health || {}),
        safeJson(toResponse(metricsFetch), metricsSummary || {}),
        safeJson(toResponse(foundersFetch), founders || []),
        safeJson(toResponse(trainingFetch), { status: 'idle', active: false, round: 0 }),
        safeJson(toResponse(opsHealthFetch), opsHealth || null)
      ]);

      setHudData(hud);
      setHealth(healthData);
      setMetricsSummary(metrics);
      setFounders(foundersData);
      setTrainingStatus(trainingData || { status: 'idle', active: false, round: 0 });
      setOpsHealth(opsHealthData);

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

  const startTraining = async () => {
    try {
      const response = await fetch(`${API_BASE}/training/start`, { method: 'POST' });
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
          const candidate = await fetch(endpoint, {
            method: 'POST',
            headers,
            body: requestBody
          });
          response = candidate;
          if (candidate.ok) {
            break;
          }
        } catch {
          // Try the next candidate endpoint.
        }
      }

      if (!response) {
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
      setPolicyMessage('Verification policy update failed: backend unreachable');
    }
  };

  return (
    <div className="hud-container">
      <header className="hud-header">
        <div>
          <h1>Sovereign Map Federated Interface</h1>
          <p>
            Use the live network operations HUD as the primary command deck, with a dedicated
            privacy-utility analytics view sourced from backend convergence data.
          </p>
        </div>

        <div className="status-bar">
          {error ? <span className="status-error">{error}</span> : <span className="status-ok">Network Reachable</span>}
          {loading && <span className="status-loading"> Syncing...</span>}
        </div>
      </header>

      <nav className="mode-nav" aria-label="View selector">
        <button
          className={mode === 'network-hud' ? 'active' : ''}
          onClick={() => setMode('network-hud')}
          type="button"
        >
          Network Operations HUD
        </button>
        <button
          className={mode === 'privacy-dashboard' ? 'active' : ''}
          onClick={() => setMode('privacy-dashboard')}
          type="button"
        >
          Privacy-Utility Analysis
        </button>
      </nav>

      {mode === 'privacy-dashboard' ? (
        <PrivacyUtilityDashboard trainingMetrics={privacyTrainingMetrics} trainingMode="real" />
      ) : (
        <>
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
            opsEvents={opsEvents}
            opsStreamStatus={opsStreamStatus}
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
        </>
      )}
    </div>
  );
}

export default App;

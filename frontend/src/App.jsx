import { useState, useEffect } from 'react';
import HUD from './HUD';
import BrowserFLDemo from './BrowserFLDemo';
import './App.css';
import PrivacyUtilityDashboard from './PrivacyUtilityDashboard';

const API_BASE = import.meta.env.VITE_HUD_API_BASE || (import.meta.env.DEV ? '/backend' : 'http://localhost:8000');
const TRUST_API_BASE = API_BASE;

function App() {
  const [mode, setMode] = useState('browser-demo');

  const [hudData, setHudData] = useState(null);
  const [health, setHealth] = useState(null);
  const [metricsSummary, setMetricsSummary] = useState(null);
  const [trustStatus, setTrustStatus] = useState(null);
  const [policyHistory, setPolicyHistory] = useState([]);
  const [founders, setFounders] = useState([]);
  const [voiceQuery, setVoiceQuery] = useState('');
  const [voiceResponse, setVoiceResponse] = useState('');
  const [trainingMetrics, setTrainingMetrics] = useState([]);
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
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState(null);

  const avgFlDuration = Number(metricsSummary?.avg_fl_duration);
  const totalStake = Number(metricsSummary?.total_stake);
  const cxlUtilization = Number(metricsSummary?.cxl_utilization);

  useEffect(() => {
    fetchData();
    const interval = setInterval(fetchData, 5000);
    return () => clearInterval(interval);
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
      const [hudRes, healthRes, metricsRes, foundersRes] = await Promise.all([
        fetch(`${API_BASE}/hud_data`),
        fetch(`${API_BASE}/health`),
        fetch(`${API_BASE}/metrics_summary`),
        fetch(`${API_BASE}/founders`)
      ]);

      const [hud, healthData, metrics, foundersData] = await Promise.all([
        hudRes.json(),
        healthRes.json(),
        metricsRes.json(),
        foundersRes.json()
      ]);

      setHudData(hud);
      setHealth(healthData);
      setMetricsSummary(metrics);
      setFounders(foundersData);

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

  const createEnclave = async () => {
    try {
      const response = await fetch(`${API_BASE}/create_enclave`, { method: 'POST' });
      if (response.ok) {
        fetchData();
      }
    } catch (err) {
      console.error('Enclave creation error:', err);
      setError('Enclave creation failed');
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

      const response = await fetch(`${TRUST_API_BASE}/verification_policy`, {
        method: 'POST',
        headers,
        body: JSON.stringify({
          ...policyDraft,
          min_confidence_bps: Number(policyDraft.min_confidence_bps) || 0
        })
      });

      if (!response.ok) {
        const failure = await response.text();
        setPolicyMessage(`Policy update failed: ${failure}`);
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
      setPolicyMessage('Verification policy update failed: backend unreachable');
    }
  };

  return (
    <div className="hud-container">
      <header className="hud-header">
        <div>
          <h1>Sovereign Map Federated Interface</h1>
          <p>
            Switch between the live network operations HUD and the in-browser WebGPU-focused
            federated learning simulator.
          </p>
        </div>

        <div className="status-bar">
          {error ? <span className="status-error">{error}</span> : <span className="status-ok">Network Reachable</span>}
          {loading && <span className="status-loading"> Syncing...</span>}
        </div>
      </header>

      <nav className="mode-nav" aria-label="View selector">
        <button
          className={mode === 'browser-demo' ? 'active' : ''}
          onClick={() => setMode('browser-demo')}
          type="button"
        >
          Browser FL Demo
        </button>
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

      {mode === 'browser-demo' ? (
        <BrowserFLDemo onMetricsUpdate={setTrainingMetrics} />
      ) : mode === 'privacy-dashboard' ? (
        <PrivacyUtilityDashboard trainingMetrics={trainingMetrics} trainingMode="real" />
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
            loading={loading}
            error={error}
            onTriggerFLRound={triggerFLRound}
            onCreateEnclave={createEnclave}
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

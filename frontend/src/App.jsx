import { useState, useEffect } from 'react'
import HUD from './HUD'
import './App.css'

const API_BASE = 'http://localhost:8000'
const TRUST_API_BASE = import.meta.env.VITE_TRUST_API_BASE || 'http://localhost:8082/api/v1'

function App() {
  const [hudData, setHudData] = useState(null)
  const [health, setHealth] = useState(null)
  const [metricsSummary, setMetricsSummary] = useState(null)
  const [trustStatus, setTrustStatus] = useState(null)
  const [founders, setFounders] = useState([])
  const [voiceQuery, setVoiceQuery] = useState('')
  const [voiceResponse, setVoiceResponse] = useState('')
  const [policyDraft, setPolicyDraft] = useState({
    require_proof: false,
    min_confidence_bps: 7000,
    reject_on_verification_failure: false,
    allow_consensus_proof: true,
    allow_zk_proof: true,
    allow_tee_proof: true,
  })
  const [policyInitialized, setPolicyInitialized] = useState(false)
  const [policyToken, setPolicyToken] = useState('')
  const [policyRole, setPolicyRole] = useState('admin')
  const [policyMessage, setPolicyMessage] = useState('')
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState(null)

  useEffect(() => {
    fetchData()
    const interval = setInterval(fetchData, 5000)
    return () => clearInterval(interval)
  }, [])

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
	    const trustRes = await fetch(`${TRUST_API_BASE}/trust_status`)
	    if (trustRes.ok) {
	      const trustData = await trustRes.json()
	      setTrustStatus(trustData)
	      if (!policyInitialized && trustData.verification_policy) {
	        setPolicyDraft({ ...trustData.verification_policy })
	        setPolicyInitialized(true)
	      }
	    }
	  } catch (trustErr) {
	    console.warn('Trust status unavailable:', trustErr)
	  }

      setError(null);
    } catch (err) {
      console.error("Fetch error:", err);
      setError("Failed to sync with Sovereign Node");
    } finally {
      setLoading(false);
    }
  };

  // Logic to trigger a Federated Learning round from the UI
  const triggerFLRound = async () => {
    try {
      const response = await fetch(`${API_BASE}/trigger_fl`, { method: 'POST' });
      if (response.ok) {
        alert("Federated Learning Round Started!");
        fetchData(); 
      }
    } catch (err) {
      console.error("Trigger FL round error:", err);
      setError("Failed to trigger FL round");
    }
  };

  // Logic to create a Secure Enclave
  const createEnclave = async () => {
    try {
      const response = await fetch(`${API_BASE}/create_enclave`, { method: 'POST' });
      if (response.ok) {
        alert("Secure Enclave Initialized");
        fetchData();
      }
    } catch (err) {
      console.error("Enclave creation error:", err);
      setError("Enclave creation failed");
    }
  };

  const submitVoiceQuery = async () => {
    if (!voiceQuery.trim()) {
      return;
    }

    try {
      const response = await fetch(`${API_BASE}/status`);
      if (response.ok) {
        const status = await response.json();
        setVoiceResponse(`System ${status.status}. Flower on ${status.flower_server_port}, Metrics API on ${status.metrics_api_port}.`);
      } else {
        setVoiceResponse("Unable to retrieve system status");
      }
    } catch (err) {
      console.error("Voice query error:", err);
      setVoiceResponse("Voice query failed: backend unreachable");
    }
  };

  const updatePolicyField = (field, value) => {
    setPolicyDraft(current => ({ ...current, [field]: value }))
  }

  const submitVerificationPolicy = async () => {
    try {
      setPolicyMessage('Submitting verification policy...')
      const headers = {
        'Content-Type': 'application/json',
        'X-API-Role': policyRole || 'admin',
      }
      if (policyToken.trim()) {
        headers.Authorization = `Bearer ${policyToken.trim()}`
      }

      const response = await fetch(`${TRUST_API_BASE}/verification_policy`, {
        method: 'POST',
        headers,
        body: JSON.stringify({
          ...policyDraft,
          min_confidence_bps: Number(policyDraft.min_confidence_bps) || 0,
        }),
      })

      if (!response.ok) {
        const failure = await response.text()
        setPolicyMessage(`Policy update failed: ${failure}`)
        return
      }

      const result = await response.json()
      setTrustStatus(current => ({
        ...current,
        trust_mode: 'p2p-reputation+governed-proof-verification',
        verification_policy: result.verification_policy,
        fl_verification: result.fl_verification,
      }))
      setPolicyDraft({ ...result.verification_policy })
      setPolicyMessage('Verification policy updated')
    } catch (err) {
      console.error('Verification policy update error:', err)
      setPolicyMessage('Verification policy update failed: backend unreachable')
    }
  }

  return (
    <div className="hud-container">
      <header className="hud-header">
        <h1>🗺️ Sovereign Map - Federated Learning HUD</h1>
        <div className="status-bar">
          {error ? (
            <span className="status-error">⚠️ {error}</span>
          ) : (
            <span className="status-ok">✅ System Online</span>
          )}
          {loading && <span className="status-loading"> 🔄 Syncing...</span>}
        </div>
      </header>

      <HUD 
        hudData={hudData} 
        health={health} 
        metricsSummary={metricsSummary} 
        trustStatus={trustStatus}
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
            <strong>Avg FL Duration:</strong> {metricsSummary.avg_fl_duration?.toFixed(2) || 'N/A'}s
          </div>
          <div className="summary-item">
            <strong>Total Stake:</strong> {metricsSummary.total_stake?.toFixed(2) || 'N/A'}
          </div>
          <div className="summary-item">
            <strong>CXL Utilization:</strong> {(metricsSummary.cxl_utilization * 100)?.toFixed(1) || 'N/A'}%
          </div>
        </div>
      )}
    </div>
  );
}

export default App;

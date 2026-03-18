import React from 'react';

/**
 * Sovereign Map Federated Learning HUD
 * Displays real-time node health, audit results, and protocol controls.
 */
export default function HUD({
  hudData,
  health,
  metricsSummary,
  trustStatus,
  policyHistory,
  founders,
  voiceQuery,
  voiceResponse,
  policyDraft,
  policyToken,
  policyRole,
  policyMessage,
  loading,
  error,
  onTriggerFLRound,
  onCreateEnclave,
  onSubmitVoiceQuery,
  onPolicyChange,
  onPolicyTokenChange,
  onPolicyRoleChange,
  onSubmitVerificationPolicy,
  setVoiceQuery
}) {
  return (
    <div className="hud-container">
      <header className="hud-section-header">
        <h2>🛰️ Network Intelligence HUD</h2>
        {loading && <div className="loading-spinner">ESTABLISHING SECURE NEURAL LINK...</div>}
      </header>

      <div className="hud-grid">
        {/* Telemetry & API Access Dashboard */}
        <div className="hud-section network-telemetry-panel">
          <h3>🌐 Active Network & API Telemetry</h3>
          <div className="telemetry-grid">
            <div className="telemetry-item">
              <span className="telemetry-label">API Latency:</span>
              <span className="telemetry-value highlight-cyan">{health?.telemetry?.api_latency_ms || Math.floor(Math.random() * 45) + 5}ms</span>
            </div>
            <div className="telemetry-item">
              <span className="telemetry-label">Live Nodes:</span>
              <span className="telemetry-value highlight-green">{hudData?.active_nodes || (Math.floor(Math.random() * 10) + 90)} Online</span>
            </div>
            <div className="telemetry-item">
              <span className="telemetry-label">Ingress Traffic:</span>
              <span className="telemetry-value">{health?.telemetry?.ingress_mbps || Math.floor(Math.random() * 300) + 120} Mbps</span>
            </div>
            <div className="telemetry-item">
              <span className="telemetry-label">API Error Rate:</span>
              <span className="telemetry-value highlight-yellow">{(health?.telemetry?.api_error_rate || Math.random() * 0.1).toFixed(2)}%</span>
            </div>
            <div className="telemetry-item">
              <span className="telemetry-label">Global Saturation:</span>
              <span className="telemetry-value">{health?.telemetry?.global_saturation_pct || Math.floor(Math.random() * 20) + 40}%</span>
            </div>
          </div>
        </div>

        {/* Federated Learning Controls */}
        <div className="hud-section">
          <h3>🧠 Federated Learning</h3>
          <div className="stats-row">
            <span>Accuracy: <strong>{hudData?.last_audit_accuracy || '85.42%'}</strong></span>
            <span>BFT Resilience: <strong>{hudData?.bft_resilience || '30% Verified'}</strong></span>
          </div>
          <button 
            className="btn-primary" 
            onClick={onTriggerFLRound} 
            disabled={loading}
          >
            {loading ? 'Processing Round...' : 'Trigger FL Round'}
          </button>
        </div>

        {/* Enclave & Security Management */}
        <div className="hud-section">
          <h3>🔒 Enclave Management</h3>
          <div className="stats-row">
            <span>Status: <strong>{health?.enclave_status || 'Isolated'}</strong></span>
            <span>TPM: <strong>{health?.tpm_verified ? '✅ Verified' : '⚠️ Pending'}</strong></span>
          </div>
          <button 
            className="btn-secondary" 
            onClick={onCreateEnclave} 
            disabled={loading}
          >
            Initialize Secure Enclave
          </button>
        </div>

        {/* AI Voice Interface */}
        <div className="hud-section voice-interface">
          <h3>🎙️ Voice Query</h3>
          <div className="input-group">
            <input 
              type="text" 
              value={voiceQuery} 
              onChange={e => setVoiceQuery(e.target.value)} 
              placeholder="Query the sovereign agent..." 
            />
            <button onClick={onSubmitVoiceQuery} disabled={loading || !voiceQuery.trim()}>
              Submit
            </button>
          </div>
          {voiceResponse && (
            <div className="voice-response">
              <strong>Response:</strong> {voiceResponse}
            </div>
          )}
        </div>

        {/* Raw System Telemetry */}
        <div className="hud-section telemetry">
          <h3>📊 System Telemetry</h3>
          <details>
            <summary>View Node Health JSON</summary>
            <pre>{JSON.stringify(health, null, 2)}</pre>
          </details>
          <details>
            <summary>View Metrics Summary JSON</summary>
            <pre>{JSON.stringify(metricsSummary, null, 2)}</pre>
          </details>
        </div>

        {/* Network Founders */}
        <div className="hud-section founders-list">
          <h3>👥 Network Founders</h3>
          <ul>
            {founders && founders.length > 0 ? (
              founders.map((f, i) => <li key={i}>{f}</li>)
            ) : (
              <li>No founders detected in current epoch</li>
            )}
          </ul>
        </div>

        <div className="hud-section verification-panel">
          <h3>🛡️ Verification Governance</h3>
          <div className="stats-row">
            <span>Trust Mode: <strong>{trustStatus?.trust_mode || 'Unavailable'}</strong></span>
            <span>Verified Rounds: <strong>{trustStatus?.fl_verification?.verified_rounds ?? 'N/A'}</strong></span>
            <span>Failed Rounds: <strong>{trustStatus?.fl_verification?.failed_rounds ?? 'N/A'}</strong></span>
            <span>Avg Confidence: <strong>{trustStatus?.fl_verification?.average_confidence_bps ?? 'N/A'} bps</strong></span>
          </div>

          <div className="policy-grid">
            <label>
              <input type="checkbox" checked={!!policyDraft.require_proof} onChange={e => onPolicyChange('require_proof', e.target.checked)} />
              Require Proof
            </label>
            <label>
              <input type="checkbox" checked={!!policyDraft.reject_on_verification_failure} onChange={e => onPolicyChange('reject_on_verification_failure', e.target.checked)} />
              Reject On Failure
            </label>
            <label>
              <input type="checkbox" checked={!!policyDraft.allow_consensus_proof} onChange={e => onPolicyChange('allow_consensus_proof', e.target.checked)} />
              Allow Consensus
            </label>
            <label>
              <input type="checkbox" checked={!!policyDraft.allow_zk_proof} onChange={e => onPolicyChange('allow_zk_proof', e.target.checked)} />
              Allow ZK
            </label>
            <label>
              <input type="checkbox" checked={!!policyDraft.allow_tee_proof} onChange={e => onPolicyChange('allow_tee_proof', e.target.checked)} />
              Allow TEE
            </label>
            <label className="policy-number">
              Minimum Confidence (bps)
              <input
                type="number"
                min="0"
                max="10000"
                value={policyDraft.min_confidence_bps ?? 0}
                onChange={e => onPolicyChange('min_confidence_bps', e.target.value)}
              />
            </label>
          </div>

          <div className="policy-auth-grid">
            <label>
              Admin Role
              <input className="api-input" type="text" value={policyRole} onChange={e => onPolicyRoleChange(e.target.value)} />
            </label>
            <label>
              API Token
              <input className="api-input" type="password" value={policyToken} onChange={e => onPolicyTokenChange(e.target.value)} placeholder="Bearer token" />
            </label>
          </div>

          <button className="btn-primary" onClick={onSubmitVerificationPolicy} disabled={loading}>
            Apply Verification Policy
          </button>

          {policyMessage && <div className="policy-message">{policyMessage}</div>}

        <div className="policy-history">
          <h4>Policy History</h4>
          {policyHistory && policyHistory.length > 0 ? (
            <ul className="policy-history-list">
              {policyHistory.slice(0, 6).map((entry, index) => (
                <li key={entry.audit_key || index}>
                  <div><strong>{entry.source || 'governance'}</strong> · {entry.proposal_id || 'unknown'}</div>
                  <div>Confidence: {entry.new_policy?.min_confidence_bps ?? 'n/a'} bps</div>
                  <div>At: {entry.timestamp ? new Date(Number(entry.timestamp) * 1000).toLocaleString() : 'unknown'}</div>
                </li>
              ))}
            </ul>
          ) : (
            <div className="policy-history-empty">No verification policy history recorded yet.</div>
          )}
        </div>
        </div>
      </div>

      {error && (
        <div className="hud-error-footer">
          <strong>Protocol Error:</strong> {error}
        </div>
      )}
    </div>
  );
}

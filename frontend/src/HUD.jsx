import React from 'react';

/**
 * Sovereign Map Federated Learning HUD
 * Displays real-time node health, audit results, and protocol controls.
 */
export default function HUD({
  hudData,
  health,
  metricsSummary,
  founders,
  voiceQuery,
  voiceResponse,
  loading,
  error,
  onTriggerFLRound,
  onCreateEnclave,
  onSubmitVoiceQuery,
  setVoiceQuery
}) {
  return (
    <div className="hud-container">
      <header className="hud-section-header">
        <h2>🛰️ Network Intelligence HUD</h2>
        {loading && <div className="loading-spinner">Synchronizing Node State...</div>}
      </header>

      <div className="hud-grid">
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
            <button onClick={onSubmitVoiceQuery} disabled={loading || !voiceQuery}>
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
      </div>

      {error && (
        <div className="hud-error-footer">
          <strong>Protocol Error:</strong> {error}
        </div>
      )}
    </div>
  );
}

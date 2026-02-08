import React from 'react';

// HUD component to display all function controls and status
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
      <h2>System HUD</h2>
      <div className="hud-section">
        <h3>Federated Learning</h3>
        <button onClick={onTriggerFLRound} disabled={loading}>Trigger FL Round</button>
      </div>
      <div className="hud-section">
        <h3>Enclave Management</h3>
        <button onClick={onCreateEnclave} disabled={loading}>Create Enclave</button>
      </div>
      <div className="hud-section">
        <h3>Voice Query</h3>
        <input
          type="text"
          value={voiceQuery}
          onChange={e => setVoiceQuery(e.target.value)}
          placeholder="Enter voice query"
        />
        <button onClick={onSubmitVoiceQuery} disabled={loading}>Submit Voice Query</button>
        <div>Response: {voiceResponse}</div>
      </div>
      <div className="hud-section">
        <h3>Health</h3>
        <pre>{JSON.stringify(health, null, 2)}</pre>
      </div>
      <div className="hud-section">
        <h3>Metrics Summary</h3>
        <pre>{JSON.stringify(metricsSummary, null, 2)}</pre>
      </div>
      <div className="hud-section">
        <h3>Founders</h3>
        <ul>
          {founders.map((f, i) => <li key={i}>{f}</li>)}
        </ul>
      </div>
      {error && <div className="hud-error">Error: {error}</div>}
    </div>
  );
}

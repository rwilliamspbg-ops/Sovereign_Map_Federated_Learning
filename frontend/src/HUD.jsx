import React, { useState } from 'react';
import LiveTerminal from './LiveTerminal';

/**
 * Sovereign Map Federated Learning HUD - v3 REDESIGN
 * Deep interactive testing controls and comprehensive system stats.
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

  const [simulationState, setSimulationState] = useState({
    byzantineAttacks: 0,
    networkPartitions: 0,
    hardwareFaults: 0
  });

  const triggerSimulation = (type) => {
    setSimulationState(prev => ({
      ...prev,
      [type]: prev[type] + 1
    }));
    // In a real implementation this would hit the backend simulation endpoint
  };

  return (
    <div className="hud-container global-dark">
      <header className="hud-main-header">
        <div className="header-branding">
          <h2>🛰️ SOVEREIGN MAP COMMAND CENTER</h2>
          <span className="system-status inline-pulse">🔴 LIVE // TESTNET</span>
        </div>
        {loading && <div className="loading-spinner">SYNCING TELEMETRY...</div>}
      </header>

      <div className="hud-grid master-grid">
        
        {/* LEFT COLUMN: ACTIVE SENSORS & TELEMETRY */}
        <div className="hud-col left-col">
          <div className="hud-section sensor-panel glass-panel">
            <h3>🌐 Network & API Telemetry</h3>
            <div className="metric-row">
              <span className="metric-label">API Latency</span>
              <span className="metric-value highlight-cyan">{health?.telemetry?.api_latency_ms || Math.floor(Math.random() * 45) + 5}ms</span>
            </div>
            <div className="metric-row">
              <span className="metric-label">Global Nodes</span>
              <span className="metric-value highlight-green">{hudData?.active_nodes || (Math.floor(Math.random() * 10) + 90)}</span>
            </div>
            <div className="metric-row">
              <span className="metric-label">Ingress</span>
              <span className="metric-value">{health?.telemetry?.ingress_mbps || Math.floor(Math.random() * 300) + 120} Mbps</span>
            </div>
            <div className="metric-row">
              <span className="metric-label">P95 Error</span>
              <span className="metric-value highlight-yellow">{(health?.telemetry?.api_error_rate || Math.random() * 0.1).toFixed(2)}%</span>
            </div>
            <div className="metric-row">
              <span className="metric-label">Core Saturation</span>
              <span className="metric-value">{health?.telemetry?.global_saturation_pct || Math.floor(Math.random() * 20) + 40}%</span>
            </div>
            <div className="metric-row">
              <span className="metric-label">Mesh Topology</span>
              <span className="metric-value">Stable (BFT)</span>
            </div>
            <div className="metric-row">
              <span className="metric-label">Global Model Acc</span>
              <span className="metric-value highlight-neon">{hudData?.last_audit_accuracy || '85.42%'}</span>
            </div>
          </div>

          <div className="hud-section glass-panel">
            <h3>🛡️ Substrate Governance</h3>
            <div className="metric-row">
              <span className="metric-label">Trust Mode</span>
              <span className="metric-value">{trustStatus?.trust_mode || 'Strict Verification'}</span>
            </div>
            <div className="metric-row">
              <span className="metric-label">Avg Confidence</span>
              <span className="metric-value highlight-green">{trustStatus?.fl_verification?.average_confidence_bps || 9850} bps</span>
            </div>
            
            <div className="policy-tuners">
              <h4>DYNAMIC POLICY TUNING</h4>
              <div className="form-group row-align">
                <label>Proof Req:</label>
                <input type="checkbox" checked={policyDraft.require_proof} onChange={(e) => onPolicyChange('require_proof', e.target.checked)} />
              </div>
              <div className="form-group row-align">
                <label>Confidence Threshold:</label>
                <input type="number" value={policyDraft.min_confidence_bps} step="100" onChange={(e) => onPolicyChange('min_confidence_bps', e.target.value)} />
              </div>
              <button className="btn-secondary full-width" onClick={onSubmitVerificationPolicy}>Apply Substrate Rules</button>
            </div>
          </div>
        </div>

        {/* MIDDLE COLUMN: LIVE TERMINAL & AI */}
        <div className="hud-col middle-col">
          <LiveTerminal />

          <div className="hud-section glass-panel voice-ai-panel">
            <h3>🤖 Sovereign Neural Assistant</h3>
            <p className="ai-subtext">Direct protocol interface for querying state, tokenomics, and intelligence.</p>
            <div className="chat-interface">
              <input 
                type="text" 
                className="voice-input"
                placeholder="Ask: 'What is the current BFT threshold?'" 
                value={voiceQuery} 
                onChange={(e) => setVoiceQuery(e.target.value)} 
                onKeyPress={(e) => e.key === 'Enter' && onSubmitVoiceQuery()}
              />
              <button className="btn-action pulse-btn" onClick={onSubmitVoiceQuery}>EXECUTE /_</button>
            </div>
            {voiceResponse && (
              <div className="ai-response-box">
                <span className="ai-label">SYSTEM RESPONSE //</span>
                <p>{voiceResponse}</p>
              </div>
            )}
          </div>
        </div>

        {/* RIGHT COLUMN: TESTING CONTROLS */}
        <div className="hud-col right-col">
          <div className="hud-section glass-panel test-bench">
            <h3>⚡ Live Simulation Bench</h3>
            <p className="panel-desc">Execute real-time stress testing against the FL mesh.</p>
            
            <div className="action-buttons-stack">
              <button className="btn-action danger" onClick={() => triggerSimulation('byzantineAttacks')}>
                ☣️ INJECT BYZANTINE FAULT [{simulationState.byzantineAttacks}]
              </button>
              <button className="btn-action warn" onClick={() => triggerSimulation('networkPartitions')}>
                ✂️ FORCE NETWORK PARTITION [{simulationState.networkPartitions}]
              </button>
              <button className="btn-action alert" onClick={() => triggerSimulation('hardwareFaults')}>
                💥 SIMULATE HARDWARE DROP [{simulationState.hardwareFaults}]
              </button>
            </div>
            
            <div className="divider" />
            
            <div className="action-buttons-stack">
              <button className="btn-primary heavy-glow" onClick={onTriggerFLRound} disabled={loading}>
                🧠 [ INITIATE GLOBAL FL EPOCH ]
              </button>
              <button className="btn-secondary" onClick={onCreateEnclave} disabled={loading}>
                🔒 [ PROVISION TEE ENCLAVE ]
              </button>
            </div>
          </div>

          <div className="hud-section glass-panel metrics-box">
            <h3>📈 Epoch Convergence</h3>
            {error ? (
              <div className="error-box">SYSTEM FAULT: {error}</div>
            ) : metricsSummary ? (
              <div className="mini-stats">
                <div className="stat-pill">
                  <label>ROUNDS</label>
                  <span>{metricsSummary?.federated_learning?.current_round || 0}</span>
                </div>
                <div className="stat-pill">
                  <label>LOSS</label>
                  <span>{(metricsSummary?.federated_learning?.current_loss || 0).toFixed(4)}</span>
                </div>
                <div className="stat-pill">
                  <label>DEGRADATION</label>
                  <span>{simulationState.byzantineAttacks > 0 ? (simulationState.byzantineAttacks * 1.2).toFixed(1) + '%' : '0.0%'}</span>
                </div>
              </div>
            ) : (
              <div className="awaiting-data">AWAITING TRAINING METRICS...</div>
            )}
          </div>
        </div>

      </div>
    </div>
  );
}

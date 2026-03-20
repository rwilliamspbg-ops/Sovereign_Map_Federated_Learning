import React, { useEffect, useState } from 'react';
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
  trainingStatus,
  opsHealth,
  opsEvents,
  opsStreamStatus,
  loading,
  error,
  onTriggerFLRound,
  onStartTraining,
  onStopTraining,
  onCreateEnclave,
  onTriggerSimulation,
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

  useEffect(() => {
    if (hudData?.simulation_counters) {
      setSimulationState((prev) => ({ ...prev, ...hudData.simulation_counters }));
    }
  }, [hudData]);

  const triggerSimulation = async (type) => {
    setSimulationState(prev => ({
      ...prev,
      [type]: prev[type] + 1
    }));
    if (onTriggerSimulation) {
      await onTriggerSimulation(type);
    }
  };

  const recentOpsEvents = (opsEvents || []).slice(-8).reverse();
  const healthStatus = opsHealth?.status || 'unknown';
  const streamLabel = opsStreamStatus || 'disconnected';

  return (
    <div className="hud-container global-dark">
      <header className="hud-main-header">
        <div className="header-branding">
          <h2>SOVEREIGN MAP COMMAND CENTER</h2>
          <span className="system-status inline-pulse">LIVE OPS // LINUX CONTROL PLANE</span>
        </div>
        {loading && <div className="loading-spinner">SYNCING TELEMETRY...</div>}
      </header>

      <section className="ops-ribbon glass-panel">
        <div className="ops-chip">
          <span className="ops-chip-label">System</span>
          <span className={`ops-chip-value ${healthStatus === 'critical' ? 'chip-critical' : healthStatus === 'degraded' ? 'chip-warning' : 'chip-ok'}`}>
            {healthStatus.toUpperCase()}
          </span>
        </div>
        <div className="ops-chip">
          <span className="ops-chip-label">Event Stream</span>
          <span className={`ops-chip-value ${streamLabel === 'connected' ? 'chip-ok' : streamLabel === 'connecting' ? 'chip-warning' : 'chip-critical'}`}>
            {streamLabel.toUpperCase()}
          </span>
        </div>
        <div className="ops-chip">
          <span className="ops-chip-label">CPU Load (1m)</span>
          <span className="ops-chip-value">{opsHealth?.system?.load_1m ?? 'N/A'}</span>
        </div>
        <div className="ops-chip">
          <span className="ops-chip-label">Memory Used</span>
          <span className="ops-chip-value">{opsHealth?.system?.memory?.used_percent ?? 'N/A'}%</span>
        </div>
        <div className="ops-chip">
          <span className="ops-chip-label">Disk Used</span>
          <span className="ops-chip-value">{opsHealth?.system?.disk?.used_percent ?? 'N/A'}%</span>
        </div>
        <div className="ops-chip">
          <span className="ops-chip-label">Ports</span>
          <span className="ops-chip-value">
            API:{opsHealth?.ports?.api_8000 ? 'up' : 'down'} | FL:{opsHealth?.ports?.flower_8080 ? 'up' : 'down'} | PROM:{opsHealth?.ports?.prometheus_9090 ? 'up' : 'down'}
          </span>
        </div>
      </section>

      <div className="hud-grid master-grid">
        
        {/* LEFT COLUMN: ACTIVE SENSORS & TELEMETRY */}
        <div className="hud-col left-col">
          <div className="hud-section sensor-panel glass-panel">
            <h3>🌐 Network & API Telemetry</h3>
            <div className="metric-row">
              <span className="metric-label">API Latency</span>
              <span className="metric-value highlight-cyan">{health?.telemetry?.api_latency_ms ?? 'N/A'}ms</span>
            </div>
            <div className="metric-row">
              <span className="metric-label">Global Nodes</span>
              <span className="metric-value highlight-green">{hudData?.active_nodes ?? 0}</span>
            </div>
            <div className="metric-row">
              <span className="metric-label">Ingress</span>
              <span className="metric-value">{health?.telemetry?.ingress_mbps ?? 'N/A'} Mbps</span>
            </div>
            <div className="metric-row">
              <span className="metric-label">P95 Error</span>
              <span className="metric-value highlight-yellow">{Number(health?.telemetry?.api_error_rate ?? 0).toFixed(2)}%</span>
            </div>
            <div className="metric-row">
              <span className="metric-label">Core Saturation</span>
              <span className="metric-value">{health?.telemetry?.global_saturation_pct ?? 'N/A'}%</span>
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
              <span className="metric-value highlight-green">{trustStatus?.fl_verification?.average_confidence_bps ?? 0} bps</span>
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
          <LiveTerminal events={opsEvents} />

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

          <div className="hud-section glass-panel timeline-panel">
            <h3>Operations Timeline</h3>
            <div className="timeline-list">
              {recentOpsEvents.length === 0 ? (
                <div className="timeline-empty">No live operations events yet.</div>
              ) : (
                recentOpsEvents.map((evt) => (
                  <div key={`${evt.id}-${evt.ts}`} className={`timeline-item severity-${evt.severity || 'info'}`}>
                    <div className="timeline-meta">
                      <span>{new Date((evt.ts || 0) * 1000).toLocaleTimeString()}</span>
                      <span>{(evt.kind || 'event').toUpperCase()}</span>
                    </div>
                    <div className="timeline-message">{evt.message}</div>
                  </div>
                ))
              )}
            </div>
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
              <button className="btn-action" onClick={onStartTraining} disabled={loading || trainingStatus?.active}>
                ▶ START REAL TRAINING LOOP
              </button>
              <button className="btn-action" onClick={onStopTraining} disabled={loading || !trainingStatus?.active}>
                ■ STOP TRAINING LOOP
              </button>
              <button className="btn-primary heavy-glow" onClick={onTriggerFLRound} disabled={loading}>
                🧠 [ INITIATE GLOBAL FL EPOCH ]
              </button>
              <button className="btn-secondary" onClick={onCreateEnclave} disabled={loading}>
                🔒 [ PROVISION TEE ENCLAVE ]
              </button>
            </div>

            <div className="metric-row">
              <span className="metric-label">Training State</span>
              <span className="metric-value highlight-cyan">{trainingStatus?.status || 'idle'} / round {trainingStatus?.round ?? 0}</span>
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

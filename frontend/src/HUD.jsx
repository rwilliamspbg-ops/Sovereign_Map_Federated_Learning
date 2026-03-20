import React, { useEffect, useState } from 'react';
import LiveTerminal from './LiveTerminal';

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
  enclaveActionMessage,
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
  const recentPolicies = (policyHistory || []).slice(-4).reverse();
  const founderList = Array.isArray(founders) ? founders.slice(0, 6) : [];
  const healthStatus = opsHealth?.status || 'unknown';
  const streamLabel = opsStreamStatus || 'disconnected';
  const trustMode = trustStatus?.trust_mode || 'governed-verification';
  const confidencePct = Number(trustStatus?.fl_verification?.average_confidence_bps || 0) / 100;
  const memUsed = Number(opsHealth?.system?.memory?.used_percent || 0);
  const promReachable = Boolean(opsHealth?.dependencies?.prometheus?.reachable ?? opsHealth?.ports?.prometheus_9090);
  const opsAlerts = Array.isArray(opsHealth?.alerts) ? opsHealth.alerts : [];

  const numberOrFallback = (value, suffix = '') => {
    const parsed = Number(value);
    return Number.isFinite(parsed) ? `${parsed}${suffix}` : 'N/A';
  };

  const fixedOrFallback = (value, decimals = 2, suffix = '') => {
    const parsed = Number(value);
    return Number.isFinite(parsed) ? `${parsed.toFixed(decimals)}${suffix}` : 'N/A';
  };

  return (
    <div className="hud-ops-root">
      <header className="hud-ops-header">
        <div>
          <p className="hud-kicker">SovereignMap Operator Console</p>
          <h2>Federated Learning Operating Deck</h2>
          <p className="hud-subline">Unified command surface for training, governance, simulation, and platform telemetry.</p>
        </div>
        <div className="header-right">
          <div className="ops-clock">{new Date().toLocaleString()}</div>
          {loading && <div className="loading-spinner">SYNCING</div>}
        </div>
      </header>

      <section className="ops-ribbon">
        <div className="ops-chip">
          <span className="ops-chip-label">Platform State</span>
          <span className={`ops-chip-value ${healthStatus === 'critical' ? 'chip-critical' : healthStatus === 'degraded' ? 'chip-warning' : 'chip-ok'}`}>
            {healthStatus.toUpperCase()}
          </span>
        </div>
        <div className="ops-chip">
          <span className="ops-chip-label">Event Bus</span>
          <span className={`ops-chip-value ${streamLabel === 'connected' ? 'chip-ok' : streamLabel === 'connecting' ? 'chip-warning' : 'chip-critical'}`}>
            {streamLabel.toUpperCase()}
          </span>
        </div>
        <div className="ops-chip">
          <span className="ops-chip-label">Training</span>
          <span className={`ops-chip-value ${trainingStatus?.active ? 'chip-ok' : 'chip-warning'}`}>
            {trainingStatus?.status || 'idle'} / round {trainingStatus?.round ?? 0}
          </span>
        </div>
        <div className="ops-chip">
          <span className="ops-chip-label">Trust Confidence</span>
          <span className="ops-chip-value">{fixedOrFallback(confidencePct, 2, '%')}</span>
        </div>
        <div className="ops-chip">
          <span className="ops-chip-label">Active Nodes</span>
          <span className="ops-chip-value">{hudData?.active_nodes ?? 0}</span>
        </div>
        <div className="ops-chip">
          <span className="ops-chip-label">Core Ports</span>
          <span className={`ops-chip-value ${promReachable ? '' : 'chip-critical'}`}>
            API:{opsHealth?.ports?.api_8000 ? 'up' : 'down'} | FL:{opsHealth?.ports?.flower_8080 ? 'up' : 'down'} | PROM:{opsHealth?.ports?.prometheus_9090 ? 'up' : 'down'}
          </span>
        </div>
      </section>

      {(healthStatus === 'degraded' || healthStatus === 'critical') && (
        <section className="ops-alert-banner">
          <div>
            <strong>Platform State {healthStatus.toUpperCase()}</strong>
            <span>
              {promReachable ? ' Prometheus reachable.' : ' Prometheus telemetry unavailable.'}
              {' '}
              Memory usage: {fixedOrFallback(memUsed, 2, '%')}.
            </span>
          </div>
          <div className="ops-alert-actions">
            {!promReachable && <a href="#prometheus-remediation">Inspect Prometheus</a>}
            {memUsed >= 88 && <a href="#memory-remediation">Inspect Memory</a>}
          </div>
        </section>
      )}

      <div className="ops-grid">
        <aside className="ops-panel control-panel">
          <h3>Command Orchestration</h3>
          <p className="panel-subtitle">Primary operator controls for federated runtime and enclave lifecycle.</p>
          <div className="button-stack">
            <button className="btn cmd-start" onClick={onStartTraining} disabled={loading || trainingStatus?.active}>Start Training Loop</button>
            <button className="btn cmd-stop" onClick={onStopTraining} disabled={loading || !trainingStatus?.active}>Stop Training Loop</button>
            <button className="btn cmd-epoch" onClick={onTriggerFLRound} disabled={loading}>Run Global FL Epoch</button>
            <button className="btn cmd-enclave" onClick={onCreateEnclave} disabled={loading}>Provision TEE Enclave</button>
          </div>

          {enclaveActionMessage && <div className="notice-box">{enclaveActionMessage}</div>}

          <div className="sim-grid">
            <button className="btn sim-byzantine" onClick={() => triggerSimulation('byzantineAttacks')}>Inject Byzantine Fault ({simulationState.byzantineAttacks})</button>
            <button className="btn sim-network" onClick={() => triggerSimulation('networkPartitions')}>Force Network Partition ({simulationState.networkPartitions})</button>
            <button className="btn sim-hardware" onClick={() => triggerSimulation('hardwareFaults')}>Simulate Hardware Drop ({simulationState.hardwareFaults})</button>
          </div>
        </aside>

        <main className="ops-panel telemetry-panel">
          <h3>Operations Telemetry</h3>
          <div className="kpi-grid">
            <div className="kpi-card">
              <span>API Latency</span>
              <strong>{numberOrFallback(health?.telemetry?.api_latency_ms, ' ms')}</strong>
            </div>
            <div className="kpi-card">
              <span>Ingress</span>
              <strong>{numberOrFallback(health?.telemetry?.ingress_mbps, ' Mbps')}</strong>
            </div>
            <div className="kpi-card">
              <span>Error Rate</span>
              <strong>{fixedOrFallback(health?.telemetry?.api_error_rate, 2, '%')}</strong>
            </div>
            <div className="kpi-card">
              <span>Saturation</span>
              <strong>{numberOrFallback(health?.telemetry?.global_saturation_pct, '%')}</strong>
            </div>
            <div className="kpi-card">
              <span>Model Accuracy</span>
              <strong>{hudData?.last_audit_accuracy || 'N/A'}</strong>
            </div>
            <div className="kpi-card">
              <span>FL Loss</span>
              <strong>{fixedOrFallback(metricsSummary?.federated_learning?.current_loss, 4)}</strong>
            </div>
            <div className="kpi-card">
              <span>CPU Load (1m)</span>
              <strong>{numberOrFallback(opsHealth?.system?.load_1m)}</strong>
            </div>
            <div className="kpi-card">
              <span>Memory Used</span>
              <strong className={memUsed >= 94 ? 'kpi-critical' : memUsed >= 88 ? 'kpi-warning' : ''}>{numberOrFallback(opsHealth?.system?.memory?.used_percent, '%')}</strong>
            </div>
            <div className="kpi-card">
              <span>Disk Used</span>
              <strong>{numberOrFallback(opsHealth?.system?.disk?.used_percent, '%')}</strong>
            </div>
          </div>

          <LiveTerminal events={opsEvents} />

          <div className="aux-panel" id="prometheus-remediation">
            <h4>Prometheus Connectivity</h4>
            <div className="audit-row">
              <span>Reachable</span>
              <span>{promReachable ? 'yes' : 'no'}</span>
            </div>
            <div className="audit-row">
              <span>Health Endpoint</span>
              <span>{opsHealth?.dependencies?.prometheus?.health_url || 'n/a'}</span>
            </div>
            <div className="audit-row">
              <span>Detail</span>
              <span>{opsHealth?.dependencies?.prometheus?.detail || 'n/a'}</span>
            </div>
          </div>

          <div className="aux-panel" id="memory-remediation">
            <h4>Memory Pressure</h4>
            <div className="audit-row">
              <span>Used %</span>
              <span>{fixedOrFallback(memUsed, 2, '%')}</span>
            </div>
            <div className="audit-row">
              <span>Used MB</span>
              <span>{numberOrFallback(opsHealth?.system?.memory?.used_mb, ' MB')}</span>
            </div>
            <div className="audit-row">
              <span>Available MB</span>
              <span>{numberOrFallback(opsHealth?.system?.memory?.available_mb, ' MB')}</span>
            </div>
          </div>

          {opsAlerts.length > 0 && (
            <div className="aux-panel">
              <h4>Remediation Hints</h4>
              {opsAlerts.map((alert, idx) => (
                <div className="notice-box" key={`${alert?.component || 'alert'}-${idx}`}>
                  <strong>{(alert?.component || 'component').toUpperCase()}:</strong> {alert?.message}
                </div>
              ))}
            </div>
          )}

          <div className="timeline-panel">
            <h4>Live Incident Timeline</h4>
            <div className="timeline-list">
              {recentOpsEvents.length === 0 ? (
                <div className="timeline-empty">No recent events from backend stream.</div>
              ) : (
                recentOpsEvents.map((evt) => (
                  <div key={`${evt.id}-${evt.ts}-${evt.kind}`} className={`timeline-item severity-${evt.severity || 'info'}`}>
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
        </main>

        <aside className="ops-panel governance-panel">
          <h3>Governance + Assistant</h3>
          <div className="governance-meta">
            <div><span>Trust Mode</span><strong>{trustMode}</strong></div>
            <div><span>Avg Confidence</span><strong>{fixedOrFallback(confidencePct, 2, '%')}</strong></div>
          </div>

          <div className="policy-form">
            <h4>Verification Policy Controls</h4>
            <label className="toggle-row"><input type="checkbox" checked={!!policyDraft.require_proof} onChange={(e) => onPolicyChange('require_proof', e.target.checked)} /> Require Proof</label>
            <label className="toggle-row"><input type="checkbox" checked={!!policyDraft.reject_on_verification_failure} onChange={(e) => onPolicyChange('reject_on_verification_failure', e.target.checked)} /> Reject on Verification Failure</label>
            <label className="toggle-row"><input type="checkbox" checked={!!policyDraft.allow_consensus_proof} onChange={(e) => onPolicyChange('allow_consensus_proof', e.target.checked)} /> Allow Consensus Proof</label>
            <label className="toggle-row"><input type="checkbox" checked={!!policyDraft.allow_zk_proof} onChange={(e) => onPolicyChange('allow_zk_proof', e.target.checked)} /> Allow ZK Proof</label>
            <label className="toggle-row"><input type="checkbox" checked={!!policyDraft.allow_tee_proof} onChange={(e) => onPolicyChange('allow_tee_proof', e.target.checked)} /> Allow TEE Proof</label>

            <label className="input-row">
              Min Confidence (bps)
              <input type="number" value={policyDraft.min_confidence_bps ?? 7000} step="100" onChange={(e) => onPolicyChange('min_confidence_bps', e.target.value)} />
            </label>

            <label className="input-row">
              Role
              <select value={policyRole} onChange={(e) => onPolicyRoleChange(e.target.value)}>
                <option value="admin">admin</option>
                <option value="operator">operator</option>
                <option value="auditor">auditor</option>
              </select>
            </label>

            <label className="input-row">
              Token
              <input type="password" value={policyToken} placeholder="Bearer token (optional)" onChange={(e) => onPolicyTokenChange(e.target.value)} />
            </label>

            <button className="btn cmd-enclave" onClick={onSubmitVerificationPolicy}>Apply Verification Policy</button>
            {policyMessage && <div className="notice-box">{policyMessage}</div>}
          </div>

          <div className="assistant-panel">
            <h4>Operator Assistant</h4>
            <div className="chat-interface">
              <input
                type="text"
                className="voice-input"
                placeholder="Query runtime state, trust, tokenomics, diagnostics..."
                value={voiceQuery}
                onChange={(e) => setVoiceQuery(e.target.value)}
                onKeyDown={(e) => e.key === 'Enter' && onSubmitVoiceQuery()}
              />
              <button className="btn cmd-epoch" onClick={onSubmitVoiceQuery}>Run Query</button>
            </div>
            {voiceResponse ? <div className="ai-response-box">{voiceResponse}</div> : <div className="timeline-empty">No assistant response yet.</div>}
          </div>

          <div className="aux-panel">
            <h4>Policy Audit Trail</h4>
            {recentPolicies.length === 0 ? (
              <div className="timeline-empty">No policy mutations recorded.</div>
            ) : (
              recentPolicies.map((entry, index) => (
                <div className="audit-row" key={`${entry?.ts || index}-${index}`}>
                  <span>{entry?.ts ? new Date(entry.ts * 1000).toLocaleTimeString() : 'time-unknown'}</span>
                  <span>{entry?.role || 'role-unknown'}</span>
                </div>
              ))
            )}
          </div>

          <div className="aux-panel">
            <h4>Founder Nodes</h4>
            {founderList.length === 0 ? (
              <div className="timeline-empty">Founder registry unavailable.</div>
            ) : (
              founderList.map((founder, index) => (
                <div className="audit-row" key={`${founder?.name || 'founder'}-${index}`}>
                  <span>{founder?.name || `Founder ${index + 1}`}</span>
                  <span>{founder?.stake != null ? `${Number(founder.stake).toFixed(2)} (${founder?.verified ? 'verified' : 'unverified'})` : 'stake N/A'}</span>
                </div>
              ))
            )}
          </div>

          {error && <div className="error-box">System fault: {error}</div>}
        </aside>
      </div>
    </div>
  );
}

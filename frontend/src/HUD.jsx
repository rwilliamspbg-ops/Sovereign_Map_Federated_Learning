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
  opsTrends,
  opsEvents,
  opsStreamStatus,
  connectionDiagnostics,
  webMetrics,
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
  const [compactMode, setCompactMode] = useState(false);
  const [isMobile, setIsMobile] = useState(false);
  const [collapsedClusters, setCollapsedClusters] = useState({
    api: false,
    llm: false,
    network: false
  });
  const [trendSeries, setTrendSeries] = useState({
    apiLatency: [],
    apiError: [],
    ingress: []
  });
  const [simulationState, setSimulationState] = useState({
    byzantineAttacks: 0,
    networkPartitions: 0,
    hardwareFaults: 0
  });
  const [requestedRounds, setRequestedRounds] = useState(10);

  useEffect(() => {
    const updateViewport = () => {
      const mobile = window.innerWidth <= 720;
      setIsMobile(mobile);
      setCollapsedClusters((prev) => {
        if (!mobile) {
          return { api: false, llm: false, network: false };
        }
        if (prev.api || prev.llm || prev.network) {
          return prev;
        }
        return { api: false, llm: true, network: true };
      });
    };

    updateViewport();
    window.addEventListener('resize', updateViewport);
    return () => window.removeEventListener('resize', updateViewport);
  }, []);

  useEffect(() => {
    if (hudData?.simulation_counters) {
      setSimulationState((prev) => ({ ...prev, ...hudData.simulation_counters }));
    }
  }, [hudData]);

  useEffect(() => {
    const target = Number(trainingStatus?.target_rounds || 0);
    if (target > 0) {
      setRequestedRounds(target);
    }
  }, [trainingStatus?.target_rounds]);

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
  const diagnostics = connectionDiagnostics || {};
  const streamLastError = diagnostics.streamLastError || '';
  const policyStatus = diagnostics.policyStatus || 'idle';
  const trustMode = trustStatus?.trust_mode || 'governed-verification';
  const confidencePct = Number(trustStatus?.fl_verification?.average_confidence_bps || 0) / 100;
  const memUsed = Number(opsHealth?.system?.memory?.used_percent || 0);
  const firstFinite = (...values) => {
    for (const value of values) {
      const parsed = Number(value);
      if (Number.isFinite(parsed)) {
        return parsed;
      }
    }
    return 0;
  };
  const latestTrendValue = (series) => {
    if (!Array.isArray(series) || series.length === 0) {
      return null;
    }
    const latest = series[series.length - 1];
    if (latest && typeof latest === 'object') {
      const parsed = Number(latest.value);
      return Number.isFinite(parsed) ? parsed : null;
    }
    const parsed = Number(latest);
    return Number.isFinite(parsed) ? parsed : null;
  };
  const trendLatency = latestTrendValue(opsTrends?.api_latency_ms);
  const trendError = latestTrendValue(opsTrends?.api_error_rate_pct);
  const trendIngress = latestTrendValue(opsTrends?.ingress_mbps);
  const apiLatencyMs = firstFinite(health?.telemetry?.api_latency_ms, opsHealth?.telemetry?.api_latency_ms, trendLatency);
  const apiErrorRatePct = firstFinite(health?.telemetry?.api_error_rate, opsHealth?.telemetry?.api_error_rate, trendError);
  const ingressMbps = firstFinite(health?.telemetry?.ingress_mbps, opsHealth?.telemetry?.ingress_mbps, trendIngress);
  const trainingRound = firstFinite(metricsSummary?.federated_learning?.current_round, trainingStatus?.round, opsHealth?.training?.round);
  const trainingLoss = firstFinite(trainingStatus?.current_metrics?.loss, metricsSummary?.federated_learning?.current_loss);
  const targetRounds = Number(trainingStatus?.target_rounds || 0);
  const remainingRounds = Number(trainingStatus?.remaining_rounds);
  const llmPolicyValid = firstFinite(hudData?.simulation_counters?.llmPolicyValid, opsHealth?.privacy_security?.llm_policy_valid);
  const llmPolicyRejected = firstFinite(hudData?.simulation_counters?.llmPolicyRejected, opsHealth?.privacy_security?.llm_policy_rejected);
  const llmPolicyTotal = llmPolicyValid + llmPolicyRejected;
  const llmPassRatePct = llmPolicyTotal > 0 ? (llmPolicyValid / llmPolicyTotal) * 100 : 100;
  const registryCount = Array.isArray(founders) ? founders.length : 0;
  const verifiedRegistryCount = Array.isArray(founders) ? founders.filter((founder) => founder?.verified).length : 0;
  const listedStakeFromRegistry = Array.isArray(founders)
    ? founders.reduce((acc, founder) => acc + Number(founder?.stake || 0), 0)
    : 0;
  const listedStake = listedStakeFromRegistry > 0
    ? listedStakeFromRegistry
    : firstFinite(metricsSummary?.total_stake, opsHealth?.governance_economics?.total_stake);
  const promReachable = Boolean(opsHealth?.dependencies?.prometheus?.reachable ?? opsHealth?.ports?.prometheus_9090);
  const apiPortUp = Boolean(opsHealth?.ports?.api_8000);
  const flPortUp = Boolean(opsHealth?.ports?.flower_8080);
  const promPortUp = Boolean(opsHealth?.ports?.prometheus_9090);
  const portHealthUpCount = [apiPortUp, flPortUp, promPortUp].filter(Boolean).length;
  const opsAlerts = Array.isArray(opsHealth?.alerts)
    ? opsHealth.alerts.filter((alert) => String(alert?.component || '').toLowerCase() !== 'privacy-budget')
    : [];
  const runbookByComponent = {
    prometheus: {
      section: 'ConsensusStatusEndpointDown',
      owner: 'platform',
      checks: [
        'Check metrics-exporter reachability and Prometheus datasource status.',
        'Validate endpoint availability before triaging dependent warnings.'
      ]
    },
    memory: {
      section: 'FLRoundDurationHighP95',
      owner: 'platform',
      checks: [
        'Inspect memory pressure alongside FL round duration p95 and queue depth.',
        'Reduce concurrency or scale workers before OOM thresholds are reached.'
      ]
    },
    'federated-network': {
      section: 'FLClientParticipationLow',
      owner: 'platform',
      checks: [
        'Inspect active peer count and partition indicators for stragglers.',
        'Trigger controlled client rejoin for degraded participants.'
      ]
    },
    'tee-enclave': {
      section: 'SlowTrustVerification',
      owner: 'tpm',
      checks: [
        'Correlate trust verification p95 with cache miss and lock contention metrics.',
        'Tune attestation cache TTL and nonce mode before scaling enclave load.'
      ]
    },
  };
  const runbookCards = opsAlerts.map((alert, idx) => {
    const key = String(alert?.component || '').toLowerCase();
    const match = runbookByComponent[key] || {
      section: 'FLRoundStalled',
      owner: 'platform',
      checks: ['Use default FL runbook triage and escalate if unresolved in 10 minutes.']
    };
    return {
      id: `${key || 'alert'}-${idx}`,
      component: key || 'unknown',
      message: alert?.message || 'No alert message provided.',
      section: match.section,
      owner: match.owner,
      checks: match.checks
    };
  });
  const browserRtt = Number(webMetrics?.rttMs || 0);
  const browserHeap = Number(webMetrics?.jsHeapUsedMB || 0);
  const epcUtilization = Number(opsHealth?.tee_hardware?.epc_utilization_pct || 0);
  const attestationLatencyMs = Number(opsHealth?.tee_hardware?.attestation_latency_ms || 0);
  const cxlThroughputGbps = Number(opsHealth?.tee_hardware?.cxl_throughput_gbps || 0);
  const npuTempC = Number(opsHealth?.tee_hardware?.npu_temp_c || 0);
  const tpmTempC = Number(opsHealth?.tee_hardware?.tpm_temp_c || 0);
  const rewardApyPct = Number(opsHealth?.governance_economics?.reward_apy_pct || 0);
  const slashingEvents = Number(opsHealth?.governance_economics?.slashing_events_total || 0);
  const recentSlashes = Array.isArray(opsHealth?.governance_economics?.recent_slashing_events)
    ? opsHealth.governance_economics.recent_slashing_events
    : [];

  useEffect(() => {
    const append = (series, value) => {
      const parsed = Number(value);
      if (!Number.isFinite(parsed)) {
        return series;
      }
      const next = [...series, parsed];
      return next.slice(-18);
    };

    setTrendSeries((prev) => ({
      apiLatency: append(prev.apiLatency, apiLatencyMs),
      apiError: append(prev.apiError, apiErrorRatePct),
      ingress: append(prev.ingress, ingressMbps)
    }));
  }, [apiLatencyMs, apiErrorRatePct, ingressMbps]);

  const normalizeTrend = (items) => {
    if (!Array.isArray(items)) {
      return [];
    }
    return items
      .map((entry) => {
        if (entry && typeof entry === 'object' && entry.value != null) {
          return Number(entry.value);
        }
        return Number(entry);
      })
      .filter((value) => Number.isFinite(value));
  };

  const latencySeries = normalizeTrend(opsTrends?.api_latency_ms);
  const errorSeries = normalizeTrend(opsTrends?.api_error_rate_pct);
  const ingressSeries = normalizeTrend(opsTrends?.ingress_mbps);
  const resolvedLatencySeries = latencySeries.length > 0 ? latencySeries : trendSeries.apiLatency;
  const resolvedErrorSeries = errorSeries.length > 0 ? errorSeries : trendSeries.apiError;
  const resolvedIngressSeries = ingressSeries.length > 0 ? ingressSeries : trendSeries.ingress;

  const numberOrFallback = (value, suffix = '') => {
    const parsed = Number(value);
    return Number.isFinite(parsed) ? `${parsed}${suffix}` : 'N/A';
  };

  const fixedOrFallback = (value, decimals = 2, suffix = '') => {
    const parsed = Number(value);
    return Number.isFinite(parsed) ? `${parsed.toFixed(decimals)}${suffix}` : 'N/A';
  };

  const toggleCluster = (key) => {
    setCollapsedClusters((prev) => ({ ...prev, [key]: !prev[key] }));
  };

  const sparklinePoints = (series) => {
    if (!Array.isArray(series) || series.length === 0) {
      return '0,22 96,22';
    }
    if (series.length === 1) {
      return `0,12 96,12`;
    }
    const min = Math.min(...series);
    const max = Math.max(...series);
    const span = max - min || 1;
    return series.map((point, index) => {
      const x = (index / (series.length - 1)) * 96;
      const y = 22 - (((point - min) / span) * 20);
      return `${x.toFixed(2)},${y.toFixed(2)}`;
    }).join(' ');
  };

  const endpointClass = (isUp) => (isUp ? 'endpoint-up' : 'endpoint-down');

  return (
    <div className={`hud-ops-root ${compactMode ? 'density-compact' : ''}`}>
      <header className="hud-ops-header">
        <div>
          <p className="hud-kicker">SovereignMap Operator Console</p>
          <h2>Marketplace + LLM Training Command Deck</h2>
          <p className="hud-subline">Readable operations cockpit for API health, model training, wallet registry, governance signals, and live network connectivity.</p>
        </div>
        <div className="header-right">
          <button
            className={`density-toggle ${compactMode ? 'active' : ''}`}
            onClick={() => setCompactMode((value) => !value)}
            type="button"
          >
            {compactMode ? 'Standard Density' : 'Wallboard Density'}
          </button>
          <div className="ops-clock">{new Date().toLocaleString()}</div>
          {loading && <div className="loading-spinner">SYNCING</div>}
        </div>
      </header>

      <section className="ops-diagnostics-strip">
        <div className="diag-item">
          <span className="diag-label">API Base</span>
          <span className="diag-value">{diagnostics.apiBase || 'N/A'}</span>
        </div>
        <div className="diag-item">
          <span className="diag-label">Events Endpoint</span>
          <span className="diag-value">{diagnostics.eventsEndpoint || 'N/A'}</span>
        </div>
        <div className="diag-item">
          <span className="diag-label">SSE Status</span>
          <span className={`diag-value ${streamLabel === 'connected' ? 'diag-ok' : streamLabel === 'connecting' ? 'diag-warning' : 'diag-critical'}`}>
            {streamLabel.toUpperCase()}
          </span>
        </div>
        <div className="diag-item">
          <span className="diag-label">Last Transport Error</span>
          <span className={`diag-value ${streamLastError ? 'diag-warning' : ''}`}>
            {streamLastError || 'none'}
          </span>
        </div>
        <div className="diag-item">
          <span className="diag-label">Policy Endpoint</span>
          <span className="diag-value">{diagnostics.policyEndpoint || 'N/A'}</span>
        </div>
        <div className="diag-item">
          <span className="diag-label">Policy Transport</span>
          <span className={`diag-value ${policyStatus === 'ok' ? 'diag-ok' : policyStatus === 'pending' ? 'diag-warning' : policyStatus === 'error' ? 'diag-critical' : ''}`}>
            {`${String(policyStatus).toUpperCase()}${diagnostics.policyDetail ? ` (${diagnostics.policyDetail})` : ''}`}
          </span>
        </div>
      </section>

      <section className="ops-ribbon">
        <div className="ops-chip">
          <span className="ops-chip-label">Platform State</span>
          <span className={`ops-chip-value ${healthStatus === 'critical' ? 'chip-critical' : healthStatus === 'degraded' ? 'chip-warning' : 'chip-ok'}`}>
            {healthStatus.toUpperCase()}
          </span>
        </div>
        <div className="ops-chip">
          <span className="ops-chip-label">API Plane</span>
          <span className={`ops-chip-value ${apiErrorRatePct >= 1 ? 'chip-warning' : 'chip-ok'}`}>
            {fixedOrFallback(apiLatencyMs, 0, ' ms')} latency | {fixedOrFallback(apiErrorRatePct, 2, '%')} errors
          </span>
        </div>
        <div className="ops-chip">
          <span className="ops-chip-label">Event Bus</span>
          <span className={`ops-chip-value ${streamLabel === 'connected' ? 'chip-ok' : streamLabel === 'connecting' ? 'chip-warning' : 'chip-critical'}`}>
            {streamLabel.toUpperCase()}
          </span>
        </div>
        <div className="ops-chip">
          <span className="ops-chip-label">LLM Guardrail</span>
          <span className={`ops-chip-value ${llmPolicyRejected > llmPolicyValid ? 'chip-warning' : 'chip-ok'}`}>
            pass {fixedOrFallback(llmPassRatePct, 1, '%')} | valid {llmPolicyValid} / rejected {llmPolicyRejected}
          </span>
        </div>
        <div className="ops-chip">
          <span className="ops-chip-label">Training Loop</span>
          <span className={`ops-chip-value ${trainingStatus?.active ? 'chip-ok' : 'chip-warning'}`}>
            {trainingStatus?.status || 'idle'} / round {trainingRound}
          </span>
        </div>
        <div className="ops-chip">
          <span className="ops-chip-label">Training Target</span>
          <span className={`ops-chip-value ${targetRounds > 0 ? 'chip-ok' : 'chip-warning'}`}>
            {targetRounds > 0
              ? `${targetRounds} rounds (${Number.isFinite(remainingRounds) ? `${remainingRounds} remaining` : 'running'})`
              : 'continuous'}
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
          <span className="ops-chip-label">Connections</span>
          <span className={`ops-chip-value ${portHealthUpCount === 3 ? 'chip-ok' : 'chip-warning'}`}>
            {portHealthUpCount}/3 core endpoints up
          </span>
        </div>
        <div className="ops-chip">
          <span className="ops-chip-label">Wallet Registry</span>
          <span className={`ops-chip-value ${verifiedRegistryCount === registryCount && registryCount > 0 ? 'chip-ok' : 'chip-warning'}`}>
            {verifiedRegistryCount}/{registryCount} verified
          </span>
        </div>
        <div className="ops-chip">
          <span className="ops-chip-label">Stake + Yield</span>
          <span className="ops-chip-value">
            {fixedOrFallback(listedStake, 2)} stake | APY {fixedOrFallback(rewardApyPct, 2, '%')}
          </span>
        </div>
        <div className="ops-chip">
          <span className="ops-chip-label">Marketplace Risk</span>
          <span className={`ops-chip-value ${slashingEvents > 0 ? 'chip-warning' : 'chip-ok'}`}>
            {slashingEvents} slashing events
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
            <label className="input-row">
              Deterministic Rounds (0 = continuous)
              <input
                type="number"
                min="0"
                step="1"
                value={requestedRounds}
                onChange={(event) => setRequestedRounds(Math.max(0, Number(event.target.value) || 0))}
                disabled={loading || trainingStatus?.active}
              />
            </label>
            <div className="round-presets" role="group" aria-label="Deterministic round presets">
              {[5, 10, 20].map((preset) => (
                <button
                  key={preset}
                  type="button"
                  className={`preset-btn ${requestedRounds === preset ? 'active' : ''}`}
                  onClick={() => setRequestedRounds(preset)}
                  disabled={loading || trainingStatus?.active}
                >
                  {preset} rounds
                </button>
              ))}
            </div>
            <button className="btn cmd-start" onClick={() => onStartTraining(requestedRounds)} disabled={loading || trainingStatus?.active}>
              {requestedRounds > 0 ? `Start Training (${requestedRounds} rounds)` : 'Start Training (continuous)'}
            </button>
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
          <h3>API + LLM Training Dashboard</h3>
          <p className="panel-subtitle">Critical service telemetry for endpoint responsiveness, model rounds, and connection quality.</p>
          <div className="kpi-grid">
            <div className="kpi-card">
              <span>API Latency</span>
              <strong>{numberOrFallback(apiLatencyMs, ' ms')}</strong>
            </div>
            <div className="kpi-card">
              <span>Ingress</span>
              <strong>{numberOrFallback(ingressMbps, ' Mbps')}</strong>
            </div>
            <div className="kpi-card">
              <span>Error Rate</span>
              <strong>{fixedOrFallback(apiErrorRatePct, 2, '%')}</strong>
            </div>
            <div className="kpi-card">
              <span>Training Round</span>
              <strong>{trainingRound}</strong>
            </div>
            <div className="kpi-card">
              <span>Model Accuracy</span>
              <strong>{hudData?.last_audit_accuracy || 'N/A'}</strong>
            </div>
            <div className="kpi-card">
              <span>FL Loss</span>
              <strong>{fixedOrFallback(trainingLoss, 4)}</strong>
            </div>
            <div className="kpi-card">
              <span>LLM Valid Updates</span>
              <strong>{llmPolicyValid}</strong>
            </div>
            <div className="kpi-card">
              <span>LLM Rejected Updates</span>
              <strong className={llmPolicyRejected > llmPolicyValid ? 'kpi-warning' : ''}>{llmPolicyRejected}</strong>
            </div>
            <div className="kpi-card">
              <span>Guardrail Pass Rate</span>
              <strong className={llmPassRatePct < 80 ? 'kpi-warning' : ''}>{fixedOrFallback(llmPassRatePct, 1, '%')}</strong>
            </div>
            <div className="kpi-card">
              <span>Browser RTT</span>
              <strong className={browserRtt >= 250 ? 'kpi-warning' : ''}>{numberOrFallback(webMetrics?.rttMs, ' ms')}</strong>
            </div>
            <div className="kpi-card">
              <span>Browser Downlink</span>
              <strong>{numberOrFallback(webMetrics?.downlinkMbps, ' Mbps')}</strong>
            </div>
            <div className="kpi-card">
              <span>JS Heap Used</span>
              <strong className={browserHeap >= 600 ? 'kpi-warning' : ''}>{numberOrFallback(webMetrics?.jsHeapUsedMB, ' MB')}</strong>
            </div>
            <div className="kpi-card">
              <span>Connection Health</span>
              <strong className={portHealthUpCount < 3 ? 'kpi-warning' : ''}>{portHealthUpCount}/3 endpoints up</strong>
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
            <div className="kpi-card">
              <span>Attestation Latency</span>
              <strong className={attestationLatencyMs >= 120 ? 'kpi-warning' : ''}>{numberOrFallback(attestationLatencyMs, ' ms')}</strong>
            </div>
            <div className="kpi-card">
              <span>CXL Throughput</span>
              <strong>{numberOrFallback(cxlThroughputGbps, ' GB/s')}</strong>
            </div>
          </div>

          <section className="domain-cluster-grid">
            <article className={`domain-cluster ${collapsedClusters.api ? 'collapsed' : ''}`}>
              <div className="cluster-head">
                <h4>API Endpoint Cluster</h4>
                {isMobile && (
                  <button type="button" className="cluster-toggle" onClick={() => toggleCluster('api')}>
                    {collapsedClusters.api ? 'Expand' : 'Collapse'}
                  </button>
                )}
              </div>
              {!collapsedClusters.api && (
                <div className="cluster-body">
                  <div className="endpoint-badges">
                    <span className={`endpoint-badge ${endpointClass(apiPortUp)}`}>API 8000</span>
                    <span className={`endpoint-badge ${endpointClass(flPortUp)}`}>FL 8080</span>
                    <span className={`endpoint-badge ${endpointClass(promPortUp)}`}>PROM 9090</span>
                  </div>
                  <div className="sparkline-row">
                    <span>Latency</span>
                    <svg viewBox="0 0 96 24" preserveAspectRatio="none" aria-label="API latency trend">
                      <polyline points={sparklinePoints(resolvedLatencySeries)} className="sparkline-latency" />
                    </svg>
                    <strong>{fixedOrFallback(apiLatencyMs, 0, ' ms')}</strong>
                  </div>
                  <div className="sparkline-row">
                    <span>Error</span>
                    <svg viewBox="0 0 96 24" preserveAspectRatio="none" aria-label="API error trend">
                      <polyline points={sparklinePoints(resolvedErrorSeries)} className="sparkline-error" />
                    </svg>
                    <strong>{fixedOrFallback(apiErrorRatePct, 2, '%')}</strong>
                  </div>
                  <div className="sparkline-row">
                    <span>Ingress</span>
                    <svg viewBox="0 0 96 24" preserveAspectRatio="none" aria-label="Ingress trend">
                      <polyline points={sparklinePoints(resolvedIngressSeries)} className="sparkline-ingress" />
                    </svg>
                    <strong>{fixedOrFallback(ingressMbps, 1, ' Mbps')}</strong>
                  </div>
                </div>
              )}
            </article>
            <article className={`domain-cluster ${collapsedClusters.llm ? 'collapsed' : ''}`}>
              <div className="cluster-head">
                <h4>LLM Training Cluster</h4>
                {isMobile && (
                  <button type="button" className="cluster-toggle" onClick={() => toggleCluster('llm')}>
                    {collapsedClusters.llm ? 'Expand' : 'Collapse'}
                  </button>
                )}
              </div>
              {!collapsedClusters.llm && (
                <div className="cluster-body">
                  <div className="audit-row"><span>Training State</span><span>{trainingStatus?.status || 'idle'}</span></div>
                  <div className="audit-row"><span>Current Round</span><span>{trainingRound}</span></div>
                  <div className="audit-row"><span>Current Loss</span><span>{fixedOrFallback(trainingLoss, 4)}</span></div>
                  <div className="audit-row"><span>Guardrail Pass</span><span>{fixedOrFallback(llmPassRatePct, 1, '%')}</span></div>
                </div>
              )}
            </article>
            <article className={`domain-cluster ${collapsedClusters.network ? 'collapsed' : ''}`}>
              <div className="cluster-head">
                <h4>Network Connections</h4>
                {isMobile && (
                  <button type="button" className="cluster-toggle" onClick={() => toggleCluster('network')}>
                    {collapsedClusters.network ? 'Expand' : 'Collapse'}
                  </button>
                )}
              </div>
              {!collapsedClusters.network && (
                <div className="cluster-body">
                  <div className="audit-row"><span>Event Stream</span><span>{streamLabel}</span></div>
                  <div className="audit-row"><span>Node Connectivity</span><span>{hudData?.active_nodes ?? 0} active</span></div>
                  <div className="audit-row"><span>Browser Downlink</span><span>{numberOrFallback(webMetrics?.downlinkMbps, ' Mbps')}</span></div>
                </div>
              )}
            </article>
          </section>

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

          <div className="aux-panel">
            <h4>Web Runtime Metrics</h4>
            <div className="audit-row">
              <span>Page Load</span>
              <span>{numberOrFallback(webMetrics?.pageLoadMs, ' ms')}</span>
            </div>
            <div className="audit-row">
              <span>DOM Interactive</span>
              <span>{numberOrFallback(webMetrics?.domInteractiveMs, ' ms')}</span>
            </div>
            <div className="audit-row">
              <span>Viewport</span>
              <span>{webMetrics?.viewport || 'N/A'}</span>
            </div>
            <div className="audit-row">
              <span>CPU Cores</span>
              <span>{webMetrics?.cores ?? 'N/A'}</span>
            </div>
            <div className="audit-row">
              <span>Device Memory</span>
              <span>{numberOrFallback(webMetrics?.deviceMemoryGB, ' GB')}</span>
            </div>
            <div className="audit-row">
              <span>Connection</span>
              <span>{[webMetrics?.connectionType, webMetrics?.effectiveType].filter(Boolean).join(' / ') || 'N/A'}</span>
            </div>
            <div className="audit-row">
              <span>Heap Capacity</span>
              <span>{numberOrFallback(webMetrics?.jsHeapTotalMB, ' MB')}</span>
            </div>
          </div>

          <div className="aux-panel">
            <h4>TEE and Hardware Trust</h4>
            <div className="audit-row">
              <span>EPC Utilization</span>
              <span>{fixedOrFallback(epcUtilization, 2, '%')}</span>
            </div>
            <div className="audit-row">
              <span>Attestation Latency</span>
              <span>{numberOrFallback(attestationLatencyMs, ' ms')}</span>
            </div>
            <div className="audit-row">
              <span>CXL Throughput</span>
              <span>{numberOrFallback(cxlThroughputGbps, ' GB/s')}</span>
            </div>
            <div className="audit-row">
              <span>NPU Temp</span>
              <span>{numberOrFallback(npuTempC, ' C')}</span>
            </div>
            <div className="audit-row">
              <span>TPM Temp</span>
              <span>{numberOrFallback(tpmTempC, ' C')}</span>
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

          {runbookCards.length > 0 && (
            <div className="aux-panel runbook-match-panel">
              <h4>Runbook Match Cards</h4>
              {runbookCards.map((card) => (
                <div className="runbook-card" key={card.id}>
                  <div className="runbook-card-head">
                    <span className="runbook-component">{card.component.toUpperCase()}</span>
                    <span className="runbook-owner">owner: {card.owner}</span>
                  </div>
                  <div className="runbook-message">{card.message}</div>
                  <div className="runbook-section">
                    Runbook: <a href="../docs/ALERT_RUNBOOKS.md" target="_blank" rel="noreferrer">{card.section}</a>
                  </div>
                  <ul className="runbook-checks">
                    {card.checks.map((check, idx) => (
                      <li key={`${card.id}-check-${idx}`}>{check}</li>
                    ))}
                  </ul>
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
          <h3>Marketplace + Wallet + Assistant</h3>
          <div className="governance-meta">
            <div><span>Trust Mode</span><strong>{trustMode}</strong></div>
            <div><span>Avg Confidence</span><strong>{fixedOrFallback(confidencePct, 2, '%')}</strong></div>
          </div>

          <div className="aux-panel marketplace-panel">
            <h4>Marketplace and Wallet Registry</h4>
            <div className="audit-row">
              <span>Listed Wallets</span>
              <span>{registryCount}</span>
            </div>
            <div className="audit-row">
              <span>Verified Wallets</span>
              <span>{verifiedRegistryCount}</span>
            </div>
            <div className="audit-row">
              <span>Visible Stake</span>
              <span>{fixedOrFallback(listedStake, 2)}</span>
            </div>
            <div className="audit-row">
              <span>Reward APY</span>
              <span>{fixedOrFallback(rewardApyPct, 2, '%')}</span>
            </div>
            <div className="audit-row">
              <span>Stake Concentration</span>
              <span>{fixedOrFallback(opsHealth?.governance_economics?.stake_concentration_pct, 2, '%')}</span>
            </div>
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

          <div className="aux-panel">
            <h4>Governance Economics</h4>
            <div className="audit-row">
              <span>Reward Yield (APY)</span>
              <span>{numberOrFallback(rewardApyPct, '%')}</span>
            </div>
            <div className="audit-row">
              <span>Slashing Events</span>
              <span>{slashingEvents}</span>
            </div>
            <div className="audit-row">
              <span>Stake Concentration</span>
              <span>{fixedOrFallback(opsHealth?.governance_economics?.stake_concentration_pct, 2, '%')}</span>
            </div>
            {recentSlashes.length === 0 ? (
              <div className="timeline-empty">No recent slashing events.</div>
            ) : (
              recentSlashes.map((slash, idx) => (
                <div className="audit-row" key={`${slash?.node || 'slash'}-${idx}`}>
                  <span>{slash?.node || `node-${idx + 1}`} ({slash?.reason || 'policy'})</span>
                  <span>-{numberOrFallback(slash?.stake_penalty)}</span>
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

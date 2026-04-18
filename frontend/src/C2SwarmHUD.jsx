import { useCallback, useEffect, useMemo, useState } from 'react';
import './C2SwarmHUD.css';

const COMMANDS = [
  'hold_position',
  'return_to_base',
  'reroute',
  'reassign_role',
  'isolate_node',
  'resume_autonomy',
  'pause_autonomy',
  'set_objective'
];

const TARGET_SCOPES = ['global', 'node', 'group'];
const ROLE_OPTIONS = ['admin', 'commander', 'operator', 'auditor', 'viewer'];

const parseTargetIds = (value) => {
  if (!value) {
    return [];
  }
  return String(value)
    .split(',')
    .map((item) => item.trim())
    .filter(Boolean)
    .slice(0, 64);
};

function C2SwarmHUD({ apiBase }) {
  const [status, setStatus] = useState(null);
  const [mapState, setMapState] = useState(null);
  const [commandLog, setCommandLog] = useState([]);
  const [auditLog, setAuditLog] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const [authToken, setAuthToken] = useState('');
  const [submitting, setSubmitting] = useState(false);
  const [submitMessage, setSubmitMessage] = useState('');
  const [form, setForm] = useState({
    command: 'hold_position',
    targetScope: 'global',
    targetIds: '',
    objective: '',
    assignmentRole: 'mapper',
    waypointLat: '',
    waypointLng: '',
    nonce: '',
    operatorRole: 'admin'
  });

  const fetchSnapshot = useCallback(async () => {
    try {
      const [statusResp, mapResp, commandsResp] = await Promise.all([
        fetch(`${apiBase}/swarm/status`),
        fetch(`${apiBase}/swarm/map?limit=280&layers=paths,risk,coverage,communications`),
        fetch(`${apiBase}/swarm/commands?limit=60`)
      ]);

      if (!statusResp.ok || !mapResp.ok || !commandsResp.ok) {
        throw new Error('C2 endpoints unavailable');
      }

      const [statusPayload, mapPayload, commandPayload] = await Promise.all([
        statusResp.json(),
        mapResp.json(),
        commandsResp.json()
      ]);

      setStatus(statusPayload || null);
      setMapState(mapPayload || null);
      setCommandLog(Array.isArray(commandPayload?.commands) ? commandPayload.commands : []);

      const auditHeaders = {};
      if (authToken.trim()) {
        auditHeaders.Authorization = `Bearer ${authToken.trim()}`;
      }
      const auditResp = await fetch(`${apiBase}/swarm/audit/recent?limit=30`, { headers: auditHeaders });
      if (auditResp.ok) {
        const auditPayload = await auditResp.json();
        setAuditLog(Array.isArray(auditPayload?.audits) ? auditPayload.audits : []);
      }

      setError('');
    } catch (snapshotErr) {
      setError(snapshotErr?.message || 'Unable to fetch C2 swarm state');
    } finally {
      setLoading(false);
    }
  }, [apiBase, authToken]);

  useEffect(() => {
    fetchSnapshot();
    const interval = setInterval(fetchSnapshot, 3000);
    return () => clearInterval(interval);
  }, [fetchSnapshot]);

  const visibleNodes = useMemo(() => {
    const nodes = Array.isArray(mapState?.nodes) ? mapState.nodes : [];
    return nodes.slice(0, 320);
  }, [mapState]);

  const pathLinks = useMemo(() => {
    const paths = Array.isArray(mapState?.paths) ? mapState.paths : [];
    const byId = new Map(visibleNodes.map((node) => [node.id, node]));
    return paths
      .map((path) => {
        const from = byId.get(path.from);
        const to = byId.get(path.to);
        if (!from || !to) {
          return null;
        }
        return {
          id: path.id,
          from: from.position,
          to: to.position
        };
      })
      .filter(Boolean)
      .slice(0, 240);
  }, [mapState, visibleNodes]);

  const riskZones = useMemo(() => {
    const zones = Array.isArray(mapState?.risk_zones) ? mapState.risk_zones : [];
    return zones.slice(0, 12);
  }, [mapState]);

  const updateForm = (field, value) => {
    setForm((current) => ({ ...current, [field]: value }));
  };

  const submitCommand = async (event) => {
    event.preventDefault();
    setSubmitting(true);
    setSubmitMessage('');

    const payload = {
      command: form.command,
      target_scope: form.targetScope,
      target_ids: parseTargetIds(form.targetIds),
      client_nonce: form.nonce,
      parameters: {}
    };

    if (form.command === 'set_objective' && form.objective.trim()) {
      payload.objective = form.objective.trim();
    }

    if (form.command === 'reassign_role') {
      payload.parameters.role = form.assignmentRole;
    }

    if (form.command === 'reroute' || form.command === 'set_objective') {
      if (form.waypointLat.trim()) {
        payload.waypoint_lat = Number(form.waypointLat);
      }
      if (form.waypointLng.trim()) {
        payload.waypoint_lng = Number(form.waypointLng);
      }
    }

    try {
      const headers = {
        'Content-Type': 'application/json',
        'X-API-Role': form.operatorRole
      };
      if (authToken.trim()) {
        headers.Authorization = `Bearer ${authToken.trim()}`;
      }

      const response = await fetch(`${apiBase}/swarm/command`, {
        method: 'POST',
        headers,
        body: JSON.stringify(payload)
      });

      const responsePayload = await response.json().catch(() => ({}));
      if (!response.ok) {
        throw new Error(responsePayload?.error || `Command failed (${response.status})`);
      }

      setSubmitMessage(`Accepted ${responsePayload?.command?.command || form.command} (${responsePayload?.action_id || 'queued'})`);
      updateForm('nonce', '');
      fetchSnapshot();
    } catch (submitErr) {
      setSubmitMessage(submitErr?.message || 'Command rejected');
    } finally {
      setSubmitting(false);
    }
  };

  const statusClass = status?.status === 'ok' ? 'ok' : 'warn';

  return (
    <div className="c2-shell">
      <header className="c2-header">
        <div>
          <h1>Autonomous Swarm C2</h1>
          <p>Security-scoped command channel with bounded map telemetry for high-node federated operations.</p>
        </div>
        <div className={`c2-pill c2-pill-${statusClass}`}>
          {status?.status ? `Status: ${status.status}` : 'Status: bootstrapping'}
        </div>
      </header>

      {error ? <div className="c2-error">{error}</div> : null}

      <section className="c2-kpis" aria-label="swarm-kpis">
        <article>
          <span>Active Nodes</span>
          <strong>{status?.nodes_active ?? mapState?.node_count ?? 0}</strong>
        </article>
        <article>
          <span>Coverage</span>
          <strong>{Number(status?.coverage_pct || mapState?.coverage?.percent || 0).toFixed(1)}%</strong>
        </article>
        <article>
          <span>API Latency</span>
          <strong>{Math.round(Number(status?.avg_latency_ms || 0))} ms</strong>
        </article>
        <article>
          <span>Error Rate</span>
          <strong>{Number(status?.error_rate_pct || 0).toFixed(2)}%</strong>
        </article>
      </section>

      <div className="c2-layout">
        <section className="c2-map-panel">
          <div className="c2-panel-head">
            <h2>Swarm Map</h2>
            <span>{visibleNodes.length} rendered nodes</span>
          </div>
          <div className="c2-map-wrap">
            <svg viewBox="0 0 100 100" role="img" aria-label="swarm map" className="c2-map-svg">
              <defs>
                <linearGradient id="riskFill" x1="0" y1="0" x2="1" y2="1">
                  <stop offset="0%" stopColor="#fda085" stopOpacity="0.45" />
                  <stop offset="100%" stopColor="#f6d365" stopOpacity="0.2" />
                </linearGradient>
              </defs>

              {riskZones.map((zone) => (
                <circle
                  key={zone.id}
                  cx={zone.x}
                  cy={zone.y}
                  r={Math.max(1, Number(zone.radius || 2))}
                  fill="url(#riskFill)"
                  stroke="#f57c00"
                  strokeWidth="0.35"
                />
              ))}

              {pathLinks.map((link) => (
                <line
                  key={link.id}
                  x1={link.from.x}
                  y1={link.from.y}
                  x2={link.to.x}
                  y2={link.to.y}
                  stroke="#69b4ff"
                  strokeWidth="0.22"
                  strokeOpacity="0.45"
                />
              ))}

              {visibleNodes.map((node) => (
                <circle
                  key={node.id}
                  cx={node.position?.x}
                  cy={node.position?.y}
                  r={node.status === 'ok' ? 0.95 : 1.35}
                  fill={node.status === 'ok' ? '#8ddf8a' : '#ffad5c'}
                  stroke="#0e2233"
                  strokeWidth="0.2"
                />
              ))}
            </svg>
          </div>
        </section>

        <section className="c2-command-panel">
          <div className="c2-panel-head">
            <h2>Command Channel</h2>
            <span>Signed admin token required</span>
          </div>

          <form onSubmit={submitCommand} className="c2-form">
            <label>
              Admin Token
              <input
                type="password"
                value={authToken}
                onChange={(event) => setAuthToken(event.target.value)}
                placeholder="Bearer token"
                autoComplete="off"
              />
            </label>

            <div className="c2-form-grid">
              <label>
                Role
                <select value={form.operatorRole} onChange={(event) => updateForm('operatorRole', event.target.value)}>
                  {ROLE_OPTIONS.map((role) => (
                    <option key={role} value={role}>{role}</option>
                  ))}
                </select>
              </label>

              <label>
                Command
                <select value={form.command} onChange={(event) => updateForm('command', event.target.value)}>
                  {COMMANDS.map((command) => (
                    <option key={command} value={command}>{command}</option>
                  ))}
                </select>
              </label>

              <label>
                Target Scope
                <select value={form.targetScope} onChange={(event) => updateForm('targetScope', event.target.value)}>
                  {TARGET_SCOPES.map((scope) => (
                    <option key={scope} value={scope}>{scope}</option>
                  ))}
                </select>
              </label>
            </div>

            <label>
              Target IDs (comma separated)
              <input
                type="text"
                value={form.targetIds}
                onChange={(event) => updateForm('targetIds', event.target.value)}
                placeholder="node-0001,node-0009"
              />
            </label>

            <label>
              Objective
              <input
                type="text"
                value={form.objective}
                onChange={(event) => updateForm('objective', event.target.value)}
                placeholder="Capture high-risk corridor"
              />
            </label>

            <div className="c2-form-grid">
              <label>
                Role (reassign)
                <select value={form.assignmentRole} onChange={(event) => updateForm('assignmentRole', event.target.value)}>
                  <option value="mapper">mapper</option>
                  <option value="scout">scout</option>
                  <option value="relay">relay</option>
                  <option value="collector">collector</option>
                  <option value="sentinel">sentinel</option>
                </select>
              </label>

              <label>
                Nonce
                <input
                  type="text"
                  value={form.nonce}
                  onChange={(event) => updateForm('nonce', event.target.value)}
                  placeholder="client nonce"
                />
              </label>
            </div>

            <div className="c2-form-grid">
              <label>
                Waypoint Lat
                <input
                  type="number"
                  value={form.waypointLat}
                  onChange={(event) => updateForm('waypointLat', event.target.value)}
                  step="0.0001"
                  placeholder="37.7749"
                />
              </label>

              <label>
                Waypoint Lng
                <input
                  type="number"
                  value={form.waypointLng}
                  onChange={(event) => updateForm('waypointLng', event.target.value)}
                  step="0.0001"
                  placeholder="-122.4194"
                />
              </label>
            </div>

            <button type="submit" disabled={submitting}>
              {submitting ? 'Submitting...' : 'Submit Command'}
            </button>
          </form>

          <div className="c2-submit-message">{submitMessage}</div>
        </section>
      </div>

      <section className="c2-log-panel">
        <div className="c2-panel-head">
          <h2>Recent Command Log</h2>
          <span>{commandLog.length} entries</span>
        </div>

        <div className="c2-log-table" role="table" aria-label="command-log">
          <div className="c2-log-row c2-log-head" role="row">
            <span>Timestamp</span>
            <span>Command</span>
            <span>Scope</span>
            <span>Targets</span>
            <span>Status</span>
          </div>
          {commandLog.slice(-24).reverse().map((item) => (
            <div className="c2-log-row" role="row" key={`${item.action_id}-${item.ts}`}>
              <span>{item.ts ? new Date(item.ts * 1000).toLocaleTimeString() : '-'}</span>
              <span>{item.command || '-'}</span>
              <span>{item.target_scope || '-'}</span>
              <span>{Array.isArray(item.target_ids) ? item.target_ids.length : 0}</span>
              <span>{item.status || '-'}</span>
            </div>
          ))}
          {!commandLog.length && !loading ? <div className="c2-empty">No commands issued yet.</div> : null}
        </div>
      </section>

      <section className="c2-log-panel">
        <div className="c2-panel-head">
          <h2>Audit Signatures</h2>
          <span>{auditLog.length} entries</span>
        </div>

        <div className="c2-log-table" role="table" aria-label="audit-log">
          <div className="c2-log-row c2-log-head" role="row">
            <span>Timestamp</span>
            <span>Command</span>
            <span>Role</span>
            <span>Action ID</span>
            <span>Signature</span>
          </div>
          {auditLog.slice(-14).reverse().map((item) => (
            <div className="c2-log-row" role="row" key={`${item.action_id}-${item.signature}`}>
              <span>{item.ts ? new Date(item.ts * 1000).toLocaleTimeString() : '-'}</span>
              <span>{item.command || '-'}</span>
              <span>{item.role || '-'}</span>
              <span>{item.action_id || '-'}</span>
              <span>{String(item.signature || '').slice(0, 14)}...</span>
            </div>
          ))}
          {!auditLog.length && !loading ? <div className="c2-empty">No signed audit entries available.</div> : null}
        </div>
      </section>
    </div>
  );
}

export default C2SwarmHUD;

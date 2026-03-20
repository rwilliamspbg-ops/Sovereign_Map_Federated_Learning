import React, { useState, useEffect, useRef } from 'react';

export default function LiveTerminal({ events = [] }) {
  const [logs, setLogs] = useState([]);
  const [hideRoutineRounds, setHideRoutineRounds] = useState(true);
  const [focusMode, setFocusMode] = useState('all');
  const [autoScroll, setAutoScroll] = useState(true);
  const containerRef = useRef(null);

  useEffect(() => {
    if (!Array.isArray(events) || events.length === 0) {
      setLogs([
        {
          line: '[BOOT] Operations terminal initialized.',
          severity: 'info',
          kind: 'boot'
        },
        {
          line: '[BOOT] Awaiting live backend events...',
          severity: 'info',
          kind: 'boot'
        }
      ]);
      return;
    }

    const filtered = events.filter((evt) => {
      const kind = (evt?.kind || '').toLowerCase();
      const severity = (evt?.severity || 'info').toLowerCase();
      const message = (evt?.message || '').toLowerCase();

      if (hideRoutineRounds && kind === 'training_round' && severity !== 'error') {
        return false;
      }

      if (focusMode === 'critical') {
        const isCriticalEvent =
          severity === 'error' ||
          severity === 'warning' ||
          kind.includes('enclave') ||
          kind.includes('policy') ||
          message.includes('error') ||
          message.includes('failed') ||
          message.includes('degraded');
        return isCriticalEvent;
      }

      if (focusMode === 'enclave') {
        return kind.includes('enclave') || message.includes('enclave');
      }

      return true;
    });

    const mapped = filtered.slice(-80).map((evt) => {
      const ts = evt?.ts ? new Date(evt.ts * 1000).toLocaleTimeString() : 'N/A';
      const kind = (evt?.kind || 'event').toUpperCase();
      return {
        line: `[${ts}] [${kind}] ${evt?.message || 'event'}`,
        severity: (evt?.severity || 'info').toLowerCase(),
        kind: (evt?.kind || 'event').toLowerCase()
      };
    });

    setLogs(mapped);
  }, [events, hideRoutineRounds, focusMode]);

  useEffect(() => {
    if (autoScroll && containerRef.current) {
      containerRef.current.scrollTop = containerRef.current.scrollHeight;
    }
  }, [logs, autoScroll]);

  return (
    <div className="terminal-container hud-section">
      <h3>System Operations Log</h3>
      <div className="terminal-toolbar">
        <label>
          <input
            type="checkbox"
            checked={hideRoutineRounds}
            onChange={(e) => setHideRoutineRounds(e.target.checked)}
          />
          Hide routine TRAINING_ROUND
        </label>
        <label>
          Focus
          <select value={focusMode} onChange={(e) => setFocusMode(e.target.value)}>
            <option value="all">all</option>
            <option value="critical">critical/enclave/policy</option>
            <option value="enclave">enclave only</option>
          </select>
        </label>
        <label>
          <input
            type="checkbox"
            checked={autoScroll}
            onChange={(e) => setAutoScroll(e.target.checked)}
          />
          Auto-scroll
        </label>
      </div>
      <div className="terminal-output" ref={containerRef}>
        {logs.map((entry, index) => (
          <div
            key={index}
            className={`terminal-line terminal-${entry?.severity || 'info'} terminal-kind-${entry?.kind || 'event'}`}
          >
            {entry?.line || ''}
          </div>
        ))}
      </div>
    </div>
  );
}

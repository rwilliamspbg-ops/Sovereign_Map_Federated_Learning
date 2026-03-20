import React, { useState, useEffect, useRef } from 'react';

export default function LiveTerminal({ events = [] }) {
  const [logs, setLogs] = useState([
    '[BOOT] Operations terminal initialized.',
    '[BOOT] Awaiting live backend events...'
  ]);
  const containerRef = useRef(null);

  useEffect(() => {
    if (!Array.isArray(events) || events.length === 0) {
      return;
    }
    const mapped = events.slice(-60).map((evt) => {
      const ts = evt?.ts ? new Date(evt.ts * 1000).toLocaleTimeString() : 'N/A';
      const kind = (evt?.kind || 'event').toUpperCase();
      return `[${ts}] [${kind}] ${evt?.message || 'event'}`;
    });
    setLogs(mapped);
  }, [events]);

  useEffect(() => {
    if (containerRef.current) {
      containerRef.current.scrollTop = containerRef.current.scrollHeight;
    }
  }, [logs]);

  return (
    <div className="terminal-container hud-section">
      <h3>🖥️ System Operations Log</h3>
      <div className="terminal-output" ref={containerRef}>
        {logs.map((log, index) => (
          <div key={index} className="terminal-line">{log}</div>
        ))}
      </div>
    </div>
  );
}

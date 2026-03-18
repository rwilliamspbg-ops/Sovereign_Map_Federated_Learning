import React, { useState, useEffect, useRef } from 'react';

const MESSAGES = [
  "[SYS] Connection securely established to Sovereign Orbit.",
  "[TEE] Remote attestation verified for Node 0x489F.",
  "[FL] Training weights securely aggregated across 12 nodes.",
  "[NET] CXL memory expanded successfully. Latency: 12ms.",
  "[TOKEN] Smart contract bridge escrow synchronized.",
  "[SYS] Global accuracy achieved convergence target (85.2%).",
  "[SEC] Anomalous gradient dropped from BFT sequence.",
  "[FL] Starting Epoch 72 on partitioned subset.",
  "[TEE] SGX Quote validation complete.",
  "[SYS] Health check passed: All services nominal."
];

export default function LiveTerminal() {
  const [logs, setLogs] = useState([]);
  const containerRef = useRef(null);

  useEffect(() => {
    // Start with a few initial logs
    setLogs([
      "Initializing Sovereign Map Systems...",
      "Connecting to mesh network...",
      "Link established."
    ]);

    const interval = setInterval(() => {
      setLogs((prev) => {
        const nextMsg = MESSAGES[Math.floor(Math.random() * MESSAGES.length)];
        const newLogs = [...prev, `[${new Date().toISOString().split('T')[1].slice(0, -1)}] ${nextMsg}`];
        if (newLogs.length > 50) newLogs.shift();
        return newLogs;
      });
    }, 1800);

    return () => clearInterval(interval);
  }, []);

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

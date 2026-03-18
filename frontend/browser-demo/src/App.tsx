import React, { useState } from "react";

const App: React.FC = () => {
  const [nodes, setNodes] = useState<
    { id: string; active: boolean; gpu: boolean }[]
  >([]);

  const addNode = () => {
    setNodes([
      ...nodes,
      { id: `node-${Date.now()}`, active: true, gpu: Math.random() > 0.5 },
    ]);
  };

  return (
    <div style={{ fontFamily: "sans-serif", margin: "40px" }}>
      <h1>Sovereign Map FL Demo</h1>
      <p>
        Interactive Browser-based Federated Learning with WebGPU Acceleration.
      </p>

      <button
        onClick={addNode}
        style={{
          padding: "10px 20px",
          fontSize: "16px",
          background: "#0070f3",
          color: "white",
          border: "none",
          borderRadius: "5px",
          cursor: "pointer",
        }}
      >
        + Add Edge Node
      </button>

      <div
        style={{
          marginTop: "30px",
          display: "grid",
          gridTemplateColumns: "repeat(4, 1fr)",
          gap: "20px",
        }}
      >
        {nodes.map((node) => (
          <div
            key={node.id}
            style={{
              border: "1px solid #ddd",
              padding: "15px",
              borderRadius: "8px",
            }}
          >
            <h3>{node.id}</h3>
            <p>Status: {node.active ? "🟢 Active" : "🔴 Inactive"}</p>
            <p>Accelerator: {node.gpu ? "🚀 WebGPU" : "💻 CPU"}</p>
            <p>Privacy Budget (\u03B5): {(Math.random() * 2).toFixed(2)}</p>
          </div>
        ))}
      </div>

      {nodes.length > 0 && (
        <div
          style={{
            marginTop: "40px",
            padding: "20px",
            background: "#f5f5f5",
            borderRadius: "8px",
          }}
        >
          <h3>Global Model Status</h3>
          <p>Clients updating: {nodes.length}</p>
          <p>
            Total Gradients Compressed: {(nodes.length * 4.2).toFixed(1)} MB
          </p>
          <p>
            Bandwidth Saved: {(nodes.length * 4.2 * 0.9).toFixed(1)} MB (90%
            Reduction)
          </p>
        </div>
      )}
    </div>
  );
};

export default App;

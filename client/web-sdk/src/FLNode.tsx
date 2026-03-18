import React, { useEffect, useState, useRef } from "react";
import {
  initializeWebGPU,
  executeDPQuantizationKernels,
} from "./webgpu-kernel";

export interface FLNodeProps {
  aggregatorUrl: string;
  tokenomicsWallet?: string;
  autoStart?: boolean;
  useSensors?: boolean; // Toggles GPS/Camera capture
}

export const FLNode: React.FC<FLNodeProps> = ({
  aggregatorUrl,
  tokenomicsWallet,
  autoStart = false,
  useSensors = true,
}) => {
  const [status, setStatus] = useState<
    "idle" | "initializing" | "active" | "error"
  >("idle");
  const [gpuDevice, setGpuDevice] = useState<any>(null);
  const [round, setRound] = useState(0);
  const [location, setLocation] = useState<{ lat: number; lon: number } | null>(
    null
  );
  const videoRef = useRef<HTMLVideoElement>(null);

  useEffect(() => {
    if (autoStart) {
      startNode();
    }
  }, [autoStart]);

  const startNode = async () => {
    setStatus("initializing");
    try {
      const device = await initializeWebGPU();
      setGpuDevice(device);
      setStatus("active");

      if (useSensors) {
        initializeSensors();
      }

      // Background thread logic would start here, listening for aggregator global weights
      setInterval(() => {
        setRound((r) => r + 1);
        // 1. Fetch global map state
        // 2. Extract local WebGL/WebGPU features from videoRef + location
        // 3. run local epoch (executeDPQuantizationKernels on difference)
        // 4. Submit via self-healing network channel
      }, 10000);
    } catch (e) {
      setStatus("error");
      console.error("FLNode Initialization Failed:", e);
    }
  };

  const initializeSensors = async () => {
    // 1. GPS Hook
    if (navigator.geolocation) {
      navigator.geolocation.watchPosition(
        (pos) => {
          setLocation({ lat: pos.coords.latitude, lon: pos.coords.longitude });
        },
        (err) => console.warn("GPS Denied: ", err)
      );
    }

    // 2. WebCam / Visual SLAM Hook
    try {
      const stream = await navigator.mediaDevices.getUserMedia({
        video: { facingMode: "environment" },
      });
      if (videoRef.current) {
        videoRef.current.srcObject = stream;
      }
    } catch (err) {
      console.warn("Camera Denied or Unavailable: ", err);
    }
  };

  if (status === "idle") {
    return (
      <button
        onClick={startNode}
        style={{ padding: "10px 20px", cursor: "pointer" }}
      >
        Join Sovereign Map Network
      </button>
    );
  }

  return (
    <div
      style={{
        padding: "20px",
        border: "1px solid #ccc",
        borderRadius: "8px",
        maxWidth: "400px",
      }}
    >
      <h3>Sovereign Map Web Node</h3>
      <p>
        Status: <strong>{status}</strong>
      </p>
      <p>Compute: {gpuDevice ? "🚀 WebGPU Accelerated" : "💻 CPU Fallback"}</p>
      <p>Current Epoch: {round}</p>

      {useSensors && (
        <div
          style={{
            marginTop: "15px",
            padding: "10px",
            background: "#f5f5f5",
            borderRadius: "4px",
          }}
        >
          <h4>Live Sensor Feed</h4>
          <p>
            📍 GPS:{" "}
            {location
              ? `${location.lat.toFixed(4)}, ${location.lon.toFixed(4)}`
              : "Waiting for lock..."}
          </p>
          <video
            ref={videoRef}
            autoPlay
            playsInline
            muted
            style={{
              width: "100%",
              height: "150px",
              background: "#000",
              borderRadius: "4px",
            }}
          />
        </div>
      )}

      {tokenomicsWallet ? (
        <p style={{ marginTop: "10px" }}>
          🪙 Rewards Routing to: {tokenomicsWallet.slice(0, 6)}...
        </p>
      ) : (
        <p style={{ marginTop: "10px" }}>Running as Volunteer</p>
      )}
    </div>
  );
};

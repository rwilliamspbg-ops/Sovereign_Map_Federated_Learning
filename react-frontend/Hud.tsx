// src/components/Hud.tsx - Neural Signal HUD with 3D Visualization
import React, { useEffect, useRef, useState } from 'react';
import * as THREE from 'three';
import { Activity, Cpu, HardDrive, Zap } from 'lucide-react';

interface HudData {
  latency_ns: number;
  latency_ms: number;
  spectral_density: number;
  mesh_nodes: number;
  active_enclaves: number;
  cxl_utilization: number;
  avg_stake: number;
  fl_round: number;
}

export const Hud: React.FC = () => {
  const mountRef = useRef<HTMLDivElement>(null);
  const [data, setData] = useState<HudData>({
    latency_ns: 0,
    latency_ms: 0,
    spectral_density: 0,
    mesh_nodes: 0,
    active_enclaves: 0,
    cxl_utilization: 0,
    avg_stake: 0,
    fl_round: 0
  });
  const [connected, setConnected] = useState(false);

  // Fetch HUD data from backend
  useEffect(() => {
    const fetchHudData = async () => {
      try {
        const res = await fetch('http://localhost:5000/hud_data');
        if (res.ok) {
          const json = await res.json();
          setData(json);
          setConnected(true);
        }
      } catch (error) {
        console.error('Failed to fetch HUD data:', error);
        setConnected(false);
      }
    };

    fetchHudData();
    const interval = setInterval(fetchHudData, 2000); // Update every 2 seconds
    return () => clearInterval(interval);
  }, []);

  // 3D Visualization with Three.js
  useEffect(() => {
    if (!mountRef.current) return;

    const scene = new THREE.Scene();
    scene.background = new THREE.Color(0x000510);
    scene.fog = new THREE.FogExp2(0x000510, 0.05);

    const camera = new THREE.PerspectiveCamera(
      75,
      window.innerWidth / (window.innerHeight * 0.7),
      0.1,
      1000
    );
    camera.position.z = 15;

    const renderer = new THREE.WebGLRenderer({ antialias: true });
    const width = mountRef.current.clientWidth;
    const height = mountRef.current.clientHeight;
    renderer.setSize(width, height);
    mountRef.current.appendChild(renderer.domElement);

    // Create neural mesh network visualization
    const nodes: THREE.Mesh[] = [];
    const nodeCount = 20;

    for (let i = 0; i < nodeCount; i++) {
      const geometry = new THREE.SphereGeometry(0.3, 32, 32);
      const material = new THREE.MeshStandardMaterial({
        color: 0x00ffff,
        emissive: 0x0088ff,
        emissiveIntensity: 0.5,
        metalness: 0.3,
        roughness: 0.4
      });
      const sphere = new THREE.Mesh(geometry, material);

      // Position nodes in a spiral
      const angle = (i / nodeCount) * Math.PI * 4;
      const radius = 5 + Math.sin(angle) * 2;
      sphere.position.x = Math.cos(angle) * radius;
      sphere.position.y = Math.sin(i * 0.5) * 3;
      sphere.position.z = Math.sin(angle) * radius;

      scene.add(sphere);
      nodes.push(sphere);
    }

    // Create connections between nodes
    const lineMaterial = new THREE.LineBasicMaterial({
      color: 0x00aaff,
      opacity: 0.3,
      transparent: true
    });

    for (let i = 0; i < nodeCount; i++) {
      const nextIndex = (i + 1) % nodeCount;
      const points = [
        nodes[i].position,
        nodes[nextIndex].position
      ];
      const geometry = new THREE.BufferGeometry().setFromPoints(points);
      const line = new THREE.Line(geometry, lineMaterial);
      scene.add(line);
    }

    // Lighting
    const ambientLight = new THREE.AmbientLight(0x404040, 2);
    scene.add(ambientLight);

    const pointLight = new THREE.PointLight(0x00ffff, 2, 100);
    pointLight.position.set(0, 10, 10);
    scene.add(pointLight);

    // Animation loop
    let animationId: number;
    const animate = () => {
      animationId = requestAnimationFrame(animate);

      // Rotate the entire scene
      scene.rotation.y += 0.002;

      // Pulse nodes based on activity
      nodes.forEach((node, i) => {
        const scale = 1 + Math.sin(Date.now() * 0.001 + i) * 0.1;
        node.scale.setScalar(scale);
      });

      renderer.render(scene, camera);
    };
    animate();

    // Cleanup
    return () => {
      cancelAnimationFrame(animationId);
      mountRef.current?.removeChild(renderer.domElement);
      scene.clear();
    };
  }, []);

  return (
    <div className="relative h-screen overflow-hidden">
      {/* 3D Visualization */}
      <div ref={mountRef} className="absolute inset-0" />

      {/* HUD Overlay */}
      <div className="absolute inset-0 pointer-events-none">
        {/* Top Left: System Status */}
        <div className="absolute top-6 left-6 space-y-4 pointer-events-auto">
          <div className="bg-black/70 backdrop-blur-lg border border-blue-500/50 rounded-lg p-4 min-w-[300px]">
            <h2 className="text-xl font-bold mb-3 text-blue-400 flex items-center gap-2">
              <Activity size={24} />
              Neural Signal HUD
            </h2>
            
            <div className="space-y-2 text-sm">
              <div className="flex justify-between">
                <span className="text-gray-400">Connection:</span>
                <span className={connected ? 'text-green-400' : 'text-red-400'}>
                  {connected ? '● ONLINE' : '● OFFLINE'}
                </span>
              </div>
              <div className="flex justify-between">
                <span className="text-gray-400">FL Round:</span>
                <span className="text-white font-mono">#{data.fl_round}</span>
              </div>
            </div>
          </div>

          {/* Performance Metrics */}
          <div className="bg-black/70 backdrop-blur-lg border border-blue-500/50 rounded-lg p-4">
            <h3 className="text-lg font-semibold mb-2 text-blue-400 flex items-center gap-2">
              <Zap size={20} />
              Performance
            </h3>
            
            <div className="space-y-2 text-sm">
              <div>
                <div className="flex justify-between mb-1">
                  <span className="text-gray-400">Latency:</span>
                  <span className="text-white font-mono">{data.latency_ms.toFixed(3)} ms</span>
                </div>
                <div className="w-full bg-gray-700 rounded-full h-2">
                  <div 
                    className="bg-gradient-to-r from-green-400 to-blue-500 h-2 rounded-full transition-all"
                    style={{ width: `${Math.min(100, (200 - data.latency_ms) / 2)}%` }}
                  />
                </div>
              </div>

              <div>
                <div className="flex justify-between mb-1">
                  <span className="text-gray-400">Signal Density:</span>
                  <span className="text-white font-mono">{(data.spectral_density * 100).toFixed(1)}%</span>
                </div>
                <div className="w-full bg-gray-700 rounded-full h-2">
                  <div 
                    className="bg-gradient-to-r from-purple-400 to-pink-500 h-2 rounded-full transition-all"
                    style={{ width: `${data.spectral_density * 100}%` }}
                  />
                </div>
              </div>
            </div>
          </div>
        </div>

        {/* Top Right: Network Status */}
        <div className="absolute top-6 right-6 space-y-4 pointer-events-auto">
          <div className="bg-black/70 backdrop-blur-lg border border-blue-500/50 rounded-lg p-4 min-w-[280px]">
            <h3 className="text-lg font-semibold mb-3 text-blue-400 flex items-center gap-2">
              <Cpu size={20} />
              Neural Mesh
            </h3>
            
            <div className="space-y-3 text-sm">
              <div className="flex items-center justify-between">
                <span className="text-gray-400">Active Nodes:</span>
                <span className="text-2xl font-bold text-green-400">{data.mesh_nodes}</span>
              </div>
              
              <div className="flex items-center justify-between">
                <span className="text-gray-400">Avg Stake:</span>
                <span className="text-lg font-mono text-yellow-400">{data.avg_stake.toFixed(0)}</span>
              </div>
            </div>
          </div>

          {/* CXL Memory Status */}
          <div className="bg-black/70 backdrop-blur-lg border border-blue-500/50 rounded-lg p-4">
            <h3 className="text-lg font-semibold mb-3 text-blue-400 flex items-center gap-2">
              <HardDrive size={20} />
              CXL 3.2 Pool
            </h3>
            
            <div className="space-y-3 text-sm">
              <div>
                <div className="flex justify-between mb-1">
                  <span className="text-gray-400">Utilization:</span>
                  <span className="text-white font-mono">{data.cxl_utilization.toFixed(1)}%</span>
                </div>
                <div className="w-full bg-gray-700 rounded-full h-2">
                  <div 
                    className="bg-gradient-to-r from-cyan-400 to-blue-600 h-2 rounded-full transition-all"
                    style={{ width: `${data.cxl_utilization}%` }}
                  />
                </div>
              </div>

              <div className="flex items-center justify-between">
                <span className="text-gray-400">Active Enclaves:</span>
                <span className="text-lg font-bold text-cyan-400">{data.active_enclaves}</span>
              </div>
            </div>
          </div>
        </div>

        {/* Bottom Center: Instructions */}
        <div className="absolute bottom-6 left-1/2 transform -translate-x-1/2 pointer-events-auto">
          <div className="bg-black/70 backdrop-blur-lg border border-blue-500/50 rounded-lg px-6 py-3">
            <p className="text-sm text-gray-300">
              Neural mesh status: <span className="text-green-400 font-semibold">NOMINAL</span> • 
              CXL 3.2 tiering: <span className="text-blue-400 font-semibold">ACTIVE</span> • 
              TSP security: <span className="text-green-400 font-semibold">ENABLED</span>
            </p>
          </div>
        </div>
      </div>
    </div>
  );
};

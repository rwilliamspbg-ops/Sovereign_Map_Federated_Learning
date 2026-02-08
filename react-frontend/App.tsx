// src/App.tsx - Main Application
import React from 'react';
import { BrowserRouter as Router, Routes, Route, Link } from 'react-router-dom';
import { Hud } from './components/Hud';
import { VoiceLink } from './components/VoiceLink';
import { DaoGovernance } from './components/DaoGovernance';
import { NetworkStatus } from './components/NetworkStatus';
import { Activity, Mic, Users, Network } from 'lucide-react';

function App() {
  return (
    <Router>
      <div className="min-h-screen bg-gradient-to-br from-gray-900 via-blue-900 to-gray-900 text-white">
        {/* Navigation */}
        <nav className="bg-black/30 backdrop-blur-lg border-b border-blue-500/30 p-4">
          <div className="container mx-auto flex items-center justify-between">
            <h1 className="text-2xl font-bold text-blue-400">
              🗺️ Sovereign Maps
            </h1>
            <div className="flex gap-4">
              <Link to="/" className="flex items-center gap-2 px-4 py-2 rounded-lg bg-blue-500/20 hover:bg-blue-500/30 transition">
                <Activity size={20} />
                <span>HUD</span>
              </Link>
              <Link to="/voice" className="flex items-center gap-2 px-4 py-2 rounded-lg bg-blue-500/20 hover:bg-blue-500/30 transition">
                <Mic size={20} />
                <span>Voice</span>
              </Link>
              <Link to="/dao" className="flex items-center gap-2 px-4 py-2 rounded-lg bg-blue-500/20 hover:bg-blue-500/30 transition">
                <Users size={20} />
                <span>DAO</span>
              </Link>
              <Link to="/network" className="flex items-center gap-2 px-4 py-2 rounded-lg bg-blue-500/20 hover:bg-blue-500/30 transition">
                <Network size={20} />
                <span>Network</span>
              </Link>
            </div>
          </div>
        </nav>

        {/* Routes */}
        <Routes>
          <Route path="/" element={<Hud />} />
          <Route path="/voice" element={<VoiceLink />} />
          <Route path="/dao" element={<DaoGovernance />} />
          <Route path="/network" element={<NetworkStatus />} />
        </Routes>
      </div>
    </Router>
  );
}

export default App;

import { useState, useEffect } from 'react'
import HUD from './HUD'
import './App.css'

const API_BASE = 'http://localhost:5000'

function App() {
  const [hudData, setHudData] = useState(null)
  const [health, setHealth] = useState(null)
  const [metricsSummary, setMetricsSummary] = useState(null)
  const [founders, setFounders] = useState([])
  const [voiceQuery, setVoiceQuery] = useState('')
  const [voiceResponse, setVoiceResponse] = useState('')
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState(null)

  useEffect(() => {
    fetchData()
    const interval = setInterval(fetchData, 5000)
    return () => clearInterval(interval)
  }, [])

  const fetchData = async () => {
    setLoading(true);
    try {
      // 1. Correctly close the Promise.all array and await the responses
      const [hudRes, healthRes, metricsRes, foundersRes] = await Promise.all([
        fetch(`${API_BASE}/hud_data`),
        fetch(`${API_BASE}/health`),
        fetch(`${API_BASE}/metrics_summary`),
        fetch(`${API_BASE}/founders`)
      ]);

      // 2. Parse the JSON from each response
      const [hud, health, metrics, foundersData] = await Promise.all([
        hudRes.json(),
        healthRes.json(),
        metricsRes.json(),
        foundersRes.json()
      ]);

      // 3. Update the state variables
      setHudData(hud);
      setHealth(health);
      setMetricsSummary(metrics);
      setFounders(foundersData);
      setError(null);
    } catch (err) {
      console.error("Fetch error:", err);
      setError("Failed to sync with Sovereign Node");
    } finally {
      setLoading(false);
    }
  }; // This closes fetchData

  return (
    <div className="hud-container">
      <header className="hud-header">
        <h1>🗺️ Sovereign Map - Federated Learning HUD</h1>
        <div className="status-bar">
          {error ? (
            <span className="status-error">⚠️ {error}</span>
          ) : (
            <span className="status-ok">✅ System Online</span>
          )}
        </div>
      </header>

      <HUD 
        hudData={hudData} 
        health={health} 
        metricsSummary={metricsSummary} 
        founders={founders} 
        voiceQuery={voiceQuery} 
        voiceResponse={voiceResponse} 
        loading={loading} 
        error={error} 
        onTriggerFLRound={() => console.log("Triggering...")} 
        onCreateEnclave={() => console.log("Creating...")} 
      />

      {metricsSummary && (
        <div className="metrics-footer">
          <div className="summary-item">
            <strong>Total FL Rounds:</strong> {metricsSummary.fl_rounds_total}
          </div>
          <div className="summary-item">
            <strong>Avg FL Duration:</strong> {metricsSummary.avg_fl_duration?.toFixed(2) || 'N/A'}s
          </div>
          <div className="summary-item">
            <strong>Total Stake:</strong> {metricsSummary.total_stake?.toFixed(2) || 'N/A'}
          </div>
          <div className="summary-item">
            <strong>CXL Utilization:</strong> {(metricsSummary.cxl_utilization * 100)?.toFixed(1) || 'N/A'}%
          </div>
        </div>
      )}
    </div>
  );
} // This closes the App function

export default App;

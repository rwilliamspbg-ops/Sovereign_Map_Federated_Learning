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
    try {
      const [hudRes, healthRes, metricsRes, foundersRes] = await Promise.all([
        fetch(`${API_BASE}/hud_data`),
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
                onTriggerFLRound={triggerFLRound}
                onCreateEnclave={createEnclave}
                onSubmitVoiceQuery={submitVoiceQuery}
                setVoiceQuery={setVoiceQuery}
              />
              {/* ...existing grid layout and cards can be kept or removed as needed... */}
            </div>
          )
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
          </div>
        )}

        <div className="card voice-interface">
          <h2>🎤 Voice Query Interface</h2>
          <div className="voice-controls">
            <input 
              type="text"
              value={voiceQuery}
              onChange={(e) => setVoiceQuery(e.target.value)}
              onKeyPress={(e) => e.key === 'Enter' && submitVoiceQuery()}
              placeholder="Enter voice command (e.g., 'scan', 'stake', 'enclave')"
              className="voice-input"
            />
            <button 
              onClick={submitVoiceQuery}
              disabled={loading || !voiceQuery.trim()}
              className="btn btn-voice"
            >
              Send Query
            </button>
          </div>
          {voiceResponse && (
            <div className="voice-response">
              <strong>Response:</strong> {voiceResponse}
            </div>
          )}
        </div>

        <div className="card monitoring-links">
          <h2>📊 Monitoring Tools</h2>
          <div className="link-buttons">
            <a href="http://localhost:9090" target="_blank" rel="noopener noreferrer" className="btn btn-link">
              📉 Prometheus
            </a>
            <a href="http://localhost:3001" target="_blank" rel="noopener noreferrer" className="btn btn-link">
              📊 Grafana
            </a>
            <a href="http://localhost:3100" target="_blank" rel="noopener noreferrer" className="btn btn-link">
              📝 Loki
            </a>
            <a href="http://localhost:5000/metrics" target="_blank" rel="noopener noreferrer" className="btn btn-link">
              🔢 Raw Metrics
            </a>
          </div>
        </div>

        {founders.length > 0 && (
          <div className="card dao-founders">
            <h2>👥 DAO Founders ({founders.length})</h2>
            <div className="founders-list">
              {founders.slice(0, 10).map(founder => (
                <div key={founder.id} className="founder-item">
                  <span className="founder-name">{founder.name}</span>
                  <span className="founder-country">{founder.country}</span>
                  <span className={`founder-verified ${founder.verified ? 'verified' : 'unverified'}`}>
                    {founder.verified ? '✓' : '✗'}
                  </span>
                </div>
              ))}
            </div>
          </div>
        )}
      </div>
    </div>
  )
}

export default App

import re

with open("frontend/src/App.css", "r") as f:
    css = f.read()

# Replace the root variables to be a hyper-modern dark HUD theme
css = re.sub(
    r":root\s*{[^}]+}",
    """@import url('https://fonts.googleapis.com/css2?family=Orbitron:wght@400;700&family=Rajdhani:wght@400;600&display=swap');

:root {
  --bg-canvas: #040814;
  --card-bg: rgba(10, 15, 30, 0.7);
  --text-main: #e2f1ff;
  --text-muted: #648ba6;
  --stroke: rgba(0, 240, 255, 0.3);
  --brand: #00f0ff;
  --brand-2: #ff0055;
  --danger: #ff2a2a;
  --panel-glow: 0 0 20px rgba(0, 240, 255, 0.1) inset;
  --font-hud: 'Orbitron', monospace;
  --font-data: 'Rajdhani', sans-serif;
}

body {
  background-color: var(--bg-canvas);
  background-image: 
    linear-gradient(rgba(0, 240, 255, 0.05) 1px, transparent 1px),
    linear-gradient(90deg, rgba(0, 240, 255, 0.05) 1px, transparent 1px);
  background-size: 30px 30px;
  background-position: center center;
  font-family: var(--font-data);
}

.hud-container {
  padding: 1.25rem;
  max-width: 1400px;
  margin: 0 auto;
  min-height: 100vh;
  color: var(--text-main);
  animation: crtFlicker 0.15s infinite;
}

h1, h2, h3, h4 {
  font-family: var(--font-hud);
  text-transform: uppercase;
  letter-spacing: 2px;
  text-shadow: 0 0 10px rgba(0, 240, 255, 0.5);
}

/* Animations */
@keyframes pulseGlow {
  0% { box-shadow: 0 0 10px rgba(0, 240, 255, 0.2); }
  50% { box-shadow: 0 0 20px rgba(0, 240, 255, 0.6), 0 0 30px rgba(0, 240, 255, 0.4) inset; }
  100% { box-shadow: 0 0 10px rgba(0, 240, 255, 0.2); }
}

@keyframes typeWriter {
  from { width: 0; }
  to { width: 100%; }
}

@keyframes datastream {
  0% { transform: translateY(-100%); opacity: 0; }
  50% { opacity: 1; }
  100% { transform: translateY(100%); opacity: 0; }
}

@keyframes crtFlicker {
  0% { opacity: 0.98; }
  50% { opacity: 1; }
  100% { opacity: 0.99; }
}

.hud-section {
  background: var(--card-bg) !important;
  border: 1px solid var(--stroke) !important;
  box-shadow: var(--panel-glow);
  backdrop-filter: blur(10px);
  border-radius: 4px !important;
  position: relative;
  overflow: hidden;
}

.hud-section::before {
  content: "";
  position: absolute;
  top: 0;
  left: -100%;
  width: 50%;
  height: 2px;
  background: linear-gradient(90deg, transparent, var(--brand), transparent);
  animation: scanline 3s linear infinite;
}

@keyframes scanline {
  0% { left: -100%; }
  100% { left: 200%; }
}

button {
  font-family: var(--font-hud) !important;
  text-transform: uppercase;
  background: rgba(0, 240, 255, 0.1) !important;
  border: 1px solid var(--brand) !important;
  color: var(--brand) !important;
  transition: all 0.3s ease !important;
  position: relative;
  overflow: hidden;
}

button:hover {
  background: var(--brand) !important;
  color: #000 !important;
  box-shadow: 0 0 15px var(--brand);
}

.btn-primary {
  animation: pulseGlow 2s infinite;
}

.btn-secondary {
  border-color: var(--brand-2) !important;
  color: var(--brand-2) !important;
}
.btn-secondary:hover {
  background: var(--brand-2) !important;
  color: #fff !important;
  box-shadow: 0 0 15px var(--brand-2);
}

.voice-response {
  border-left: 3px solid var(--brand-2);
  padding-left: 1rem;
  font-family: monospace;
  color: var(--brand-2);
  background: rgba(255, 0, 85, 0.05);
  display: inline-block;
  overflow: hidden;
  white-space: pre-wrap;
  animation: typeWriter 2s steps(60, end);
}

.active-visualizer {
  height: 20px;
  display: flex;
  align-items: center;
  gap: 3px;
  margin: 10px 0;
}
.active-visualizer .bar {
  width: 4px;
  background: var(--brand);
  height: 100%;
  animation: equalize 1s infinite;
}
.active-visualizer .bar:nth-child(1) { animation-duration: 0.9s; }
.active-visualizer .bar:nth-child(2) { animation-duration: 0.6s; }
.active-visualizer .bar:nth-child(3) { animation-duration: 1.2s; }
.active-visualizer .bar:nth-child(4) { animation-duration: 0.8s; }
.active-visualizer .bar:nth-child(5) { animation-duration: 1.1s; }

@keyframes equalize {
  0%, 100% { height: 20%; }
  50% { height: 100%; }
}
""",
    css,
    flags=re.DOTALL | re.IGNORECASE,
)

with open("frontend/src/App.css", "w") as f:
    f.write(css)

# Phase 3C: Browser Federated Learning Demo

## Status (2026-03-23)

This document is retained as implementation history.

- Browser FL Demo components remain in the repository for reference and targeted testing.
- The primary operator UX is now HUD-first.
- The app shell no longer exposes Browser FL Demo as a primary navigation mode.

## Overview

Phase 3C introduced an in-browser federated learning simulation UI inside the frontend app. The production-facing app flow has since been consolidated to HUD-first operation.

The Browser FL Demo provides an interactive way to tune federated training inputs and observe convergence, communication cost, privacy/compression tradeoffs, and runtime backend state (WebGPU vs CPU fallback) when run directly for exploration.

## Where It Lives

- Frontend app shell: frontend/src/App.jsx
- Browser simulator: frontend/src/BrowserFLDemo.jsx
- Browser simulator styles: frontend/src/BrowserFLDemo.css
- Shared app styles: frontend/src/App.css
- Global theme styles: frontend/src/index.css

## Key Capabilities

### 1. Runtime Detection

The simulator detects browser runtime support for WebGPU:

- If navigator.gpu and adapter are available: runtime badge shows WebGPU Active
- Otherwise: runtime badge shows CPU Fallback

### 2. Scenario Controls

Users can tune:

- Participants (10-500)
- Local epochs (1-10)
- Privacy epsilon (0.2-3.0)
- Compression bits (4-16)
- Learning rate (0.005-0.08)
- Target rounds (5-400)

### 3. Live Simulation Metrics

The panel updates each round with:

- Current round
- Accuracy
- Loss
- Compression ratio
- Bandwidth per round (KB)
- Round latency (ms)
- Progress toward target rounds

### 4. Visualizations

Two charts are rendered using Recharts:

- Convergence chart (accuracy and loss over rounds)
- Communication cost chart (bandwidth over rounds)

### 5. Existing HUD Preserved

The prior operational HUD and governance controls remain available in the second tab, including:

- FL round trigger
- Enclave controls
- Voice query
- Trust verification policy controls
- Policy history and telemetry JSON viewers

## Simulation Model Notes

The simulator is intentionally lightweight and deterministic enough for UX exploration:

- Accuracy gain decays over time with exponential convergence
- Participant count and local epochs increase convergence rate
- Lower epsilon can add privacy-related utility penalty
- Lower compression bit depth can add reconstruction penalty
- WebGPU mode reduces latency and adds a small throughput boost
- Compression ratio and transferred bandwidth are updated each round

This is a controlled product demo model, not a scientific benchmark implementation.

## Run Locally

From repository root:

```bash
cd frontend
npm install
npm run dev
```

Build production bundle:

```bash
cd frontend
npm run build
```

## Validation Status

- Frontend build: passing
- Frontend test command: no tests are currently defined in frontend, so vitest exits with code 1 when run with --run

## Recommended Next Enhancements

1. Add integration tests for tab switching and chart updates.
2. Add real backend hooks for round telemetry instead of simulation-only data.
3. Add dynamic import code splitting for chart-heavy bundle optimization.
4. Add scenario presets (edge device, regional cluster, global scale).

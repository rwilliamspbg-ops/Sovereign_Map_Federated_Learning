# PySyft x Sovereign Mohawk PoC

This folder is a ready-to-demo starter for integrating PySyft Datasite FL workflows with Sovereign Mohawk aggregation.

## What this delivers

- PySyft-compatible round loop (mock mode and Datasite mode)
- Mohawk aggregation hook with optional proof verification and gradient compression
- Docker Compose one-command demo path
- Notebook for show-and-tell

## Quick start (mock mode)

```bash
cd /workspaces/Sovereign_Map_Federated_Learning
python -m venv .venv
source .venv/bin/activate
pip install -r examples/pysyft-integration/requirements-pysyft-demo.txt
python examples/pysyft-integration/pysyft_mohawk_poc.py --mode mock --rounds 5 --participants 5
```

## One-command demo with Docker Compose

```bash
cd /workspaces/Sovereign_Map_Federated_Learning/examples/pysyft-integration
docker compose -f docker-compose.pysyft-demo.yml up --build
```

## Datasite mode (live PySyft)

```bash
python examples/pysyft-integration/pysyft_mohawk_poc.py \
  --mode datasite \
  --rounds 3 \
  --server-url http://localhost:8080 \
  --email you@example.com \
  --password 'your-password' \
  --dataset YourDataset \
  --asset YourData
```

Notes:

- Datasite mode expects a reachable Syft server and a dataset asset already registered.
- Depending on your Syft deployment policy, remote code may require explicit approval before execution.
- If the Python `mohawk` package is not available, the script falls back to numpy averaging in mock mode.

## Demo script architecture

- `pysyft_mohawk_poc.py`
- `requirements-pysyft-demo.txt`
- `docker-compose.pysyft-demo.yml`
- `pysyft_mohawk_poc.ipynb`

## Suggested 10-minute demo flow

1. Run mock mode and show round-by-round metrics + proof status.
2. Open the notebook and walk through the aggregation boundary (`mohawk_aggregate`).
3. Switch to Datasite mode command to show production-ready path.
4. Close with why this is high leverage: PySyft policy layer + Mohawk verifiable BFT aggregation.

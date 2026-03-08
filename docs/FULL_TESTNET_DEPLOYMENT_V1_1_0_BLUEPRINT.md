# Full-Testnet-Deployment V1.1.0 Blueprint

This document turns the 200-node testnet goals into an implementable repository plan.

## Current Reality

The repository already includes meaningful components:
- `internal/p2p` (networking foundation)
- `monitoring/` and Grafana assets
- `genesis-launch.sh`
- Byzantine/zk/privacy research and testing artifacts

The main remaining gap is operational integration: node runtime UX, bootstrapping, scheduler orchestration, model distribution workflows, and repeatable deployment entrypoints.

## V1.1.0 Deliverables Added

1. Bootstrap network configs:
- `network/bootstrap/bootstrap_nodes.json`
- `network/bootstrap/seed_peers.json`
- `network/bootstrap/network_config.json`

2. Operator node entrypoint:
- `cmd/sovereign-node/main.go`
- `scripts/install_node.sh`

3. Training coordinator starter:
- `internal/scheduler/scheduler.go`

4. Model distribution starter:
- `internal/modeldist/store.go`

5. Genesis wrapper and deploy artifact:
- `scripts/launch_genesis.sh`
- `deploy/kubernetes/bootstrap_cluster.yaml`

## 30-45 Day Build Sequence

1. Week 1:
- Wire `internal/p2p` into `sovereign-node`
- Join/bootstrap flow from `network/bootstrap/*.json`

2. Week 2:
- Implement scheduler timeout/straggler policy
- Connect model checkpoint publish/fetch lifecycle

3. Week 3:
- Add Prometheus counters for rounds, peers, and model versions
- Deploy bootstrap + workers in k8s

4. Week 4:
- 50-node soak test with fault injection

5. Week 5-6:
- 200-node testnet launch

## Success Criteria

- 200 active nodes with stable peer discovery
- 95%+ round completion under configured timeout
- Checkpoint version consistency across nodes
- Full observability for peer health and training rounds

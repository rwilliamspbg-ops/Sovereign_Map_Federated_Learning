# Bootstrap Network Assets

These files provide the baseline runtime inputs expected by `cmd/sovereign-node` and the Ansible playbook.

- `network_config.json` is a usable default for a first bootstrap node.
- `bootstrap_nodes.json` and `seed_peers.json` intentionally start empty so a `start` node can come up without stale peer metadata.

To add joiner nodes after the first bootstrap node is running:

1. Capture the bootstrap node's reachable listen address and peer ID.
2. Add a full multiaddr entry to `bootstrap_nodes.json`, for example:

```json
[
  {
    "id": "bootstrap-1",
    "multiaddr": "/ip4/203.0.113.10/udp/4001/quic-v1/p2p/12D3KooWExamplePeerIDReplaceMe",
    "region": "us-east-1",
    "role": "bootstrap"
  }
]
```

Then:

1. Optionally add the same reachable addresses as plain strings in `seed_peers.json`.
2. Run the Ansible playbook with `-e node_mode=join` for worker nodes.

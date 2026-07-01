# Ansible Deployment

This directory contains the playbook and inventory scaffold for deploying `sovereign-node` onto Linux hosts with `systemd`.

## Files

- `deploy_nodes.yml`: Builds or uploads the `sovereign-node` binary, installs runtime config, writes a `systemd` unit, and enables the service.
- `inventory.example.ini`: Example inventory layout for one bootstrap node and two joiner nodes.

## Prerequisites

1. SSH access to the target hosts.
2. Privilege escalation on those hosts for writing `/usr/local/bin`, `/opt/sovereign-map`, and `/etc/systemd/system`.
3. A reachable bootstrap address and peer ID before deploying joiner nodes.

## Bootstrap Node

Copy the sample inventory and adjust hostnames, SSH users, and IPs:

```bash
cp deploy/ansible/inventory.example.ini deploy/ansible/inventory.ini
```

Deploy the first bootstrap node using the repo-managed defaults:

```bash
PATH="$HOME/.local/bin:$PATH" ansible-playbook \
  deploy/ansible/deploy_nodes.yml \
  -i deploy/ansible/inventory.ini \
  --limit bootstrap \
  --become
```

After it starts, capture the reachable libp2p address and peer ID from the node logs or runtime output.

## Joiner Nodes

Update `network/bootstrap/bootstrap_nodes.json` with the bootstrap node multiaddr and peer ID, then optionally mirror those addresses in `network/bootstrap/seed_peers.json`.

Example bootstrap entry:

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

Deploy the joiner nodes:

```bash
PATH="$HOME/.local/bin:$PATH" ansible-playbook \
  deploy/ansible/deploy_nodes.yml \
  -i deploy/ansible/inventory.ini \
  --limit workers \
  --become \
  -e node_mode=join
```

## Optional Overrides

- Set `-e local_build_binary=false -e local_binary_path=/path/to/sovereign-node` to reuse a prebuilt binary.
- Set `-e service_name=custom-sovereign-node` to install under a different `systemd` unit name.
- Set `-e remote_install_root=/custom/path` if the target hosts require a different runtime directory.

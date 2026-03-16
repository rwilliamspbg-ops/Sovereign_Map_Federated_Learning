# Local Participant Join and Training

This guide tests self-serve participant onboarding locally using the Join API and participant compose profile.

## Prerequisites

- Docker with compose
- curl and jq

## 1. Start coordinator stack

```bash
export JOIN_API_ADMIN_TOKEN=local-dev-admin-token
export PUBLIC_AGGREGATOR_HOST=backend
export PUBLIC_AGGREGATOR_PORT=8080

./scripts/run_channel.sh dev
```

The backend exposes:

- Join API: http://localhost:8000/join/*
- Flower aggregator: localhost:8080

## 2. Bootstrap a participant

```bash
scripts/participant_bootstrap.sh \
  --backend-url http://localhost:8000 \
  --participant-name alice \
  --admin-token local-dev-admin-token
```

This creates:

- participants/alice/certs/node-cert.pem
- participants/alice/certs/node-key.pem
- participants/alice/certs/ca-cert.pem
- participants/alice/.participant.env
- participants/alice/join-registration.json

And starts a local participant client using docker-compose.participant.yml.

## 3. Verify policy and registrations

```bash
curl -s http://localhost:8000/llm_policy | jq

curl -s -H "X-Join-Admin-Token: local-dev-admin-token" \
  http://localhost:8000/join/registrations | jq
```

## 4. Revoke a participant

```bash
curl -s -X POST \
  -H "X-Join-Admin-Token: local-dev-admin-token" \
  http://localhost:8000/join/revoke/1 | jq
```

## 5. Stop participant

```bash
docker compose -f docker-compose.participant.yml \
  --env-file participants/alice/.participant.env down
```

# Despaddosy

Enterprise-style defensive edge gateway + demo social network.

## Quick start
```bash
make up
```

Then open: http://localhost:8080

Admin login:
- `admin@example.com`
- `admin1234`

## Services
- `gateway`: L7 proxy with policy, rate limits, risk scoring.
- `riskd`: Rule-based risk scoring with Redis state.
- `sociald`: Demo social network with admin console.

## Development
```bash
make build
make test
```

## Docs
- `docs/ARCHITECTURE.md`
- `docs/THREAT_MODEL.md`
- `docs/RUNBOOK.md`
- `docs/API.md`
- `docs/CONFIG.md`

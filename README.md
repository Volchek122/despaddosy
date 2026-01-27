# Despaddosy

Enterprise-style defensive edge gateway + demo social network.

## Quick start (Linux/macOS)
```bash
./scripts/up.sh
```

## Quick start (Windows PowerShell)
```powershell
.\scripts\up.ps1
```

Then open: http://localhost:8080

Admin login:
- `admin@example.com`
- `admin1234`

## Stop stack
```bash
./scripts/down.sh
```

```powershell
.\scripts\down.ps1
```

## Services
- `gateway`: L7 proxy with policy, rate limits, risk scoring.
- `riskd`: Rule-based risk scoring with Redis state.
- `sociald`: Demo social network with admin console.

## Development (Linux/macOS)
```bash
go build ./cmd/gateway
go build ./cmd/riskd
go build ./cmd/sociald
go build ./cmd/migrate

go test ./...
```

## Development (Windows)
```powershell
go build ./cmd/gateway
go build ./cmd/riskd
go build ./cmd/sociald
go build ./cmd/migrate

go test ./...
```

## Docs
- `docs/ARCHITECTURE.md`
- `docs/THREAT_MODEL.md`
- `docs/RUNBOOK.md`
- `docs/API.md`
- `docs/CONFIG.md`

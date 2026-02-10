# Despaddosy

Enterprise-style defensive edge gateway + demo social network.

## Prerequisites
- Go 1.22+
- Docker Desktop / Docker Engine
- Internet access for `go mod download` on first build
  - If you see `missing go.sum entry`, run `go mod download` or `go mod tidy`.

## Quick start (Linux/macOS)
```bash
./scripts/up.sh
```

## Quick start (Windows PowerShell)
```powershell
.\scripts\up.ps1
```
Note: use `dir` or `Get-ChildItem` instead of `ls` in Windows PowerShell.

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
# fetch deps once
go mod download

go build ./cmd/gateway
go build ./cmd/riskd
go build ./cmd/sociald
go build ./cmd/migrate

go test ./...
```

## Development (Windows)
```powershell
# fetch deps once
go mod download

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

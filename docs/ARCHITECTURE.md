# Architecture

## Components
- **gateway**: L7 reverse proxy with policy enforcement, rate limiting, and risk scoring decisions.
- **riskd**: Risk scoring service with rule-based decisions and Redis-backed reputation/blocks.
- **sociald**: Demo social network backend with SSR UI and admin console.
- **redis/postgres**: State for counters, blocks, and application data.

## Data Flow
1. Client connects to gateway (TLS termination optional in demo).
2. Gateway normalizes request, enforces size/timeout/auth, checks rate limits.
3. Features are extracted without logging raw bodies or secrets.
4. Riskd returns a decision and reasons; gateway enforces allow/monitor/challenge/block.
5. On allow, gateway proxies to sociald.

## No MITM packet sniffing
Traffic analysis is performed at L7 after TLS termination by the gateway. The system does **not** inspect raw packets or attempt to decrypt user TLS sessions outside of gateway termination.

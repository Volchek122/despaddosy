# Threat Model

## Threats
- Automated abuse: rapid posting, like spam, scraping.
- Brute-force login attempts.
- Injection attempts via query or form parameters.

## Mitigations
- JWT auth + RBAC for admin routes.
- Rate limiting per IP/user/route.
- Request size limits and timeouts.
- WAF-like pattern flags and entropy checks feeding risk scoring.
- Temporary TTL blocks for suspicious behavior.

# API

## Sociald
- `GET /api/posts`
- `POST /api/posts`
- `POST /api/posts/{id}/like`
- `DELETE /api/posts/{id}/like`
- `GET /api/posts/{id}/comments`
- `POST /api/posts/{id}/comments`
- `POST /api/reports`

## Riskd
- `POST /v1/score`
  - Request: `{ "features": { ... }, "auth_hint": bool, "max_body_hint": int }`
  - Response: `{ "decision": "allow|monitor|challenge|block", "score": 0-100, "reasons": [], "ttl_seconds": 120 }`

#!/usr/bin/env bash
set -euo pipefail

docker compose -f deploy/docker-compose/docker-compose.yml up --build

.PHONY: build test lint up down

build:
	go build ./cmd/gateway
	go build ./cmd/riskd
	go build ./cmd/sociald
	go build ./cmd/migrate

test:
	go test ./...

lint:
	golangci-lint run

up:
	docker compose -f deploy/docker-compose/docker-compose.yml up --build

down:
	docker compose -f deploy/docker-compose/docker-compose.yml down

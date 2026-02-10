FROM golang:1.22 AS builder
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download || true
COPY . .
ARG SERVICE
RUN CGO_ENABLED=0 go build -o /out/${SERVICE} ./cmd/${SERVICE}

FROM alpine:3.19
WORKDIR /app
COPY --from=builder /out/${SERVICE} /app/${SERVICE}
COPY configs /app/configs
COPY web /app/web
COPY migrations /app/migrations
EXPOSE 8080 8081 8082
ENTRYPOINT ["/app/${SERVICE}"]

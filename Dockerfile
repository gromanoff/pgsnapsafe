FROM golang:1.23 AS builder

ARG POSTGRES_VERSION=15
RUN apt-get update && apt-get install -y \
    postgresql-client-${POSTGRES_VERSION} \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod tidy

COPY . .
RUN go build -o pgsnapsafe cmd/app/main.go

FROM debian:bookworm-slim

ARG POSTGRES_VERSION=15
RUN apt-get update && apt-get install -y postgresql-client-${POSTGRES_VERSION} && rm -rf /var/lib/apt/lists/*

WORKDIR /app

COPY --from=builder /app/pgsnapsafe /usr/local/bin/pgsnapsafe

COPY /pkg/email/template /app

COPY config-example.yml .env /app/

RUN chmod +x /usr/local/bin/pgsnapsafe

CMD ["/usr/local/bin/pgsnapsafe"]

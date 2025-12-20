# syntax=docker/dockerfile:1

FROM golang:1.24-bookworm AS dev
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

# utilities for scripts (mysqladmin for wait-mysql.sh)
RUN apt-get update \
 && apt-get install -y --no-install-recommends default-mysql-client \
 && rm -rf /var/lib/apt/lists/*

COPY . .

EXPOSE 8080
CMD ["go", "run", "./cmd/server"]

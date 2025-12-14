# syntax=docker/dockerfile:1

FROM golang:1.24-bookworm AS dev
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

EXPOSE 8080
CMD ["go", "run", "./cmd/server"]

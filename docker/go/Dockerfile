#   Dev stages
FROM golang:1.26.3-alpine AS base

WORKDIR /app

RUN apk add --no-cache ca-certificates git

COPY go.mod ./
RUN go mod download

COPY . .

FROM base AS dev

RUN go install github.com/air-verse/air@latest

EXPOSE 8080

CMD ["air", "-c", ".air.toml"]

#   Prod stages
FROM base AS builder

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /app/bin/cloud_storage ./cmd/cloud_storage

FROM alpine:3.22 AS prod

WORKDIR /app

RUN apk add --no-cache ca-certificates

COPY --from=builder /app/bin/cloud_storage ./bin/cloud_storage

EXPOSE 8080

CMD ["./bin/cloud_storage"]

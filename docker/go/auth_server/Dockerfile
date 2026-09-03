FROM golang:1.26.3-alpine AS base

WORKDIR /app

RUN apk add --no-cache ca-certificates git

COPY go.mod go.sum ./
RUN go mod download

COPY . .

FROM base AS dev

RUN go install github.com/air-verse/air@latest

EXPOSE 7070

CMD ["air", "-c", "./docker/go/auth_server/.air.toml"]

FROM base AS builder

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /bin/auth_server ./cmd/auth_server

FROM alpine:3.22 AS prod

WORKDIR /app

RUN apk add --no-cache ca-certificates

COPY ./config/prod.yml ./config/prod.yml

COPY --from=builder /bin/auth_server ./bin/auth_server

EXPOSE 7070

CMD ["./bin/auth_server"]

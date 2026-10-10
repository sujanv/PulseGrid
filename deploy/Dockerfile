FROM golang:1.22-alpine AS builder

WORKDIR /app

RUN apk add --no-cache git ca-certificates tzdata

COPY go.mod go.sum* ./
RUN go mod download || true

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /pulsegrid ./cmd/server/main.go
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /pulsegrid-cli ./cmd/cli/main.go

FROM alpine:3.19

RUN apk --no-cache add ca-certificates tzdata && adduser -D -u 1000 pulsegrid

WORKDIR /app
COPY --from=builder /pulsegrid /app/pulsegrid
COPY --from=builder /pulsegrid-cli /app/pulsegrid-cli

USER pulsegrid

EXPOSE 8080 9090

ENTRYPOINT ["/app/pulsegrid"]

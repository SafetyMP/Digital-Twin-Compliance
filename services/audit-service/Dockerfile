FROM golang:1.25-alpine AS builder
WORKDIR /app
RUN apk add --no-cache git
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /audit-service ./cmd/server

FROM alpine:3.24
RUN apk add --no-cache ca-certificates wget
WORKDIR /app
COPY --from=builder /audit-service /app/audit-service
COPY migrations /app/migrations
EXPOSE 8090
ENTRYPOINT ["/app/audit-service"]

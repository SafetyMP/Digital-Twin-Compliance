FROM golang:1.26-alpine AS builder
WORKDIR /app
RUN apk add --no-cache git
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /alert-service ./cmd/server

FROM alpine:3.24
RUN apk add --no-cache ca-certificates wget
WORKDIR /app
COPY --from=builder /alert-service /app/alert-service
COPY migrations /app/migrations
EXPOSE 8085
ENTRYPOINT ["/app/alert-service"]

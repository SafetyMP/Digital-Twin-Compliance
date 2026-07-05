FROM golang:1.26-alpine AS builder
WORKDIR /app
RUN apk add --no-cache git
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /graph-service ./cmd/server

FROM alpine:3.24
RUN apk add --no-cache ca-certificates wget
WORKDIR /app
COPY --from=builder /graph-service /app/graph-service
EXPOSE 8093
ENTRYPOINT ["/app/graph-service"]

FROM golang:1.25-alpine AS builder
WORKDIR /app
RUN apk add --no-cache git
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /cedar-service ./cmd/server

FROM alpine:3.24
RUN apk add --no-cache ca-certificates wget
WORKDIR /app
COPY --from=builder /cedar-service /app/cedar-service
EXPOSE 8091
ENTRYPOINT ["/app/cedar-service"]

# Build Stage
FROM golang:alpine AS builder

WORKDIR /app

ENV GOTOOLCHAIN=auto

# Install certificates and git
RUN apk add --no-cache ca-certificates git

# Cache dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy source code and build
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /app/api ./cmd/api

# Run Stage
FROM alpine:3.20

WORKDIR /app

RUN apk add --no-cache ca-certificates tzdata

COPY --from=builder /app/api /app/api

EXPOSE 8080

CMD ["/app/api"]

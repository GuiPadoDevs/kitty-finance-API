# Build Stage
FROM golang:alpine AS builder

WORKDIR /app

ENV CGO_ENABLED=0
ENV GOOS=linux

# Install certificates and git
RUN apk add --no-cache ca-certificates git

# Copy source code
COPY . .

# Download dependencies & compile binary
RUN go mod tidy
RUN go build -ldflags="-w -s" -o /app/api ./cmd/api

# Run Stage
FROM alpine:3.20

WORKDIR /app

RUN apk add --no-cache ca-certificates tzdata

COPY --from=builder /app/api /app/api

EXPOSE 8080

CMD ["/app/api"]

# Build stage
FROM golang:1.24-alpine AS builder

WORKDIR /app

# Install dependencies first for better caching
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o titan-ledger ./cmd/api

# Runtime stage
FROM alpine:3.19

WORKDIR /app

# Install certificates for HTTPS calls
RUN apk --no-cache add ca-certificates

# Copy binary from builder
COPY --from=builder /app/titan-ledger .

# Copy static files for API playground
COPY --from=builder /app/static ./static

# Copy migrations for database setup
COPY --from=builder /app/migrations ./migrations

# Expose port
EXPOSE 8080

# Run the application
CMD ["./titan-ledger"]

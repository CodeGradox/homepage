# Build stage
FROM golang:1.25.5-alpine AS builder

WORKDIR /app

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o main .

# Runtime stage
FROM alpine:latest

WORKDIR /app

# Copy binary and required files from builder
COPY --from=builder /app/main .
COPY --from=builder /app/views ./views
COPY --from=builder /app/assets ./assets
COPY --from=builder /app/public ./public

# Expose port
EXPOSE 8080

# Run the application
CMD ["./main"]

# Stage 1: Build the Go binary
FROM golang:1.26-alpine AS builder

WORKDIR /app

# Install build tools if needed
RUN apk add --no-cache git ca-certificates

# Copy go module files and download dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy application source code
COPY . .

# Build lightweight static binary
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o server cmd/server/main.go

# Stage 2: Minimal runtime image
FROM alpine:3.19

RUN apk add --no-cache ca-certificates tzdata

WORKDIR /app

# Copy binary and assets from builder
COPY --from=builder /app/server /app/server
COPY --from=builder /app/knowledge /app/knowledge
COPY --from=builder /app/.source_of_truth /app/.source_of_truth
COPY --from=builder /app/.agents /app/.agents

# Expose default Cloud Run port
ENV PORT=8080
EXPOSE 8080

CMD ["/app/server"]

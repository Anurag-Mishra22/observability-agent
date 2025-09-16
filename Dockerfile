# Build stage
FROM golang:1.24-alpine AS builder

WORKDIR /app

# Copy go.mod and go.sum
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the agent
RUN go build -o agent ./cmd/agent

# Run stage
FROM alpine:3.18

WORKDIR /app

# Copy the compiled binary from builder
COPY --from=builder /app/agent .

# Run the agent
CMD ["./agent"]

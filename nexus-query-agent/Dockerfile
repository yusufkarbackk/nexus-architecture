# Nexus Query Agent - Dockerfile
# Connects to SAP HANA and executes queries via WebSocket

FROM golang:1.24-alpine AS builder

WORKDIR /build

# Install dependencies (ca-certificates already included in golang image)
RUN apk add --no-cache git

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the query agent
RUN CGO_ENABLED=0 GOOS=linux go build -o nexus-query-agent ./cmd/query-agent

# ----------------------
# Final Stage
# ----------------------
FROM golang:1.22-alpine

WORKDIR /app

# Copy binary from builder
COPY --from=builder /build/nexus-query-agent .
COPY --from=builder /build/config/config.yml ./config/config.yml

# Set timezone
ENV TZ=Asia/Jakarta

CMD ["./nexus-query-agent"]


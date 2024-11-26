# Stage 1: Build the Subnet binary
FROM golang:1.23 AS builder

# Set the working directory inside the container
WORKDIR /app

# Copy the Go modules manifests
COPY go.mod go.sum ./

# Download the dependencies
RUN go mod download

# Copy the source code
COPY . .

# Build the Subnet binary
RUN GOOS=linux go build -o subnet-node -buildvcs=false ./cmd/subnet

# Stage 2: Create a minimal runtime container
FROM ubuntu:22.04

# Set the working directory inside the container
WORKDIR /root

# Install runtime dependencies
RUN apt-get update && apt-get install -y \
    ca-certificates && \
    rm -rf /var/lib/apt/lists/*

# Copy the binary from the builder stage
COPY --from=builder /app/subnet-node ./subnet

# Copy default configuration (optional)
# COPY config.yaml ./config.yaml

# Expose necessary ports
EXPOSE 4001 8080

# # Set the command to run Subnet
CMD ["./subnet", "--repo", "/data"]

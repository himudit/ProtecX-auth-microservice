# Use Go 1.25.4
FROM golang:1.25.4-alpine

# Set working directory inside container
WORKDIR /app

# Copy go.mod and go.sum first (for caching)
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy rest of the source code
COPY . .

# Build the app
RUN go build -o authservice

# Expose port (Gin uses 8080)
EXPOSE 8080

# Run the binary
CMD ["./authservice"]

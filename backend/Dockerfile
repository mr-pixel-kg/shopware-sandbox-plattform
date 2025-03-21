# Step 1: Build stage
FROM golang:1.23-alpine as builder

WORKDIR /app

# Install build dependencies für SQLite
RUN apk add --no-cache gcc musl-dev

# Copy go modules and source files
COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Step 2: Build the application
# Set CGO_ENABLED=1 for SQLite
ENV CGO_ENABLED=1
RUN go build -o main .

# Step 3: Production stage
FROM alpine:latest

WORKDIR /app/

# Install SQLite runtime dependencies
RUN apk add --no-cache sqlite-libs

# Copy go application build
COPY --from=builder /app/main ./sandbox-server

# Export application port
EXPOSE 8080

# Run application
CMD ["./sandbox-server"]
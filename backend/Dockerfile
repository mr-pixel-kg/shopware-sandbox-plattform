# Step 1: Build stage
FROM golang:1.23-alpine as builder

WORKDIR /app

# Copy go modules and source files
COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Step 2: Build the application
RUN go build -o main .

# Step 3: Production stage
FROM alpine:latest

WORKDIR /root/

# Copy go application build
COPY --from=builder /app/main .

# Export application port
EXPOSE 8080

# Run application
CMD ["./main"]
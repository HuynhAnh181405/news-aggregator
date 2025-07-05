# news-aggregator/services/api/Dockerfile

# Stage 1: Build the Go application
FROM golang:1.24.3-alpine AS builder
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Build the API service
RUN CGO_ENABLED=0 GOOS=linux go build -o /api-app ./main.go

# Stage 2: Create the final lightweight image
FROM alpine:latest
WORKDIR /app

# Copy the compiled binary from the builder stage
COPY --from=builder /api-app .

# Expose the port the API listens on
EXPOSE 8080

# Command to run the application
CMD ["./api-app"]
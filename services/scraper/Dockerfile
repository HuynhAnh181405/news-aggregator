# news-aggregator/services/scraper/Dockerfile

# Stage 1: Build the Go application
FROM golang:1.24.3-alpine AS builder
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Build the scraper service
RUN CGO_ENABLED=0 GOOS=linux go build -o /scraper-app ./main.go

# Stage 2: Create the final lightweight image
FROM alpine:latest
WORKDIR /app

# Copy the compiled binary from the builder stage
COPY --from=builder /scraper-app .

CMD ["./scraper-app"]
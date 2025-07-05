package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"news-aggregator/services/api/handlers"
	"news-aggregator/services/api/kafka"
	"news-aggregator/services/api/middleware"
	"news-aggregator/services/api/storage"
)

func main() {
	// --- Configuration from Environment Variables ---
	apiPort := os.Getenv("API_PORT")
	if apiPort == "" {
		apiPort = "8080"
	}
	kafkaBroker := os.Getenv("KAFKA_BROKER")
	if kafkaBroker == "" {
		kafkaBroker = "kafka:9092" // Default for Docker Compose
	}
	kafkaTopic := os.Getenv("KAFKA_TOPIC")
	if kafkaTopic == "" {
		kafkaTopic = "news-updates"
	}
	kafkaGroupID := os.Getenv("KAFKA_GROUP_ID")
	if kafkaGroupID == "" {
		kafkaGroupID = "news-aggregator-group"
	}
	storageCapacity := 100 // Max articles to store in memory

	// --- Initialize Storage ---
	s := storage.NewInMemoryStorage(storageCapacity)

	// --- Initialize Kafka Consumer ---
	consumerCtx, cancelConsumer := context.WithCancel(context.Background())
	consumer := kafka.NewConsumer(kafkaBroker, kafkaTopic, kafkaGroupID, s)
	go consumer.StartConsuming(consumerCtx) // Run consumer in a goroutine

	// --- Setup API Handlers ---
	articleHandler := handlers.NewArticleHandler(s)

	// --- Setup HTTP Server ---
	mux := http.NewServeMux()

	// Apply Rate Limiting middleware
	// Allow 5 requests per second, with a burst of 10 requests
	rateLimitAppliedHandler := middleware.RateLimitMiddleware(5, 10)(
		http.HandlerFunc(articleHandler.GetLatestArticles),
	)

	mux.Handle("/articles/latest", rateLimitAppliedHandler)
	mux.HandleFunc("/health", handlers.HealthCheck) // Health check endpoint

	server := &http.Server{
		Addr:    fmt.Sprintf(":%s", apiPort),
		Handler: mux,
		// Good practice for production servers:
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	// --- Graceful Shutdown ---
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	go func() {
		log.Printf("API Server listening on :%s", apiPort)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Could not listen on %s: %v\n", apiPort, err)
		}
	}()

	<-stop // Block until a signal is received

	log.Println("Shutting down API server...")
	cancelConsumer() // Tell consumer to stop
	if err := consumer.Close(); err != nil {
		log.Printf("Error closing Kafka consumer: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server Shutdown Failed:%+v", err)
	}
	log.Println("API Server gracefully stopped.")
}

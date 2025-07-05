package main

import (
	"log"
	"os"
	"time"

	"news-aggregator/services/scraper/kafka"
	"news-aggregator/services/scraper/scraper" // <--- DÒNG NÀY PHẢI ĐÚNG VỚI CẤU TRÚC THƯ MỤC CỦA BẠN
)

func main() {
	kafkaBroker := os.Getenv("KAFKA_BROKER")
	if kafkaBroker == "" {
		kafkaBroker = "kafka:9092"
	}
	kafkaTopic := os.Getenv("KAFKA_TOPIC")
	if kafkaTopic == "" {
		kafkaTopic = "news-updates"
	}

	producer := kafka.NewProducer(kafkaBroker, kafkaTopic)
	defer func() {
		if err := producer.Close(); err != nil {
			log.Fatalf("Failed to close Kafka producer: %v", err)
		}
	}()

	s := scraper.NewScraper(producer) // <--- Gọi NewScraper từ package "scraper" (từ thư mục con scraper)

	s.StartScraping(5 * time.Minute)
}

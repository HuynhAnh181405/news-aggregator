package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"news-aggregator/services/scraper/models"

	"github.com/segmentio/kafka-go"
)

// Producer quản lý việc gửi (produce) các message (bài viết) đến Kafka.
type Producer struct {
	writer *kafka.Writer // Kafka writer để gửi message đến một topic nhất định
}

// NewProducer khởi tạo một Kafka Producer mới.
// Tham số: broker là địa chỉ Kafka, topic là nơi gửi dữ liệu.
func NewProducer(broker string, topic string) *Producer {
	writer := &kafka.Writer{
		Addr:         kafka.TCP(broker),
		Topic:        topic,
		Balancer:     &kafka.LeastBytes{}, // Cân bằng partition theo kích thước nhỏ nhất
		RequiredAcks: kafka.RequireOne,    // Đảm bảo ít nhất 1 replica xác nhận đã nhận message
		Async:        false,               // Gửi đồng bộ để đảm bảo reliability (đồng bộ = chậm hơn)
	}
	log.Printf("Kafka Producer initialized for broker %s, topic %s", broker, topic)
	return &Producer{writer: writer}
}

// ProduceArticle gửi một bài viết (Article) đến Kafka.
// Bài viết sẽ được marshal sang JSON trước khi gửi.
func (p *Producer) ProduceArticle(article models.Article) error {
	// Convert article sang JSON
	articleJSON, err := json.Marshal(article)
	if err != nil {
		return fmt.Errorf("failed to marshal article: %w", err)
	}

	// Tạo Kafka message: sử dụng ID làm key để đảm bảo message được gửi tới cùng partition
	msg := kafka.Message{
		Key:   []byte(article.ID), // Key được dùng để xác định partition
		Value: articleJSON,        // Nội dung message
		Time:  time.Now(),         // Thời gian gửi message
	}

	// Gửi message tới Kafka
	err = p.writer.WriteMessages(context.Background(), msg)
	if err != nil {
		return fmt.Errorf("failed to write message to Kafka: %w", err)
	}
	log.Printf("Produced article '%s' (ID: %s) to Kafka", article.Title, article.ID)
	return nil
}

// Close đóng Kafka writer để giải phóng tài nguyên.
func (p *Producer) Close() error {
	log.Println("Closing Kafka Producer...")
	return p.writer.Close()
}

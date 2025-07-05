package kafka

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"news-aggregator/services/api/models"
	"news-aggregator/services/api/storage"

	"github.com/segmentio/kafka-go"
)

// Consumer quản lý việc nhận message từ Kafka và lưu vào bộ nhớ trong (InMemoryStorage).
type Consumer struct {
	reader  *kafka.Reader            // Kafka reader để đọc message từ topic
	storage *storage.InMemoryStorage // Bộ nhớ để lưu các bài báo đã nhận
}

// NewConsumer khởi tạo một Kafka Consumer mới với thông tin broker, topic, groupID và nơi lưu trữ.
func NewConsumer(broker, topic, groupID string, storage *storage.InMemoryStorage) *Consumer {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:  []string{broker},
		Topic:    topic,
		GroupID:  groupID,
		MinBytes: 10e3,            // 10KB: kích thước tối thiểu của message batch
		MaxBytes: 10e6,            // 10MB: kích thước tối đa của message batch
		MaxWait:  1 * time.Second, // Thời gian tối đa chờ trước khi gửi batch về client
	})
	log.Printf("Kafka Consumer initialized for broker %s, topic %s, group %s", broker, topic, groupID)
	return &Consumer{reader: reader, storage: storage}
}

// StartConsuming bắt đầu vòng lặp nhận message từ Kafka.
// Gọi hàm này trong một goroutine và truyền vào context để dừng an toàn.
func (c *Consumer) StartConsuming(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			// Dừng tiêu thụ nếu context bị hủy (ví dụ như Ctrl+C hoặc shutdown)
			log.Println("Kafka Consumer shutting down...")
			return

		default:
			// Lấy message từ Kafka topic
			m, err := c.reader.FetchMessage(ctx)
			if err != nil {
				log.Printf("Error fetching message from Kafka: %v", err)
				time.Sleep(time.Second) // Tạm nghỉ trước khi thử lại
				continue
			}

			// Parse message từ JSON sang Article
			var article models.Article
			err = json.Unmarshal(m.Value, &article)
			if err != nil {
				log.Printf("Error unmarshaling article: %v, message: %s", err, string(m.Value))

				// Vẫn commit offset để tránh lặp lại message hỏng
				if err := c.reader.CommitMessages(ctx, m); err != nil {
					log.Printf("Error committing offset after unmarshal error: %v", err)
				}
				continue
			}

			// Lưu bài viết vào bộ nhớ
			c.storage.AddArticle(article)
			log.Printf("Consumed article '%s' (ID: %s) from Kafka. Total articles: %d",
				article.Title, article.ID, c.storage.GetArticleCount())

			// Commit offset sau khi xử lý xong message
			if err := c.reader.CommitMessages(ctx, m); err != nil {
				log.Printf("Error committing offset: %v", err)
			}
		}
	}
}

// Close đóng Kafka reader khi không sử dụng nữa.
func (c *Consumer) Close() error {
	log.Println("Closing Kafka Consumer...")
	return c.reader.Close()
}

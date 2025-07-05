package handlers

import (
	"encoding/json"
	"log"
	"net/http"

	"news-aggregator/services/api/storage"
)

// ArticleHandler là struct chịu trách nhiệm xử lý các request liên quan đến bài viết.
type ArticleHandler struct {
	Storage *storage.InMemoryStorage // Lưu trữ bài viết trong bộ nhớ
}

// NewArticleHandler tạo một handler mới với bộ nhớ đã được khởi tạo.
func NewArticleHandler(s *storage.InMemoryStorage) *ArticleHandler {
	return &ArticleHandler{Storage: s}
}

// GetLatestArticles xử lý yêu cầu HTTP GET để lấy danh sách các bài viết mới nhất.
func (h *ArticleHandler) GetLatestArticles(w http.ResponseWriter, r *http.Request) {
	// Chỉ cho phép phương thức GET
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Lấy danh sách bài viết mới nhất từ bộ nhớ
	articles := h.Storage.GetLatestArticles()

	// Thiết lập header để phản hồi là JSON
	w.Header().Set("Content-Type", "application/json")

	// Mã hóa và gửi danh sách bài viết về client
	if err := json.NewEncoder(w).Encode(articles); err != nil {
		log.Printf("Error encoding response: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
	log.Printf("Served %d latest articles", len(articles))
}

// HealthCheck là endpoint dùng để kiểm tra tình trạng hoạt động của API.
// Trả về status 200 OK nếu hệ thống hoạt động bình thường.
func HealthCheck(w http.ResponseWriter, r *http.Request) {
	// Chỉ cho phép phương thức GET
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Phản hồi "OK" nếu hệ thống đang hoạt động
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("API Service is healthy!"))
}

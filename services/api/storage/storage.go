package storage

import (
	"log"
	"sync"

	"news-aggregator/services/api/models"
)

// InMemoryStorage stores news articles in memory.
type InMemoryStorage struct {
	mu       sync.RWMutex
	articles map[string]models.Article // Key: Article ID
	latest   []models.Article          // Store latest articles for quick retrieval
	capacity int                       // Max number of latest articles to store
}

// NewInMemoryStorage creates a new in-memory storage.
func NewInMemoryStorage(capacity int) *InMemoryStorage {
	return &InMemoryStorage{
		articles: make(map[string]models.Article),
		latest:   make([]models.Article, 0, capacity),
		capacity: capacity,
	}
}

// AddArticle adds a new article to the storage.
func (s *InMemoryStorage) AddArticle(article models.Article) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Only add if it doesn't exist (basic deduplication by ID)
	if _, exists := s.articles[article.ID]; exists {
		return
	}

	s.articles[article.ID] = article

	// Add to latest articles, maintaining capacity
	if len(s.latest) < s.capacity {
		s.latest = append(s.latest, article)
	} else {
		// Remove the oldest and add the new one
		s.latest = append(s.latest[1:], article)
	}

	log.Printf("Added article: '%s'", article.Title)
}

// GetLatestArticles returns a slice of the most recent articles.
func (s *InMemoryStorage) GetLatestArticles() []models.Article {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Return a copy to prevent external modification
	result := make([]models.Article, len(s.latest))
	copy(result, s.latest)
	return result
}

// GetArticleCount returns the total number of unique articles stored.
func (s *InMemoryStorage) GetArticleCount() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.articles)
}

// news-aggregator/services/scraper/models/article.go
// news-aggregator/services/api/models/article.go
package models

import "time"

// Article represents a news article.
type Article struct {
	ID        string    `json:"id"`
	Title     string    `json:"title"`
	URL       string    `json:"url"`
	Source    string    `json:"source"`
	Published time.Time `json:"published"`
	Content   string    `json:"content,omitempty"` // Optional content
}

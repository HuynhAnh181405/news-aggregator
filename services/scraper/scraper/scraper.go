package scraper

import (
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"news-aggregator/services/scraper/kafka"
	"news-aggregator/services/scraper/models"

	"github.com/PuerkitoBio/goquery"
)

// Scraper chịu trách nhiệm lấy bài viết từ web và gửi lên Kafka.
type Scraper struct {
	producer *kafka.Producer // Dùng để publish các bài viết đã scrape vào Kafka
}

// NewScraper tạo mới một đối tượng Scraper với Kafka producer truyền vào.
func NewScraper(producer *kafka.Producer) *Scraper {
	return &Scraper{producer: producer}
}

// StartScraping bắt đầu quá trình lấy bài viết theo chu kỳ interval định sẵn.
func (s *Scraper) StartScraping(interval time.Duration) {
	log.Printf("Scraper started, fetching news every %v...", interval)
	ticker := time.NewTicker(interval) // Tạo ticker định kỳ
	defer ticker.Stop()

	for range ticker.C {
		log.Println("Scraping for new articles from VnExpress...")
		articles, err := s.scrapeVnExpress() // Gọi hàm scrape từng bài viết
		if err != nil {
			log.Printf("Error scraping VnExpress: %v", err)
			continue
		}

		// Gửi từng bài viết vào Kafka
		for _, article := range articles {
			err := s.producer.ProduceArticle(article)
			if err != nil {
				log.Printf("Error producing article '%s': %v", article.Title, err)
			}
		}
		log.Printf("Finished scraping, %d articles processed from VnExpress.", len(articles))
	}
}

// scrapeVnExpress lấy danh sách bài viết từ chuyên mục Thời Sự của VnExpress.
func (s *Scraper) scrapeVnExpress() ([]models.Article, error) {
	var articles []models.Article
	targetURL := "https://vnexpress.net/thoi-su"

	// Gửi HTTP request lấy nội dung HTML
	res, err := http.Get(targetURL)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch URL %s: %w", targetURL, err)
	}
	defer res.Body.Close()

	// Kiểm tra mã trạng thái HTTP
	if res.StatusCode != 200 {
		return nil, fmt.Errorf("status code error: %d %s", res.StatusCode, res.Status)
	}

	// Phân tích HTML bằng goquery
	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to parse HTML: %w", err)
	}

	// Tìm các thẻ chứa bài viết
	doc.Find("article.item-news").Each(func(i int, sel *goquery.Selection) {
		linkSelection := sel.Find("h3.title-news a, h2.title-news a")
		href, exists := linkSelection.Attr("href")       // Lấy đường dẫn bài viết
		title := strings.TrimSpace(linkSelection.Text()) // Lấy tiêu đề bài viết

		if exists && title != "" {
			fullURL := href
			if !strings.HasPrefix(fullURL, "http") {
				fullURL = "https://vnexpress.net" + fullURL
			}

			// Thử lấy thời gian đăng bài (có thể không chính xác nếu VnExpress thay đổi HTML)
			publishedStr := sel.Find("span.time, p.description").First().Text()
			published := parseVnExpressTime(publishedStr)

			article := models.Article{
				ID:        generateArticleID(fullURL),
				Title:     title,
				URL:       fullURL,
				Source:    "VnExpress",
				Published: published,
			}

			// Gọi đến hàm lấy nội dung chi tiết của bài viết
			fullContent, err := s.ScrapeArticleContent(fullURL)
			if err != nil {
				log.Printf("Warning: Could not scrape full content for '%s': %v", title, err)
			} else {
				article.Content = fullContent
				log.Printf("Scraped full content for '%s'", title)
			}

			// Thêm bài viết vào danh sách
			articles = append(articles, article)
		}
	})

	return articles, nil
}

// generateArticleID tạo một ID duy nhất dựa trên URL của bài viết.
// Dùng để tránh gửi trùng bài.
func generateArticleID(url string) string {
	return fmt.Sprintf("vnexpress-%x", strings.ReplaceAll(url, "/", "-"))
}

// parseVnExpressTime chuyển đổi chuỗi thời gian thành kiểu time.Time.
// Hiện tại chỉ là placeholder, bạn có thể viết parser chính xác hơn.
func parseVnExpressTime(timeStr string) time.Time {
	log.Printf("Attempting to parse time: '%s'", timeStr)
	return time.Now() // TODO: nên parse đúng theo format nếu cần
}

// ScrapeArticleContent lấy nội dung chi tiết của một bài viết dựa vào URL.
func (s *Scraper) ScrapeArticleContent(articleURL string) (string, error) {
	res, err := http.Get(articleURL)
	if err != nil {
		return "", fmt.Errorf("failed to fetch article URL %s: %w", articleURL, err)
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		return "", fmt.Errorf("status code error for content: %d %s", res.StatusCode, res.Status)
	}

	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		return "", fmt.Errorf("failed to parse article HTML: %w", err)
	}

	// Tìm nội dung chính của bài viết
	content := ""
	doc.Find("article.fck_detail p").Each(func(i int, sel *goquery.Selection) {
		paragraphText := strings.TrimSpace(sel.Text())
		if paragraphText != "" {
			content += paragraphText + "\n\n" // Thêm đoạn văn
		}
	})

	return strings.TrimSpace(content), nil
}

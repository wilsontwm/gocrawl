package news

import (
	"encoding/json"
	"fmt"
	"github.com/gocolly/colly"
	"github.com/gocolly/colly/queue"
	"gocrawl/models"
	"log"
	"strings"
	"time"
)

var (
	newStraitsTimesArticleUrls map[string]bool
)

func init() {
	// Initialize the article URLs
	existingLinks := models.GetArticlesBySource(models.NewStraitsTimes)
	newStraitsTimesArticleUrls = map[string]bool{}

	for _, link := range existingLinks {
		newStraitsTimesArticleUrls[link] = true
	}
}

func CrawlNewStraitsTimes() {
	log.Println("Starting to scrape New Straits Times news")

	type Content struct {
		Title     string
		Url       string
		Timestamp int    `json:"created"`
		Thumbnail string `json:"field_image_listing_featured_v2"`
		Body      string
	}

	// Instantiate the collector
	c := colly.NewCollector(
		colly.AllowedDomains("www.nst.com.my"),
	)

	q, _ := queue.New(
		1, // Number of consumer threads
		&queue.InMemoryQueueStorage{MaxSize: 10000}, // Use default queue storage
	)

	c.OnHTML("article-teaser", func(e *colly.HTMLElement) {
		rawContent := e.Attr(":article")
		var content Content
		json.Unmarshal([]byte(rawContent), &content)

		if strings.Index(content.Url, "https://www.nst.com.my") == -1 {
			return
		}

		if _, found := newStraitsTimesArticleUrls[content.Url]; !found {

			newStraitsTimesArticleUrls[content.Url] = true

			title := content.Title
			publishedAt := time.Unix(int64(content.Timestamp), 0)
			body := content.Body
			thumbnail := content.Thumbnail
			url := content.Url

			// Post processing of the content
			body = strings.ReplaceAll(body, "<p />", "\n")
			body = strings.ReplaceAll(body, "</p>", "\n")
			body = strings.ReplaceAll(body, "<p>", "")

			article := &models.Article{
				Source:      models.NewStraitsTimes,
				Title:       title,
				Content:     body,
				URL:         url,
				Thumbnail:   thumbnail,
				PublishedAt: publishedAt,
			}

			models.CreateArticle(article)

		}
	})

	for pageIndex := 0; pageIndex < 3; pageIndex++ {
		// Add URLs to the queue
		q.AddURL("https://www.nst.com.my/business?page=" + fmt.Sprintf("%d", pageIndex))
	}

	// Consume URLs
	q.Run(c)
	log.Println("Ending to scrape New Straits Times news")
}

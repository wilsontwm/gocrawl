package news

import (
	"fmt"
	"github.com/gocolly/colly"
	"github.com/gocolly/colly/queue"
	"gocrawl/models"
	"log"
	"strings"
	"time"
)

var (
	theEdgeArticleUrls map[string]bool
)

func init() {
	// Initialize the article URLs
	existingLinks := models.GetArticlesBySource(models.TheEdge)
	theEdgeArticleUrls = map[string]bool{}

	for _, link := range existingLinks {
		theEdgeArticleUrls[link] = true
	}
}

func CrawlTheEdge() {
	log.Println("Starting to scrape The Edge news")
	const (
		datetimeFormat = "January 02, 2006 15:04 pm +08"
	)

	// Instantiate the collector
	c := colly.NewCollector(
		colly.AllowedDomains("www.theedgemarkets.com"),
	)

	q, _ := queue.New(
		1, // Number of consumer threads
		&queue.InMemoryQueueStorage{MaxSize: 10000}, // Use default queue storage
	)

	detailCollector := c.Clone()

	c.OnHTML(".grid-inner", func(e *colly.HTMLElement) {
		link := e.ChildAttr(".field-content a[href]", "href")
		if strings.Index(link, "/article/") == -1 {
			return
		}

		if !strings.HasPrefix(link, "https://www.theedgemarkets.com") {
			link = "https://www.theedgemarkets.com" + link
		}

		// start scaping the page under the link found if not scraped before
		if _, found := theEdgeArticleUrls[link]; !found {
			detailCollector.Visit(link)
			theEdgeArticleUrls[link] = true
		}
	})

	// Before making request
	// c.OnRequest(func(r *colly.Request) {
	// 	log.Println("Visiting", r.URL.String())
	// })

	// detailCollector.OnRequest(func(r *colly.Request) {
	// 	log.Println("Sub Visiting", r.URL.String())
	// })

	// Extract details of the course
	detailCollector.OnHTML("article", func(e *colly.HTMLElement) {
		title := e.ChildText(".post-title h1")
		datetime := e.ChildText(".post-created")
		thumbnail := e.ChildAttr(".article-getimage center img", "src")
		publishedAt := time.Now()

		var paragraphs []string
		e.ForEach(".field-item p", func(_ int, el *colly.HTMLElement) {
			paragraphs = append(paragraphs, el.Text)
		})
		content := strings.Join(paragraphs, "\n\n")

		loc, err := time.LoadLocation("Asia/Kuala_Lumpur")
		if err == nil {
			if t, err := time.ParseInLocation(datetimeFormat, datetime, loc); err == nil {
				publishedAt = t
			}
		}

		article := &models.Article{
			Source:      models.TheEdge,
			Title:       title,
			Content:     content,
			URL:         e.Request.URL.String(),
			Thumbnail:   thumbnail,
			PublishedAt: publishedAt,
		}

		models.CreateArticle(article)
	})

	for pageIndex := 0; pageIndex < 3; pageIndex++ {
		// Add URLs to the queue
		q.AddURL("https://www.theedgemarkets.com/categories/malaysia?page=" + fmt.Sprintf("%d", pageIndex))
	}

	// Consume URLs
	q.Run(c)
	log.Println("Ending to scrape The Edge news")
}

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
	malayMailArticleUrls map[string]bool
)

func init() {
	// Initialize the article URLs
	existingLinks := models.GetArticlesBySource(models.MalayMail)
	malayMailArticleUrls = map[string]bool{}

	for _, link := range existingLinks {
		malayMailArticleUrls[link] = true
	}
}

func CrawlMalayMail() {
	log.Println("Starting to scrape Malay Mail news")
	const (
		datetimeFormat = "Monday, 02 Jan 2006 03:04 PM MYT"
	)

	// Instantiate the collector
	c := colly.NewCollector(
		colly.AllowedDomains("www.malaymail.com"),
	)

	q, _ := queue.New(
		1, // Number of consumer threads
		&queue.InMemoryQueueStorage{MaxSize: 10000}, // Use default queue storage
	)

	detailCollector := c.Clone()

	c.OnHTML("#news-list-items", func(e *colly.HTMLElement) {
		e.ForEach("a[href]", func(_ int, el *colly.HTMLElement) {
			link := el.Attr("href")
			if strings.Index(link, "/news/money/") == -1 {
				return
			}

			// start scaping the page under the link found if not scraped before
			if _, found := malayMailArticleUrls[link]; !found {
				detailCollector.Visit(link)
				malayMailArticleUrls[link] = true
			}
		})

	})

	// Before making request
	// c.OnRequest(func(r *colly.Request) {
	// 	log.Println("Visiting", r.URL.String())
	// })

	// detailCollector.OnRequest(func(r *colly.Request) {
	// 	log.Println("Sub Visiting", r.URL.String())
	// })

	// Extract details of the course
	detailCollector.OnHTML(".article", func(e *colly.HTMLElement) {
		title := e.ChildText("h1")
		datetime := e.ChildText(".byline .meta")
		thumbnail := e.ChildAttr("article figure img[src]", "src")
		publishedAt := time.Now()

		var paragraphs []string
		e.ForEach("article p", func(_ int, el *colly.HTMLElement) {
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
			Source:      models.MalayMail,
			Title:       title,
			Content:     content,
			URL:         e.Request.URL.String(),
			Thumbnail:   thumbnail,
			PublishedAt: publishedAt,
		}

		models.CreateArticle(article)
	})

	for pageIndex := 1; pageIndex <= 3; pageIndex++ {
		// Add URLs to the queue
		q.AddURL("https://www.malaymail.com/news/money?page=" + fmt.Sprintf("%d", pageIndex))
	}

	// Consume URLs
	q.Run(c)
	log.Println("Ending to scrape Malay Mail news")
}

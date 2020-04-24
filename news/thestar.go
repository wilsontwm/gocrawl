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
	theStarArticleUrls map[string]bool
)

func init() {
	// Initialize the article URLs
	existingLinks := models.GetArticlesBySource(models.TheStar)
	theStarArticleUrls = map[string]bool{}

	for _, link := range existingLinks {
		theStarArticleUrls[link] = true
	}
}

func CrawlTheStar() {
	log.Println("Starting to scrape The Star news")
	const (
		datetimeFormat = "Monday, 02 Jan 2006, 3:04 PM MST"
		dateFormat     = "Monday, 02 Jan 2006"
	)

	// Instantiate the collector
	c := colly.NewCollector(
		colly.AllowedDomains("www.thestar.com.my"),
	)

	q, _ := queue.New(
		1, // Number of consumer threads
		&queue.InMemoryQueueStorage{MaxSize: 10000}, // Use default queue storage
	)

	detailCollector := c.Clone()

	c.OnHTML("a[data-content-category]", func(e *colly.HTMLElement) {
		link := e.Attr("href")
		if strings.Index(link, "/business/business-news/") == -1 {
			return
		}
		// start scaping the page under the link found if not scraped before
		if _, found := theStarArticleUrls[link]; !found {
			detailCollector.Visit(link)
			theStarArticleUrls[link] = true
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
	detailCollector.OnHTML("html", func(e *colly.HTMLElement) {
		title := e.ChildAttr("meta[name=content_title]", "content")
		date := e.ChildText("p.date")
		timestamp := e.ChildText("time.timestamp")
		content := strings.ReplaceAll(e.ChildText("div#story-body"), "  ", "\n")
		thumbnail := e.ChildAttr("meta[name=thumbnailUrl]", "content")
		publishedAt := time.Now()

		// If no timestamp is given, store the current time
		if len(timestamp) == 0 {
			timestamp = time.Now().Format("3:04 PM MST")
		}

		loc, err := time.LoadLocation("Asia/Kuala_Lumpur")
		if err == nil {
			datetime := date + ", " + timestamp
			if t, err := time.ParseInLocation(datetimeFormat, datetime, loc); err == nil {
				publishedAt = t
			}
		}

		article := &models.Article{
			Source:      models.TheStar,
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
		q.AddURL("https://www.thestar.com.my/news/latest?tag=Business&pgno=" + fmt.Sprintf("%d", pageIndex))
	}

	// Consume URLs
	q.Run(c)
	log.Println("Ending to scrape The Star news")
}

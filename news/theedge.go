package news

import (
	"encoding/json"
	"fmt"
	"github.com/gocolly/colly"
	"github.com/gocolly/colly/queue"
	"log"
	"os"
	"strings"
	"time"
)

var (
	theEdgeArticleUrls map[string]bool
	theEdgeArticles    []Article
)

func init() {
	// Initialize the article URLs
	theEdgeArticleUrls = map[string]bool{}
}

func CrawlTheEdge() {
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
		// start scaping the page under the link found if not scraped before
		if _, found := theEdgeArticleUrls[link]; !found {
			if !strings.HasPrefix(link, "https://www.theedgemarkets.com") {
				link = "https://www.theedgemarkets.com" + link
			}
			detailCollector.Visit(link)
			theEdgeArticleUrls[link] = true
		}
	})

	// Before making request
	c.OnRequest(func(r *colly.Request) {
		log.Println("Visiting", r.URL.String())
	})

	detailCollector.OnRequest(func(r *colly.Request) {
		log.Println("Sub Visiting", r.URL.String())
	})

	// Extract details of the course
	detailCollector.OnHTML("article", func(e *colly.HTMLElement) {
		title := e.ChildText(".post-title h1")
		datetime := e.ChildText(".post-created")
		content := e.ChildText(".field-item p")
		thumbnail := e.ChildAttr(".article-getimage center img", "src")
		publishedAt := time.Now()
		loc, err := time.LoadLocation("Asia/Kuala_Lumpur")
		if err == nil {
			if t, err := time.ParseInLocation(datetimeFormat, datetime, loc); err == nil {
				publishedAt = t
			}
		}

		article := Article{
			Title:       title,
			Content:     content,
			URL:         e.Request.URL.String(),
			Thumbnail:   thumbnail,
			PublishedAt: publishedAt,
		}
		fmt.Printf("%+v\n", article)
		theEdgeArticles = append(theEdgeArticles, article)
	})

	for pageIndex := 0; pageIndex < 3; pageIndex++ {
		// Add URLs to the queue
		q.AddURL("https://www.theedgemarkets.com/categories/malaysia?page=" + fmt.Sprintf("%d", pageIndex))
	}

	// Consume URLs
	q.Run(c)

}

func OutputTheEdge() {

	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", " ")
	enc.Encode(theEdgeArticles)
}

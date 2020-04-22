package news

import (
	"encoding/json"
	"fmt"
	"github.com/gocolly/colly"
	"github.com/gocolly/colly/queue"
	"log"
	"os"
	"regexp"
	"strings"
	"time"
)

var (
	sinChewArticleUrls map[string]bool
	sinChewArticles    []Article
)

func init() {
	// Initialize the article URLs
	sinChewArticleUrls = map[string]bool{}
}

func CrawlSinChew() {
	const (
		datetimeFormat = "2006-01-02 15:04:05"
	)

	// Instantiate the collector
	c := colly.NewCollector(
		colly.AllowedDomains("www.sinchew.com.my"),
	)

	q, _ := queue.New(
		1, // Number of consumer threads
		&queue.InMemoryQueueStorage{MaxSize: 10000}, // Use default queue storage
	)

	detailCollector := c.Clone()

	c.OnHTML("#articlenum", func(e *colly.HTMLElement) {
		e.ForEach("a[href]", func(_ int, el *colly.HTMLElement) {
			link := strings.Trim(el.Attr("href"), " ")
			if strings.Index(link, "/content/") == -1 {
				return
			}

			// start scaping the page under the link found if not scraped before
			if _, found := sinChewArticleUrls[link]; !found {
				detailCollector.Visit(link)
				sinChewArticleUrls[link] = true
			}
		})
	})

	// Before making request
	c.OnRequest(func(r *colly.Request) {
		log.Println("Visiting", r.URL.String())
	})

	detailCollector.OnRequest(func(r *colly.Request) {
		log.Println("Sub Visiting", r.URL.String())
	})

	// Extract details of the course
	detailCollector.OnHTML("#articlenum", func(e *colly.HTMLElement) {
		title, _ := e.DOM.ParentsUntil("body").Find("#forsharebutton").Attr("data-a2a-title")
		content := e.ChildText("p")
		thumbnail := e.ChildAttr("p img", "src")
		publishedAt := time.Now()

		loc, err := time.LoadLocation("Asia/Kuala_Lumpur")
		if err == nil {
			if t, err := time.ParseInLocation(datetimeFormat, getDateString(e.Text), loc); err == nil {
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

		sinChewArticles = append(sinChewArticles, article)
	})

	for pageIndex := 1; pageIndex <= 3; pageIndex++ {
		// Add URLs to the queue
		page := ""
		if pageIndex > 1 {
			page = "_" + fmt.Sprintf("%d", pageIndex)
		}

		url := "https://www.sinchew.com.my/column/node_33" + page + ".html"
		q.AddURL(url)
	}

	// Consume URLs
	q.Run(c)

}

func getDateString(str string) string {
	re := regexp.MustCompile(`\d{4}[-]\d{2}[-]\d{2}[\s]\d{2}[:]\d{2}[:]\d{2}`)
	result := ""
	submatchall := re.FindAllString(str, -1)
	for _, element := range submatchall {
		result += element
	}

	return result
}

func OutputSinChew() {

	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", " ")
	enc.Encode(theEdgeArticles)
}

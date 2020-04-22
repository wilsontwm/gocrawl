package news

import (
	"encoding/json"
	"fmt"
	"github.com/gocolly/colly"
	"github.com/gocolly/colly/queue"
	"log"
	"os"
	"strconv"
	"strings"
	"time"
)

var (
	nanYangArticleUrls map[string]bool
	nanYangArticles    []Article
)

func init() {
	// Initialize the article URLs
	nanYangArticleUrls = map[string]bool{}
}

func CrawlNanYang() {
	// Instantiate the collector
	c := colly.NewCollector(
		colly.AllowedDomains("www.enanyang.my"),
	)

	q, _ := queue.New(
		1, // Number of consumer threads
		&queue.InMemoryQueueStorage{MaxSize: 10000}, // Use default queue storage
	)

	detailCollector := c.Clone()

	c.OnHTML(".cat-post-item", func(e *colly.HTMLElement) {
		link := e.ChildAttr("a[href]", "href")
		if strings.Index(link, "/news/") == -1 {
			return
		}
		// start scaping the page under the link found if not scraped before
		if _, found := nanYangArticleUrls[link]; !found {
			detailCollector.Visit(link)
			nanYangArticleUrls[link] = true
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
	detailCollector.OnHTML(".article-content", func(e *colly.HTMLElement) {

		title := e.ChildText(".post-content-title h1")
		datetime := e.ChildText(".entry-date")
		content := e.ChildText(".entry-content p")
		thumbnail := e.ChildAttr("p img", "src")

		article := Article{
			Title:       title,
			Content:     content,
			URL:         e.Request.URL.String(),
			Thumbnail:   thumbnail,
			PublishedAt: getNYPublishedTime(datetime),
		}

		nanYangArticles = append(nanYangArticles, article)
	})

	for pageIndex := 1; pageIndex <= 3; pageIndex++ {
		// Add URLs to the queue
		q.AddURL("https://www.enanyang.my/category/%E8%B4%A2%E7%BB%8F%E6%96%B0%E9%97%BB/page/" + fmt.Sprintf("%d", pageIndex))
	}

	// Consume URLs
	q.Run(c)

}

func getNYPublishedTime(datetime string) time.Time {
	now := time.Now()
	hour, min, sec := now.Clock()

	splitted := strings.Split(datetime, "年")
	year, _ := strconv.Atoi(splitted[0])
	splitted = strings.Split(splitted[1], "月")
	month, _ := strconv.Atoi(splitted[0])
	splitted = strings.Split(splitted[1], "日")
	day, _ := strconv.Atoi(splitted[0])
	location, _ := time.LoadLocation("Local")

	publishedAt := time.Date(year, time.Month(month), day, hour, min, sec, 0, location)

	return publishedAt
}

func OutputNanYang() {

	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", " ")
	enc.Encode(theEdgeArticles)
}

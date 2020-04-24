package news

import (
	"fmt"
	"github.com/gocolly/colly"
	"github.com/gocolly/colly/queue"
	"gocrawl/models"
	"log"
	"strconv"
	"strings"
	"time"
)

var (
	chinaPressArticleUrls map[string]bool
)

func init() {
	// Initialize the article URLs
	existingLinks := models.GetArticlesBySource(models.ChinaPress)
	chinaPressArticleUrls = map[string]bool{}

	for _, link := range existingLinks {
		chinaPressArticleUrls[link] = true
	}
}

func CrawlChinaPress() {
	log.Println("Starting to scrape China Press news")

	// Instantiate the collector
	c := colly.NewCollector(
		colly.AllowedDomains("www.chinapress.com.my"),
	)

	q, _ := queue.New(
		1, // Number of consumer threads
		&queue.InMemoryQueueStorage{MaxSize: 10000}, // Use default queue storage
	)

	detailCollector := c.Clone()

	c.OnHTML(".category_page_post", func(e *colly.HTMLElement) {
		link := e.ChildAttr("a[href]", "href")

		// start scaping the page under the link found if not scraped before
		if _, found := chinaPressArticleUrls[link]; !found {
			detailCollector.Visit(link)
			chinaPressArticleUrls[link] = true
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
	detailCollector.OnHTML(".status-publish", func(e *colly.HTMLElement) {
		title := e.ChildText(".entry-title")
		datetime := e.ChildText(".entry-date")
		content := e.ChildText(".entry-content p")
		thumbnail := e.ChildAttr("p img", "src")

		article := &models.Article{
			Source:      models.ChinaPress,
			Title:       title,
			Content:     content,
			URL:         e.Request.URL.String(),
			Thumbnail:   thumbnail,
			PublishedAt: getCPPublishedTime(datetime),
		}

		models.CreateArticle(article)
	})

	for pageIndex := 1; pageIndex <= 1; pageIndex++ {
		// Add URLs to the queue
		url := "https://www.chinapress.com.my/category/%e8%b4%a2%e7%bb%8f%e6%96%b0%e9%97%bb/page/" + fmt.Sprintf("%d", pageIndex)
		q.AddURL(url)
	}

	// Consume URLs
	q.Run(c)
	log.Println("Ending to scrape China Press news")
}

func getCPPublishedTime(datetime string) time.Time {
	now := time.Now()
	hour, min, sec := now.Clock()

	splitted := strings.Split(datetime, "年")
	year, _ := strconv.Atoi(splitted[0])
	splitted = strings.Split(splitted[1], "月")
	month, _ := strconv.Atoi(splitted[0])
	splitted = strings.Split(splitted[1], "日")
	day, _ := strconv.Atoi(splitted[0])
	location, _ := time.LoadLocation("Asia/Kuala_Lumpur")

	publishedAt := time.Date(year, time.Month(month), day, hour, min, sec, 0, location)

	return publishedAt
}

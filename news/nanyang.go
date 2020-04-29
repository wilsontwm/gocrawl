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
	nanYangArticleUrls map[string]bool
)

func init() {
	// Initialize the article URLs
	existingLinks := models.GetArticlesBySource(models.NanYang)
	nanYangArticleUrls = map[string]bool{}

	for _, link := range existingLinks {
		nanYangArticleUrls[link] = true
	}
}

func CrawlNanYang() {
	log.Println("Starting to scrape NanYang news")
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
	// c.OnRequest(func(r *colly.Request) {
	// 	log.Println("Visiting", r.URL.String())
	// })

	// detailCollector.OnRequest(func(r *colly.Request) {
	// 	log.Println("Sub Visiting", r.URL.String())
	// })

	// Extract details of the course
	detailCollector.OnHTML(".article-content", func(e *colly.HTMLElement) {

		title := e.ChildText(".post-content-title h1")
		datetime := e.ChildText(".entry-date")
		thumbnail := e.ChildAttr("p img", "src")

		var paragraphs []string
		e.ForEach(".entry-content p", func(_ int, el *colly.HTMLElement) {
			paragraphs = append(paragraphs, el.Text)
		})
		content := strings.Join(paragraphs, "\n\n")

		if thumbnail != "" && strings.Index(thumbnail, "www.enanyang.my") == -1 {
			thumbnail = "https://www.enanyang.my" + thumbnail
		}

		article := &models.Article{
			Source:      models.NanYang,
			Title:       title,
			Content:     content,
			URL:         e.Request.URL.String(),
			Thumbnail:   thumbnail,
			PublishedAt: getNYPublishedTime(datetime),
		}

		models.CreateArticle(article)
	})

	for pageIndex := 1; pageIndex <= 3; pageIndex++ {
		// Add URLs to the queue
		q.AddURL("https://www.enanyang.my/category/%E8%B4%A2%E7%BB%8F%E6%96%B0%E9%97%BB/page/" + fmt.Sprintf("%d", pageIndex))
	}

	// Consume URLs
	q.Run(c)
	log.Println("Ending to scrape NanYang news")
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
	location, _ := time.LoadLocation("Asia/Kuala_Lumpur")

	publishedAt := time.Date(year, time.Month(month), day, hour, min, sec, 0, location)

	return publishedAt
}

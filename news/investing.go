package news

import (
	"fmt"
	"github.com/gocolly/colly"
	"github.com/gocolly/colly/queue"
	"gocrawl/models"
	"log"
	"regexp"
	"strings"
	"time"
)

var (
	investingArticleUrls map[string]bool
	investingArticles    []Article
)

func init() {
	// Initialize the article URLs
	existingLinks := models.GetArticlesBySource(models.Investing)
	investingArticleUrls = map[string]bool{}

	for _, link := range existingLinks {
		investingArticleUrls[link] = true
	}
}

func CrawlInvesting() {
	log.Println("Starting to scrape Investing.com news")
	const (
		datetimeFormat = "Jan 02, 2006 03:04PM"
	)

	// Instantiate the collector
	c := colly.NewCollector(
		colly.AllowedDomains("www.investing.com"),
	)

	q, _ := queue.New(
		1, // Number of consumer threads
		&queue.InMemoryQueueStorage{MaxSize: 10000}, // Use default queue storage
	)

	detailCollector := c.Clone()

	c.OnHTML("#leftColumn", func(e *colly.HTMLElement) {
		e.ForEach(".articleItem", func(_ int, el *colly.HTMLElement) {
			link := el.ChildAttr("a.title", "href")
			if strings.Index(link, "/news/") == -1 {
				return
			}

			link = "https://www.investing.com" + link
			// start scaping the page under the link found if not scraped before
			if _, found := investingArticleUrls[link]; !found {
				detailCollector.Visit(link)
				investingArticleUrls[link] = true
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
	detailCollector.OnHTML("#leftColumn", func(e *colly.HTMLElement) {
		title := e.ChildText(".articleHeader")
		date := e.ChildText(".contentSectionDetails")
		thumbnail := e.ChildAttr("#carouselImage", "src")
		publishedAt := time.Now()

		var paragraphs []string
		e.ForEach(".articlePage p", func(_ int, el *colly.HTMLElement) {
			paragraphs = append(paragraphs, el.Text)
		})
		content := strings.Join(paragraphs, "\n\n")

		loc, err := time.LoadLocation("America/New_York")
		if err == nil {
			if t, err := time.ParseInLocation(datetimeFormat, getInvestingDateString(date), loc); err == nil {
				publishedAt = t
			}
		}

		article := &models.Article{
			Source:      models.Investing,
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
		q.AddURL("https://www.investing.com/news/economy/" + fmt.Sprintf("%d", pageIndex))
	}

	// Consume URLs
	q.Run(c)
	log.Println("Ending to scrape Investing.com news")
}

func getInvestingDateString(str string) string {
	re := regexp.MustCompile(`\w{3}[\s]\d{1,2}[,][\s]\d{4}[\s]\d{2}[:]\d{2}[AP][M]`)
	result := ""
	submatchall := re.FindAllString(str, -1)
	for _, element := range submatchall {
		result = element
	}

	return result
}

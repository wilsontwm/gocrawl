package news

import (
	"encoding/json"
	"fmt"
	"github.com/gocolly/colly"
	"github.com/gocolly/colly/queue"
	"gocrawl/models"
	"golang.org/x/net/html"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

var (
	businessInsiderArticleUrls map[string]bool
	businessInsiderArticles    []Article
)

func init() {
	// Initialize the article URLs
	existingLinks := models.GetArticlesBySource(models.BusinessInsider)
	businessInsiderArticleUrls = map[string]bool{}

	for _, link := range existingLinks {
		businessInsiderArticleUrls[link] = true
	}
}

func CrawlBusinessInsider() {
	log.Println("Starting to scrape Business Insider news")
	var links []string

	const (
		datetimeFormat = "2006-01-02T15:04:05+00:00"
	)

	// Instantiate the collector
	c := colly.NewCollector(
		colly.AllowedDomains("www.businessinsider.my"),
	)

	q, _ := queue.New(
		1, // Number of consumer threads
		&queue.InMemoryQueueStorage{MaxSize: 10000}, // Use default queue storage
	)

	c.OnHTML(".status-publish", func(e *colly.HTMLElement) {
		title := e.ChildText(".entry-title")
		datetime := e.ChildAttr(".entry-date", "datetime")
		content := e.ChildText("p")
		thumbnail := e.ChildAttr("img[src]", "src")
		publishedAt := time.Now()

		if t, err := time.Parse(datetimeFormat, datetime); err == nil {
			publishedAt = t
		}

		if thumbnail != "" && !strings.HasPrefix(thumbnail, "http") {
			thumbnail = "https://www.businessinsider.my" + thumbnail
		}

		article := &models.Article{
			Source:      models.BusinessInsider,
			Title:       title,
			Content:     content,
			URL:         e.Request.URL.String(),
			Thumbnail:   thumbnail,
			PublishedAt: publishedAt,
		}

		models.CreateArticle(article)
	})

	for i := 1; i <= 3; i++ {
		newLinks := getLinksOnPage(i)
		links = append(links, newLinks...)
	}

	for _, link := range links {
		if _, found := businessInsiderArticleUrls[link]; !found {
			q.AddURL(link)
			businessInsiderArticleUrls[link] = true
		}
	}

	// Consume URLs
	q.Run(c)

	log.Println("Ending to scrape Business Insider news")
}

func getLinksOnPage(page int) (links []string) {
	type Content struct {
		Data string `json:"td_data"`
	}

	apiUrl := "https://www.businessinsider.my/wp-admin/admin-ajax.php?td_theme_name=Newspaper&v=7.8"
	data := url.Values{}
	data.Set("action", "td_ajax_block")
	data.Set("td_column_number", strconv.Itoa(2))
	data.Set("td_current_page", strconv.Itoa(page))
	data.Set("block_type", "td_block_11")
	data.Set("td_atts", `{"limit":"","sort":"","post_ids":"","tag_slug":"","autors_id":"","installed_post_types":"","category_id":"36","category_ids":"","custom_title":"","custom_url":"","show_child_cat":"","sub_cat_ajax":"","ajax_pagination":"infinite","header_color":"","header_text_color":"","ajax_pagination_infinite_stop":"","td_column_number":2,"td_ajax_preloading":"","td_ajax_filter_type":"","td_ajax_filter_ids":"","td_filter_default_txt":"All","color_preset":"","border_top":"","class":"td_uid_2_5ea438b14b40e_rand","el_class":"home-list-block","offset":"","css":"","tdc_css":"","tdc_css_class":"td_uid_2_5ea438b14b40e_rand","live_filter":"","live_filter_cur_post_id":"","live_filter_cur_post_author":"","block_template_id":""}`)

	client := &http.Client{}
	r, _ := http.NewRequest("POST", apiUrl, strings.NewReader(data.Encode()))
	r.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	resp, err := client.Do(r)
	if err != nil {
		fmt.Println(err)
	}
	body, _ := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()

	var content Content
	json.Unmarshal([]byte(string(body)), &content)

	z := html.NewTokenizer(strings.NewReader(content.Data))

	for {
		tt := z.Next()
		link := ""
		ok := false
		switch {
		case tt == html.ErrorToken:
			// End of the document, we're done
			return
		case tt == html.StartTagToken:
			t := z.Token()
			link, ok = getHref(t)

		}
		if !ok {
			continue
		}

		links = append(links, link)
	}
}

// Helper function to pull the href attribute from a Token
func getHref(t html.Token) (href string, ok bool) {
	// Check if the token is an <a> tag
	if t.Data != "a" {
		return
	}
	hasBookmarkRel := false
	hasValidHref := false
	// Iterate over token attributes until we find an "href"
	for _, a := range t.Attr {
		switch a.Key {
		case "href":
			href = a.Val
			hasValidHref = true
		case "rel":
			hasBookmarkRel = true
		}
	}
	ok = hasBookmarkRel && hasValidHref

	return
}

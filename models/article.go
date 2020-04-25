package models

import (
	"fmt"
	"time"
)

const (
	ChinaPress      = 1
	NanYang         = 2
	SinChew         = 3
	TheStar         = 4
	TheEdge         = 5
	Investing       = 6
	NewStraitsTimes = 7
	MalayMail       = 8
	BusinessInsider = 9
)

type Article struct {
	ID          int `gorm:"primary_key;"`
	Title       string
	Content     string
	Source      int
	URL         string
	Thumbnail   string
	PublishedAt time.Time
	CreatedAt   time.Time
}

var NewsSources map[int]string

func init() {
	NewsSources = map[int]string{
		ChinaPress:      "China Press (中国报)",
		NanYang:         "Nanyang Siang Pau (南洋商报)",
		SinChew:         "Sin Chew Daily (星洲日报)",
		TheStar:         "The Star",
		TheEdge:         "The Edge",
		Investing:       "Investing.com",
		NewStraitsTimes: "New Straits Times",
		MalayMail:       "Malay Mail",
		BusinessInsider: "Business Insider",
	}
}

// Create new article
func CreateArticle(input *Article) (*Article, error) {
	article := input

	db := GetDB()
	defer db.Close()

	// Create the user
	db.Create(article)

	if article.ID <= 0 {
		return nil, fmt.Errorf("Article is not created.")
	}

	return article, nil
}

// Get article links by new source
func GetArticlesBySource(sourceID int) []string {
	var links []string

	db := GetDB()
	defer db.Close()

	db.Table("articles").Where("source = ?", sourceID).Pluck("url", &links)

	return links
}

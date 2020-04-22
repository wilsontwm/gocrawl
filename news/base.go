package news

import (
	"time"
)

type Article struct {
	Title       string
	Content     string
	URL         string
	Thumbnail   string
	PublishedAt time.Time
}

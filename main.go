package main

import (
	"github.com/robfig/cron/v3"
	"gocrawl/news"
)

func main() {
	// go startCron()

	// select {}
}

// Start the cron jobs to scrape news from news press
func startCron() {
	c := cron.New()
	// International press
	c.AddFunc("0,15,30,45 * * * *", news.CrawlInvesting)

	// Local press (scraping at different interval)
	c.AddFunc("5,20,35,50 6-23 * * *", news.CrawlTheStar)
	c.AddFunc("6,21,36,51 6-23 * * *", news.CrawlTheEdge)
	c.AddFunc("7,22,37,52 6-23 * * *", news.CrawlNanYang)
	c.AddFunc("8,23,38,53 6-23 * * *", news.CrawlChinaPress)
	c.AddFunc("9,24,39,54 6-23 * * *", news.CrawlSinChew)
	c.AddFunc("10,25,40,55 6-23 * * *", news.CrawlNewStraitsTimes)

	c.Start()
}

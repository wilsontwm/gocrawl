package main

import (
	"fmt"
	"github.com/robfig/cron/v3"
	"gocrawl/news"
)

func main() {
	// go startCron()

	// select {}
	getNews()
}

func startCron() {
	c := cron.New()
	//c.AddFunc("@every 0h2m", getNews)
	//c.AddFunc("0,10,20,30,40,50 6-23 * * *", getNews)
	c.Start()
	getNews()
}

func getNews() {
	fmt.Println("Getting news")
	news.CrawlChinaPress()
}

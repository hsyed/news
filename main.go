package main

import (
	"github.com/hsyed/news/pkg/news"
	"log"
	"net/http"
)

func main() {
	feedConfig, err := news.LoadFeedServiceConfig("feeds.json")
	if err != nil {
		log.Panicf("could not load field config: %s", err)
	}
	svc, err := news.NewFeedService(feedConfig)
	if err != nil {
		log.Panicf("could not create feed service: %s", err)
	}
	err = http.ListenAndServe(":3000", news.NewFeedServeMux(svc))
	if err != nil {
		log.Panicf("could not launch feed service: %s", err)
	}
}

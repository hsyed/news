package news

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

var config = &FeedServiceConfig{
	Sources: []FeedSourceConfig{
		{
			Id: "bbc",
			Feeds: []*FeedConfig{
				{Id: "uk", URL: "http://feedMap.bbci.co.uk/news/uk/rss.xml"},
				{Id: "world", URL: "http://feedMap.bbci.co.uk/news/world/rss.xml"},
				{Id: "technology", URL: "http://feedMap.bbci.co.uk/news/technology/rss.xml"},
			},
		}, {
			Id: "reuters",
			Feeds: []*FeedConfig{
				{Id: "uk", URL: "http://feedMap.reuters.com/reuters/UKdomesticNews"},
				{Id: "world", URL: "http://feedMap.reuters.com/reuters/UKWorldNews"},
				{Id: "technology", URL: "http://feedMap.reuters.com/reuters/technologyNews"},
			},
		},
	},
	Topics: []FeedTopicConfig{
		{Id: "uk"},
		{Id: "technology"},
		{Id: "world"},
	},
}

func TestFeedService_SelectFeeds(t *testing.T) {
	svc, _ := NewFeedService(config)

	feeds := svc.selectFeeds(nil, nil)
	assert.Equal(t, 6, len(feeds))

	feeds = svc.selectFeeds([]string{"reuters"}, nil)
	assert.Equal(t, 3, len(feeds))

	feeds = svc.selectFeeds(nil, []string{"technology"})
	assert.Equal(t, 2, len(feeds))

	feeds = svc.selectFeeds([]string{"bbc"}, []string{"technology"})
	assert.Equal(t, 1, len(feeds))
}


package news

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/mmcdole/gofeed"
	"io/ioutil"
	"log"
	"os"
	"sort"
	"sync"
	"time"
)

type FeedConfig struct {
	Id       string `json:"id"`
	URL      string `json:"url"`
	SourceId string `json:"source_id,omitempty"`
}

type FeedSourceConfig struct {
	Id          string        `json:"id"`
	Description string        `json:"description,omitempty"`
	Feeds       []*FeedConfig `json:"feeds"`
}

type FeedTopicConfig struct {
	Id          string `json:"id"`
	Description string `json:"description,omitempty"`
}

type FeedServiceConfig struct {
	Sources []FeedSourceConfig `json:"sources"`
	Topics  []FeedTopicConfig  `json:"topics"`
}

func LoadFeedServiceConfig(path string) (*FeedServiceConfig, error) {
	jsonFile, err := os.Open(path)

	if err != nil {
		return nil, fmt.Errorf("could not load field config from %s: %s", path, err)
	}

	defer jsonFile.Close()

	byteValue, err := ioutil.ReadAll(jsonFile)
	if err != nil {
		return nil, fmt.Errorf("could not load field config from %s: %s", path, err)
	}

	feedConfig := FeedServiceConfig{}

	err = json.Unmarshal([]byte(byteValue), &feedConfig)
	if err != nil {
		return nil, err
	}

	return &feedConfig, nil
}

type FeedServiceSourceMeta struct {
	Id          string
	Description string
}

// FeedServiceMeta contains the metadata for client discovery
type FeedServiceMeta struct {
	Sources []FeedServiceSourceMeta `json:"sources"`
	Topics  []FeedTopicConfig       `json:"topics"`
}

type FeedService struct {
	feedMap    map[string]map[string]*FeedConfig
	config     *FeedServiceConfig
	meta       FeedServiceMeta
	allSources []string
	allTopics  []string
}

func NewFeedService(config *FeedServiceConfig) (*FeedService, error) {
	service := FeedService{
		feedMap: map[string]map[string]*FeedConfig{},
	}

	for _, source := range config.Sources {
		feeds := map[string]*FeedConfig{}
		for _, feed := range source.Feeds {
			feed.SourceId = source.Id
			feeds[feed.Id] = feed
		}
		service.feedMap[source.Id] = feeds

		// add the source id to allSources
		service.allSources = append(service.allSources, source.Id)
	}

	for _, topic := range config.Topics {
		service.allTopics = append(service.allTopics, topic.Id)
	}

	for _, source := range config.Sources {
		service.meta.Sources = append(service.meta.Sources, FeedServiceSourceMeta{
			Id:          source.Id,
			Description: source.Description,
		})
	}

	service.meta.Topics = config.Topics

	return &service, nil
}

// selectFeeds returns a list of feeds based on the selectors. If the selectors are nil then all relevant feeds are
// returned.
func (fs *FeedService) selectFeeds(sources []string, topics []string) (feeds []*FeedConfig) {
	if sources == nil {
		sources = fs.allSources
	}

	if topics == nil {
		topics = fs.allTopics
	}

	for _, source := range sources {
		sourceTopics := fs.feedMap[source]
		if sourceTopics != nil {
			for _, topicKey := range topics {
				topic := sourceTopics[topicKey]
				if topic != nil {
					feeds = append(feeds, topic)
				}
			}
		}
	}
	return feeds
}

type FeedItem struct {
	SourceId    string     `json:"source_id"`
	TopicId     string     `json:"topic_id"`
	Title       string     `json:"title"`
	Description string     `json:"description"`
	Link        string     `json:"link"`
	Published   *time.Time `json:"published"`
	Thumbnail   string     `json:"thumbnail,omitempty"`
}

func selectFeedImage(item *gofeed.Item) string {
	// select the image field if it is set
	if item.Image != nil {
		return item.Image.URL
	}

	// otherwise look for the thumbnail extension and select the url field.
	if len(item.Extensions) != 0 {
		if media, ok := item.Extensions["media"]; ok {
			if thumbnail, ok := media["thumbnail"]; ok {
				if len(thumbnail) > 0 {
					if url, ok := thumbnail[0].Attrs["url"]; ok {
						return url
					}
				}
			}
		}
	}
	return ""
}

type GetFeedItemsResponse struct {
	Items []FeedItem `json:"items"`
}

type GetFeedItemsRequest struct {
	Sources []string
	Topics  []string
}

func (fs *FeedService) GetFeedItems(_ context.Context, req *GetFeedItemsRequest) (*GetFeedItemsResponse, error) {
	// The context is currently unused.
	outFeed := GetFeedItemsResponse{}
	feeds := fs.selectFeeds(req.Sources, req.Topics)

	var lock sync.Mutex
	var wg sync.WaitGroup

	for _, feed := range feeds {
		wg.Add(1)

		go func(feed *FeedConfig) {
			defer wg.Done()
			fp := gofeed.NewParser()
			// ParseUrl should be wrapped with a timeout watchdog, or the feed body should be retrieved manually.
			resp, err := fp.ParseURL(feed.URL)
			if err != nil {
				log.Printf("error retrieving %s: %s", feed.URL, err)
			} else {
				lock.Lock()
				defer lock.Unlock()
				for _, item := range resp.Items {
					outItem := FeedItem{
						SourceId:    feed.SourceId,
						TopicId:     feed.Id,
						Title:       item.Title,
						Description: item.Description,
						Link:        item.Link,
						Published:   item.PublishedParsed,
						Thumbnail:   selectFeedImage(item),
					}

					outFeed.Items = append(outFeed.Items, outItem)
				}
			}
		}(feed)
	}
	wg.Wait()

	// sort the feed items in ascending order by publication time.
	sort.Slice(outFeed.Items, func(i, j int) bool {
		return outFeed.Items[i].Published.Before(*outFeed.Items[j].Published)
	})

	return &outFeed, nil
}

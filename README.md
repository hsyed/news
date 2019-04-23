This repo contains a simple service for aggregating RSS feeds for a mobile client.

To keep things simple the service must be launched from the repo directory and the service will bind to port 3000. 
Launch the service as follows:

```
go get
go run main.go
```

# Overview

The client aggregates feeds using two selectors, feed `sources` --e.g., `bbc` or `reuters` and `topics` --e.g., `uk`, 
`world` or `technology`. Feeds are returned in a simple JSON format suited to this specific application.

`feeds.json` contains the metadata defining the external feed sources and the topics to aggregate. 

# Endpoints:
 
## `/feeds/meta`

This is a "discovery" url for the client to discover the keys needed to retrieve feeds.

```json
{
    "sources": [
        {
            "Id": "bbc",
            "Description": "The BBC news feeds."
        },
        {
            "Id": "reuters",
            "Description": ""
        }
    ],
    "topics": [
        {
            "id": "technology",
            "description": "News articles about technology."
        },
        {
            "id": "uk"
        },
        {
            "id": "world"
        }
    ]
}
```

## `/feeds`

Retrieve feed data two string parameters are provided `sources` and `topics`. `/feeds?sources=bbc` would retrieve all 
topics from the BBC source only. `/feeds/topics=uk,world` would retrieve the two topics from all sources. An except of 
output looks as follows:

```json
{
    "items": [
        {
            "source_id": "bbc",
            "topic_id": "uk",
            "title": "BBC News Channel",
            "description": "Britain's most-watched news channel, delivering breaking news and analysis all day, every day.",
            "link": "https://www.bbc.co.uk/news/10318089",
            "published": "2019-03-11T22:33:15Z",
            "thumbnail": "http://c.files.bbci.co.uk/059C/production/_106463410_ecb0b10f-d372-402e-aced-e7b92ab4ebd1.png"
        },
        {
            "source_id": "reuters",
            "topic_id": "uk",
            "title": "Northern Ireland journalist killed by gunman during riot",
            "description": "A 29-year-old Northern Irish journalist was shot dead during rioting in Londonderry overnight, an attack that shocked the region and police said was likely the work of Irish nationalist militants opposed to the 1998 Good Friday peace deal.\u003cdiv class=\"feedflare\"\u003e\r\n\u003ca href=\"http://feeds.reuters.com/~ff/reuters/UKDomesticNews?a=YVUS7f2fR0Y:541Tfq1itxc:yIl2AUoC8zA\"\u003e\u003cimg src=\"http://feeds.feedburner.com/~ff/reuters/UKDomesticNews?d=yIl2AUoC8zA\" border=\"0\"\u003e\u003c/img\u003e\u003c/a\u003e \u003ca href=\"http://feeds.reuters.com/~ff/reuters/UKDomesticNews?a=YVUS7f2fR0Y:541Tfq1itxc:F7zBnMyn0Lo\"\u003e\u003cimg src=\"http://feeds.feedburner.com/~ff/reuters/UKDomesticNews?i=YVUS7f2fR0Y:541Tfq1itxc:F7zBnMyn0Lo\" border=\"0\"\u003e\u003c/img\u003e\u003c/a\u003e \u003ca href=\"http://feeds.reuters.com/~ff/reuters/UKDomesticNews?a=YVUS7f2fR0Y:541Tfq1itxc:V_sGLiPBpWU\"\u003e\u003cimg src=\"http://feeds.feedburner.com/~ff/reuters/UKDomesticNews?i=YVUS7f2fR0Y:541Tfq1itxc:V_sGLiPBpWU\" border=\"0\"\u003e\u003c/img\u003e\u003c/a\u003e\r\n\u003c/div\u003e\u003cimg src=\"http://feeds.feedburner.com/~r/reuters/UKDomesticNews/~4/YVUS7f2fR0Y\" height=\"1\" width=\"1\" alt=\"\"/\u003e",
            "link": "http://feeds.reuters.com/~r/reuters/UKDomesticNews/~3/YVUS7f2fR0Y/northern-ireland-journalist-killed-by-gunman-during-riot-idUKKCN1RV02F",
            "published": "2019-04-19T20:20:27Z"
        }
    ]
}
``` 

# Design & Notes

* The only external library used is `github.com/mmcdole/gofeed` for rendering feed formats.
* Since the service aggregates over topics and sources each feed item embeds the source and topic Ids so the client may 
further sort them as needed.
* `github.com/gorilla/feeds` could be used to construct feed outputs that conform to various feed standards. It 
was not used in this repository to keep things simple. 
* Reuters feeds contain images embedded in the description whereas BBC feeds use an extension to provide a thumbnail. 
More advanced heuristics for cleansing and normalizing the data are needed.
* The external services are queried in parallel however the code is not production ready. Proper timeout handling needs 
to be added. Specifically `ParseURL` from `gofeeds` does not use the HTTP client in a way that enables proper timeouts 
and cancellation. 
* To add to the previous point caching should be added to the feed client. The simplest way would be to just cache at 
the transport layer using something like `github.com/gregjones/httpcache` -- this assumes that the feed providers set 
correct caching headers.
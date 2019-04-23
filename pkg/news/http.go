package news

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
)

type HttpFeedMux struct {
	*http.ServeMux
	svc *FeedService
}

// NewFeedServeMux creates a ServeMux for servicing feed requests.
func NewFeedServeMux(svc *FeedService) *HttpFeedMux {
	httpSvc := HttpFeedMux{svc: svc, ServeMux: http.NewServeMux()}
	httpSvc.HandleFunc("/feeds", httpSvc.feedsHandler)
	httpSvc.HandleFunc("/feeds/meta", httpSvc.feedsMetaHandler)
	return &httpSvc
}

func (fhs *HttpFeedMux) feedsHandler(w http.ResponseWriter, r *http.Request) {
	items, err := fhs.svc.GetFeedItems(
		context.Background(),
		&GetFeedItemsRequest{
			Sources: mergedDelimitedUrlParams(r, "sources"),
			Topics: mergedDelimitedUrlParams(r, "topics"),
		},
	)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	renderJson(w, items)
}

func (fhs *HttpFeedMux) feedsMetaHandler(w http.ResponseWriter, r *http.Request) {
	renderJson(w, fhs.svc.meta)
}

// Merges query params including splitting with `,` as the delimiter. So `?topics=uk,world&topics=technology` would
// return the slice `["uk","world","technology"].
func mergedDelimitedUrlParams(r *http.Request, key string) (values []string) {
	for _, str := range r.URL.Query()[key] {
		for _, value := range strings.Split(str, ",") {
			values = append(values, value)
		}
	}
	return
}

// renderJson writes out a JSON struct via the http.ResponseWriter.
func renderJson(w http.ResponseWriter, data interface{}) {
	js, err := json.MarshalIndent(data, "", "    ")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write(js)
}

package main

import (
	"context"
	"encoding/xml"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/friskywombat/blog-aggregator/internal/database"
)

// RssData was generated 2024-08-17 21:17:31 by https://xml-to-go.github.io/ in Ukraine.
type RssData struct {
	XMLName xml.Name `xml:"rss" json:"rss,omitempty"`
	Text    string   `xml:",chardata" json:"text,omitempty"`
	Version string   `xml:"version,attr" json:"version,omitempty"`
	Atom    string   `xml:"atom,attr" json:"atom,omitempty"`
	Channel struct {
		Text  string `xml:",chardata" json:"text,omitempty"`
		Title string `xml:"title"`
		Link  struct {
			Text string `xml:",chardata" json:"text,omitempty"`
			Href string `xml:"href,attr" json:"href,omitempty"`
			Rel  string `xml:"rel,attr" json:"rel,omitempty"`
			Type string `xml:"type,attr" json:"type,omitempty"`
		} `xml:"link" json:"link,omitempty"`
		Description   string `xml:"description"`
		Generator     string `xml:"generator"`
		Language      string `xml:"language"`
		LastBuildDate string `xml:"lastBuildDate"`
		Item          []struct {
			Text        string `xml:",chardata" json:"text,omitempty"`
			Title       string `xml:"title"`
			Link        string `xml:"link"`
			PubDate     string `xml:"pubDate"`
			GUID        string `xml:"guid"`
			Description string `xml:"description"`
		} `xml:"item" json:"item,omitempty"`
	} `xml:"channel" json:"channel,omitempty"`
}

func (cfg *apiConfig) fetchFromFeed(ctxt context.Context, feed database.Feed) (RssData, error) {
	r, err := http.Get(feed.Url)
	if err != nil {
		return RssData{}, fmt.Errorf("Error '%v' from Url: %s", err.Error(), feed.Url)
	}
	dat := RssData{}
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return RssData{}, fmt.Errorf("Malformed response from Url: " + feed.Url)
	}
	err = xml.Unmarshal(body, &dat)
	if err != nil {
		return dat, fmt.Errorf("Malformed XML from Url: " + feed.Url)
	}
	cfg.DB.MarkFeedFetched(ctxt, feed.ID)
	return dat, nil
}

func (cfg *apiConfig) fetchLoop(ctxt context.Context) {
	ticker := time.NewTicker(time.Second * 30)
	for range ticker.C {
		feeds, err := cfg.DB.GetNextFeedsToFetch(ctxt, 5)
		if err != nil {
			log.Println("!Fetch Error!:", err.Error())
			continue
		}
		c := make(chan RssData)
		for _, f := range feeds {
			go func(ctxt context.Context, feed database.Feed) {
				dat, err := cfg.fetchFromFeed(ctxt, f)
				if err != nil {
					log.Println("!Fetch Error!:", err.Error())
					c <- RssData{}
					return
				}
				c <- dat
				log.Println(" Fetched <-", dat.Channel.Title)
			}(ctxt, f)
		}
		rssDatas := make([]RssData, len(feeds))
		for i := 0; i < len(feeds); i++ {
			r := <-c
			rssDatas[i] = r
		}
		log.Println("Done fetching for now...")
	}
}

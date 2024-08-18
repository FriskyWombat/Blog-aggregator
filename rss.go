package main

import (
	"context"
	"database/sql"
	"encoding/xml"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/friskywombat/blog-aggregator/internal/database"
	"github.com/google/uuid"
)

// RssData was generated 2024-08-17 21:17:31 by https://xml-to-go.github.io/ in Ukraine.
type RssData struct {
	feedID  uuid.UUID
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

// Post with labeled fields for json marshalling
type Post struct {
	ID          uuid.UUID      `json:"id"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	Title       string         `json:"title"`
	URL         string         `json:"url"`
	Description sql.NullString `json:"description"`
	PublishedAt time.Time      `json:"published_at"`
	FeedID      uuid.UUID      `json:"feed_id"`
}

func toPost(p database.Post) Post {
	return Post{
		ID:          p.ID,
		CreatedAt:   p.CreatedAt,
		UpdatedAt:   p.UpdatedAt,
		Title:       p.Title,
		URL:         p.Url,
		Description: p.Description,
		PublishedAt: p.PublishedAt,
		FeedID:      p.FeedID,
	}
}

// 2024-08-09T05:54:00.000-05:00
const numLayout = "2006-01-02T03:04:05.000-07:00"

// Sat, 29 Oct 2022 15:24:25 +0000
const wordyLayout = "Mon, 02 Jan 2006 15:04:05 -0700"

func parseDate(date string) (time.Time, error) {
	if len(date) < 15 {
		return time.Now(), fmt.Errorf("Bad timestamp reading : " + date)
	}
	var layout string
	if date[0] >= '0' && date[0] <= '9' {
		layout = numLayout
	} else {
		layout = wordyLayout
	}
	t, err := time.Parse(layout, date)
	if err != nil {
		log.Printf("Time formatting error: %v\n", err)
		return time.Now(), err
	}
	return t, nil
}

func (cfg *apiConfig) newPost(ctxt context.Context, title string, url string, description string, publishedAt string, feedID uuid.UUID) (Post, error) {
	id := uuid.New()
	pub, err := parseDate(publishedAt)
	if err != nil {
		pub = time.Now()
	}
	param := database.CreatePostParams{
		ID:          id,
		Title:       title,
		Url:         url,
		Description: sql.NullString{description, true},
		PublishedAt: pub,
		FeedID:      feedID,
	}
	p, err := cfg.DB.CreatePost(ctxt, param)
	if err != nil {
		return Post{}, err
	}
	if err != nil {
		return Post{}, err
	}
	return toPost(p), nil
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
				dat.feedID = f.ID
				c <- dat
				log.Println(" Fetched <-", dat.Channel.Title)
			}(ctxt, f)
		}
		for i := 0; i < len(feeds); i++ {
			r := <-c
			for _, p := range r.Channel.Item {
				res, err := cfg.newPost(ctxt, p.Title, p.Link, p.Description, p.PubDate, r.feedID)
				if err != nil {
					// We can ignore errors caused by duplicate URLs because we expect it to happen very often
					if !strings.Contains(err.Error(), "duplicate key value") {
						log.Println(err)
					}
					continue
				}
				log.Printf("Post created:\n\tTitle: %s\n\tDescription: %s\n", res.Title, res.Description.String)
			}
		}
		log.Println("Done fetching for now...")
	}
}

func (cfg *apiConfig) getPostsHandleFunc(w http.ResponseWriter, r *http.Request) {
	user, ok := r.Context().Value(userKey).(database.User)
	if !ok {
		respondWithError(w, 500, "Failed to retrieve user data from context")
		return
	}
	limit, err := strconv.Atoi(r.URL.Query().Get("limit"))
	if err != nil {
		limit = 10
	}
	param := database.GetPostsByUserParams{UserID: user.ID, Limit: int32(limit)}
	posts, err := cfg.DB.GetPostsByUser(r.Context(), param)
	if err != nil {
		respondWithError(w, 400, err.Error())
		return
	}
	ps := []Post{}
	for _, p := range posts {
		ps = append(ps, toPost(p))
	}
	respondWithJSON(w, 200, ps)
}

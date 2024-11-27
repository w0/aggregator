package main

import (
	"context"
	"database/sql"
	"encoding/xml"
	"fmt"
	"html"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/w0/aggregator/internal/database"
)

type RSSFeed struct {
	Channel struct {
		Title       string    `xml:"title"`
		Link        string    `xml:"link"`
		Description string    `xml:"description"`
		Item        []RSSItem `xml:"item"`
	} `xml:"channel"`
}

type RSSItem struct {
	Title       string `xml:"title"`
	Link        string `xml:"link"`
	Description string `xml:"description"`
	PubDate     string `xml:"pubDate"`
}

func fetchFeed(ctx context.Context, feedURL string) (*RSSFeed, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", feedURL, nil)

	if err != nil {
		return &RSSFeed{}, err
	}

	req.Header.Add("User-Agent", "gator")

	client := http.Client{}

	res, err := client.Do(req)

	if err != nil {
		return &RSSFeed{}, err
	}

	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)

	if err != nil {
		return &RSSFeed{}, err
	}

	var feed RSSFeed

	err = xml.Unmarshal(body, &feed)

	if err != nil {
		return &RSSFeed{}, err
	}

	unescapeHTML(&feed)

	return &feed, nil

}

func unescapeHTML(feed *RSSFeed) {
	feed.Channel.Title = html.UnescapeString(feed.Channel.Title)
	feed.Channel.Description = html.UnescapeString(feed.Channel.Description)

	for i := 0; i < len(feed.Channel.Item); i++ {
		feed.Channel.Item[i].Title = html.UnescapeString(feed.Channel.Item[i].Title)
		feed.Channel.Item[i].Description = html.UnescapeString(feed.Channel.Item[i].Description)
	}
}

func scrapeFeeds(s *state) error {
	feed, err := s.db.GetNextFeedToFetch(context.Background())

	if err != nil {
		return fmt.Errorf("failed getting next feed. %w", err)
	}

	now := time.Now()

	err = s.db.MarkFeedFetched(context.Background(),
		database.MarkFeedFetchedParams{
			ID:        feed.ID,
			UpdatedAt: now,
			LastFetchedAt: sql.NullTime{
				Time:  now,
				Valid: true,
			},
		})

	if err != nil {
		return fmt.Errorf("failed updating fetched time. %w", err)
	}

	rssFeed, err := fetchFeed(context.Background(), feed.Url)

	if err != nil {
		return fmt.Errorf("failed to fetch feed. %w", err)
	}

	return saveFeeds(s, rssFeed, feed)
}

func saveFeeds(s *state, rss *RSSFeed, feed database.Feed) error {

	for _, v := range rss.Channel.Item {
		publishedAt := sql.NullTime{}

		if time, err := time.Parse(time.RFC1123Z, v.PubDate); err == nil {
			publishedAt.Time = time
			publishedAt.Valid = true
		}

		now := time.Now()
		_, err := s.db.CreatePost(context.Background(),
			database.CreatePostParams{
				ID:        uuid.New(),
				CreatedAt: now,
				UpdatedAt: now,
				Title:     v.Title,
				Url:       v.Link,
				Description: sql.NullString{
					String: v.Description,
					Valid:  true,
				},
				PublishedAt: publishedAt,
				FeedID:      feed.ID,
			})
		if err != nil {
			if strings.Contains(err.Error(), "duplicate key value violates unique constraint") {
				continue
			}
			return err
		}
	}

	return nil
}

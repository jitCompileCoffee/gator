package main

import (
	"context"
	"database/sql"
	"encoding/xml"
	"errors"
	"fmt"
	"html"
	"io"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jitCompileCoffee/blog-agg/internal/database"
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
	client := http.Client{
		Timeout: 10 * time.Second,
	}
	req, err := http.NewRequestWithContext(ctx, "GET", feedURL, nil)
	if err != nil {
		return &RSSFeed{}, err
	}
	req.Header.Set("User-Agent", "gator")
	res, err := client.Do(req)
	if err != nil {
		return &RSSFeed{}, err
	}
	defer res.Body.Close()
	rawData, err := io.ReadAll(res.Body)
	if err != nil {
		return &RSSFeed{}, err
	}
	feed := RSSFeed{}
	if err := xml.Unmarshal(rawData, &feed); err != nil {
		return &RSSFeed{}, err
	}

	feed.Channel.Title = html.UnescapeString(feed.Channel.Title)
	feed.Channel.Description = html.UnescapeString(feed.Channel.Description)

	for i := range feed.Channel.Item {
		feed.Channel.Item[i].Title = html.UnescapeString(feed.Channel.Item[i].Title)
		feed.Channel.Item[i].Description = html.UnescapeString(feed.Channel.Item[i].Description)
	}

	return &feed, nil
}

func scrapeFeeds(s *state) error {
	feed, err := s.db.GetNextFeedToFetch(context.Background())
	if err != nil {
		return fmt.Errorf("could not get next feed to fetch: %w", err)
	}
	if err := s.db.MarkFeedFetched(context.Background(), feed.ID); err != nil {
		return fmt.Errorf("could not mark feed as fetched: %w", err)
	}
	fetchedFeed, err := fetchFeed(context.Background(), feed.Url)
	if err != nil {
		return fmt.Errorf("failed to fetch feed from URL %s: %w", feed.Url, err)
	}
	fmt.Printf("\n--- Found %d posts in feed: %s ---\n", len(fetchedFeed.Channel.Item), feed.Name)
	for _, item := range fetchedFeed.Channel.Item {
		// Define a list of common time layouts found in RSS feeds
		timeLayouts := []string{time.RFC1123Z, time.RFC1123, time.RFC3339, time.ANSIC}

		pubAt := sql.NullTime{Valid: false} // Default to NULL

		if item.PubDate != "" {
			for _, layout := range timeLayouts {
				// Try parsing the string using the current layout layout
				parsedTime, err := time.Parse(layout, item.PubDate)
				if err == nil {
					pubAt.Time = parsedTime
					pubAt.Valid = true
					break // We successfully parsed it, stop trying other layouts!
				}
			}
		}
		description := sql.NullString{}
		if item.Description != "" {
			description.String = item.Description
			description.Valid = true
		} else {
			description.Valid = false
		}
		_, err := s.db.CreatePost(context.Background(), database.CreatePostParams{
			ID:          uuid.New(),
			CreatedAt:   time.Now().UTC(),
			UpdatedAt:   time.Now().UTC(),
			Title:       item.Title,
			Url:         item.Link,
			Description: description,
			PublishedAt: pubAt,
			FeedID:      feed.ID,
		})
		if err != nil {
			// 1. Create an empty pointer for the pgx connection error type
			var pgErr *pgconn.PgError

			// 2. Check if the error is actually a Postgres system error
			if errors.As(err, &pgErr) {
				// 3. Code "23505" is the strict standard code for unique_violation
				if pgErr.Code == "23505" {
					// This is a duplicate post! We can safely skip it.
					continue
				}
			}

			// If it's any OTHER kind of error (e.g. lost connection), log it or return it
			fmt.Printf("Failed to create post: %v\n", err)
		}
	}
	return nil
}

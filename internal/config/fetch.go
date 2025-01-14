package config

import (
	"context"
	"encoding/xml"
	"fmt"
	"html"
	"io"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/tholho/aggreGator/internal/database"
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
		return nil, err
	}
	req.Header.Set("User-Agent", "gator")
	client := http.DefaultClient
	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	rawData, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	feed := RSSFeed{}
	err = xml.Unmarshal(rawData, &feed)
	if err != nil {
		feed = RSSFeed{}
		stringedData := html.UnescapeString(string(rawData))
		rawData = []byte(stringedData)
		err = xml.Unmarshal(rawData, &feed)
		if err != nil {
			fmt.Println(string(rawData))
			return nil, fmt.Errorf("error parsing XML or HTML: %w", err)
		}
	}
	feed.Channel.Title = html.UnescapeString(feed.Channel.Title)
	feed.Channel.Description = html.UnescapeString(feed.Channel.Description)
	for i, val := range feed.Channel.Item {
		feed.Channel.Item[i].Description = html.UnescapeString(val.Description)
		feed.Channel.Item[i].Title = html.UnescapeString(val.Title)
	}
	return &feed, nil
}

func scrapeFeeds(s *State) error {
	feed, err := s.Db.GetNextFeedToFetch(context.Background())
	if err != nil {
		fmt.Println("Error in scrapFeeds (fetch.go) - get feed id")
		fmt.Println(err)
		return err
	}
	err = s.Db.MarkFeedFetched(context.Background(), feed.ID)
	if err != nil {
		fmt.Println("Error in scrapFeeds (fetch.go) - marking")
		fmt.Println(err)
		return err
	}
	rssFeed, err := fetchFeed(context.Background(), feed.Url)
	if err != nil {
		fmt.Println("Error in scrapFeeds (fetch.go) - fetchfeed")
		fmt.Println(err)
		return err
	}

	for _, val := range rssFeed.Channel.Item {
		layout := "Mon, 02 Jan 2006 15:04:05 +0000"
		value := val.PubDate
		pubVal, err := time.Parse(layout, value)
		if err != nil {
			fmt.Println(err)
			return err
		}
		postParams := database.CreatePostParams{
			ID:          uuid.New(),
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
			Title:       val.Title,
			Description: val.Description,
			Url:         val.Link,
			PublishedAt: pubVal,
			FeedID:      feed.ID,
		}
		err = s.Db.CreatePost(context.Background(), postParams)
		if err != nil {
			fmt.Println(err)
			return err
		}
	}
	return nil
}

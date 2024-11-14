package database

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/deatil/go-encoding/base62"
	"github.com/gorilla/feeds"
	"github.com/mmcdole/gofeed"
)

var (
	dataDir      = "./data"
	feedFilename = "feed.json"
)

func init() {
	if v := os.Getenv("DATA_DIR"); v != "" {
		dataDir = v
	}
}

func GetFeed(url string) (*feeds.Feed, error) {
	feedDir := filepath.Join(dataDir, base62encode(url))

	// open feed file
	f, err := os.Open(filepath.Join(feedDir, feedFilename))
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, &ErrFeedNotFound{URL: url}
		}

		return nil, fmt.Errorf("cannot open feed: %w", err)
	}
	defer f.Close()

	var feed gofeed.Feed
	if err := json.NewDecoder(f).Decode(&feed); err != nil {
		return nil, fmt.Errorf("cannot decode feed: %w", err)
	}

	itemPaths, err := filepath.Glob(filepath.Join(feedDir, "items", "*.json"))
	if err != nil {
		return nil, fmt.Errorf("cannot read items: %w", err)
	}

	for _, itemPath := range itemPaths {
		item, err := readItem(itemPath)
		if err != nil {
			return nil, err
		}

		feed.Items = append(feed.Items, item)
	}

	writeableFeed := makeWriteable(&feed)

	return &writeableFeed, nil
}

func makeWriteable(feed *gofeed.Feed) feeds.Feed {
	image := &feeds.Image{}
	if feed.Image != nil {
		image.Url = feed.Image.URL
		image.Title = feed.Image.Title
		image.Link = feed.Image.URL
	}

	var author *feeds.Author
	if feed.Authors != nil && len(feed.Authors) > 0 {
		author = &feeds.Author{
			Name:  feed.Authors[0].Name,
			Email: feed.Authors[0].Email,
		}
	}

	items := []*feeds.Item{}
	for _, item := range feed.Items {
		enclosure := &feeds.Enclosure{}
		if item.Enclosures != nil && len(item.Enclosures) > 0 {
			enclosure.Url = item.Enclosures[0].URL
			enclosure.Type = item.Enclosures[0].Type
			enclosure.Length = item.Enclosures[0].Length
		}

		var updated time.Time
		if item.UpdatedParsed != nil {
			updated = *item.UpdatedParsed
		}

		var created time.Time
		if item.PublishedParsed != nil {
			created = *item.PublishedParsed
		}

		var author *feeds.Author
		if item.Author != nil {
			author = &feeds.Author{
				Name:  item.Author.Name,
				Email: item.Author.Email,
			}
		}

		items = append(items, &feeds.Item{
			Title: item.Title,

			Link: &feeds.Link{
				Href: item.Link,
			},
			Source: &feeds.Link{
				Href: item.Link,
			},
			Author:      author,
			Description: item.Description,
			Id:          firstNonEmpty(item.GUID, item.Link),
			IsPermaLink: "true",
			Updated:     updated,
			Created:     created,
			Enclosure:   enclosure,
			Content:     item.Content,
		})
	}

	var updated time.Time
	if feed.UpdatedParsed != nil {
		updated = *feed.UpdatedParsed
	}

	var created time.Time
	if feed.PublishedParsed != nil {
		created = *feed.PublishedParsed
	}

	f := feeds.Feed{
		Title:       feed.Title,
		Link:        &feeds.Link{Href: feed.FeedLink},
		Description: feed.Description,
		Author:      author,
		Updated:     updated,
		Created:     created,
		Id:          feed.FeedLink,
		Items:       items,
		Copyright:   feed.Copyright,
		Image:       image,
	}

	return f
}

func ParseAndSaveFeed(r io.Reader) (url string, count int, err error) {
	feed, err := gofeed.NewParser().Parse(r)
	if err != nil {
		return "", 0, fmt.Errorf("cannot parse feed: %w", err)
	}

	if err := saveFeed(*feed); err != nil {
		return "", 0, fmt.Errorf("cannot save feed: %w", err)
	}

	return feed.FeedLink, feed.Len(), nil
}

func saveFeed(feed gofeed.Feed) error {
	feedDir := filepath.Join(dataDir, base62encode(feed.FeedLink))

	// mkdir
	if err := os.MkdirAll(filepath.Join(feedDir, "items"), 0755); err != nil {
		return fmt.Errorf("cannot create feed directory: %w", err)
	}

	// save items
	for _, item := range feed.Items {
		itemSavePath := filepath.Join(feedDir, "items", itemFilename(item))

		if err := saveItem(item, itemSavePath); err != nil {
			return err
		}
	}

	// empty items
	feed.Items = nil

	// open feed file
	f, err := os.Create(filepath.Join(feedDir, feedFilename))
	if err != nil {
		return fmt.Errorf("cannot create feed file: %w", err)
	}
	defer f.Close()

	// write feed
	if err := json.NewEncoder(f).Encode(feed); err != nil {
		return fmt.Errorf("cannot encode feed: %w", err)
	}

	return nil
}

func saveItem(item *gofeed.Item, path string) error {
	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("cannot create item: %w", err)
	}
	defer f.Close()

	if err := json.NewEncoder(f).Encode(item); err != nil {
		return fmt.Errorf("cannot encode item: %w", err)
	}

	return nil
}

func readItem(path string) (*gofeed.Item, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var item gofeed.Item

	if err := json.NewDecoder(f).Decode(&item); err != nil {
		return nil, err
	}

	return &item, nil
}

func itemFilename(item *gofeed.Item) string {
	return fmt.Sprintf("%s-%s.json", item.PublishedParsed.Format("2006-01-02"), md5sum(item.Link))
}

func md5sum(s string) string {
	hash := md5.Sum([]byte(s))

	return hex.EncodeToString(hash[:])
}

func base62encode(s string) string {
	return base62.StdEncoding.EncodeToString([]byte(s))
}

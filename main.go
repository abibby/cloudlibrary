package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/abibby/cloudlibrary/cloudlibrary"
	"github.com/abibby/cloudlibrary/libgourou"
	"github.com/joho/godotenv"
)

var ebookTitlePattern = regexp.MustCompile(`[^a-zA-Z0-9]+`)

func main() {
	godotenv.Load()

	adeptDir := "/data/adept"

	activate := sync.OnceValue(func() error {
		return libgourou.Activate(adeptDir)
	})

	c := cloudlibrary.New()

	login := func() error {
		return c.Login(os.Getenv("CLOUDLIBRARY_LIBRARY"), os.Getenv("CLOUDLIBRARY_USERNAME"), os.Getenv("CLOUDLIBRARY_PASSWORD"))
	}
	if err := login(); err != nil {
		log.Fatalf("Login: %v", err)
	}

	downloadTicker := time.NewTicker(5 * time.Minute)
	defer downloadTicker.Stop()
	loginTicker := time.NewTicker(24 * time.Hour)
	defer loginTicker.Stop()

	if err := downloadBooks(c, adeptDir, activate); err != nil {
		log.Printf("Download books: %v", err)
	}
	for {
		select {
		case <-loginTicker.C:
			if err := login(); err != nil {
				log.Printf("Login: %v", err)
			}
		case <-downloadTicker.C:
			if err := downloadBooks(c, adeptDir, activate); err != nil {
				log.Printf("Download books: %v", err)
			}
		}
	}
}

func downloadBooks(c *cloudlibrary.Client, adeptDir string, activate func() error) error {
	books, err := c.MyBooks()
	if err != nil {
		return fmt.Errorf("retrieve borrowed books: %w", err)
	}

	for _, b := range books {
		author := b.Contributors[0].Name
		if last, first, ok := strings.Cut(author, ","); ok {
			author = strings.TrimSpace(strings.TrimSpace(first) + " " + strings.TrimSpace(last))
		}
		dir := filepath.Join("/books", author, b.Title)
		metadata := MetadataManager{Dir: dir}
		state, err := metadata.State()
		if err != nil {
			return fmt.Errorf("read metadata for %q: %w", b.Title, err)
		}
		if b.MediaCategory == "AudioBook" {
			if state.Audiobook {
				log.Printf("Skipping already downloaded audiobook %s\n", b.Title)
				continue
			}
			log.Printf("Downloading audiobook %s\n", b.Title)
			err = c.DownloadAudiobook(b.ItemID, dir)
			if err != nil {
				return fmt.Errorf("download audiobook %q: %w", b.Title, err)
			}
			if err := metadata.Update(false, true); err != nil {
				return fmt.Errorf("update metadata for %q: %w", b.Title, err)
			}
		}
		if b.MediaCategory == "eBook" {
			if state.Ebook {
				log.Printf("Skipping already downloaded ebook %s\n", b.Title)
				continue
			}
			log.Printf("Downloading ebook %s\n", b.Title)
			err = activate()
			if err != nil {
				return fmt.Errorf("activate Adobe device: %w", err)
			}

			title := ebookTitlePattern.ReplaceAllString(b.Title, "_")
			filename := filepath.Join(dir, title+".acsm")
			err = c.DownloadEBook(b, filename)
			if err != nil {
				return fmt.Errorf("download ebook %q: %w", b.Title, err)
			}
			err = libgourou.DownloadEPUB(
				adeptDir,
				filename,
				filepath.Join(dir, title+".epub"),
			)
			if err != nil {
				return fmt.Errorf("download EPUB for %q: %w", b.Title, err)
			}
			if err := metadata.Update(true, false); err != nil {
				return fmt.Errorf("update metadata for %q: %w", b.Title, err)
			}
		}
	}
	return nil
}

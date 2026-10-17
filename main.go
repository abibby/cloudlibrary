package main

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/abibby/cloudlibrary/cloudlibrary"
	"github.com/abibby/cloudlibrary/libgourou"
)

var ebookTitlePattern = regexp.MustCompile(`[^a-zA-Z0-9]+`)

func main() {
	adeptDir := "/data/adept"
	err := libgourou.Activate(adeptDir)
	if err != nil {
		panic(err)
	}

	c := cloudlibrary.New()

	books, err := c.MyBooks()
	if err != nil {
		panic(err)
	}

	for _, b := range books {
		author := b.Contributors[0].Name
		if last, first, ok := strings.Cut(author, ","); ok {
			author = strings.TrimSpace(strings.TrimSpace(first) + " " + strings.TrimSpace(last))
		}
		dir := filepath.Join("/books", author, b.Title)
		if b.MediaCategory == "AudioBook" {
			fmt.Printf("Downloading audiobook %s\n", b.Title)
			err = c.DownloadAudiobook(b.ItemID, dir)
			if err != nil {
				panic(err)
			}
		}
		if b.MediaCategory == "eBook" {
			fmt.Printf("Downloading ebook %s\n", b.Title)
			title := ebookTitlePattern.ReplaceAllString(b.Title, "_")
			filename := filepath.Join(dir, title+".acsm")
			err = c.DownloadEBook(b, filename)
			if err != nil {
				panic(err)
			}
			err = libgourou.DownloadEPUB(
				adeptDir,
				filename,
				filepath.Join(dir, title+".epub"),
			)
			if err != nil {
				panic(err)
			}
		}
	}
}

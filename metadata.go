package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type downloadMetadata struct {
	Ebook     bool `json:"ebook"`
	Audiobook bool `json:"audiobook"`
}

// MetadataManager manages the .cloudlibrary.json file in a book's directory.
type MetadataManager struct {
	Dir string
}

// State reads the saved download status. Missing metadata means neither format
// has been downloaded.
func (m MetadataManager) State() (downloadMetadata, error) {
	filename := filepath.Join(m.Dir, ".cloudlibrary.json")
	metadata := downloadMetadata{}
	data, err := os.ReadFile(filename)
	if err == nil {
		if err := json.Unmarshal(data, &metadata); err != nil {
			return downloadMetadata{}, fmt.Errorf("decode download metadata: %w", err)
		}
	} else if !os.IsNotExist(err) {
		return downloadMetadata{}, fmt.Errorf("read download metadata: %w", err)
	}
	return metadata, nil
}

// Update creates or updates the metadata with successfully downloaded formats.
// A false argument preserves that format's existing status.
func (m MetadataManager) Update(ebook, audiobook bool) error {
	metadata, err := m.State()
	if err != nil {
		return err
	}
	metadata.Ebook = metadata.Ebook || ebook
	metadata.Audiobook = metadata.Audiobook || audiobook
	data, err := json.Marshal(metadata)
	if err != nil {
		return fmt.Errorf("encode download metadata: %w", err)
	}
	if err := os.MkdirAll(m.Dir, 0755); err != nil {
		return fmt.Errorf("create metadata directory: %w", err)
	}

	// Replace the metadata only after the complete update has been written.
	file, err := os.CreateTemp(m.Dir, ".cloudlibrary-*.json")
	if err != nil {
		return fmt.Errorf("create metadata file: %w", err)
	}
	defer os.Remove(file.Name())
	_, writeErr := file.Write(data)
	closeErr := file.Close()
	if writeErr != nil {
		return fmt.Errorf("write download metadata: %w", writeErr)
	}
	if closeErr != nil {
		return fmt.Errorf("close metadata file: %w", closeErr)
	}
	if err := os.Rename(file.Name(), filepath.Join(m.Dir, ".cloudlibrary.json")); err != nil {
		return fmt.Errorf("publish download metadata: %w", err)
	}
	return nil
}

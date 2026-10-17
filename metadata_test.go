package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestMetadataManagerState(t *testing.T) {
	for _, tt := range []struct {
		name    string
		data    string
		want    downloadMetadata
		wantErr bool
	}{
		{name: "missing"},
		{name: "neither", data: `{"ebook":false,"audiobook":false}`},
		{name: "ebook", data: `{"ebook":true,"audiobook":false}`, want: downloadMetadata{Ebook: true}},
		{name: "audiobook", data: `{"ebook":false,"audiobook":true}`, want: downloadMetadata{Audiobook: true}},
		{name: "both", data: `{"ebook":true,"audiobook":true}`, want: downloadMetadata{Ebook: true, Audiobook: true}},
		{name: "invalid JSON", data: `{"ebook":`, wantErr: true},
		{name: "invalid status", data: `{"ebook":"true"}`, wantErr: true},
	} {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			filename := filepath.Join(dir, ".cloudlibrary.json")
			if tt.data != "" {
				if err := os.WriteFile(filename, []byte(tt.data), 0600); err != nil {
					t.Fatal(err)
				}
			}
			manager := MetadataManager{Dir: dir}
			got, err := manager.State()
			if (err != nil) != tt.wantErr {
				t.Fatalf("State() error = %v, want error = %v", err, tt.wantErr)
			}
			if got != tt.want {
				t.Fatalf("State() = %+v, want %+v", got, tt.want)
			}
			if tt.data == "" {
				if _, err := os.Stat(filename); !os.IsNotExist(err) {
					t.Fatalf("State() should not create metadata, stat error = %v", err)
				}
			}
		})
	}
}

func TestMetadataManagerUpdate(t *testing.T) {
	for _, tt := range []struct {
		name    string
		updates [][2]bool
		want    string
	}{
		{"empty", [][2]bool{{false, false}}, `{"ebook":false,"audiobook":false}`},
		{"ebook", [][2]bool{{true, false}}, `{"ebook":true,"audiobook":false}`},
		{"audiobook", [][2]bool{{false, true}}, `{"ebook":false,"audiobook":true}`},
		{"ebook then audiobook", [][2]bool{{true, false}, {false, true}}, `{"ebook":true,"audiobook":true}`},
		{"audiobook then ebook", [][2]bool{{false, true}, {true, false}}, `{"ebook":true,"audiobook":true}`},
		{"preserve existing flags", [][2]bool{{true, true}, {false, false}}, `{"ebook":true,"audiobook":true}`},
	} {
		t.Run(tt.name, func(t *testing.T) {
			dir := filepath.Join(t.TempDir(), "book")
			for _, update := range tt.updates {
				// New instances must preserve status saved by previous instances.
				manager := MetadataManager{Dir: dir}
				if err := manager.Update(update[0], update[1]); err != nil {
					t.Fatal(err)
				}
			}
			data, err := os.ReadFile(filepath.Join(dir, ".cloudlibrary.json"))
			if err != nil {
				t.Fatal(err)
			}
			if string(data) != tt.want {
				t.Fatalf("metadata = %s, want %s", data, tt.want)
			}
		})
	}
}

func TestMetadataManagerPreservesInvalidFile(t *testing.T) {
	dir := t.TempDir()
	filename := filepath.Join(dir, ".cloudlibrary.json")
	const invalid = `{"ebook":`
	if err := os.WriteFile(filename, []byte(invalid), 0600); err != nil {
		t.Fatal(err)
	}
	manager := MetadataManager{Dir: dir}
	if err := manager.Update(true, false); err == nil {
		t.Fatal("expected an error for invalid metadata")
	}
	data, err := os.ReadFile(filename)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != invalid {
		t.Fatalf("invalid metadata was overwritten: %s", data)
	}
}

func TestMetadataManagerReadError(t *testing.T) {
	dir := t.TempDir()
	if err := os.Mkdir(filepath.Join(dir, ".cloudlibrary.json"), 0755); err != nil {
		t.Fatal(err)
	}
	manager := MetadataManager{Dir: dir}
	if err := manager.Update(false, true); err == nil {
		t.Fatal("expected an error when metadata path is a directory")
	}
}

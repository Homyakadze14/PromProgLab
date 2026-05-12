package file

import (
	"context"
	"os"
	"playlist/internal/entity"
	"testing"
	"time"
)

func TestSaveAndLoadPlaylist(t *testing.T) {
	storage := NewBase64PlaylistStorage()
	ctx := context.Background()
	filename := "test_playlist.tmp"
	defer os.Remove(filename)

	original := []entity.Track{
		{Title: "Track1", Duration: 180 * time.Second, Genre: "Rock", Rating: 4.5},
		{Title: "Track2", Duration: 240 * time.Second, Genre: "Pop", Rating: 3.8},
	}
	err := storage.SavePlaylist(ctx, original, filename)
	if err != nil {
		t.Fatalf("save failed: %v", err)
	}
	loaded, err := storage.LoadPlaylist(ctx, filename)
	if err != nil {
		t.Fatalf("load failed: %v", err)
	}
	if len(loaded) != len(original) {
		t.Errorf("length mismatch: got %d, want %d", len(loaded), len(original))
	}
	for i := range original {
		if original[i].Title != loaded[i].Title ||
			original[i].Duration != loaded[i].Duration ||
			original[i].Genre != loaded[i].Genre ||
			original[i].Rating != loaded[i].Rating {
			t.Errorf("track %d mismatch: %+v vs %+v", i, original[i], loaded[i])
		}
	}
}

func TestLoadNonExistentFile(t *testing.T) {
	storage := NewBase64PlaylistStorage()
	ctx := context.Background()
	_, err := storage.LoadPlaylist(ctx, "does_not_exist.txt")
	if err == nil {
		t.Error("expected error on missing file")
	}
}

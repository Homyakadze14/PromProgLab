package file

import (
	"bufio"
	"context"
	"encoding/base64"
	"fmt"
	"os"
	"playlist/internal/entity"
	"strings"
	"time"
)

type Base64PlaylistStorage struct{}

func NewBase64PlaylistStorage() *Base64PlaylistStorage {
	return &Base64PlaylistStorage{}
}

func (s *Base64PlaylistStorage) SavePlaylist(ctx context.Context, tracks []entity.Track, filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("create file: %w", err)
	}
	defer file.Close()
	writer := bufio.NewWriter(file)
	for _, t := range tracks {
		line := fmt.Sprintf("%s|%d|%s|%.1f\n", t.Title, int(t.Duration.Seconds()), t.Genre, t.Rating)
		encoded := base64.StdEncoding.EncodeToString([]byte(line))
		if _, err := writer.WriteString(encoded + "\n"); err != nil {
			return fmt.Errorf("write line: %w", err)
		}
	}
	return writer.Flush()
}

func (s *Base64PlaylistStorage) LoadPlaylist(ctx context.Context, filename string) ([]entity.Track, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("open file: %w", err)
	}
	defer file.Close()
	var tracks []entity.Track
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		encoded := strings.TrimSpace(scanner.Text())
		if encoded == "" {
			continue
		}
		decoded, err := base64.StdEncoding.DecodeString(encoded)
		if err != nil {
			return nil, fmt.Errorf("base64 decode: %w", err)
		}
		parts := strings.Split(string(decoded), "|")
		if len(parts) != 4 {
			return nil, fmt.Errorf("invalid format: %s", decoded)
		}
		var durSec int
		var rating float32
		fmt.Sscanf(parts[1], "%d", &durSec)
		fmt.Sscanf(parts[3], "%f", &rating)
		tracks = append(tracks, entity.Track{
			Title:    parts[0],
			Duration: time.Duration(durSec) * time.Second,
			Genre:    parts[2],
			Rating:   rating,
		})
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("scan file: %w", err)
	}
	return tracks, nil
}

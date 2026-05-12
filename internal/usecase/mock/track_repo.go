package mock

import (
	"context"
	"playlist/internal/entity"
	"sync"
	"time"
)

type MockTrackRepository struct {
	mu            sync.Mutex
	tracks        []entity.Track
	shuffleCalled bool
}

func NewMockTrackRepository() *MockTrackRepository {
	return &MockTrackRepository{
		tracks: make([]entity.Track, 0),
	}
}

func (m *MockTrackRepository) Create(ctx context.Context, track entity.Track) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.tracks = append(m.tracks, track)
	return nil
}

func (m *MockTrackRepository) Delete(ctx context.Context, idx int) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if idx < 0 || idx >= len(m.tracks) {
		return nil
	}
	m.tracks = append(m.tracks[:idx], m.tracks[idx+1:]...)
	return nil
}

func (m *MockTrackRepository) GetAll(ctx context.Context) ([]entity.Track, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	result := make([]entity.Track, len(m.tracks))
	copy(result, m.tracks)
	return result, nil
}

func (m *MockTrackRepository) GetByGenre(ctx context.Context, genre string) ([]entity.Track, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	var res []entity.Track
	for _, t := range m.tracks {
		if t.Genre == genre {
			res = append(res, t)
		}
	}
	return res, nil
}

func (m *MockTrackRepository) GetByRating(ctx context.Context, minRating float32) ([]entity.Track, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	var res []entity.Track
	for _, t := range m.tracks {
		if t.Rating >= minRating {
			res = append(res, t)
		}
	}
	return res, nil
}

func (m *MockTrackRepository) GetByDurationRange(ctx context.Context, minDur, maxDur time.Duration) ([]entity.Track, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	var res []entity.Track
	for _, t := range m.tracks {
		if t.Duration >= minDur && t.Duration <= maxDur {
			res = append(res, t)
		}
	}
	return res, nil
}

func (m *MockTrackRepository) Shuffle(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.shuffleCalled = true
	return nil
}

func (m *MockTrackRepository) GetLen(ctx context.Context) (int, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	return len(m.tracks), nil
}

func (m *MockTrackRepository) ReplaceAll(ctx context.Context, tracks []entity.Track) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.tracks = make([]entity.Track, len(tracks))
	copy(m.tracks, tracks)
	return nil
}

func (m *MockTrackRepository) GetTracks() []entity.Track {
	m.mu.Lock()
	defer m.mu.Unlock()
	result := make([]entity.Track, len(m.tracks))
	copy(result, m.tracks)
	return result
}

func (m *MockTrackRepository) IsShuffleCalled() bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.shuffleCalled
}

package mock

import (
	"context"
	"playlist/internal/entity"
	"sync"
)

type MockFileStorage struct {
	mu           sync.Mutex
	saveCalled   bool
	loadCalled   bool
	loadResult   []entity.Track
	loadErr      error
	saveErr      error
	lastFilename string
}

func NewMockFileStorage() *MockFileStorage {
	return &MockFileStorage{}
}

func (m *MockFileStorage) SavePlaylist(ctx context.Context, tracks []entity.Track, filename string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.saveCalled = true
	m.lastFilename = filename
	if m.saveErr != nil {
		return m.saveErr
	}
	return nil
}

func (m *MockFileStorage) LoadPlaylist(ctx context.Context, filename string) ([]entity.Track, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.loadCalled = true
	if m.loadErr != nil {
		return nil, m.loadErr
	}
	return m.loadResult, nil
}

func (m *MockFileStorage) IsSaveCalled() bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.saveCalled
}

func (m *MockFileStorage) IsLoadCalled() bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.loadCalled
}

func (m *MockFileStorage) GetLastFilename() string {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.lastFilename
}

func (m *MockFileStorage) SetLoadResult(tracks []entity.Track) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.loadResult = tracks
}

func (m *MockFileStorage) SetLoadError(err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.loadErr = err
}

func (m *MockFileStorage) SetSaveError(err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.saveErr = err
}

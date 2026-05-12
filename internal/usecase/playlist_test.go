package usecase

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"playlist/internal/entity"
	"playlist/internal/usecase/mock"
	"testing"
	"time"
)

func newTestService(repo *mock.MockTrackRepository, fs *mock.MockFileStorage) *PlaylistService {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	return NewPlaylistService(logger, repo, fs)
}

func TestAddTrack(t *testing.T) {
	repo := mock.NewMockTrackRepository()
	fs := mock.NewMockFileStorage()
	service := newTestService(repo, fs)
	ctx := context.Background()

	track := entity.Track{Title: "Test", Duration: 200 * time.Second, Genre: "Rock", Rating: 4.5}
	err := service.AddTrack(ctx, track)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	tracks := repo.GetTracks()
	if len(tracks) != 1 || tracks[0].Title != "Test" {
		t.Errorf("track not added correctly, got %+v", tracks)
	}
}

func TestRemoveTrack(t *testing.T) {
	repo := mock.NewMockTrackRepository()
	fs := mock.NewMockFileStorage()
	service := newTestService(repo, fs)
	ctx := context.Background()

	service.AddTrack(ctx, entity.Track{Title: "A"})
	service.AddTrack(ctx, entity.Track{Title: "B"})

	err := service.RemoveTrack(ctx, 0)
	if err != nil {
		t.Fatalf("remove error: %v", err)
	}
	tracks := repo.GetTracks()
	if len(tracks) != 1 || tracks[0].Title != "B" {
		t.Errorf("wrong track after removal, got %+v", tracks)
	}
}

func TestNextTrack_OffMode(t *testing.T) {
	repo := mock.NewMockTrackRepository()
	fs := mock.NewMockFileStorage()
	service := newTestService(repo, fs)
	ctx := context.Background()

	service.AddTrack(ctx, entity.Track{Title: "A"})
	service.AddTrack(ctx, entity.Track{Title: "B"})
	service.SetRepeatMode("off")

	track, err := service.NextTrack(ctx)
	if err != nil || track.Title != "A" {
		t.Errorf("first NextTrack: got %v, err %v", track, err)
	}
	track, err = service.NextTrack(ctx)
	if err != nil || track.Title != "B" {
		t.Errorf("second NextTrack: got %v, err %v", track, err)
	}
	_, err = service.NextTrack(ctx)
	if err == nil {
		t.Error("expected error on third NextTrack (playlist ended)")
	}
}

func TestNextTrack_RepeatOne(t *testing.T) {
	repo := mock.NewMockTrackRepository()
	fs := mock.NewMockFileStorage()
	service := newTestService(repo, fs)
	ctx := context.Background()

	service.AddTrack(ctx, entity.Track{Title: "X"})
	service.AddTrack(ctx, entity.Track{Title: "Y"})
	service.SetRepeatMode("one")

	track, _ := service.NextTrack(ctx)
	if track.Title != "X" {
		t.Errorf("first NextTrack: expected X, got %s", track.Title)
	}
	track, _ = service.NextTrack(ctx)
	if track.Title != "X" {
		t.Errorf("repeat one: expected X again, got %s", track.Title)
	}
}

func TestNextTrack_RepeatAll(t *testing.T) {
	repo := mock.NewMockTrackRepository()
	fs := mock.NewMockFileStorage()
	service := newTestService(repo, fs)
	ctx := context.Background()

	service.AddTrack(ctx, entity.Track{Title: "A"})
	service.AddTrack(ctx, entity.Track{Title: "B"})
	service.SetRepeatMode("all")

	track, _ := service.NextTrack(ctx)
	if track.Title != "A" {
		t.Errorf("first: expected A, got %s", track.Title)
	}
	track, _ = service.NextTrack(ctx)
	if track.Title != "B" {
		t.Errorf("second: expected B, got %s", track.Title)
	}
	track, _ = service.NextTrack(ctx)
	if track.Title != "A" {
		t.Errorf("third (wrap): expected A, got %s", track.Title)
	}
}

func TestShuffleResetsCurrentTrack(t *testing.T) {
	repo := mock.NewMockTrackRepository()
	fs := mock.NewMockFileStorage()
	service := newTestService(repo, fs)
	ctx := context.Background()

	service.AddTrack(ctx, entity.Track{Title: "1"})
	service.AddTrack(ctx, entity.Track{Title: "2"})
	service.NextTrack(ctx)
	service.Shuffle(ctx)

	track, _ := service.NextTrack(ctx)
	if track == nil {
		t.Error("next track after shuffle returned nil")
	}
	if !repo.IsShuffleCalled() {
		t.Error("Shuffle did not call repo.Shuffle")
	}
}

func TestGetReport(t *testing.T) {
	repo := mock.NewMockTrackRepository()
	fs := mock.NewMockFileStorage()
	service := newTestService(repo, fs)
	ctx := context.Background()

	service.AddTrack(ctx, entity.Track{Duration: 60 * time.Second, Rating: 4.0})
	service.AddTrack(ctx, entity.Track{Duration: 120 * time.Second, Rating: 5.0})
	report, err := service.GetReport(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if report.CountTracks != 2 {
		t.Errorf("CountTracks = %d, want 2", report.CountTracks)
	}
	if report.CommonDuration != 180*time.Second {
		t.Errorf("CommonDuration = %v, want 3m", report.CommonDuration)
	}
	if report.AvgRating != 4.5 {
		t.Errorf("AvgRating = %v, want 4.5", report.AvgRating)
	}
}

func TestFilterByGenre(t *testing.T) {
	repo := mock.NewMockTrackRepository()
	fs := mock.NewMockFileStorage()
	service := newTestService(repo, fs)
	ctx := context.Background()

	service.AddTrack(ctx, entity.Track{Title: "R1", Genre: "Rock"})
	service.AddTrack(ctx, entity.Track{Title: "P1", Genre: "Pop"})
	res, err := service.FilterByGenre(ctx, "Rock")
	if err != nil || len(res) != 1 || res[0].Title != "R1" {
		t.Errorf("filter by Rock failed, got %+v", res)
	}
}

func TestFilterByMinRating(t *testing.T) {
	repo := mock.NewMockTrackRepository()
	fs := mock.NewMockFileStorage()
	service := newTestService(repo, fs)
	ctx := context.Background()

	service.AddTrack(ctx, entity.Track{Title: "Low", Rating: 3.0})
	service.AddTrack(ctx, entity.Track{Title: "High", Rating: 4.8})
	res, err := service.FilterByMinRating(ctx, 4.5)
	if err != nil || len(res) != 1 || res[0].Title != "High" {
		t.Errorf("filter by rating failed, got %+v", res)
	}
}

func TestGetTracksInDurationRange(t *testing.T) {
	repo := mock.NewMockTrackRepository()
	fs := mock.NewMockFileStorage()
	service := newTestService(repo, fs)
	ctx := context.Background()

	service.AddTrack(ctx, entity.Track{Title: "Short", Duration: 100 * time.Second})
	service.AddTrack(ctx, entity.Track{Title: "Long", Duration: 250 * time.Second})
	res, err := service.GetTracksInDurationRange(ctx, 150, 350)
	if err != nil || len(res) != 1 || res[0].Title != "Long" {
		t.Errorf("duration range failed, got %+v", res)
	}
}

func TestSaveToFile(t *testing.T) {
	repo := mock.NewMockTrackRepository()
	fs := mock.NewMockFileStorage()
	service := newTestService(repo, fs)
	ctx := context.Background()

	service.AddTrack(ctx, entity.Track{Title: "SaveMe"})
	err := service.SaveToFile(ctx, "test.txt")
	if err != nil {
		t.Errorf("SaveToFile error: %v", err)
	}
	if !fs.IsSaveCalled() || fs.GetLastFilename() != "test.txt" {
		t.Error("SaveToFile did not call underlying storage")
	}
}

func TestLoadFromFile(t *testing.T) {
	repo := mock.NewMockTrackRepository()
	fs := mock.NewMockFileStorage()
	fs.SetLoadResult([]entity.Track{{Title: "LoadedTrack"}})
	service := newTestService(repo, fs)
	ctx := context.Background()

	err := service.LoadFromFile(ctx, "load.txt")
	if err != nil {
		t.Fatal(err)
	}
	if !fs.IsLoadCalled() {
		t.Error("LoadFromFile did not call storage")
	}
	tracks, _ := repo.GetAll(ctx)
	if len(tracks) != 1 || tracks[0].Title != "LoadedTrack" {
		t.Errorf("after load, repo has %+v", tracks)
	}
	service.mu.Lock()
	curr := service.currTrack
	mode := service.mode
	service.mu.Unlock()
	if curr != 0 || mode != "off" {
		t.Errorf("state not reset after load: currTrack=%d, mode=%s", curr, mode)
	}
}

func TestSetRepeatModeInvalid(t *testing.T) {
	repo := mock.NewMockTrackRepository()
	fs := mock.NewMockFileStorage()
	service := newTestService(repo, fs)
	err := service.SetRepeatMode("invalid")
	if err == nil {
		t.Error("expected error for invalid mode")
	}
}

func TestRemoveTrack_IndexLessThanCurrTrack(t *testing.T) {
	repo := mock.NewMockTrackRepository()
	fs := mock.NewMockFileStorage()
	service := newTestService(repo, fs)
	ctx := context.Background()
	service.AddTrack(ctx, entity.Track{Title: "A"})
	service.AddTrack(ctx, entity.Track{Title: "B"})
	service.AddTrack(ctx, entity.Track{Title: "C"})
	service.NextTrack(ctx)
	service.NextTrack(ctx)
	err := service.RemoveTrack(ctx, 1)
	if err != nil {
		t.Fatal(err)
	}
	tracks, _ := repo.GetAll(ctx)
	if len(tracks) != 2 {
		t.Fatalf("expected 2 tracks, got %d", len(tracks))
	}
	track, err := service.NextTrack(ctx)
	if err != nil || track.Title != "C" {
		t.Errorf("expected track C, got %v, err %v", track, err)
	}
}

func TestSetRepeatMode_Valid(t *testing.T) {
	repo := mock.NewMockTrackRepository()
	fs := mock.NewMockFileStorage()
	service := newTestService(repo, fs)
	err := service.SetRepeatMode("all")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	ctx := context.Background()
	service.AddTrack(ctx, entity.Track{Title: "X"})
	service.AddTrack(ctx, entity.Track{Title: "Y"})
	t1, _ := service.NextTrack(ctx)
	t2, _ := service.NextTrack(ctx)
	t3, _ := service.NextTrack(ctx)
	if t1.Title != "X" || t2.Title != "Y" || t3.Title != "X" {
		t.Errorf("repeat all not working: got %s, %s, %s", t1.Title, t2.Title, t3.Title)
	}
}

func TestSaveToFile_Error(t *testing.T) {
	repo := mock.NewMockTrackRepository()
	fs := mock.NewMockFileStorage()
	fs.SetSaveError(fmt.Errorf("disk full"))
	service := newTestService(repo, fs)
	ctx := context.Background()
	service.AddTrack(ctx, entity.Track{Title: "T"})
	err := service.SaveToFile(ctx, "test.txt")
	if err == nil {
		t.Error("expected error, got nil")
	}
}

func TestLoadFromFile_Error(t *testing.T) {
	repo := mock.NewMockTrackRepository()
	fs := mock.NewMockFileStorage()
	fs.SetLoadError(fmt.Errorf("file not found"))
	service := newTestService(repo, fs)
	ctx := context.Background()
	err := service.LoadFromFile(ctx, "missing.txt")
	if err == nil {
		t.Error("expected error, got nil")
	}
}

func TestGetAllTracks(t *testing.T) {
	repo := mock.NewMockTrackRepository()
	fs := mock.NewMockFileStorage()
	service := newTestService(repo, fs)
	ctx := context.Background()
	service.AddTrack(ctx, entity.Track{Title: "A"})
	service.AddTrack(ctx, entity.Track{Title: "B"})
	tracks, err := service.GetAllTracks(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(tracks) != 2 || tracks[0].Title != "A" || tracks[1].Title != "B" {
		t.Errorf("unexpected tracks: %+v", tracks)
	}
}

func TestGetReport_EmptyPlaylist(t *testing.T) {
	repo := mock.NewMockTrackRepository()
	fs := mock.NewMockFileStorage()
	service := newTestService(repo, fs)
	ctx := context.Background()
	report, err := service.GetReport(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if report.CountTracks != 0 || report.CommonDuration != 0 || report.AvgRating != 0 {
		t.Errorf("empty report wrong: %+v", report)
	}
}

func TestNextTrack_EmptyPlaylist(t *testing.T) {
	repo := mock.NewMockTrackRepository()
	fs := mock.NewMockFileStorage()
	service := newTestService(repo, fs)
	ctx := context.Background()
	_, err := service.NextTrack(ctx)
	if err == nil {
		t.Error("expected error on empty playlist")
	}
}

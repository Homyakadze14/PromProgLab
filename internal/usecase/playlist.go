package usecase

import (
	"context"
	"fmt"
	"log/slog"
	"playlist/internal/entity"
	"sync"
	"time"
)

type TrackRepository interface {
	Create(ctx context.Context, track entity.Track) error
	Delete(ctx context.Context, idx int) error
	GetAll(ctx context.Context) ([]entity.Track, error)
	GetByGenre(ctx context.Context, genre string) ([]entity.Track, error)
	GetByRating(ctx context.Context, minRating float32) ([]entity.Track, error)
	GetByDurationRange(ctx context.Context, minDur, maxDur time.Duration) ([]entity.Track, error)
	Shuffle(ctx context.Context) error
	GetLen(ctx context.Context) (int, error)
	ReplaceAll(ctx context.Context, tracks []entity.Track) error
}

type FileStorage interface {
	SavePlaylist(ctx context.Context, tracks []entity.Track, filename string) error
	LoadPlaylist(ctx context.Context, filename string) ([]entity.Track, error)
}

type PlaylistService struct {
	repo      TrackRepository
	fs        FileStorage
	mode      string
	currTrack int
	mu        sync.Mutex
	log       *slog.Logger
}

func NewPlaylistService(log *slog.Logger, repo TrackRepository, fs FileStorage) *PlaylistService {
	return &PlaylistService{
		repo:      repo,
		log:       log,
		mode:      "off",
		fs:        fs,
		currTrack: 0,
	}
}

func (s *PlaylistService) AddTrack(ctx context.Context, track entity.Track) error {
	const op = "usecase.PlaylistService.AddTrack"
	log := s.log.With(slog.String("op", op), slog.String("title", track.Title))
	if err := s.repo.Create(ctx, track); err != nil {
		log.Error("failed to add track", slog.Any("error", err))
		return fmt.Errorf("%s: %w", op, err)
	}
	log.Info("track added")
	return nil
}

func (s *PlaylistService) RemoveTrack(ctx context.Context, idx int) error {
	const op = "usecase.PlaylistService.RemoveTrack"
	log := s.log.With(slog.String("op", op), slog.Int("index", idx))
	if err := s.repo.Delete(ctx, idx); err != nil {
		log.Error("failed to remove track", slog.Any("error", err))
		return fmt.Errorf("%s: %w", op, err)
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	lenTracks, err := s.repo.GetLen(ctx)
	if err != nil {
		log.Warn("failed to get playlist length after removal", slog.Any("error", err))
		return nil
	}
	if lenTracks == 0 {
		s.currTrack = 0
	} else if idx < s.currTrack && s.currTrack > 0 {
		s.currTrack--
	}
	log.Info("track removed", slog.Int("new_length", lenTracks))
	return nil
}

func (s *PlaylistService) GetAllTracks(ctx context.Context) ([]entity.Track, error) {
	const op = "usecase.PlaylistService.GetAllTracks"
	log := s.log.With(slog.String("op", op))
	tracks, err := s.repo.GetAll(ctx)
	if err != nil {
		log.Error("failed to get all tracks", slog.Any("error", err))
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	log.Debug("tracks retrieved", slog.Int("count", len(tracks)))
	return tracks, nil
}

func (s *PlaylistService) SetRepeatMode(mode string) error {
	const op = "usecase.PlaylistService.SetRepeatMode"
	log := s.log.With(slog.String("op", op), slog.String("new_mode", mode))
	if mode != "off" && mode != "all" && mode != "one" {
		log.Warn("invalid repeat mode")
		return fmt.Errorf("%s: wrong mode '%s'", op, mode)
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	oldMode := s.mode
	s.mode = mode
	log.Info("repeat mode changed", slog.String("old_mode", oldMode))
	return nil
}

func (s *PlaylistService) NextTrack(ctx context.Context) (*entity.Track, error) {
	const op = "usecase.PlaylistService.NextTrack"
	log := s.log.With(slog.String("op", op))
	tracks, err := s.repo.GetAll(ctx)
	if err != nil {
		log.Error("failed to get tracks", slog.Any("error", err))
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	if len(tracks) == 0 {
		log.Warn("next track called on empty playlist")
		return nil, fmt.Errorf("%s: playlist is empty", op)
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	log = log.With(slog.Int("current_position", s.currTrack), slog.String("repeat_mode", s.mode))
	switch s.mode {
	case "one":
		if s.currTrack >= len(tracks) {
			s.currTrack = 0
		}
		track := tracks[s.currTrack]
		log.Debug("next track (repeat one)", slog.String("title", track.Title))
		return &track, nil
	case "all":
		if s.currTrack >= len(tracks) {
			s.currTrack = 0
		}
		track := tracks[s.currTrack]
		s.currTrack = (s.currTrack + 1) % len(tracks)
		log.Debug("next track (repeat all)", slog.String("title", track.Title))
		return &track, nil
	default:
		if s.currTrack >= len(tracks) {
			log.Warn("playlist ended")
			return nil, fmt.Errorf("%s: playlist ended", op)
		}
		track := tracks[s.currTrack]
		s.currTrack++
		log.Debug("next track (off)", slog.String("title", track.Title))
		return &track, nil
	}
}

func (s *PlaylistService) GetReport(ctx context.Context) (entity.Report, error) {
	const op = "usecase.PlaylistService.GetReport"
	log := s.log.With(slog.String("op", op))
	tracks, err := s.repo.GetAll(ctx)
	if err != nil {
		log.Error("failed to get tracks for report", slog.Any("error", err))
		return entity.Report{}, fmt.Errorf("%s: %w", op, err)
	}
	if len(tracks) == 0 {
		return entity.Report{CountTracks: 0, AvgRating: 0, CommonDuration: 0}, nil
	}
	var totalDuration time.Duration
	var ratingSum float32
	for _, t := range tracks {
		totalDuration += t.Duration
		ratingSum += t.Rating
	}
	avgRating := ratingSum / float32(len(tracks))
	log.Info("report generated", slog.Int("track_count", len(tracks)), slog.Duration("total_duration", totalDuration), slog.Float64("avg_rating", float64(avgRating)))
	return entity.Report{
		CountTracks:    len(tracks),
		CommonDuration: totalDuration,
		AvgRating:      avgRating,
	}, nil
}

func (s *PlaylistService) FilterByGenre(ctx context.Context, genre string) ([]entity.Track, error) {
	const op = "usecase.PlaylistService.FilterByGenre"
	log := s.log.With(slog.String("op", op), slog.String("genre", genre))
	tracks, err := s.repo.GetByGenre(ctx, genre)
	if err != nil {
		log.Error("failed to filter by genre", slog.Any("error", err))
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	log.Info("filter by genre", slog.Int("count", len(tracks)))
	return tracks, nil
}

func (s *PlaylistService) FilterByMinRating(ctx context.Context, minRating float32) ([]entity.Track, error) {
	const op = "usecase.PlaylistService.FilterByMinRating"
	log := s.log.With(slog.String("op", op), slog.Float64("min_rating", float64(minRating)))
	tracks, err := s.repo.GetByRating(ctx, minRating)
	if err != nil {
		log.Error("failed to filter by rating", slog.Any("error", err))
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	log.Info("filter by min rating", slog.Int("count", len(tracks)))
	return tracks, nil
}

func (s *PlaylistService) GetTracksInDurationRange(ctx context.Context, minSec, maxSec int) ([]entity.Track, error) {
	const op = "usecase.PlaylistService.GetTracksInDurationRange"
	minDur := time.Duration(minSec) * time.Second
	maxDur := time.Duration(maxSec) * time.Second
	log := s.log.With(slog.String("op", op), slog.Duration("min", minDur), slog.Duration("max", maxDur))
	tracks, err := s.repo.GetByDurationRange(ctx, minDur, maxDur)
	if err != nil {
		log.Error("failed to filter by duration", slog.Any("error", err))
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	log.Info("filter by duration", slog.Int("count", len(tracks)))
	return tracks, nil
}

func (s *PlaylistService) Shuffle(ctx context.Context) error {
	const op = "usecase.PlaylistService.Shuffle"
	log := s.log.With(slog.String("op", op))
	if err := s.repo.Shuffle(ctx); err != nil {
		log.Error("failed to shuffle", slog.Any("error", err))
		return err
	}
	s.mu.Lock()
	s.currTrack = 0
	s.mu.Unlock()
	log.Info("playlist shuffled")
	return nil
}

func (s *PlaylistService) SaveToFile(ctx context.Context, filename string) error {
	const op = "usecase.PlaylistService.SaveToFile"
	log := s.log.With(slog.String("op", op), slog.String("filename", filename))
	tracks, err := s.repo.GetAll(ctx)
	if err != nil {
		log.Error("failed to get tracks for save", slog.Any("error", err))
		return fmt.Errorf("%s: %w", op, err)
	}
	if err := s.fs.SavePlaylist(ctx, tracks, filename); err != nil {
		log.Error("failed to save playlist", slog.Any("error", err))
		return fmt.Errorf("%s: %w", op, err)
	}
	log.Info("playlist saved", slog.Int("track_count", len(tracks)))
	return nil
}

func (s *PlaylistService) LoadFromFile(ctx context.Context, filename string) error {
	const op = "usecase.PlaylistService.LoadFromFile"
	log := s.log.With(slog.String("op", op), slog.String("filename", filename))
	tracks, err := s.fs.LoadPlaylist(ctx, filename)
	if err != nil {
		log.Error("failed to load playlist", slog.Any("error", err))
		return fmt.Errorf("%s: %w", op, err)
	}
	if err := s.repo.ReplaceAll(ctx, tracks); err != nil {
		log.Error("failed to replace tracks", slog.Any("error", err))
		return fmt.Errorf("%s: %w", op, err)
	}
	s.mu.Lock()
	s.currTrack = 0
	s.mode = "off"
	s.mu.Unlock()
	log.Info("playlist loaded", slog.Int("track_count", len(tracks)))
	return nil
}

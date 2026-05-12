package array

import (
	"context"
	"fmt"
	"math/rand/v2"
	"playlist/internal/entity"
	"sync"
	"time"
)

const defaultCapacity = 10

type TrackArrRepository struct {
	arr []entity.Track
	mu  sync.Mutex
}

func NewTrackArrRepositoryRepo() *TrackArrRepository {
	return &TrackArrRepository{
		arr: make([]entity.Track, 0, defaultCapacity),
	}
}

func (r *TrackArrRepository) Create(ctx context.Context, track entity.Track) error {
	const op = "infra.array.TrackArrRepository.Create"

	r.mu.Lock()
	defer r.mu.Unlock()

	r.arr = append(r.arr, track)

	return nil
}

func (r *TrackArrRepository) Delete(ctx context.Context, idx int) error {
	const op = "infra.array.TrackArrRepository.Delete"

	r.mu.Lock()
	defer r.mu.Unlock()

	if idx < 0 || idx >= len(r.arr) {
		return fmt.Errorf("%s: wrong index", op)
	}

	r.arr = append(r.arr[:idx], r.arr[idx+1:]...)

	return nil
}

func (r *TrackArrRepository) GetLen(ctx context.Context) (int, error) {
	const op = "infra.array.TrackArrRepository.GetLen"

	r.mu.Lock()
	defer r.mu.Unlock()

	return len(r.arr), nil
}

func (r *TrackArrRepository) GetAll(ctx context.Context) ([]entity.Track, error) {
	const op = "infra.array.TrackArrRepository.GetAll"

	r.mu.Lock()
	defer r.mu.Unlock()

	result := make([]entity.Track, len(r.arr))
	copy(result, r.arr)
	return result, nil
}

// функция получения ограниченого количества треков
// возвращает ошибку, если указан неправильный лимит
func (r *TrackArrRepository) GetAllWithLimit(ctx context.Context, limit int) ([]entity.Track, error) {
	const op = "infra.array.TrackArrRepository.GetAllWithLimit"

	r.mu.Lock()
	defer r.mu.Unlock()

	if limit < 0 || limit > len(r.arr) {
		return nil, fmt.Errorf("limit must be between 0 and %d", len(r.arr))
	}

	result := make([]entity.Track, limit)
	copy(result, r.arr[:limit])
	return result, nil
}

func (r *TrackArrRepository) GetByGenre(ctx context.Context, genre string) ([]entity.Track, error) {
	const op = "infra.array.TrackArrRepository.GetByGenre"

	r.mu.Lock()
	defer r.mu.Unlock()
	result := make([]entity.Track, 0, defaultCapacity)
	for _, t := range r.arr {
		if t.Genre == genre {
			result = append(result, t)
		}
	}
	return result, nil
}

func (r *TrackArrRepository) GetByRating(ctx context.Context, rating float32) ([]entity.Track, error) {
	const op = "infra.array.TrackArrRepository.GetByRating"

	r.mu.Lock()
	defer r.mu.Unlock()
	result := make([]entity.Track, 0, defaultCapacity)
	for _, t := range r.arr {
		if t.Rating >= rating {
			result = append(result, t)
		}
	}
	return result, nil
}

func (r *TrackArrRepository) GetByDurationRange(ctx context.Context, minDur, maxDur time.Duration) ([]entity.Track, error) {
	const op = "infra.array.TrackArrRepository.GetByDurationRange"

	r.mu.Lock()
	defer r.mu.Unlock()

	result := make([]entity.Track, 0, defaultCapacity)
	for _, track := range r.arr {
		if track.Duration >= minDur && track.Duration <= maxDur {
			result = append(result, track)
		}
	}
	return result, nil
}

func (r *TrackArrRepository) Shuffle(ctx context.Context) error {
	const op = "infra.array.TrackArrRepository.Shuffle"

	r.mu.Lock()
	defer r.mu.Unlock()

	rand.Shuffle(len(r.arr), func(i, j int) {
		r.arr[i], r.arr[j] = r.arr[j], r.arr[i]
	})

	return nil
}

func (r *TrackArrRepository) ReplaceAll(ctx context.Context, tracks []entity.Track) error {
	const op = "infra.array.TrackArrRepository.ReplaceAll"
	r.mu.Lock()
	defer r.mu.Unlock()
	r.arr = make([]entity.Track, len(tracks))
	copy(r.arr, tracks)
	return nil
}

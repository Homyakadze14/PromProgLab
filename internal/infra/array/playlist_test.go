package array

import (
	"context"
	"playlist/internal/entity"
	"testing"
	"time"
)

func TestCreateAndGetAll(t *testing.T) {
	repo := NewTrackArrRepositoryRepo()
	ctx := context.Background()
	track := entity.Track{Title: "Song", Duration: 3 * time.Minute, Genre: "Pop", Rating: 4.2}

	err := repo.Create(ctx, track)
	if err != nil {
		t.Fatal(err)
	}
	all, _ := repo.GetAll(ctx)
	if len(all) != 1 || all[0].Title != "Song" {
		t.Errorf("expected 1 track, got %+v", all)
	}
}

func TestDelete(t *testing.T) {
	repo := NewTrackArrRepositoryRepo()
	ctx := context.Background()
	repo.Create(ctx, entity.Track{Title: "A"})
	repo.Create(ctx, entity.Track{Title: "B"})

	err := repo.Delete(ctx, 0)
	if err != nil {
		t.Fatal(err)
	}
	all, _ := repo.GetAll(ctx)
	if len(all) != 1 || all[0].Title != "B" {
		t.Errorf("after delete: %+v", all)
	}

	err = repo.Delete(ctx, 5)
	if err == nil {
		t.Error("expected error on wrong index")
	}
}

func TestGetLen(t *testing.T) {
	repo := NewTrackArrRepositoryRepo()
	ctx := context.Background()
	repo.Create(ctx, entity.Track{})
	repo.Create(ctx, entity.Track{})
	l, _ := repo.GetLen(ctx)
	if l != 2 {
		t.Errorf("GetLen = %d, want 2", l)
	}
}

func TestGetByGenre(t *testing.T) {
	repo := NewTrackArrRepositoryRepo()
	ctx := context.Background()
	repo.Create(ctx, entity.Track{Title: "Rock1", Genre: "Rock"})
	repo.Create(ctx, entity.Track{Title: "Rock2", Genre: "Rock"})
	repo.Create(ctx, entity.Track{Title: "Pop1", Genre: "Pop"})

	res, err := repo.GetByGenre(ctx, "Rock")
	if err != nil || len(res) != 2 {
		t.Errorf("expected 2 rock tracks, got %d", len(res))
	}
}

func TestGetByRating(t *testing.T) {
	repo := NewTrackArrRepositoryRepo()
	ctx := context.Background()
	repo.Create(ctx, entity.Track{Rating: 3.5})
	repo.Create(ctx, entity.Track{Rating: 4.9})
	repo.Create(ctx, entity.Track{Rating: 2.0})

	res, err := repo.GetByRating(ctx, 4.0)
	if err != nil || len(res) != 1 || res[0].Rating != 4.9 {
		t.Errorf("expected 1 track with rating>=4.0, got %d", len(res))
	}
}

func TestGetByDurationRange(t *testing.T) {
	repo := NewTrackArrRepositoryRepo()
	ctx := context.Background()
	repo.Create(ctx, entity.Track{Duration: 100 * time.Second})
	repo.Create(ctx, entity.Track{Duration: 250 * time.Second})
	repo.Create(ctx, entity.Track{Duration: 400 * time.Second})

	res, err := repo.GetByDurationRange(ctx, 150*time.Second, 300*time.Second)
	if err != nil || len(res) != 1 || res[0].Duration != 250*time.Second {
		t.Errorf("duration range failed, got %d tracks", len(res))
	}
}

func TestShuffle(t *testing.T) {
	repo := NewTrackArrRepositoryRepo()
	ctx := context.Background()
	tracks := []entity.Track{
		{Title: "A"}, {Title: "B"}, {Title: "C"}, {Title: "D"},
	}
	for _, tr := range tracks {
		repo.Create(ctx, tr)
	}
	err := repo.Shuffle(ctx)
	if err != nil {
		t.Fatal(err)
	}
	all, _ := repo.GetAll(ctx)

	if len(all) != 4 {
		t.Errorf("after shuffle length changed to %d", len(all))
	}
}

func TestReplaceAll(t *testing.T) {
	repo := NewTrackArrRepositoryRepo()
	ctx := context.Background()
	newTracks := []entity.Track{{Title: "X"}, {Title: "Y"}}
	err := repo.ReplaceAll(ctx, newTracks)
	if err != nil {
		t.Fatal(err)
	}
	all, _ := repo.GetAll(ctx)
	if len(all) != 2 || all[0].Title != "X" {
		t.Errorf("ReplaceAll failed: %+v", all)
	}
}

package cmd

import (
	"context"
	"fmt"
	"log/slog"
	"playlist/internal/entity"
	"playlist/internal/usecase"
	"strings"
	"time"
)

type ConsoleController struct {
	service *usecase.PlaylistService
	log     *slog.Logger
}

func NewConsoleController(log *slog.Logger, service *usecase.PlaylistService) *ConsoleController {
	return &ConsoleController{
		service: service,
		log:     log,
	}
}

func (c *ConsoleController) Run() {
	ctx := context.Background()
	c.log.Info("=== Starting playlist demo ===")

	// 1. Добавление треков
	tracks := []entity.Track{
		{Title: "Bohemian Rhapsody", Duration: 355 * time.Second, Genre: "Rock", Rating: 4.9},
		{Title: "Shape of You", Duration: 233 * time.Second, Genre: "Pop", Rating: 4.5},
		{Title: "Stairway to Heaven", Duration: 482 * time.Second, Genre: "Rock", Rating: 5.0},
		{Title: "Blinding Lights", Duration: 200 * time.Second, Genre: "Pop", Rating: 4.2},
		{Title: "Symphony No.5", Duration: 450 * time.Second, Genre: "Classical", Rating: 4.8},
	}
	for _, track := range tracks {
		if err := c.service.AddTrack(ctx, track); err != nil {
			c.log.Error("failed to add track", "track", track.Title, "error", err)
		} else {
			c.log.Info("added track", "title", track.Title)
		}
	}

	// 2. Вывод плейлиста в виде таблицы
	c.printPlaylist(ctx, "Current playlist")

	// 3. Отчёт (общая длительность, средний рейтинг)
	report, err := c.service.GetReport(ctx)
	if err != nil {
		c.log.Error("failed to get report", "error", err)
	} else {
		c.log.Info("Report",
			"total_tracks", report.CountTracks,
			"total_duration", formatDuration(report.CommonDuration),
			"avg_rating", report.AvgRating,
		)
	}

	// 4. Фильтрация по жанру "Rock"
	rockTracks, err := c.service.FilterByGenre(ctx, "Rock")
	if err != nil {
		c.log.Error("filter by genre failed", "error", err)
	} else {
		c.log.Info("Filtered by genre 'Rock'", "count", len(rockTracks))
		for _, t := range rockTracks {
			c.log.Info(" - track", "title", t.Title, "rating", t.Rating)
		}
	}

	// 5. Список треков с рейтингом не ниже 4.7
	highRated, _ := c.service.FilterByMinRating(ctx, 4.7)
	c.log.Info("Tracks with rating >= 4.7", "count", len(highRated))
	for _, t := range highRated {
		c.log.Info(" - high rated track", "title", t.Title, "rating", t.Rating)
	}

	// 6. Поиск треков в диапазоне времени 3-5 минут (180-300 сек)
	rangeTracks, _ := c.service.GetTracksInDurationRange(ctx, 180, 300)
	c.log.Info("Tracks between 3 and 5 minutes", "count", len(rangeTracks))
	for _, t := range rangeTracks {
		c.log.Info(" - track in range", "title", t.Title, "duration_sec", t.Duration.Seconds())
	}

	// 7. Сохранение в Base64-файл
	filename := "playlist_base64.txt"
	if err := c.service.SaveToFile(ctx, filename); err != nil {
		c.log.Error("save to file failed", "error", err)
	} else {
		c.log.Info("Playlist saved to file", "filename", filename)
	}

	// 8. Загрузка из Base64-файла
	if err := c.service.LoadFromFile(ctx, filename); err != nil {
		c.log.Error("load from file failed", "error", err)
	} else {
		c.log.Info("Playlist loaded from file")
		c.printPlaylist(ctx, "After loading from file")
	}

	// 9. Перемешивание (shuffle)
	if err := c.service.Shuffle(ctx); err != nil {
		c.log.Error("shuffle failed", "error", err)
	} else {
		c.log.Info("Playlist shuffled")
		c.printPlaylist(ctx, "After shuffle")
	}

	// 10. Режим повтора "one" – несколько вызовов NextTrack
	c.service.SetRepeatMode("one")
	c.log.Info("Repeat mode set to 'one'")
	for i := 0; i < 3; i++ {
		track, err := c.service.NextTrack(ctx)
		if err != nil {
			c.log.Error("NextTrack failed", "error", err)
		} else {
			c.log.Info("NextTrack (repeat one)", "track", track.Title, "position", i+1)
		}
	}

	// 11. Режим повтора "all" – несколько шагов (зацикливание)
	c.service.SetRepeatMode("all")
	c.log.Info("Repeat mode set to 'all'")
	for i := 0; i < 6; i++ {
		track, err := c.service.NextTrack(ctx)
		if err != nil {
			c.log.Error("NextTrack failed", "error", err)
		} else {
			c.log.Info("NextTrack (repeat all)", "track", track.Title, "step", i+1)
		}
	}

	// 12. Удаление трека (индекс 2)
	if err := c.service.RemoveTrack(ctx, 2); err != nil {
		c.log.Error("remove track failed", "error", err)
	} else {
		c.log.Info("Removed track at index 2")
		c.printPlaylist(ctx, "After removal")
	}

	// 13. Сброс режима повтора на "off" и проход до конца
	c.service.SetRepeatMode("off")
	c.log.Info("Repeat mode set to 'off'")
	for {
		track, err := c.service.NextTrack(ctx)
		if err != nil {
			c.log.Info("Playlist ended", "reason", err.Error())
			break
		}
		c.log.Info("NextTrack (off)", "track", track.Title)
	}

	c.log.Info("=== Demo completed ===")
}

func (c *ConsoleController) printPlaylist(ctx context.Context, title string) {
	tracks, err := c.service.GetAllTracks(ctx)
	if err != nil {
		c.log.Error("failed to get tracks for table", "error", err)
		return
	}
	if len(tracks) == 0 {
		fmt.Println("\n", title, "[empty]")
		return
	}

	headers := []string{"#", "Title", "Duration", "Genre", "Rating"}
	rows := make([][]string, len(tracks))
	for i, t := range tracks {
		rows[i] = []string{
			fmt.Sprintf("%d", i),
			t.Title,
			formatDuration(t.Duration),
			t.Genre,
			fmt.Sprintf("%.1f", t.Rating),
		}
	}

	colWidths := make([]int, len(headers))
	for i, h := range headers {
		colWidths[i] = len(h)
	}
	for _, row := range rows {
		for i, cell := range row {
			if len(cell) > colWidths[i] {
				colWidths[i] = len(cell)
			}
		}
	}

	fmt.Printf("\n=== %s ===\n", title)
	separator := "+" + strings.Join(func() []string {
		parts := make([]string, len(colWidths))
		for i, w := range colWidths {
			parts[i] = strings.Repeat("-", w+2)
		}
		return parts
	}(), "+") + "+"

	fmt.Println(separator)
	fmt.Print("|")
	for i, h := range headers {
		fmt.Printf(" %-*s |", colWidths[i], h)
	}
	fmt.Println()
	fmt.Println(separator)
	for _, row := range rows {
		fmt.Print("|")
		for i, cell := range row {
			fmt.Printf(" %-*s |", colWidths[i], cell)
		}
		fmt.Println()
	}
	fmt.Println(separator)
}

func formatDuration(d time.Duration) string {
	totalSec := int(d.Seconds())
	minutes := totalSec / 60
	seconds := totalSec % 60
	return fmt.Sprintf("%02d:%02d", minutes, seconds)
}

func (c *ConsoleController) Stop() {
	c.log.Info("Console controller stopped")
}

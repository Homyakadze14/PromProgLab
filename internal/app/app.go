package app

import (
	"log/slog"
	"playlist/internal/controller/cmd"
	"playlist/internal/infra/array"
	"playlist/internal/infra/file"
	"playlist/internal/usecase"
)

type App struct {
	CMD *cmd.ConsoleController
}

func Run(
	log *slog.Logger,
) *App {
	// Repository
	fsRepo := file.NewBase64PlaylistStorage()
	trackRepo := array.NewTrackArrRepositoryRepo()

	// Usecases
	playlist := usecase.NewPlaylistService(log, trackRepo, fsRepo)

	// Controller
	cmdCtrl := cmd.NewConsoleController(log, playlist)
	go func() { cmdCtrl.Run() }()

	return &App{CMD: cmdCtrl}
}

func (s *App) Shutdown() {
	s.CMD.Stop()
	slog.Info("The application is closed")
}

package app

import (
	"miniflux-digest/internal/config"
)

type App struct {
	Config *config.Config
	ArchiveService ArchiveService
	EmailService EmailService
	MinifluxClientService MinifluxClientService
	DigestService DigestService
}

func NewApp(cfg *config.Config, client MinifluxClientService, archiveSvc ArchiveService, emailSvc EmailService, digestSvc DigestService) *App {
	return &App{
		Config: cfg,
		ArchiveService: archiveSvc,
		EmailService: emailSvc,
		MinifluxClientService: client,
		DigestService: digestSvc,
	}
}

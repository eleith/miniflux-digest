package app

import (
	"miniflux-digest/internal/config"
)

type App struct {
	Config *config.Config
	ArchiveService ArchiveService
	EmailService EmailService
	MinifluxClientService MinifluxClientService
	CategoryService CategoryService
}

func NewApp(cfg *config.Config, client MinifluxClientService, archiveSvc ArchiveService, emailSvc EmailService, categorySvc CategoryService) *App {
	return &App{
		Config: cfg,
		ArchiveService: archiveSvc,
		EmailService: emailSvc,
		MinifluxClientService: client,
		CategoryService: categorySvc,
	}
}

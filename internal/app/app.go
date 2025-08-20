package app

import (
	"miniflux-digest/internal/app/services"
	"miniflux-digest/internal/config"
)

type App struct {
	Config                *config.Config
	ArchiveService        services.ArchiveService
	EmailService          services.EmailService
	MinifluxClientService services.MinifluxClientService
	DigestService         services.DigestService
	LLMService            services.LLMService
}

type Option func(*App)

func NewApp(opts ...Option) *App {
	app := &App{}
	for _, opt := range opts {
		opt(app)
	}
	return app
}

func WithConfig(cfg *config.Config) Option {
	return func(a *App) {
		a.Config = cfg
	}
}

func WithArchiveService(s services.ArchiveService) Option {
	return func(a *App) {
		a.ArchiveService = s
	}
}

func WithEmailService(s services.EmailService) Option {
	return func(a *App) {
		a.EmailService = s
	}
}

func WithMinifluxClientService(s services.MinifluxClientService) Option {
	return func(a *App) {
		a.MinifluxClientService = s
	}
}

func WithDigestService(s services.DigestService) Option {
	return func(a *App) {
		a.DigestService = s
	}
}

func WithLLMService(s services.LLMService) Option { // Changed
	return func(a *App) {
		a.LLMService = s
	}
}

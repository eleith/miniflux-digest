package app

import (
	"miniflux-digest/internal/config"
	"miniflux-digest/internal/digest"
	"miniflux-digest/internal/llm/service"
)

type App struct {
	Config                *config.Config
	ArchiveService        ArchiveService
	EmailService          EmailService
	MinifluxClientService MinifluxClientService
	DigestService         digest.DigestService
	LLMService            service.LLMService
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

func WithArchiveService(s ArchiveService) Option {
	return func(a *App) {
		a.ArchiveService = s
	}
}

func WithEmailService(s EmailService) Option {
	return func(a *App) {
		a.EmailService = s
	}
}

func WithMinifluxClientService(s MinifluxClientService) Option {
	return func(a *App) {
		a.MinifluxClientService = s
	}
}

func WithDigestService(s digest.DigestService) Option {
	return func(a *App) {
		a.DigestService = s
	}
}

func WithLLMService(s service.LLMService) Option { // Changed
	return func(a *App) {
		a.LLMService = s
	}
}

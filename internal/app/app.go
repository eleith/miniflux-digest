package app

import (
	"miniflux-digest/internal/config"
	"miniflux-digest/internal/llm"
)

type App struct {
	Config                *config.Config
	ArchiveService        ArchiveService
	EmailService          EmailService
	MinifluxClientService MinifluxClientService
	DigestService         DigestService
	LLMService            llm.LLMService
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

func WithDigestService(s DigestService) Option {
	return func(a *App) {
		a.DigestService = s
	}
}

func WithLLMService(s llm.LLMService) Option {
	return func(a *App) {
		a.LLMService = s
	}
}

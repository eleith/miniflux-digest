package config

import (
	"errors"
	"fmt"

	"github.com/go-playground/validator/v10"
	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/v2"
	"github.com/robfig/cron/v3"
)

// https://github.com/go-co-op/gocron/issues/826
func IsValidGocronSchedule(s string) bool {
	standardParser := cron.NewParser(cron.SecondOptional | cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow | cron.Descriptor)
	_, err := standardParser.Parse(s)
	if err == nil {
		return true
	}

	_, err = cron.ParseStandard(s)
	return err == nil
}

type ConfigMiniflux struct {
	Host     string `koanf:"host" validate:"required"`
	ApiToken string `koanf:"api_token" validate:"required"`
}

type ConfigDigestEmail struct {
	To   string `koanf:"to" validate:"omitempty,email"`
	From string `koanf:"from" validate:"omitempty,email"`
}

type ConfigSmtp struct {
	Host     string `koanf:"host"`
	Port     int    `koanf:"port" validate:"omitempty,min=1,max=65535"`
	User     string `koanf:"user"`
	Password string `koanf:"password"`
}

type ConfigCategory struct {
	Title       string `koanf:"title" validate:"required"`
	Description string `koanf:"description"`
}

type ConfigFilters struct {
	FeedTitles            []string `koanf:"feed_titles"`
	CategoryTitles        []string `koanf:"category_titles"`
	SiteURLs              []string `koanf:"site_urls"`
	EntryURLs             []string `koanf:"entry_urls"`
	FeedTitlePatterns     []string `koanf:"feed_title_patterns"`
	CategoryTitlePatterns []string `koanf:"category_title_patterns"`
	SiteURLPatterns       []string `koanf:"site_url_patterns"`
	EntryURLPatterns      []string `koanf:"entry_url_patterns"`
}

type ConfigDigest struct {
	Title        string            `koanf:"title" validate:"required"`
	Email        ConfigDigestEmail `koanf:"email"`
	Schedule     string            `koanf:"schedule" validate:"required,gocron"`
	Host         string            `koanf:"host" validate:"required,url"`
	Compress     bool              `koanf:"compress"`
	View         string            `koanf:"view" validate:"required,oneof=date category ai"`
	MarkAsRead   bool              `koanf:"mark_as_read"`
	RunOnStartup bool              `koanf:"run_on_startup"`
	Filters      ConfigFilters     `koanf:"filters"`
	Categories   []ConfigCategory  `koanf:"categories"`
}

type ConfigAI struct {
	ApiKey string `koanf:"api_key"`
}

type Config struct {
	Miniflux ConfigMiniflux `koanf:"miniflux"`
	Smtp     ConfigSmtp     `koanf:"smtp"`
	Digests  []ConfigDigest `koanf:"digests" validate:"min=1,dive"`
	AI       ConfigAI       `koanf:"ai"`
}

func (c *Config) Validate() error {
	validate := validator.New()
	if err := validate.RegisterValidation("gocron", func(fl validator.FieldLevel) bool {
		return IsValidGocronSchedule(fl.Field().String())
	}); err != nil {
		return fmt.Errorf("failed to register gocron validator: %w", err)
	}

	validate.RegisterStructValidation(func(sl validator.StructLevel) {
		cfg := sl.Current().Interface().(Config)
		for _, digest := range cfg.Digests {
			if digest.View == "ai" && cfg.AI.ApiKey == "" {
				sl.ReportError(cfg.AI.ApiKey, "AI.ApiKey", "ApiKey", "required_if", "Digest.View is 'ai'")
			}
		}
	}, Config{})

	err := validate.Struct(c)
	if err == nil {
		return nil
	}

	var validationErrors validator.ValidationErrors
	if errors.As(err, &validationErrors) {
		return fmt.Errorf("configuration validation failed: %v", validationErrors)
	}

	return err
}

func Load(path string) (*Config, error) {
	k := koanf.New(".")
	parser := yaml.Parser()

	// Default values are mostly relevant for single-digest setups which we've moved away from.
	// We could re-implement them but for now, we rely on the user providing a valid config.

	if err := k.Load(file.Provider(path), parser); err != nil {
		return nil, err
	}

	cfg := &Config{}
	if err := k.Unmarshal("", &cfg); err != nil {
		return nil, err
	}

	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}
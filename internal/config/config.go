package config

import (
	"errors"
	"fmt"
	"regexp"

	"miniflux-digest/internal/utils"

	"github.com/go-playground/validator/v10"
	"github.com/go-viper/mapstructure/v2"
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
	FeedURLs              []string `koanf:"feed_urls"`
	EntryURLs             []string `koanf:"entry_urls"`
	FeedTitlePatterns     []string `koanf:"feed_title_patterns"`
	CategoryTitlePatterns []string `koanf:"category_title_patterns"`
	FeedURLPatterns       []string `koanf:"feed_url_patterns"`
	EntryURLPatterns      []string `koanf:"entry_url_patterns"`
}

type ConfigDigest struct {
	Title        string            `koanf:"title" validate:"required"`
	Email        ConfigDigestEmail `koanf:"email"`
	Schedule     string            `koanf:"schedule" validate:"required,gocron"`
	Host         string            `koanf:"host" validate:"omitempty,url"`
	Compress     *bool             `koanf:"compress"`
	View         string            `koanf:"view" validate:"required,oneof=date category ai"`
	MarkAsRead   bool              `koanf:"mark_as_read"`
	RunOnStartup bool              `koanf:"run_on_startup"`
	Filters      ConfigFilters     `koanf:"filters"`
	Categories   []ConfigCategory  `koanf:"categories" validate:"dive"`
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
		slugs := make(map[string]string)

		for _, digest := range cfg.Digests {
			if digest.View == "ai" && cfg.AI.ApiKey == "" {
				sl.ReportError(cfg.AI.ApiKey, "AI.ApiKey", "ApiKey", "required_if", "Digest.View is 'ai'")
			}

			slug := utils.Slugify(digest.Title)
			if slug == "" {
				sl.ReportError(digest.Title, "Digests.Title", "Title", "invalid_slug", fmt.Sprintf("Digest title '%s' results in an empty slug", digest.Title))
			}

			if originalTitle, exists := slugs[slug]; exists {
				sl.ReportError(digest.Title, "Digests.Title", "Title", "unique_slug", fmt.Sprintf("Digest title '%s' conflicts with '%s' (both slugify to '%s')", digest.Title, originalTitle, slug))
			} else {
				slugs[slug] = digest.Title
			}

			// Validate regex patterns
			validatePatterns := func(patterns []string, fieldName string) {
				for _, p := range patterns {
					if _, err := regexp.Compile(p); err != nil {
						sl.ReportError(digest.Filters, fieldName, fieldName, "regex", fmt.Sprintf("Invalid regex '%s': %v", p, err))
					}
				}
			}

			validatePatterns(digest.Filters.FeedTitlePatterns, "Filters.FeedTitlePatterns")
			validatePatterns(digest.Filters.CategoryTitlePatterns, "Filters.CategoryTitlePatterns")
			validatePatterns(digest.Filters.FeedURLPatterns, "Filters.FeedURLPatterns")
			validatePatterns(digest.Filters.EntryURLPatterns, "Filters.EntryURLPatterns")
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

	if err := k.Load(file.Provider(path), parser); err != nil {
		return nil, err
	}

	cfg := &Config{}
	if err := k.UnmarshalWithConf("", &cfg, koanf.UnmarshalConf{
		DecoderConfig: &mapstructure.DecoderConfig{
			ErrorUnused: true,
			TagName:     "koanf",
			ZeroFields:  true,
			Result:      &cfg,
		},
	}); err != nil {
		return nil, err
	}

	// Apply defaults
	if cfg.Smtp.Port == 0 {
		cfg.Smtp.Port = 587
	}

	for i := range cfg.Digests {
		if cfg.Digests[i].Compress == nil {
			def := true
			cfg.Digests[i].Compress = &def
		}
	}

	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

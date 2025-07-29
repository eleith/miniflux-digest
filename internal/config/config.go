package config

import (
	"errors"
	"fmt"

	"github.com/go-playground/validator/v10"
	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/confmap"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/v2"
	"github.com/robfig/cron/v3"

	"miniflux-digest/internal/digest"
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
	To			string `koanf:"to" validate:"omitempty,email"`
	From		string `koanf:"from" validate:"omitempty,email"`
}

type ConfigSmtp struct {
	Host     string `koanf:"host"`
	Port     int    `koanf:"port" validate:"omitempty,min=1,max=65535"`
	User     string `koanf:"user"`
	Password string `koanf:"password"`
}

type ConfigDigest struct {
	Email        ConfigDigestEmail         `koanf:"email"`
	Schedule     string                    `koanf:"schedule" validate:"gocron"`
	Host         string                    `koanf:"host"`
	Compress     bool                      `koanf:"compress"`
	GroupBy      digest.GroupingType       `koanf:"group_by" validate:"omitempty,oneof=day feed ai"`
	MarkAsRead   bool                      `koanf:"mark_as_read"`
	RunOnStartup bool                      `koanf:"run_on_startup"`
}

type ConfigAI struct {
	ApiKey string `koanf:"api_key"`
}

type Config struct {
	Miniflux ConfigMiniflux `koanf:"miniflux"`
	Smtp     ConfigSmtp     `koanf:"smtp"`
	Digest   ConfigDigest   `koanf:"digest"`
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
		if cfg.Digest.GroupBy == "ai" && cfg.AI.ApiKey == "" {
			sl.ReportError(cfg.AI.ApiKey, "AI.ApiKey", "ApiKey", "required_if", "Digest.GroupBy is 'ai'")
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

	if err := k.Load(confmap.Provider(map[string]any{
		"digest.compress":     true,
		"digest.group_by":     "day",
		"digest.schedule":     "@weekly",
		"digest.mark_as_read": true,
		"digest.run_on_startup": false,
	}, "."), nil); err != nil {
		return nil, err
	}

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

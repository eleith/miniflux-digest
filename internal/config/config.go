package config

import (
	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/confmap"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/v2"

	"miniflux-digest/internal/digest"
)

type Config struct {
	MinifluxHost     string
	MinifluxApiToken string
	SmtpHost         string
	SmtpPort         int
	SmtpUser         string
	SmtpPassword     string
	DigestEmailTo    string
	DigestEmailFrom  string
	DigestSchedule   string
	DigestHost       string
	DigestCompress   bool
	DigestGroupBy    digest.GroupingType
}

var k = koanf.New(".")

func Load(path string) (*Config, error) {
	projectDefaults := map[string]any{
		"digest.compress": true,
		"digest.group_by": digest.GroupingTypeDay.String(),
	}

	if err := k.Load(confmap.Provider(projectDefaults, "."), nil); err != nil {
		return nil, err
	}

	if err := k.Load(file.Provider(path), yaml.Parser()); err != nil {
		return nil, err
	}

	cfg := &Config{
		MinifluxHost:     k.String("miniflux.host"),
		MinifluxApiToken: k.String("miniflux.api_token"),
		SmtpHost:         k.String("smtp.host"),
		SmtpPort:         k.Int("smtp.port"),
		SmtpUser:         k.String("smtp.user"),
		SmtpPassword:     k.String("smtp.password"),
		DigestEmailTo:    k.String("digest.email.to"),
		DigestEmailFrom:  k.String("digest.email.from"),
		DigestSchedule:   k.String("digest.schedule"),
		DigestHost:       k.String("digest.host"),
		DigestCompress:   k.Bool("digest.compress"),
		DigestGroupBy:    digest.GroupingType(k.String("digest.group_by")),
	}

	return cfg, nil
}

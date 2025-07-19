package config

import (
	"log"

	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/v2"
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
}

var k = koanf.New(".")

func Load() (*Config, error) {

	err := k.Load(file.Provider("./config.yaml"), yaml.Parser())

	if err != nil {
		log.Fatalf("error loading config file: %v", err)
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
		DigestSchedule:		k.String("digest.schedule"),
	}

	return cfg, nil
}

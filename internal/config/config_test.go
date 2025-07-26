package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoad(t *testing.T) {
	tests := []struct {
		name    string
		content string
		want    *Config
		wantErr bool
	}{
		{
			name: "compress not set",
			content: `
miniflux:
  host: "miniflux.example.com"
  api_token: "test-token"
smtp:
  host: "smtp.example.com"
  port: 587
  user: "test-user"
  password: "test-password"
digest:
  email:
    to: "to@example.com"
    from: "from@example.com"
  schedule: "@daily"
  host: "https://example.com"
`,
			want: &Config{
				MinifluxHost:     "miniflux.example.com",
				MinifluxApiToken: "test-token",
				SmtpHost:         "smtp.example.com",
				SmtpPort:         587,
				SmtpUser:         "test-user",
				SmtpPassword:     "test-password",
				DigestEmailTo:    "to@example.com",
				DigestEmailFrom:  "from@example.com",
				DigestSchedule:   "@daily",
				DigestHost:       "https://example.com",
				DigestCompress:   true,
			},
			wantErr: false,
		},
		{
			name: "compress set to false",
			content: `
miniflux:
  host: "miniflux.example.com"
  api_token: "test-token"
smtp:
  host: "smtp.example.com"
  port: 587
  user: "test-user"
  password: "test-password"
digest:
  email:
    to: "to@example.com"
    from: "from@example.com"
  schedule: "@daily"
  host: "https://example.com"
  compress: false
`,
			want: &Config{
				MinifluxHost:     "miniflux.example.com",
				MinifluxApiToken: "test-token",
				SmtpHost:         "smtp.example.com",
				SmtpPort:         587,
				SmtpUser:         "test-user",
				SmtpPassword:     "test-password",
				DigestEmailTo:    "to@example.com",
				DigestEmailFrom:  "from@example.com",
				DigestSchedule:   "@daily",
				DigestHost:       "https://example.com",
				DigestCompress:   false,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir, err := os.MkdirTemp("", "config-test")
			if err != nil {
				t.Fatalf("Failed to create temp dir: %v", err)
			}
			defer func() {
				if err := os.RemoveAll(tmpDir); err != nil {
					t.Errorf("Failed to remove temp dir: %v", err)
				}
			}()

			configPath := filepath.Join(tmpDir, "config.yaml")
			if err := os.WriteFile(configPath, []byte(tt.content), 0644); err != nil {
				t.Fatalf("Failed to write dummy config file: %v", err)
			}

			cfg, err := Load(configPath)

			if (err != nil) != tt.wantErr {
				t.Errorf("Load() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if *cfg != *tt.want {
				t.Errorf("Load() = %v, want %v", *cfg, *tt.want)
			}
		})
	}
}
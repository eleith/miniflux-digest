package config

import (
	"os"
	"path/filepath"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestLoad(t *testing.T) {
	tests := []struct {
		name    string
		config  map[string]any
		wantErr bool
	}{
		{
			name: "valid config",
			config: map[string]any{
				"miniflux": map[string]any{
					"host":      "miniflux.example.com",
					"api_token": "test-token",
				},
				"digest": map[string]any{
					"schedule": "@daily",
				},
			},
			wantErr: false,
		},
		{
			name: "invalid config missing miniflux.host",
			config: map[string]any{
				"miniflux": map[string]any{
					"api_token": "test-token",
				},
				"digest": map[string]any{
					"schedule": "@daily",
				},
			},
			wantErr: true,
		},
		{
			name: "invalid config missing miniflux.api_token",
			config: map[string]any{
				"miniflux": map[string]any{
					"host": "miniflux.example.com",
				},
				"digest": map[string]any{
					"schedule": "@daily",
				},
			},
			wantErr: true,
		},
		{
			name: "invalid smtp.port too high",
			config: map[string]any{
				"miniflux": map[string]any{
					"host":      "miniflux.example.com",
					"api_token": "test-token",
				},
				"digest": map[string]any{
					"schedule": "@daily",
				},
				"smtp": map[string]any{
					"port": 65536,
				},
			},
			wantErr: true,
		},
		{
			name: "valid smtp.port",
			config: map[string]any{
				"miniflux": map[string]any{
					"host":      "miniflux.example.com",
					"api_token": "test-token",
				},
				"digest": map[string]any{
					"schedule": "@daily",
				},
				"smtp": map[string]any{
					"port": 587,
				},
			},
			wantErr: false,
		},
		{
			name: "invalid digest.email.to format",
			config: map[string]any{
				"miniflux": map[string]any{
					"host":      "miniflux.example.com",
					"api_token": "test-token",
				},
				"digest": map[string]any{
					"schedule":  "@daily",
					"email.to": "invalid-email",
				},
			},
			wantErr: true,
		},
		{
			name: "invalid digest.email.from format",
			config: map[string]any{
				"miniflux": map[string]any{
					"host":      "miniflux.example.com",
					"api_token": "test-token",
				},
				"digest": map[string]any{
					"schedule":   "@daily",
					"email.from": "another-invalid-email",
				},
			},
			wantErr: true,
		},
		{
			name: "valid digest.email.to and from",
			config: map[string]any{
				"miniflux": map[string]any{
					"host":      "miniflux.example.com",
					"api_token": "test-token",
				},
				"digest": map[string]any{
					"schedule":   "@daily",
					"email.to":   "test@example.com",
					"email.from": "sender@example.com",
				},
			},
							wantErr: false,
		},
		{
			name: "valid cron schedule",
			config: map[string]any{
				"miniflux": map[string]any{
					"host":      "miniflux.example.com",
					"api_token": "test-token",
				},
				"digest": map[string]any{
					"schedule": "* * 1 * *",
				},
			},
			wantErr: false,
		},
		{
			name: "invalid cron schedule",
			config: map[string]any{
				"miniflux": map[string]any{
					"host":      "miniflux.example.com",
					"api_token": "test-token",
				},
				"digest": map[string]any{
					"schedule": "* * * * * * *",
				},
			},
			wantErr: true,
		},
		{
			name: "valid @every schedule",
			config: map[string]any{
				"miniflux": map[string]any{
					"host":      "miniflux.example.com",
					"api_token": "test-token",
				},
				"digest": map[string]any{
					"schedule": "@every 1h30m",
				},
			},
			wantErr: false,
		},
		{
			name: "invalid @every schedule",
			config: map[string]any{
				"miniflux": map[string]any{
					"host":      "miniflux.example.com",
					"api_token": "test-token",
				},
				"digest": map[string]any{
					"schedule": "@every bad-duration",
				},
			},
			wantErr: true,
		},
		{
			name: "invalid group_by value",
			config: map[string]any{
				"miniflux": map[string]any{
					"host":      "miniflux.example.com",
					"api_token": "test-token",
				},
				"digest": map[string]any{
					"schedule": "@daily",
					"group_by": "magic",
				},
			},
			wantErr: true,
		},
		{
			name: "missing ai.api_key when group_by is ai",
			config: map[string]any{
				"miniflux": map[string]any{
					"host":      "miniflux.example.com",
					"api_token": "test-token",
				},
				"digest": map[string]any{
					"schedule": "@daily",
					"group_by": "ai",
				},
			},
			wantErr: true,
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
			data, err := yaml.Marshal(tt.config)
			if err != nil {
				t.Fatalf("Failed to marshal test config: %v", err)
			}

			if err := os.WriteFile(configPath, data, 0644); err != nil {
				t.Fatalf("Failed to write dummy config file: %v", err)
			}

			cfg, err := Load(configPath)

			if (err != nil) != tt.wantErr {
				t.Errorf("Load() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && tt.name == "default digest.schedule is @weekly" {
				if cfg.Digest.Schedule != "@weekly" {
					t.Errorf("Expected digest.schedule to be @weekly, got %s", cfg.Digest.Schedule)
				}
			}
		})
	}

}

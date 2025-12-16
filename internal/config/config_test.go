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
				"digests": []map[string]any{{
					"title":    "Daily Digest",
					"schedule": "@daily",
					"host":     "http://localhost:8080",
					"view":     "category",
				}},
				"ai": map[string]any{
					"api_key": "dummy-key",
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
				"digests": []map[string]any{{
					"title":    "Daily Digest",
					"schedule": "@daily",
					"host":     "http://localhost:8080",
					"view":     "category",
				}},
			},
			wantErr: true,
		},
		{
			name: "invalid config missing miniflux.api_token",
			config: map[string]any{
				"miniflux": map[string]any{
					"host": "miniflux.example.com",
				},
				"digests": []map[string]any{{
					"title":    "Daily Digest",
					"schedule": "@daily",
					"host":     "http://localhost:8080",
					"view":     "category",
				}},
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
				"digests": []map[string]any{{
					"title":    "Daily Digest",
					"schedule": "@daily",
					"host":     "http://localhost:8080",
					"view":     "category",
				}},
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
				"digests": []map[string]any{{
					"title":    "Daily Digest",
					"schedule": "@daily",
					"host":     "http://localhost:8080",
					"view":     "category",
				}},
				"smtp": map[string]any{
					"port": 587,
				},
				"ai": map[string]any{
					"api_key": "dummy-key",
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
				"digests": []map[string]any{{
					"title":    "Daily Digest",
					"schedule": "@daily",
					"host":     "http://localhost:8080",
					"view":     "category",
					"email": map[string]any{
						"to": "invalid-email",
					},
				}},
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
				"digests": []map[string]any{{
					"title":    "Daily Digest",
					"schedule": "@daily",
					"host":     "http://localhost:8080",
					"view":     "category",
					"email": map[string]any{
						"from": "another-invalid-email",
					},
				}},
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
				"digests": []map[string]any{{
					"title":    "Daily Digest",
					"schedule": "@daily",
					"host":     "http://localhost:8080",
					"view":     "category",
					"email": map[string]any{
						"to":   "test@example.com",
						"from": "sender@example.com",
					},
				}},
				"ai": map[string]any{
					"api_key": "dummy-key",
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
				"digests": []map[string]any{{
					"title":    "Daily Digest",
					"schedule": "* * 1 * *",
					"host":     "http://localhost:8080",
					"view":     "category",
				}},
				"ai": map[string]any{
					"api_key": "dummy-key",
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
				"digests": []map[string]any{{
					"title":    "Daily Digest",
					"schedule": "* * * * * * *",
					"host":     "http://localhost:8080",
					"view":     "category",
				}},
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
				"digests": []map[string]any{{
					"title":    "Daily Digest",
					"schedule": "@every 1h30m",
					"host":     "http://localhost:8080",
					"view":     "category",
				}},
				"ai": map[string]any{
					"api_key": "dummy-key",
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
				"digests": []map[string]any{{
					"title":    "Daily Digest",
					"schedule": "@every bad-duration",
					"host":     "http://localhost:8080",
					"view":     "category",
				}},
			},
			wantErr: true,
		},

		{
			name: "invalid view value",
			config: map[string]any{
				"miniflux": map[string]any{
					"host":      "miniflux.example.com",
					"api_token": "test-token",
				},
				"digests": []map[string]any{{
					"title":    "Daily Digest",
					"schedule": "@daily",
					"host":     "http://localhost:8080",
					"view":     "magic",
				}},
			},
			wantErr: true,
		},
		{
			name: "missing ai.api_key when view is ai",
			config: map[string]any{
				"miniflux": map[string]any{
					"host":      "miniflux.example.com",
					"api_token": "test-token",
				},
				"digests": []map[string]any{{
					"title":    "Daily Digest",
					"schedule": "@daily",
					"host":     "http://localhost:8080",
					"view":     "ai",
				}},
			},
			wantErr: true,
		},
		{
			name: "missing digests",
			config: map[string]any{
				"miniflux": map[string]any{
					"host":      "miniflux.example.com",
					"api_token": "test-token",
				},
				"digests": []map[string]any{},
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

			_, err = Load(configPath)

			if (err != nil) != tt.wantErr {
				t.Errorf("Load() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}
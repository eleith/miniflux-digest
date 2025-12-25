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
			name: "valid lowercase filter logic",
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
					"filters": map[string]any{
						"logic": "and", // lowercase should be valid
					},
				}},
			},
			wantErr: false,
		},
		{
			name: "invalid filter logic",
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
					"filters": map[string]any{
						"logic": "XOR", // invalid logic
					},
				}},
			},
			wantErr: true,
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
				}},
			},
			wantErr: true,
		},
		{
			name: "valid config missing view defaults to date",
			config: map[string]any{
				"miniflux": map[string]any{
					"host":      "miniflux.example.com",
					"api_token": "test-token",
				},
				"digests": []map[string]any{{
					"title":    "Daily Digest",
					"schedule": "@daily",
					"host":     "http://localhost:8080",
				}},
			},
			wantErr: false,
		},
		{
			name: "email configured but missing smtp host",
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
						"to": "test@example.com",
					},
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
				"smtp": map[string]any{
					"host": "smtp.example.com",
				},
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
		{
			name: "duplicate digest slugs",
			config: map[string]any{
				"miniflux": map[string]any{
					"host":      "miniflux.example.com",
					"api_token": "test-token",
				},
				"digests": []map[string]any{
					{
						"title":    "My Digest",
						"schedule": "@daily",
						"host":     "http://localhost:8080",
						"view":     "category",
					},
					{
						"title":    "My-Digest",
						"schedule": "@daily",
						"host":     "http://localhost:8080",
						"view":     "category",
					},
				},
			},
			wantErr: true,
		},
		{
			name: "invalid regex pattern",
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
					"filters": map[string]any{
						"feed_title_patterns": []string{"(unclosed parenthesis"},
					},
				}},
			},
			wantErr: true,
		},
		{
			name: "invalid category missing title",
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
					"categories": []map[string]any{
						{"description": "No title here"},
					},
				}},
				"ai": map[string]any{
					"api_key": "dummy-key",
				},
			},
			wantErr: true,
		},
		{
			name: "unknown filter field",
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
					"filters": map[string]any{
						"feed_title": []string{"Some Feed"}, // Incorrect field name
					},
				}},
			},
			wantErr: true,
		},
		{
			name: "invalid digest title empty slug",
			config: map[string]any{
				"miniflux": map[string]any{
					"host":      "miniflux.example.com",
					"api_token": "test-token",
				},
				"digests": []map[string]any{{
					"title":    "!!!",
					"schedule": "@daily",
					"host":     "http://localhost:8080",
					"view":     "category",
				}},
			},
			wantErr: true,
		},
		{
			name: "valid config with top-level unknown key (anchor)",
			config: map[string]any{
				"miniflux": map[string]any{
					"host":      "miniflux.example.com",
					"api_token": "test-token",
				},
				"common_settings": map[string]any{ // This should be ignored
					"schedule": "@daily",
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

func TestYamlAnchors(t *testing.T) {
	yamlContent := `
miniflux:
  host: http://miniflux.app
  api_token: secret

smtp:
  host: smtp.example.com

# Define an anchor for common digest settings
common_digest: &common_digest
  schedule: "@daily"
  view: date
  host: http://example.com
  filters:
    feed_titles:
      - "TechCrunch"

digests:
  - title: "Daily Digest"
    <<: *common_digest
    email:
      to: user@example.com

  - title: "Weekly Digest"
    <<: *common_digest
    schedule: "@weekly"
    email:
      to: user@example.com
`

	tmpfile, err := os.CreateTemp("", "config_anchor_*.yaml")
	if err != nil {
		t.Fatal(err)
	}

	err = os.Remove(tmpfile.Name())
	if err != nil {
		t.Fatalf("Failed to remove temp file: %v", err)
	}

	if _, err := tmpfile.Write([]byte(yamlContent)); err != nil {
		t.Fatal(err)
	}

	err = tmpfile.Close()
	if err != nil {
		t.Fatalf("Failed to close temp file: %v", err)
	}

	cfg, err := Load(tmpfile.Name())
	if err != nil {
		t.Fatalf("Load() unexpected error: %v", err)
	}

	if len(cfg.Digests) != 2 {
		t.Fatalf("Expected 2 digests, got %d", len(cfg.Digests))
	}

	// Check Daily Digest (inherits schedule)
	d1 := cfg.Digests[0]
	if d1.Title != "Daily Digest" {
		t.Errorf("Expected title 'Daily Digest', got '%s'", d1.Title)
	}
	if d1.Schedule != "@daily" {
		t.Errorf("Expected schedule '@daily', got '%s'", d1.Schedule)
	}
	if d1.View != "date" {
		t.Errorf("Expected view 'date', got '%s'", d1.View)
	}
	if len(d1.Filters.FeedTitles) != 1 || d1.Filters.FeedTitles[0] != "TechCrunch" {
		t.Errorf("Expected feed titles ['TechCrunch'], got %v", d1.Filters.FeedTitles)
	}

	// Check Weekly Digest (overrides schedule)
	d2 := cfg.Digests[1]
	if d2.Title != "Weekly Digest" {
		t.Errorf("Expected title 'Weekly Digest', got '%s'", d2.Title)
	}
	if d2.Schedule != "@weekly" {
		t.Errorf("Expected schedule '@weekly', got '%s'", d2.Schedule)
	}
	if d2.View != "date" {
		t.Errorf("Expected view 'date', got '%s'", d2.View)
	}
	if len(d2.Filters.FeedTitles) != 1 || d2.Filters.FeedTitles[0] != "TechCrunch" {
		t.Errorf("Expected feed titles ['TechCrunch'], got %v", d2.Filters.FeedTitles)
	}
}

package manager

import (
	"testing"

	"miniflux-digest/internal/config"
	"miniflux-digest/internal/models"
)

func TestGetOwningDigest(t *testing.T) {
	digests := []config.ConfigDigest{
		{
			Title: "High Priority Tech",
			Filters: config.ConfigFilters{
				FeedTitles: []string{"Tech News"},
			},
		},
		{
			Title: "General News",
			Filters: config.ConfigFilters{
				// Empty filters match everything (catch-all)
			},
		},
	}

	dm, err := NewDigestManager(digests)
	if err != nil {
		t.Fatalf("Failed to create DigestManager: %v", err)
	}

	tests := []struct {
		name      string
		entry     *models.Entry
		wantIndex int
	}{
		{
			name: "Match first digest",
			entry: &models.Entry{
				FeedTitle: "Tech News",
				SiteURL:   "http://technews.com",
			},
			wantIndex: 0,
		},
		{
			name: "Match second digest (catch-all)",
			entry: &models.Entry{
				FeedTitle: "Cooking Weekly",
				SiteURL:   "http://cooking.com",
			},
			wantIndex: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := dm.GetOwningDigest(tt.entry)
			if got != tt.wantIndex {
				t.Errorf("GetOwningDigest() = %v, want %v", got, tt.wantIndex)
			}
		})
	}
}

func TestGetOwningDigest_NoMatch(t *testing.T) {
	digests := []config.ConfigDigest{
		{
			Title: "Exclusive Club",
			Filters: config.ConfigFilters{
				FeedTitles: []string{"VIP"},
			},
		},
	}

	dm, err := NewDigestManager(digests)
	if err != nil {
		t.Fatalf("Failed to create DigestManager: %v", err)
	}

	entry := &models.Entry{FeedTitle: "Plebeian News"}
	got := dm.GetOwningDigest(entry)
	if got != -1 {
		t.Errorf("GetOwningDigest() = %v, want -1", got)
	}
}

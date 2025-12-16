package filter

import (
	"testing"

	"miniflux-digest/internal/config"
	"miniflux-digest/internal/models"
)

func TestMatcher_Matches(t *testing.T) {
	tests := []struct {
		name     string
		entry    *models.Entry
		filters  config.ConfigFilters
		wantMatch bool
	}{
		{
			name: "Match exact feed title",
			entry: &models.Entry{
				FeedTitle: "Tech News", GroupTitle: "Some Category", SiteURL: "http://example.com",
				URL: "http://example.com/post/1",
			},
			filters: config.ConfigFilters{
				FeedTitles: []string{"Tech News"},
			},
			wantMatch: true,
		},
		{
			name: "No match exact feed title",
			entry: &models.Entry{
				FeedTitle: "Cooking", GroupTitle: "Some Category", SiteURL: "http://example.com",
				URL: "http://example.com/post/1",
			},
			filters: config.ConfigFilters{
				FeedTitles: []string{"Tech News"},
			},
			wantMatch: false,
		},
		{
			name: "Match prefix site url",
			entry: &models.Entry{
				FeedTitle: "A", GroupTitle: "Some Category", SiteURL: "https://github.com/user/repo",
				URL: "http://example.com/post/1",
			},
			filters: config.ConfigFilters{
				SiteURLs: []string{"https://github.com"},
			},
			wantMatch: true,
		},
		{
			name: "No match prefix site url",
			entry: &models.Entry{
				FeedTitle: "A", GroupTitle: "Some Category", SiteURL: "https://google.com",
				URL: "http://example.com/post/1",
			},
			filters: config.ConfigFilters{
				SiteURLs: []string{"https://github.com"},
			},
			wantMatch: false,
		},
		{
			name: "Match regex feed title",
			entry: &models.Entry{
				FeedTitle: "My Tech Blog", GroupTitle: "Some Category", SiteURL: "http://example.com",
				URL: "http://example.com/post/1",
			},
			filters: config.ConfigFilters{
				FeedTitlePatterns: []string{".*Tech.*"},
			},
			wantMatch: true,
		},
		{
			name: "Mixed AND logic (Feed Title AND Category)",
			entry: &models.Entry{
				FeedTitle: "Tech News", GroupTitle: "Programming", SiteURL: "http://example.com",
				URL: "http://example.com/post/1",
			},
			filters: config.ConfigFilters{
				FeedTitles:     []string{"Tech News"},
				CategoryTitles: []string{"Programming"},
			},
			wantMatch: true,
		},
		{
			name: "Mixed AND logic fail (Feed Title match, Category mismatch)",
			entry: &models.Entry{
				FeedTitle: "Tech News", GroupTitle: "Cooking", SiteURL: "http://example.com",
				URL: "http://example.com/post/1",
			},
			filters: config.ConfigFilters{
				FeedTitles:     []string{"Tech News"},
				CategoryTitles: []string{"Programming"},
			},
			wantMatch: false,
		},
		{
			name: "Empty filters matches all",
			entry: &models.Entry{
				FeedTitle: "Any", GroupTitle: "Any", SiteURL: "Any",
				URL: "Any",
			},
			filters: config.ConfigFilters{},
			wantMatch: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m, err := NewMatcher(tt.filters)
			if err != nil {
				t.Fatalf("Failed to create matcher: %v", err)
			}
			if got := m.Matches(tt.entry, tt.filters); got != tt.wantMatch {
				t.Errorf("Matches() = %v, want %v", got, tt.wantMatch)
			}
		})
	}
}

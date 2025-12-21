package filter

import (
	"regexp"
	"strings"

	"miniflux-digest/internal/config"
	"miniflux-digest/internal/models"
)

type Matcher struct {
	// Cache regex patterns to avoid recompiling
	feedTitleRegex     []*regexp.Regexp
	categoryTitleRegex []*regexp.Regexp
	siteURLRegex       []*regexp.Regexp
	entryURLRegex      []*regexp.Regexp
}

func NewMatcher(cfg config.ConfigFilters) (*Matcher, error) {
	m := &Matcher{}
	var err error

	m.feedTitleRegex, err = compileRegexes(cfg.FeedTitlePatterns)
	if err != nil {
		return nil, err
	}

	m.categoryTitleRegex, err = compileRegexes(cfg.CategoryTitlePatterns)
	if err != nil {
		return nil, err
	}

	m.siteURLRegex, err = compileRegexes(append(cfg.SiteURLPatterns, cfg.FeedURLPatterns...))
	if err != nil {
		return nil, err
	}

	m.entryURLRegex, err = compileRegexes(cfg.EntryURLPatterns)
	if err != nil {
		return nil, err
	}

	return m, nil
}

func compileRegexes(patterns []string) ([]*regexp.Regexp, error) {
	regexes := make([]*regexp.Regexp, 0, len(patterns))
	for _, p := range patterns {
		r, err := regexp.Compile(p)
		if err != nil {
			return nil, err
		}
		regexes = append(regexes, r)
	}
	return regexes, nil
}

// Matches returns true if the entry matches the filter criteria.
// Logic:
// 1. If no filters are defined, it's a match (default allow).
// 2. Uses cfg.Logic to determine strategy ("AND" vs "OR").
//    - AND: Entry must match ALL filter groups that are defined.
//    - OR:  Entry must match AT LEAST ONE filter group that is defined.
func (m *Matcher) Matches(entry *models.Entry, cfg config.ConfigFilters) bool {
	// If all filters are empty, return true (default allow if "filters" block is missing/empty)
	if isEmpty(cfg) {
		return true
	}

	if cfg.Logic == "AND" {
		// AND Logic: Reject if any defined filter fails to match
		
		if (len(cfg.FeedTitles) > 0 || len(m.feedTitleRegex) > 0) &&
			!matchString(entry.FeedTitle, cfg.FeedTitles, m.feedTitleRegex, true) {
			return false
		}

		if (len(cfg.CategoryTitles) > 0 || len(m.categoryTitleRegex) > 0) &&
			!matchString(entry.GroupTitle, cfg.CategoryTitles, m.categoryTitleRegex, true) {
			return false
		}

		if (len(cfg.SiteURLs) > 0 || len(cfg.FeedURLs) > 0 || len(m.siteURLRegex) > 0) &&
			!matchString(entry.SiteURL, append(cfg.SiteURLs, cfg.FeedURLs...), m.siteURLRegex, false) {
			return false
		}

		// EntryURL corresponds to entry.URL
		if (len(cfg.EntryURLs) > 0 || len(m.entryURLRegex) > 0) &&
			!matchString(entry.URL, cfg.EntryURLs, m.entryURLRegex, false) {
			return false
		}
		
		return true
	}

	// Default to OR Logic
	// OR Logic: Accept if ANY defined filter matches

	if (len(cfg.FeedTitles) > 0 || len(m.feedTitleRegex) > 0) && 
		matchString(entry.FeedTitle, cfg.FeedTitles, m.feedTitleRegex, true) {
		return true
	}

	if (len(cfg.CategoryTitles) > 0 || len(m.categoryTitleRegex) > 0) &&
		matchString(entry.GroupTitle, cfg.CategoryTitles, m.categoryTitleRegex, true) {
		return true
	}

	if (len(cfg.SiteURLs) > 0 || len(cfg.FeedURLs) > 0 || len(m.siteURLRegex) > 0) &&
		matchString(entry.SiteURL, append(cfg.SiteURLs, cfg.FeedURLs...), m.siteURLRegex, false) {
		return true
	}

	// EntryURL corresponds to entry.URL
	if (len(cfg.EntryURLs) > 0 || len(m.entryURLRegex) > 0) &&
		matchString(entry.URL, cfg.EntryURLs, m.entryURLRegex, false) {
		return true
	}

	// If filters were defined but none matched, reject the entry.
	return false
}

// matchString checks if value matches any of the exact strings or regex patterns.
// exactMatch: true for Full Match (Titles), false for Prefix Match (URLs).
func matchString(value string, exacts []string, patterns []*regexp.Regexp, exactMatch bool) bool {
	// If no filters for this specific field are set, we "pass" this field.
	// We only check for a match if there IS something to match against.
	if len(exacts) == 0 && len(patterns) == 0 {
		return true
	}

	for _, e := range exacts {
		if exactMatch {
			if value == e {
				return true
			}
		} else {
			if strings.HasPrefix(value, e) {
				return true
			}
		}
	}

	for _, r := range patterns {
		if r.MatchString(value) {
			return true
		}
	}

	return false
}

func isEmpty(cfg config.ConfigFilters) bool {
	return len(cfg.FeedTitles) == 0 &&
		len(cfg.CategoryTitles) == 0 &&
		len(cfg.SiteURLs) == 0 &&
		len(cfg.FeedURLs) == 0 &&
		len(cfg.EntryURLs) == 0 &&
		len(cfg.FeedTitlePatterns) == 0 &&
		len(cfg.CategoryTitlePatterns) == 0 &&
		len(cfg.SiteURLPatterns) == 0 &&
		len(cfg.FeedURLPatterns) == 0 &&
		len(cfg.EntryURLPatterns) == 0
}

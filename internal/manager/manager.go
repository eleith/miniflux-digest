package manager

import (
	"miniflux-digest/internal/config"
	"miniflux-digest/internal/filter"
	"miniflux-digest/internal/models"
)

type DigestManager struct {
	digests  []config.ConfigDigest
	matchers []*filter.Matcher
}

func NewDigestManager(digests []config.ConfigDigest) (*DigestManager, error) {
	matchers := make([]*filter.Matcher, len(digests))
	for i, d := range digests {
		m, err := filter.NewMatcher(d.Filters)
		if err != nil {
			return nil, err
		}
		matchers[i] = m
	}

	return &DigestManager{
		digests:  digests,
		matchers: matchers,
	}, nil
}

// GetOwningDigest returns the index of the first digest that matches the entry.
// Returns -1 if no digest matches.
func (dm *DigestManager) GetOwningDigest(entry *models.Entry) int {
	for i, m := range dm.matchers {
		if m.Matches(entry, dm.digests[i].Filters) {
			return i
		}
	}
	return -1
}

package digest

import (
	"miniflux-digest/internal/testutil"
	"testing"
)

func TestGroupByDate(t *testing.T) {
	entries := testutil.CreateMockEntries(0)
	groups := GroupByDate(entries)

	if len(groups) != 3 {
		t.Fatalf("Expected 3 date groups, got %d", len(groups))
	}

	if groups[0].Title != "Jan 1, 2024" {
		t.Errorf("Expected first group to be Jan 1, 2024, got %s", groups[0].Title)
	}

	if len(groups[0].Entries) != 3 {
		t.Errorf("Expected 3 entries in first group, got %d", len(groups[0].Entries))
	}
}

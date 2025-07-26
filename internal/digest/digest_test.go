package digest

import (
	"miniflux-digest/internal/models"
	"testing"
	"time"

	miniflux "miniflux.app/v2/client"
)

func TestDayGrouper_GroupEntries(t *testing.T) {
	entries := &miniflux.Entries{
		{
			ID:    1,
			Title: "Entry 1",
			Date:  time.Date(2024, time.January, 2, 10, 0, 0, 0, time.UTC),
		},
		{
			ID:    2,
			Title: "Entry 2",
			Date:  time.Date(2024, time.January, 1, 10, 0, 0, 0, time.UTC),
		},
		{
			ID:    3,
			Title: "Entry 3",
			Date:  time.Date(2024, time.January, 2, 11, 0, 0, 0, time.UTC),
		},
	}

	grouper := &DayGrouper{}
	groups := grouper.GroupEntries(entries)

	if len(groups) != 2 {
		t.Fatalf("Expected 2 groups, got %d", len(groups))
	}

	// Check sorting (older to newer)
	if groups[0].Title != "Jan 1, 2024" {
		t.Errorf("Expected first group to be Jan 1, 2024, got %s", groups[0].Title)
	}
	if groups[1].Title != "Jan 2, 2024" {
		t.Errorf("Expected second group to be Jan 2, 2024, got %s", groups[1].Title)
	}

	// Check entries within groups (older to newer)
	if len(groups[0].Entries) != 1 || groups[0].Entries[0].ID != 2 {
		t.Errorf("Incorrect entries for Jan 1, 2024 group")
	}
	if len(groups[1].Entries) != 2 || groups[1].Entries[0].ID != 1 || groups[1].Entries[1].ID != 3 {
		t.Errorf("Incorrect entries for Jan 2, 2024 group")
	}
}

func TestFeedGrouper_GroupEntries(t *testing.T) {
	entries := &miniflux.Entries{
		{
			ID:     1,
			Title:  "Entry 1",
			FeedID: 100,
			Feed:   &miniflux.Feed{ID: 100, Title: "Feed A"},
		},
		{
			ID:     2,
			Title:  "Entry 2",
			FeedID: 200,
			Feed:   &miniflux.Feed{ID: 200, Title: "Feed B"},
		},
		{
			ID:     3,
			Title:  "Entry 3",
			FeedID: 100,
			Feed:   &miniflux.Feed{ID: 100, Title: "Feed A"},
		},
	}

	grouper := &FeedGrouper{}
	groups := grouper.GroupEntries(entries)

	if len(groups) != 2 {
		t.Fatalf("Expected 2 groups, got %d", len(groups))
	}

	// Find groups by title for easier assertion
	groupA := findGroup(groups, "Feed A")
	groupB := findGroup(groups, "Feed B")

	if groupA == nil || groupB == nil {
		t.Fatalf("Expected groups for Feed A and Feed B")
	}

	if len(groupA.Entries) != 2 || groupA.Entries[0].ID != 1 || groupA.Entries[1].ID != 3 {
		t.Errorf("Incorrect entries for Feed A group")
	}
	if len(groupB.Entries) != 1 || groupB.Entries[0].ID != 2 {
		t.Errorf("Incorrect entries for Feed B group")
	}
}

func TestNewGrouper(t *testing.T) {
	if _, ok := NewGrouper("day").(*DayGrouper); !ok {
		t.Error("Expected DayGrouper for 'day' grouping")
	}
	if _, ok := NewGrouper("feed").(*FeedGrouper); !ok {
		t.Error("Expected FeedGrouper for 'feed' grouping")
	}
	if _, ok := NewGrouper("invalid").(*DayGrouper); !ok {
		t.Error("Expected DayGrouper for invalid grouping")
	}
}

func findGroup(groups []*models.EntryGroup, title string) *models.EntryGroup {
	for _, group := range groups {
		if group.Title == title {
			return group
		}
	}
	return nil
}
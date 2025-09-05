package view

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

func TestDayGrouper_GroupEntries(t *testing.T) {
	entries := testutil.CreateMockEntries(0)
	primaryGroupsMap := GroupByCategory(entries)

	groups := SubGroupByDay(primaryGroupsMap)

	if len(groups) != 3 {
		t.Fatalf("Expected 3 primary groups, got %d", len(groups))
	}

	catAGroup := testutil.FindPrimaryGroup(groups, "Category A")
	if catAGroup == nil || len(catAGroup.SubGroups) != 1 {
		t.Fatalf("Incorrect sub-groups for Category A: %+v", catAGroup)
	}

	subGroup := testutil.FindSubGroup(catAGroup, "Jan 2, 2024")
	if subGroup == nil || len(subGroup.Entries) != 4 {
		t.Fatalf("Incorrect entries for sub-group: %+v", subGroup)
	}
}
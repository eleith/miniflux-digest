package view

import (
	"miniflux-digest/internal/models"
	"miniflux-digest/internal/testutil"
	
	"testing"

	"github.com/stretchr/testify/assert"
)



func TestSubGroupByFeed(t *testing.T) {
	entries := testutil.CreateMockEntries(0)
	primaryGroups := GroupByCategory(entries) // Use GroupByCategory to get initial primary groups

	groups := SubGroupByFeed(primaryGroups)
	assert.Len(t, groups, 3, "Expected 3 primary groups after SubGroupByFeed")

	// Check "Category A" group
	catAGroup := findPrimaryGroupDigestData(groups, "Category A")
	assert.NotNil(t, catAGroup, "Category A group not found")
	assert.Len(t, catAGroup.SubGroups, 3, "Expected 3 sub-groups for Category A") // Feed A, Feed C, Feed D

	// Check "Category B" group
	catBGroup := findPrimaryGroupDigestData(groups, "Category B")
	assert.NotNil(t, catBGroup, "Category B group not found")
	assert.Len(t, catBGroup.SubGroups, 2, "Expected 2 sub-groups for Category B") // Feed B, Feed E

	// Check "Category C" group
	catCGroup := findPrimaryGroupDigestData(groups, "Category C")
	assert.NotNil(t, catCGroup, "Category C group not found")
	assert.Len(t, catCGroup.SubGroups, 2, "Expected 2 sub-groups for Category C") // Feed F, Feed G
}

// findPrimaryGroupDigestData is a helper function to find a PrimaryGroupDigestData by title
func findPrimaryGroupDigestData(groups []*models.PrimaryGroupDigestData, title string) *models.PrimaryGroupDigestData {
	for _, group := range groups {
		if group.Title == title {
			return group
		}
	}
	return nil
}

func TestFeedGrouper_GroupEntries(t *testing.T) {
	entries := testutil.CreateMockEntries(0)
	primaryGroupsMap := GroupByCategory(entries)

	groups := SubGroupByFeed(primaryGroupsMap)

	if len(groups) != 3 {
		t.Fatalf("Expected 3 primary groups, got %d", len(groups))
	}

	catAGroup := testutil.FindPrimaryGroup(groups, "Category A")
	if catAGroup == nil || len(catAGroup.SubGroups) != 3 {
		t.Fatalf("Incorrect sub-groups for Category A: %+v", catAGroup)
	}
}

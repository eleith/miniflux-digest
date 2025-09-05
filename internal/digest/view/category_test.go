package view

import (
	"miniflux-digest/internal/models" // Added this line
	"miniflux-digest/internal/testutil"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGroupByCategory(t *testing.T) {
	entries := testutil.CreateMockEntries(0) // Use mock entries

	groups := GroupByCategory(entries)

	assert.Len(t, groups, 3, "Expected 3 category groups")

	// Check "Category A" group
	catAGroup := findPrimaryGroup(groups, "Category A")
	assert.NotNil(t, catAGroup, "Category A group not found")
	assert.Len(t, catAGroup.Entries, 4, "Expected 4 entries in Category A group")

	// Check "Category B" group
	catBGroup := findPrimaryGroup(groups, "Category B")
	assert.NotNil(t, catBGroup, "Category B group not found")
	assert.Len(t, catBGroup.Entries, 3, "Expected 3 entries in Category B group")

	// Check "Category C" group
	catCGroup := findPrimaryGroup(groups, "Category C")
	assert.NotNil(t, catCGroup, "Category C group not found")
	assert.Len(t, catCGroup.Entries, 2, "Expected 2 entries in Category C group")
}

// findPrimaryGroup is a helper function to find a primary group by title
func findPrimaryGroup(groups []*models.PrimaryGroup, title string) *models.PrimaryGroup {
	for _, group := range groups {
		if group.Title == title {
			return group
		}
	}
	return nil
}
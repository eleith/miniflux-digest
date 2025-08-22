package digest

import (
	"context"
	"errors"
	"miniflux-digest/internal/models"
	"miniflux-digest/internal/testutil"
	"testing"

	"google.golang.org/genai"
)

type mockLLMService struct {
	GenerateContentFunc func(ctx context.Context, prompt string, schema *genai.Schema) (string, error)
}

func (m *mockLLMService) GenerateContent(ctx context.Context, prompt string, schema *genai.Schema) (string, error) {
	if m.GenerateContentFunc != nil {
		return m.GenerateContentFunc(ctx, prompt, schema)
	}
	return "", errors.New("GenerateContentFunc not implemented")
}

func findPrimaryGroup(groups []*models.PrimaryGroupDigestData, title string) *models.PrimaryGroupDigestData {
	for _, group := range groups {
		if group.Title == title {
			return group
		}
	}
	return nil
}

func findSubGroup(primaryGroup *models.PrimaryGroupDigestData, title string) *models.EntryGroup {
	for _, group := range primaryGroup.SubGroups {
		if group.Title == title {
			return group
		}
	}
	return nil
}

func createLLMGrouperMockEntries() []*models.Entry {
	return []*models.Entry{
		{ID: 1, Title: "Entry 1", FeedID: 100, FeedTitle: "Feed A", GroupID: 1, GroupTitle: "Category A"},
		{ID: 2, Title: "Entry 2", FeedID: 100, FeedTitle: "Feed A", GroupID: 1, GroupTitle: "Category A"},
		{ID: 3, Title: "Entry 3", FeedID: 200, FeedTitle: "Feed B", GroupID: 1, GroupTitle: "Category A"},
	}
}

func TestDayGrouper_GroupEntries(t *testing.T) {
	entries := testutil.CreateMockEntries(0)
	primaryGroupsMap := GroupEntries(entries, "category")

	grouper := &DayGrouper{}
	groups, summary := grouper.GroupEntries(primaryGroupsMap)

	if summary != nil {
		t.Errorf("Expected an empty summary for DayGrouper, got %q", *summary)
	}

	if len(groups) != 3 {
		t.Fatalf("Expected 3 primary groups, got %d", len(groups))
	}

	catAGroup := findPrimaryGroup(groups, "Category A")
	if catAGroup == nil || len(catAGroup.SubGroups) != 1 {
		t.Fatalf("Incorrect sub-groups for Category A: %+v", catAGroup)
	}

	subGroup := findSubGroup(catAGroup, "Category A - Jan 2, 2024")
	if subGroup == nil || len(subGroup.Entries) != 4 {
		t.Fatalf("Incorrect entries for sub-group: %+v", subGroup)
	}
}

func TestFeedGrouper_GroupEntries(t *testing.T) {
	entries := testutil.CreateMockEntries(0)
	primaryGroupsMap := GroupEntries(entries, "category")

	grouper := &FeedGrouper{}
	groups, summary := grouper.GroupEntries(primaryGroupsMap)

	if summary != nil {
		t.Errorf("Expected an empty summary for FeedGrouper, got %q", *summary)
	}

	if len(groups) != 3 {
		t.Fatalf("Expected 3 primary groups, got %d", len(groups))
	}

	catAGroup := findPrimaryGroup(groups, "Category A")
	if catAGroup == nil || len(catAGroup.SubGroups) != 3 {
		t.Fatalf("Incorrect sub-groups for Category A: %+v", catAGroup)
	}
}

func TestLLMGrouper_GroupEntries(t *testing.T) {
	entries := createLLMGrouperMockEntries()
	primaryGroupsMap := GroupEntries(entries, "category")

	mockLLM := &mockLLMService{
		GenerateContentFunc: func(ctx context.Context, prompt string, schema *genai.Schema) (string, error) {
			return `{
				"overview": "This is a summary of all entries.",
				"primary_group_summaries": [
					{
						"primary_group_id": 1,
						"summary": "This is a summary of Category A."
					}
				],
				"sub_groups": [
					{
						"title": "Sub-group 1",
						"entry_ids": [1, 2]
					}
				]
			}`,
				nil
		},
	}

	grouper := &LLMGrouper{LLMService: mockLLM}
	groups, summary := grouper.GroupEntries(primaryGroupsMap)

	if summary == nil || *summary != "This is a summary of all entries." {
		t.Errorf("Expected summary 'This is a summary of all entries.', got %q", *summary)
	}

	if len(groups) != 1 {
		t.Fatalf("Expected 1 group, got %d", len(groups))
	}

	catA := findPrimaryGroup(groups, "Category A")
	if catA == nil || len(catA.SubGroups) != 2 { // Sub-group 1 and Uncategorized
		t.Fatalf("Incorrect number of sub-groups for Category A: %+v", catA)
	}

	if catA.Summary != "This is a summary of Category A." {
		t.Errorf("Expected summary 'This is a summary of Category A.', got %q", catA.Summary)
	}

	subGroup1 := findSubGroup(catA, "Sub-group 1")
	if subGroup1 == nil || len(subGroup1.Entries) != 2 {
		t.Fatalf("Incorrect entries for sub-group 1: %+v", subGroup1)
	}

	uncategorized := findSubGroup(catA, "Uncategorized")
	if uncategorized == nil || len(uncategorized.Entries) != 1 { // Entry 3
		t.Fatalf("Incorrect entries for Uncategorized sub-group: %+v", uncategorized)
	}
}

func TestDigestService_BuildDigestData_NonAI(t *testing.T) {
	mockLLM := &mockLLMService{}
	digestService := NewDigestService(mockLLM)

	entries := testutil.CreateMockEntries(0)
	icons := map[int64]*models.FeedIcon{
		101: {FeedID: 101, Data: "iconA"},
		102: {FeedID: 102, Data: "iconB"},
	}

	overviewData := digestService.BuildDigestData(entries, icons, "category", "day", "date", "http://miniflux.test")

	if overviewData == nil {
		t.Fatal("overviewData is nil")
	}
	if len(overviewData.PrimaryGroups) != 3 {
		t.Errorf("Expected 3 primary groups, got %d", len(overviewData.PrimaryGroups))
	}
}

func TestDigestService_BuildDigestData_LLMFallback(t *testing.T) {
	entries := testutil.CreateMockEntries(0)
	icons := map[int64]*models.FeedIcon{
		100: {FeedID: 100, Data: "icon100"},
		200: {FeedID: 200, Data: "icon200"},
	}

	mockLLM := &mockLLMService{
		GenerateContentFunc: func(ctx context.Context, prompt string, schema *genai.Schema) (string, error) {
			return "", errors.New("LLM service error during test")
		},
	}
	digestService := NewDigestService(mockLLM)

	overviewData := digestService.BuildDigestData(entries, icons, "category", "ai", "date", "http://miniflux.test")

	if overviewData == nil {
		t.Fatal("overviewData is nil")
	}

	if overviewData.OverviewSummary != "" {
		t.Errorf("Expected OverviewSummary to be empty, got %q", overviewData.OverviewSummary)
	}

	if len(overviewData.PrimaryGroups) != 3 { // Category A, Category B, Category C
		t.Fatalf("Expected 3 primary groups after fallback, got %d", len(overviewData.PrimaryGroups))
	}
}

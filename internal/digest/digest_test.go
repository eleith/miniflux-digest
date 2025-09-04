package digest

import (
	"context"
	"errors"
	"miniflux-digest/internal/models"
	"miniflux-digest/internal/testutil"
	"sort"
	"testing"

	"google.golang.org/genai"
)

type mockLLMService struct {
	GenerateContentFunc func(ctx context.Context, prompt string, schema *genai.Schema) ([]byte, error)
}

func (m *mockLLMService) GenerateContent(ctx context.Context, prompt string, schema *genai.Schema) ([]byte, error) {
	if m.GenerateContentFunc != nil {
		return m.GenerateContentFunc(ctx, prompt, schema)
	}
	return nil, errors.New("GenerateContentFunc not implemented")
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

	groups, summary := SubGroupByDay(primaryGroupsMap)

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

	subGroup := findSubGroup(catAGroup, "Jan 2, 2024")
	if subGroup == nil || len(subGroup.Entries) != 4 {
		t.Fatalf("Incorrect entries for sub-group: %+v", subGroup)
	}
}

func TestFeedGrouper_GroupEntries(t *testing.T) {
	entries := testutil.CreateMockEntries(0)
	primaryGroupsMap := GroupEntries(entries, "category")

	groups, summary := SubGroupByFeed(primaryGroupsMap)

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
		GenerateContentFunc: func(ctx context.Context, prompt string, schema *genai.Schema) ([]byte, error) {
			return []byte(`{
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
			}`),
				nil
		},
	}

	groups, summary := SubGroupByAI(primaryGroupsMap, mockLLM)

	if summary == nil || *summary != "This is a summary of all entries." {
		t.Errorf("Expected summary 'This is a summary of all entries.', got %q", *summary)
	}

	if len(groups) != 1 {
		t.Fatalf("Expected 1 group, got %d", len(groups))
	}

	catA := findPrimaryGroup(groups, "Category A")
	if catA == nil || len(catA.SubGroups) != 2 {
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
	if uncategorized == nil || len(uncategorized.Entries) != 1 {
		t.Fatalf("Incorrect entries for Uncategorized sub-group: %+v", uncategorized)
	}
}

func TestDigestService_BuildDigestData(t *testing.T) {
	mockLLM := &mockLLMService{
		GenerateContentFunc: func(ctx context.Context, prompt string, schema *genai.Schema) ([]byte, error) {
			return []byte(`{
				"overview": "This is a summary of all entries.",
				"primary_group_summaries": [],
				"sub_groups": [
					{
						"title": "Sub-group 1",
						"entry_ids": [1, 2]
					}
				]
			}`), nil
		},
	}
	digestService := NewDigestService(mockLLM)

	entries := testutil.CreateMockEntries(0)
	icons := map[int64]*models.FeedIcon{
		101: {FeedID: 101, Data: "iconA"},
		102: {FeedID: 102, Data: "iconB"},
	}

	t.Run("view=date", func(t *testing.T) {
		overviewData := digestService.BuildDigestData(entries, icons, "date", "http://miniflux.test")
		if overviewData == nil {
			t.Fatal("overviewData is nil")
		}
		if len(overviewData.PrimaryGroups) != 3 {
			t.Errorf("Expected 3 primary groups, got %d", len(overviewData.PrimaryGroups))
		}
		dateGroup := findPrimaryGroup(overviewData.PrimaryGroups, "Jan 1, 2024")
		if dateGroup == nil {
			t.Fatalf("Did not find expected primary group 'Jan 1, 2024'")
		}
		if len(dateGroup.SubGroups) != 2 {
			t.Fatalf("Incorrect sub-groups for Jan 1, 2024: %+v", dateGroup)
		}
		firstSubGroup := dateGroup.SubGroups[0]
		if !sort.SliceIsSorted(firstSubGroup.Entries, func(i, j int) bool {
			return firstSubGroup.Entries[i].Date.Before(firstSubGroup.Entries[j].Date)
		}) {
			t.Errorf("Entries in sub-group are not sorted by date")
		}
	})

	t.Run("view=category", func(t *testing.T) {
		overviewData := digestService.BuildDigestData(entries, icons, "category", "http://miniflux.test")
		if overviewData == nil {
			t.Fatal("overviewData is nil")
		}
		if len(overviewData.PrimaryGroups) != 3 {
			t.Errorf("Expected 3 primary groups, got %d", len(overviewData.PrimaryGroups))
		}
		catAGroup := findPrimaryGroup(overviewData.PrimaryGroups, "Category A")
		if catAGroup == nil || len(catAGroup.SubGroups) != 3 {
			t.Fatalf("Incorrect sub-groups for Category A: %+v", catAGroup)
		}
		firstSubGroup := catAGroup.SubGroups[0]
		if !sort.SliceIsSorted(firstSubGroup.Entries, func(i, j int) bool {
			return firstSubGroup.Entries[i].Date.Before(firstSubGroup.Entries[j].Date)
		}) {
			t.Errorf("Entries in sub-group are not sorted by date")
		}
	})

	t.Run("view=ai", func(t *testing.T) {
		aiEntries := createLLMGrouperMockEntries()
		overviewData := digestService.BuildDigestData(aiEntries, icons, "ai", "http://miniflux.test")
		if overviewData == nil {
			t.Fatal("overviewData is nil")
		}
		if overviewData.OverviewSummary != "This is a summary of all entries." {
			t.Errorf("Expected summary, got %q", overviewData.OverviewSummary)
		}
		if len(overviewData.PrimaryGroups) != 1 {
			t.Errorf("Expected 1 primary group, got %d", len(overviewData.PrimaryGroups))
		}
	})

	t.Run("view=ai fallback", func(t *testing.T) {
		llmErrorService := NewDigestService(&mockLLMService{
			GenerateContentFunc: func(ctx context.Context, prompt string, schema *genai.Schema) ([]byte, error) {
				return nil, errors.New("LLM service error")
			},
		})
		overviewData := llmErrorService.BuildDigestData(entries, icons, "ai", "http://miniflux.test")
		if overviewData == nil {
			t.Fatal("overviewData is nil")
		}
		if overviewData.OverviewSummary != "" {
			t.Errorf("Expected empty summary on fallback, got %q", overviewData.OverviewSummary)
		}
		catAGroup := findPrimaryGroup(overviewData.PrimaryGroups, "Category A")
		if catAGroup == nil || len(catAGroup.SubGroups) != 3 {
			t.Fatalf("Incorrect sub-groups for Category A on fallback: %+v", catAGroup)
		}
	})
}
package digest

import (
	"context"
	"errors"
	"miniflux-digest/internal/models"
	"testing"
	"time"

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

func findGroup(groups []*models.EntryGroup, title string) *models.EntryGroup {
	for _, group := range groups {
		if group.Title == title {
			return group
		}
	}
	return nil
}



func createDayGrouperMockEntries() []*models.Entry {
	return []*models.Entry{
		{
			ID:     1,
			Title:  "Entry 1 - Jan 2",
			Date:   time.Date(2024, time.January, 2, 10, 0, 0, 0, time.UTC),
			FeedID: 100,
			FeedTitle: "Feed A",
			GroupTitle: "Category A",
		},
		{
			ID:     2,
			Title:  "Entry 2 - Jan 1",
			Date:   time.Date(2024, time.January, 1, 10, 0, 0, 0, time.UTC),
			FeedID: 200,
			FeedTitle: "Feed B",
			GroupTitle: "Category B",
		},
		{
			ID:     3,
			Title:  "Entry 3 - Jan 2",
			Date:   time.Date(2024, time.January, 2, 11, 0, 0, 0, time.UTC),
			FeedID: 100,
			FeedTitle: "Feed A",
			GroupTitle: "Category A",
		},
		{
			ID:      4,
			Title:   "Entry 4 - Jan 1",
			Content: "Content of entry 4 about Go concurrency.",
			Date:    time.Date(2024, time.January, 1, 12, 0, 0, 0, time.UTC),
			FeedID:  200,
			FeedTitle: "Feed B",
			GroupTitle: "Category B",
		},
	}
}

func createFeedGrouperMockEntries() []*models.Entry {
	return []*models.Entry{
		{
			ID:     1,
			Title:  "Entry 1 - Feed A",
			FeedID: 100,
			FeedTitle: "Feed A",
			Date:   time.Date(2024, time.January, 1, 10, 0, 0, 0, time.UTC),
		},
		{
			ID:     2,
			Title:  "Entry 2 - Feed B",
			FeedID: 200,
			FeedTitle: "Feed B",
			Date:   time.Date(2024, time.January, 1, 11, 0, 0, 0, time.UTC),
		},
		{
			ID:     3,
			Title:  "Entry 3 - Feed A",
			FeedID: 100,
			FeedTitle: "Feed A",
			Date:   time.Date(2024, time.January, 1, 12, 0, 0, 0, time.UTC),
		},
		{
			ID:     4,
			Title:  "Entry 4 - Feed B",
			FeedID: 200,
			FeedTitle: "Feed B",
			Date:   time.Date(2024, time.January, 1, 13, 0, 0, 0, time.UTC),
		},
	}
}

func TestDayGrouper_GroupEntries(t *testing.T) {
	entries := createDayGrouperMockEntries()

	grouper := &DayGrouper{}
	groups, summary := grouper.GroupEntries(entries)

	if summary != nil {
		t.Errorf("Expected an empty summary for DayGrouper, got %q", *summary)
	}

	if len(groups) != 2 {
		t.Fatalf("Expected 2 groups, got %d", len(groups))
	}

	if groups[0].Title != "Jan 1, 2024" {
		t.Errorf("Expected first group to be Jan 1, 2024, got %s", groups[0].Title)
	}
	if groups[1].Title != "Jan 2, 2024" {
		t.Errorf("Expected second group to be Jan 2, 2024, got %s", groups[1].Title)
	}

	jan1Group := findGroup(groups, "Jan 1, 2024")
	if jan1Group == nil || len(jan1Group.Entries) != 2 || jan1Group.Entries[0].ID != 2 || jan1Group.Entries[1].ID != 4 {
		t.Errorf("Incorrect entries for Jan 1, 2024 group: %+v", jan1Group)
	}

	jan2Group := findGroup(groups, "Jan 2, 2024")
	if jan2Group == nil || len(jan2Group.Entries) != 2 || jan2Group.Entries[0].ID != 1 || jan2Group.Entries[1].ID != 3 {
		t.Errorf("Incorrect entries for Jan 2, 2024 group: %+v", jan2Group)
	}
}

func TestFeedGrouper_GroupEntries(t *testing.T) {
	entries := createFeedGrouperMockEntries()

	grouper := &FeedGrouper{}
	groups, summary := grouper.GroupEntries(entries)

	if summary != nil {
		t.Errorf("Expected an empty summary for FeedGrouper, got %q", *summary)
	}

	if len(groups) != 2 {
		t.Fatalf("Expected 2 groups, got %d", len(groups))
	}

	feedAGroup := findGroup(groups, "Feed A")
	if feedAGroup == nil || len(feedAGroup.Entries) != 2 || feedAGroup.Entries[0].ID != 1 || feedAGroup.Entries[1].ID != 3 {
		t.Errorf("Incorrect entries for Feed A group: %+v", feedAGroup)
	}

	feedBGroup := findGroup(groups, "Feed B")
	if feedBGroup == nil || len(feedBGroup.Entries) != 2 || feedBGroup.Entries[0].ID != 2 || feedBGroup.Entries[1].ID != 4 {
		t.Errorf("Incorrect entries for Feed B group: %+v", feedBGroup)
	}
}

func TestLLMGrouper_GroupEntries(t *testing.T) {
	entries := createDayGrouperMockEntries()

	mockLLM := &mockLLMService{
		GenerateContentFunc: func(ctx context.Context, prompt string, schema *genai.Schema) (string, error) {
			return `{
				"summary": "This is a summary of all entries.",
				"groups": [
					{
						"title": "Go Programming",
						"summary": "summary",
						"entry_ids": [1, 2, 4]
					},
					{
						"title": "Python Programming",
						"summary": "summary",
						"entry_ids": [3]
					}
				]
			}`, nil
		},
	}

	grouper := &LLMGrouper{LLMService: mockLLM}
	groups, summary := grouper.GroupEntries(entries)

	if summary == nil || *summary != "This is a summary of all entries." {
		t.Errorf("Expected summary 'This is a summary of all entries.', got %q", *summary)
	}

	if len(groups) != 2 {
		t.Fatalf("Expected 2 groups, got %d", len(groups))
	}

	goGroup := findGroup(groups, "Go Programming")
	if goGroup == nil || len(goGroup.Entries) != 3 || goGroup.Entries[0].ID != 1 || goGroup.Entries[1].ID != 2 || goGroup.Entries[2].ID != 4 {
		t.Errorf("Incorrect Go Programming group: %+v", goGroup)
	}

	pythonGroup := findGroup(groups, "Python Programming")
	if pythonGroup == nil || len(pythonGroup.Entries) != 1 || pythonGroup.Entries[0].ID != 3 {
		t.Errorf("Incorrect Python Programming group: %+v", pythonGroup)
	}

	mockLLM.GenerateContentFunc = func(ctx context.Context, prompt string, schema *genai.Schema) (string, error) {
		return "", errors.New("LLM API error")
	}
	groups, summary = grouper.GroupEntries(entries)
	if len(groups) == 0 || summary != nil { 
		t.Error("Expected fallback to DayGrouper on LLM error")
	}

	mockLLM.GenerateContentFunc = func(ctx context.Context, prompt string, schema *genai.Schema) (string, error) {
		return "invalid json", nil
	}
	groups, summary = grouper.GroupEntries(entries)
	if len(groups) == 0 || summary != nil { 
		t.Error("Expected fallback to DayGrouper on invalid JSON")
	}

	mockLLM.GenerateContentFunc = func(ctx context.Context, prompt string, schema *genai.Schema) (string, error) {
			return `{
				"summary": "Summary with missing entry.",
				"groups": [
					{
						"title": "Go Programming",
						"summary": "summary",
						"entry_ids": [1, 2]
					}
				]
			}`, nil
		}
	groups, _ = grouper.GroupEntries(entries)

	if len(groups) != 2 {
		t.Fatalf("Expected 2 groups including uncategorized, got %d", len(groups))
	}

	uncategorizedGroup := findGroup(groups, "Uncategorized")
	if uncategorizedGroup == nil || len(uncategorizedGroup.Entries) != 2 || uncategorizedGroup.Entries[0].ID != 3 || uncategorizedGroup.Entries[1].ID != 4 {
		t.Errorf("Incorrect Uncategorized group: %+v", uncategorizedGroup)
	}
}

func TestLLMGrouper_GroupEntries_WithDuplicateEntries(t *testing.T) {
	entries := createDayGrouperMockEntries()

	mockLLM := &mockLLMService{
		GenerateContentFunc: func(ctx context.Context, prompt string, schema *genai.Schema) (string, error) {
			return `{
				"summary": "This is a summary of all entries.",
				"groups": [
					{
						"title": "Go Programming",
						"summary": "summary",
						"entry_ids": [1, 2, 1, 4]
					},
					{
						"title": "Python Programming",
						"summary": "summary",
						"entry_ids": [3, 3]
					}
				]
			}`, nil
		},
	}

	grouper := &LLMGrouper{LLMService: mockLLM}
	groups, summary := grouper.GroupEntries(entries)

	if summary == nil || *summary != "This is a summary of all entries." {
		t.Errorf("Expected summary 'This is a summary of all entries.', got %q", *summary)
	}

	if len(groups) != 2 {
		t.Fatalf("Expected 2 groups, got %d", len(groups))
	}

	goGroup := findGroup(groups, "Go Programming")
	if goGroup == nil || len(goGroup.Entries) != 3 {
		t.Errorf("Incorrect Go Programming group: %+v", goGroup)
	}

	pythonGroup := findGroup(groups, "Python Programming")
	if pythonGroup == nil || len(pythonGroup.Entries) != 1 {
		t.Errorf("Incorrect Python Programming group: %+v", pythonGroup)
	}
}

func TestLLMGrouper_GroupEntries_Counts(t *testing.T) {
	entries := createDayGrouperMockEntries()

	mockLLM := &mockLLMService{
		GenerateContentFunc: func(ctx context.Context, prompt string, schema *genai.Schema) (string, error) {
			return `{
				"summary": "Test summary for counts.",
				"groups": [
					{
						"title": "Group A (2 entries, 2 feeds)",
						"summary": "Summary A",
						"entry_ids": [1, 2]
					},
					{
						"title": "Group B (1 entry, 1 feed)",
						"summary": "Summary B",
						"entry_ids": [3]
					},
					{
						"title": "Group C (0 entries, 0 feeds)",
						"summary": "Summary C",
						"entry_ids": []
					}
				]
			}`, nil
		},
	}

	grouper := &LLMGrouper{LLMService: mockLLM}
	groups, summary := grouper.GroupEntries(entries)

	if summary == nil || *summary != "Test summary for counts." {
		t.Errorf("Expected overview summary 'Test summary for counts.', got %q", *summary)
	}

	if len(groups) != 4 {
		t.Fatalf("Expected 4 groups, got %d", len(groups))
	}

	groupA := findGroup(groups, "Group A (2 entries, 2 feeds)")
	if groupA == nil {
		t.Fatal("Expected 'Group A' not found")
	}
	if groupA.TotalEntries != 2 {
		t.Errorf("Expected TotalEntries 2 for 'Group A', got %d", groupA.TotalEntries)
	}
	if groupA.TotalFeeds != 2 {
		t.Errorf("Expected TotalFeeds 2 for 'Group A', got %d", groupA.TotalFeeds)
	}

	groupB := findGroup(groups, "Group B (1 entry, 1 feed)")
	if groupB == nil {
		t.Fatal("Expected 'Group B' not found")
	}
	if groupB.TotalEntries != 1 {
		t.Errorf("Expected TotalEntries 1 for 'Group B', got %d", groupB.TotalEntries)
	}
	if groupB.TotalFeeds != 1 {
		t.Errorf("Expected TotalFeeds 1 for 'Group B', got %d", groupB.TotalFeeds)
	}

	groupC := findGroup(groups, "Group C (0 entries, 0 feeds)")
	if groupC == nil {
		t.Fatal("Expected 'Group C' not found")
	}
	if groupC.TotalEntries != 0 {
		t.Errorf("Expected TotalEntries 0 for 'Group C', got %d", groupC.TotalEntries)
	}
	if groupC.TotalFeeds != 0 {
		t.Errorf("Expected TotalFeeds 0 for 'Group C', got %d", groupC.TotalFeeds)
	}
}

func TestGroupEntries(t *testing.T) {
	categoryEntries := []*models.Entry{
		{GroupTitle: "Category A"},
		{GroupTitle: "Category B"},
		{GroupTitle: "Category A"},
		{GroupTitle: ""},
	}

	categoryGroups := GroupEntries(categoryEntries, "category")
	if len(categoryGroups) != 3 {
		t.Errorf("Expected 3 category groups, got %d", len(categoryGroups))
	}
	if len(categoryGroups["Category A"]) != 2 {
		t.Errorf("Expected 2 entries in Category A, got %d", len(categoryGroups["Category A"]))
	}
	if len(categoryGroups["Category B"]) != 1 {
		t.Errorf("Expected 1 entry in Category B, got %d", len(categoryGroups["Category B"]))
	}
	if len(categoryGroups["Uncategorized"]) != 1 {
		t.Errorf("Expected 1 entry in Uncategorized, got %d", len(categoryGroups["Uncategorized"]))
	}

	feedEntries := []*models.Entry{
		{FeedTitle: "Feed X"},
		{FeedTitle: "Feed Y"},
		{FeedTitle: "Feed X"},
		{FeedTitle: "Feed Z"},
	}

	feedGroups := GroupEntries(feedEntries, "feed")
	if len(feedGroups) != 3 {
		t.Errorf("Expected 3 feed groups, got %d", len(feedGroups))
	}
	if len(feedGroups["Feed X"]) != 2 {
		t.Errorf("Expected 2 entries in Feed X, got %d", len(feedGroups["Feed X"]))
	}
	if len(feedGroups["Feed Y"]) != 1 {
		t.Errorf("Expected 1 entry in Feed Y, got %d", len(feedGroups["Feed Y"]))
	}
	if len(feedGroups["Feed Z"]) != 1 {
		t.Errorf("Expected 1 entry in Feed Z, got %d", len(feedGroups["Feed Z"]))
	}
}

func TestLLMGrouper_GroupEntries_CorrectCounts(t *testing.T) {
	entries := createDayGrouperMockEntries()

	mockLLM := &mockLLMService{
		GenerateContentFunc: func(ctx context.Context, prompt string, schema *genai.Schema) (string, error) {
			return `{
				"summary": "Test summary for correct counts.",
				"groups": [
					{
						"title": "Group Alpha",
						"summary": "Summary Alpha",
						"entry_ids": [1, 2]
					},
					{
						"title": "Group Beta",
						"summary": "Summary Beta",
						"entry_ids": [3]
					},
					{
						"title": "Group Gamma",
						"summary": "Summary Gamma",
						"entry_ids": []
					}
				]
			}`, nil
		},
	}

	grouper := &LLMGrouper{LLMService: mockLLM}
	groups, summary := grouper.GroupEntries(entries)

	if summary == nil || *summary != "Test summary for correct counts." {
		t.Errorf("Expected overview summary 'Test summary for correct counts.', got %q", *summary)
	}

	if len(groups) != 4 { 
		t.Fatalf("Expected 3 groups, got %d", len(groups))
	}

	groupAlpha := findGroup(groups, "Group Alpha")
	if groupAlpha == nil {
		t.Fatal("Expected 'Group Alpha' not found")
	}
	if groupAlpha.TotalEntries != 2 {
		t.Errorf("Expected TotalEntries 2 for 'Group Alpha', got %d", groupAlpha.TotalEntries)
	}
	if groupAlpha.TotalFeeds != 2 { 
		t.Errorf("Expected TotalFeeds 2 for 'Group Alpha', got %d", groupAlpha.TotalFeeds)
	}

	groupBeta := findGroup(groups, "Group Beta")
	if groupBeta == nil {
		t.Fatal("Expected 'Group Beta' not found")
	}
	if groupBeta.TotalEntries != 1 {
		t.Errorf("Expected TotalEntries 1 for 'Group Beta', got %d", groupBeta.TotalEntries)
	}
	if groupBeta.TotalFeeds != 1 { 
		t.Errorf("Expected TotalFeeds 1 for 'Group Beta', got %d", groupBeta.TotalFeeds)
	}

	groupGamma := findGroup(groups, "Group Gamma")
	if groupGamma == nil {
		t.Fatal("Expected 'Group Gamma' not found")
	}
	if groupGamma.TotalEntries != 0 {
		t.Errorf("Expected TotalEntries 0 for 'Group Gamma', got %d", groupGamma.TotalEntries)
	}
	if groupGamma.TotalFeeds != 0 {
		t.Errorf("Expected TotalFeeds 0 for 'Group Gamma', got %d", groupGamma.TotalFeeds)
	}
}

func TestDigestService_BuildDigestData_NonAI(t *testing.T) {
	mockLLM := &mockLLMService{}
	digestService := NewDigestService(mockLLM)

	entries1 := []*models.Entry{
		{ID: 1, Title: "Entry 1", FeedID: 101, FeedTitle: "Feed A", GroupTitle: "Category X", Date: time.Date(2024, time.January, 1, 0, 0, 0, 0, time.UTC)},
		{ID: 2, Title: "Entry 2", FeedID: 102, FeedTitle: "Feed B", GroupTitle: "Category X", Date: time.Date(2024, time.January, 2, 0, 0, 0, 0, time.UTC)},
		{ID: 3, Title: "Entry 3", FeedID: 101, FeedTitle: "Feed A", GroupTitle: "Category Y", Date: time.Date(2024, time.January, 1, 0, 0, 0, 0, time.UTC)},
	}
	icons1 := map[int64]*models.FeedIcon{
		101: {FeedID: 101, Data: "iconA"},
		102: {FeedID: 102, Data: "iconB"},
	}

	overviewData1 := digestService.BuildDigestData(entries1, icons1, "category", "day", "date", "http://miniflux.test")

	if overviewData1 == nil {
		t.Fatal("overviewData1 is nil")
	}
	if len(overviewData1.PrimaryGroups) != 2 {
		t.Errorf("Expected 2 primary groups, got %d", len(overviewData1.PrimaryGroups))
	}

	catX := overviewData1.PrimaryGroups[0]
	if catX.Title != "Category X" || catX.TotalEntries != 2 || catX.TotalFeeds != 2 {
		t.Errorf("Category X mismatch: %+v", catX)
	}
	if len(catX.SubGroups) != 2 {
		t.Errorf("Expected 2 sub-groups for Category X, got %d", len(catX.SubGroups))
	}

	catY := overviewData1.PrimaryGroups[1]
	if catY.Title != "Category Y" || catY.TotalEntries != 1 || catY.TotalFeeds != 1 {
		t.Errorf("Category Y mismatch: %+v", catY)
	}
	if len(catY.SubGroups) != 1 {
		t.Errorf("Expected 1 sub-group for Category Y, got %d", len(catY.SubGroups))
	}

	entries2 := []*models.Entry{
		{ID: 4, Title: "Entry 4", FeedID: 103, FeedTitle: "Feed C", GroupTitle: "Category Z", Date: time.Date(2024, time.January, 3, 0, 0, 0, 0, time.UTC)},
		{ID: 5, Title: "Entry 5", FeedID: 104, FeedTitle: "Feed D", GroupTitle: "Category Z", Date: time.Date(2024, time.January, 3, 0, 0, 0, 0, time.UTC)},
	}
	icons2 := map[int64]*models.FeedIcon{
		103: {FeedID: 103, Data: "iconC"},
		104: {FeedID: 104, Data: "iconD"},
	}

	overviewData2 := digestService.BuildDigestData(entries2, icons2, "feed", "feed", "date", "http://miniflux.test")

	if overviewData2 == nil {
		t.Fatal("overviewData2 is nil")
	}
	if len(overviewData2.PrimaryGroups) != 2 {
		t.Errorf("Expected 2 primary groups, got %d", len(overviewData2.PrimaryGroups))
	}

	feedC := overviewData2.PrimaryGroups[0]
	if feedC.Title != "Feed C" || feedC.TotalEntries != 1 || feedC.TotalFeeds != 1 {
		t.Errorf("Feed C mismatch: %+v", feedC)
	}
	if len(feedC.SubGroups) != 1 {
		t.Errorf("Expected 1 sub-group for Feed C, got %d", len(feedC.SubGroups))
	}

	feedD := overviewData2.PrimaryGroups[1]
	if feedD.Title != "Feed D" || feedD.TotalEntries != 1 || feedD.TotalFeeds != 1 {
		t.Errorf("Feed D mismatch: %+v", feedD)
	}
	if len(feedD.SubGroups) != 1 {
		t.Errorf("Expected 1 sub-group for Feed D, got %d", len(feedD.SubGroups))
	}
}

package digest

import (
	"context"
	"errors"
	"miniflux-digest/internal/models"
	"testing"
	"time"

	"google.golang.org/genai"
	miniflux "miniflux.app/v2/client"
)

type mockLLMService struct {
	GenerateContentFunc func(ctx context.Context, prompt string, schema *genai.Schema) (string, error)
}

func findGroup(groups []*models.EntryGroup, title string) *models.EntryGroup {
	for _, group := range groups {
		if group.Title == title {
			return group
		}
	}
	return nil
}

func (m *mockLLMService) GenerateContent(ctx context.Context, prompt string, schema *genai.Schema) (string, error) {
	if m.GenerateContentFunc != nil {
		return m.GenerateContentFunc(ctx, prompt, schema)
	}
	return "", nil
}

func createDayGrouperMockEntries() *miniflux.Entries {
	return &miniflux.Entries{
		{
			ID:    1,
			Title: "Entry 1 - Jan 2",
			Date:  time.Date(2024, time.January, 2, 10, 0, 0, 0, time.UTC),
			Feed:  &miniflux.Feed{ID: 100, Title: "Feed A"},
		},
		{
			ID:    2,
			Title: "Entry 2 - Jan 1",
			Date:  time.Date(2024, time.January, 1, 10, 0, 0, 0, time.UTC),
			Feed:  &miniflux.Feed{ID: 200, Title: "Feed B"},
		},
		{
			ID:    3,
			Title: "Entry 3 - Jan 2",
			Date:  time.Date(2024, time.January, 2, 11, 0, 0, 0, time.UTC),
			Feed:  &miniflux.Feed{ID: 100, Title: "Feed A"},
		},
		{
			ID:    4,
			Title: "Entry 4 - Jan 1",
			Content: "Content of entry 4 about Go concurrency.",
			Date:  time.Date(2024, time.January, 1, 12, 0, 0, 0, time.UTC),
			Feed:  &miniflux.Feed{ID: 200, Title: "Feed B"},
		},
	}
}

func createFeedGrouperMockEntries() *miniflux.Entries {
	return &miniflux.Entries{
		{
			ID:     1,
			Title:  "Entry 1 - Feed A",
			FeedID: 100,
			Feed:   &miniflux.Feed{ID: 100, Title: "Feed A"},
			Date:   time.Date(2024, time.January, 1, 10, 0, 0, 0, time.UTC),
		},
		{
			ID:     2,
			Title:  "Entry 2 - Feed B",
			FeedID: 200,
			Feed:   &miniflux.Feed{ID: 200, Title: "Feed B"},
			Date:   time.Date(2024, time.January, 1, 11, 0, 0, 0, time.UTC),
		},
		{
			ID:     3,
			Title:  "Entry 3 - Feed A",
			FeedID: 100,
			Feed:   &miniflux.Feed{ID: 100, Title: "Feed A"},
			Date:   time.Date(2024, time.January, 1, 12, 0, 0, 0, time.UTC),
		},
		{
			ID:     4,
			Title:  "Entry 4 - Feed B",
			FeedID: 200,
			Feed:   &miniflux.Feed{ID: 200, Title: "Feed B"},
			Date:   time.Date(2024, time.January, 1, 13, 0, 0, 0, time.UTC),
		},
	}
}

func TestDayGrouper_GroupEntries(t *testing.T) {
	entries := createDayGrouperMockEntries()

	grouper := &DayGrouper{}
	groups, summary := grouper.GroupEntries(entries)

	if summary == "" {
		t.Error("Expected a non-empty summary for DayGrouper")
	}

	// Expect 2 groups: Jan 1, 2024 and Jan 2, 2024
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

	if summary == "" {
		t.Error("Expected a non-empty summary for FeedGrouper")
	}

	// Expect 2 groups: Feed A and Feed B
	if len(groups) != 2 {
		t.Fatalf("Expected 2 groups, got %d", len(groups))
	}

	// Find groups by title for easier assertion
	feedAGroup := findGroup(groups, "Feed A")
	if feedAGroup == nil || len(feedAGroup.Entries) != 2 || feedAGroup.Entries[0].ID != 1 || feedAGroup.Entries[1].ID != 3 {
		t.Errorf("Incorrect entries for Feed A group: %+v", feedAGroup)
	}

	feedBGroup := findGroup(groups, "Feed B")
	if feedBGroup == nil || len(feedBGroup.Entries) != 2 || feedBGroup.Entries[0].ID != 2 || feedBGroup.Entries[1].ID != 4 {
		t.Errorf("Incorrect entries for Feed B group: %+v", feedBGroup)
	}
}

func TestNewGrouper(t *testing.T) {
	mockLLM := &mockLLMService{}

	if _, ok := NewGrouper(GroupingTypeDay, mockLLM).(*DayGrouper); !ok {
		t.Error("Expected DayGrouper for 'day' grouping")
	}
	if _, ok := NewGrouper(GroupingTypeFeed, mockLLM).(*FeedGrouper); !ok {
		t.Error("Expected FeedGrouper for 'feed' grouping")
	}
	if _, ok := NewGrouper("invalid", mockLLM).(*DayGrouper); !ok {
		t.Error("Expected DayGrouper for invalid grouping")
	}
	if _, ok := NewGrouper("ai", mockLLM).(*LLMGrouper); !ok {
		t.Error("Expected LLMGrouper for 'ai' grouping")
	}
}

func TestLLMGrouper_GroupEntries(t *testing.T) {
	entries := createDayGrouperMockEntries()

	expectedLLMResponse := `{
	"summary": "This is a summary of all entries.",
	"groups": [
		{
			"title": "Go Programming",
			"entries": [1, 2, 4]
		},
		{
			"title": "Python Programming",
			"entries": [3]
		}
	]
}`

	mockLLM := &mockLLMService{
		GenerateContentFunc: func(ctx context.Context, prompt string, schema *genai.Schema) (string, error) {
			return expectedLLMResponse, nil
		},
	}

	grouper := &LLMGrouper{LLMService: mockLLM}
	groups, summary := grouper.GroupEntries(entries)

	if summary != "This is a summary of all entries." {
		t.Errorf("Expected summary \"This is a summary of all entries.\", got \"%s\"", summary)
	}

	if len(groups) != 2 {
		t.Fatalf("Expected 2 groups, got %d", len(groups))
	}

	// Check group titles and entries
	goGroup := findGroup(groups, "Go Programming")
	if goGroup == nil || len(goGroup.Entries) != 3 || goGroup.Entries[0].ID != 1 || goGroup.Entries[1].ID != 2 || goGroup.Entries[2].ID != 4 {
		t.Errorf("Incorrect Go Programming group: %+v", goGroup)
	}

	pythonGroup := findGroup(groups, "Python Programming")
	if pythonGroup == nil || len(pythonGroup.Entries) != 1 || pythonGroup.Entries[0].ID != 3 {
		t.Errorf("Incorrect Python Programming group: %+v", pythonGroup)
	}

	// Test fallback to DayGrouper on LLM error
	mockLLM.GenerateContentFunc = func(ctx context.Context, prompt string, schema *genai.Schema) (string, error) {
		return "", errors.New("LLM API error")
	}
	groups, summary = grouper.GroupEntries(entries)
	if len(groups) == 0 || summary == "" {
		t.Error("Expected fallback to DayGrouper on LLM error")
	}

	// Test fallback to DayGrouper on invalid JSON
	mockLLM.GenerateContentFunc = func(ctx context.Context, prompt string, schema *genai.Schema) (string, error) {
		return "invalid json", nil
	}
	groups, summary = grouper.GroupEntries(entries)
	if len(groups) == 0 || summary == "" {
		t.Error("Expected fallback to DayGrouper on invalid JSON")
	}

	// Test ungrouped entries
	expectedLLMResponseWithMissingEntry := `{
	"summary": "Summary with missing entry.",
	"groups": [
		{
			"title": "Go Programming",
			"entries": [1, 2]
		}
	]
}`
	mockLLM.GenerateContentFunc = func(ctx context.Context, prompt string, schema *genai.Schema) (string, error) {
		return expectedLLMResponseWithMissingEntry, nil
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
package digest

import (
	"context"
	"errors"
	"miniflux-digest/internal/models"
	"testing"
	"time"

	miniflux "miniflux.app/v2/client"
)

type mockLLMService struct {
	GenerateContentFunc func(ctx context.Context, prompt string) (string, error)
}

func findGroup(groups []*models.EntryGroup, title string) *models.EntryGroup {
	for _, group := range groups {
		if group.Title == title {
			return group
		}
	}
	return nil
}

func (m *mockLLMService) GenerateContent(ctx context.Context, prompt string) (string, error) {
	if m.GenerateContentFunc != nil {
		return m.GenerateContentFunc(ctx, prompt)
	}
	return "", nil
}

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
	groups, summary := grouper.GroupEntries(entries)

	if summary == "" {
		t.Error("Expected a non-empty summary for DayGrouper")
	}

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
	groups, summary := grouper.GroupEntries(entries)

	if summary == "" {
		t.Error("Expected a non-empty summary for FeedGrouper")
	}

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
	entries := &miniflux.Entries{
		{
			ID:    1,
			Title: "Entry 1",
			Content: "Content of entry 1 about Go programming.",
		},
		{
			ID:    2,
			Title: "Entry 2",
			Content: "Content of entry 2 about Go testing.",
		},
		{
			ID:    3,
			Title: "Entry 3",
			Content: "Content of entry 3 about Python programming.",
		},
		{
			ID:    4,
			Title: "Entry 4",
			Content: "Content of entry 4 about Go concurrency.",
		},
	}

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
		GenerateContentFunc: func(ctx context.Context, prompt string) (string, error) {
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
	mockLLM.GenerateContentFunc = func(ctx context.Context, prompt string) (string, error) {
		return "", errors.New("LLM API error")
	}
	groups, summary = grouper.GroupEntries(entries)
	if len(groups) == 0 || summary == "" {
		t.Error("Expected fallback to DayGrouper on LLM error")
	}

	// Test fallback to DayGrouper on invalid JSON
	mockLLM.GenerateContentFunc = func(ctx context.Context, prompt string) (string, error) {
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
	mockLLM.GenerateContentFunc = func(ctx context.Context, prompt string) (string, error) {
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

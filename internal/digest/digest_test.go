package digest

import (
	"context"
	"errors"
	"miniflux-digest/internal/digest/view"
	"miniflux-digest/internal/models"
	"miniflux-digest/internal/testutil"
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
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

func createLLMGrouperMockEntries() []*models.Entry {
	return []*models.Entry{
		{ID: 1, Title: "Entry 1", FeedID: 100, FeedTitle: "Feed A", GroupID: 1, GroupTitle: "Category A"},
		{ID: 2, Title: "Entry 2", FeedID: 100, FeedTitle: "Feed A", GroupID: 1, GroupTitle: "Category A"},
		{ID: 3, Title: "Entry 3", FeedID: 200, FeedTitle: "Feed B", GroupID: 1, GroupTitle: "Category A"},
	}
}

func TestDigestService_BuildDigestData(t *testing.T) {
	mockLLM := &mockLLMService{
		GenerateContentFunc: func(ctx context.Context, prompt string, schema *genai.Schema) ([]byte, error) {
			switch schema {
			case view.InitialGroupingResponseSchema:
				return []byte(`{
					"groups": [
						{
							"title": "Go Lang",
							"entry_ids": [2, 4]
						},
						{
							"title": "AI News",
							"entry_ids": [1, 3]
						}
					]
				}`), nil
			case view.SummaryResponseSchema:
				return []byte(`{"summary": "This is a summary."}`), nil
			case view.SubGroupingResponseSchema:
				return []byte(`{
					"sub_groups": [
						{"title": "Sub-group 1", "entry_ids": [1]}
					]
				}`), nil
			default:
				return nil, errors.New("unexpected LLM call")
			}
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
		dateGroup := testutil.FindPrimaryGroup(overviewData.PrimaryGroups, "Jan 1, 2024")
		if dateGroup == nil {
			t.Fatalf("Did not find expected primary group 'Jan 1, 2024'")
		}
		if len(dateGroup.SubGroups) != 1 {
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
		catAGroup := testutil.FindPrimaryGroup(overviewData.PrimaryGroups, "Category A")
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
		if len(overviewData.PrimaryGroups) != 2 {
			t.Errorf("Expected 2 primary groups, got %d", len(overviewData.PrimaryGroups))
		}
		// Assert on sorted order
		assert.Equal(t, "AI News", overviewData.PrimaryGroups[0].Title)
		assert.Equal(t, "Go Lang", overviewData.PrimaryGroups[1].Title)
	})

	t.Run("view=ai - GroupAIEntries error fallback", func(t *testing.T) {
		mockLLM := &mockLLMService{
			GenerateContentFunc: func(ctx context.Context, prompt string, schema *genai.Schema) ([]byte, error) {
				if schema == view.InitialGroupingResponseSchema {
					return nil, errors.New("mock LLM grouping error")
				}
				return nil, errors.New("unexpected LLM call")
			},
		}
		digestService := NewDigestService(mockLLM)
		aiEntries := createLLMGrouperMockEntries()
		overviewData := digestService.BuildDigestData(aiEntries, icons, "ai", "http://miniflux.test")

		if overviewData == nil {
			t.Fatal("overviewData is nil")
		}
		if len(overviewData.PrimaryGroups) != 1 {
			t.Errorf("Expected 1 primary group from date fallback, got %d", len(overviewData.PrimaryGroups))
		}
		dateGroup := testutil.FindPrimaryGroup(overviewData.PrimaryGroups, "Jan 1, 0001")
		if dateGroup == nil {
			t.Fatalf("Did not find expected primary group 'Jan 1, 0001' after date fallback")
		}
	})

	t.Run("view=ai - sub-grouping error", func(t *testing.T) {
		mockLLM := &mockLLMService{
			GenerateContentFunc: func(ctx context.Context, prompt string, schema *genai.Schema) ([]byte, error) {
				if schema == view.InitialGroupingResponseSchema {
					return []byte(`{
						"groups": [
							{
								"title": "Test Group",
								"entry_ids": [1, 2, 3]
							}
						]
					}`), nil
				}
				if schema == view.SummaryResponseSchema {
					return []byte(`{"summary": "This is a summary."}`), nil
				}
				if schema == view.SubGroupingResponseSchema {
					return nil, errors.New("mock LLM sub-grouping error")
				}
				return nil, errors.New("unexpected LLM call")
			},
		}
		digestService := NewDigestService(mockLLM)
		aiEntries := createLLMGrouperMockEntries()
		overviewData := digestService.BuildDigestData(aiEntries, icons, "ai", "http://miniflux.test")

		if overviewData == nil {
			t.Fatal("overviewData is nil")
		}
		if len(overviewData.PrimaryGroups) != 1 {
			t.Errorf("Expected 1 primary group, got %d", len(overviewData.PrimaryGroups))
		}
		group := overviewData.PrimaryGroups[0]
		if group.Summary != "This is a summary." {
			t.Errorf("Expected group summary to be present, got %q", group.Summary)
		}
		if len(group.SubGroups) != 1 || group.SubGroups[0].Title != "Uncategorized" {
			t.Errorf("Expected a single 'Uncategorized' sub-group on error, got %+v", group.SubGroups)
		}
	})
}
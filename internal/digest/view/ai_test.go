
package view

import (
	"context"
	"encoding/json"
	"errors"
	
	"miniflux-digest/internal/models"
	"miniflux-digest/internal/testutil"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"google.golang.org/genai"
)

// mockLLMService is a mock implementation of the LLMService for testing.
type mockLLMService struct {
	GenerateContentFunc func(ctx context.Context, prompt string, schema *genai.Schema) ([]byte, error)
}

func (m *mockLLMService) GenerateContent(ctx context.Context, prompt string, schema *genai.Schema) ([]byte, error) {
	if m.GenerateContentFunc != nil {
		return m.GenerateContentFunc(ctx, prompt, schema)
	}
	return nil, errors.New("mockLLMService.GenerateContentFunc is not implemented")
}

func TestGroupAIEntries(t *testing.T) {
	entries := []*models.Entry{
		{ID: 1, Title: "Entry 1", Content: "Content 1", FeedTitle: "Feed 1"},
		{ID: 2, Title: "Entry 2", Content: "Content 2", FeedTitle: "Feed 2"},
		{ID: 3, Title: "Entry 3", Content: "Content 3", FeedTitle: "Feed 1"},
		{ID: 4, Title: "Entry 4", Content: "Content 4", FeedTitle: "Feed 3"},
	}

	t.Run("happy path", func(t *testing.T) {
		mockResp := GroupingResponse{
			Groups: []struct {
				Title    string  `json:"title"`
				EntryIDs []int64 `json:"entry_ids"`
			}{
				{Title: "Group 1", EntryIDs: []int64{1, 2}},
				{Title: "Group 2", EntryIDs: []int64{3}},
			},
		}
		mockJSON, _ := json.Marshal(mockResp)

		llmService := &mockLLMService{
			GenerateContentFunc: func(ctx context.Context, prompt string, schema *genai.Schema) ([]byte, error) {
				return mockJSON, nil
			},
		}

		groups, err := GroupAIEntries(context.Background(), entries, llmService)
		assert.NoError(t, err)
		assert.NotNil(t, groups)
		assert.Len(t, groups, 3, "Expected 2 grouped and 1 uncategorized group")

		group1 := testutil.FindPrimaryGroup(groups, "Group 1")
		assert.NotNil(t, group1)
		assert.Len(t, group1.SubGroups[0].Entries, 2)
		assert.Equal(t, 2, group1.TotalEntries)

		group2 := testutil.FindPrimaryGroup(groups, "Group 2")
		assert.NotNil(t, group2)
		assert.Len(t, group2.SubGroups[0].Entries, 1)
		assert.Equal(t, 1, group2.TotalEntries)

		uncategorized := testutil.FindPrimaryGroup(groups, "Uncategorized")
		assert.NotNil(t, uncategorized)
		assert.Len(t, uncategorized.SubGroups[0].Entries, 1)
		assert.Equal(t, int64(4), uncategorized.SubGroups[0].Entries[0].ID)
	})

	t.Run("llm service error", func(t *testing.T) {
		llmService := &mockLLMService{
			GenerateContentFunc: func(ctx context.Context, prompt string, schema *genai.Schema) ([]byte, error) {
				return nil, errors.New("llm error")
			},
		}

		_, err := GroupAIEntries(context.Background(), entries, llmService)
		assert.Error(t, err)
	})

	t.Run("llm returns invalid json", func(t *testing.T) {
		llmService := &mockLLMService{
			GenerateContentFunc: func(ctx context.Context, prompt string, schema *genai.Schema) ([]byte, error) {
				return []byte("invalid json"), nil
			},
		}

		_, err := GroupAIEntries(context.Background(), entries, llmService)
		assert.Error(t, err)
	})
}

func TestProcessPrimaryGroupWithAI(t *testing.T) {
	pg := &models.PrimaryGroup{
		ID:    1,
		Title: "Test Primary Group",
		Entries: []*models.Entry{
			{ID: 1, Title: "Entry 1", Content: "Content 1"},
			{ID: 2, Title: "Entry 2", Content: "Content 2"},
			{ID: 3, Title: "Entry 3", Content: "Content 3"},
		},
	}

	t.Run("happy path", func(t *testing.T) {
		summaryResp := AIGroupSummaryResponse{Summary: "This is a summary."}
		summaryJSON, _ := json.Marshal(summaryResp)

		subGroupingResp := AISubGroupingResponse{
			SubGroups: []struct {
				Title    string  `json:"title"`
				EntryIDs []int64 `json:"entry_ids"`
			}{
				{Title: "Sub-group 1", EntryIDs: []int64{1, 2}},
			},
		}
		subGroupingJSON, _ := json.Marshal(subGroupingResp)

		llmService := &mockLLMService{
			GenerateContentFunc: func(ctx context.Context, prompt string, schema *genai.Schema) ([]byte, error) {
				if schema == SummaryResponseSchema {
					return summaryJSON, nil
				}
				if schema == SubGroupingResponseSchema {
					return subGroupingJSON, nil
				}
				return nil, errors.New("unexpected schema in happy path test")
			},
		}

		summary, subGroups, err := ProcessPrimaryGroupWithAI(context.Background(), pg, llmService)
		assert.NoError(t, err)
		assert.Equal(t, "This is a summary.", summary)
		assert.Len(t, subGroups, 2, "Expected 1 sub-group and 1 uncategorized")

		sg1 := subGroups[0]
		assert.Equal(t, "Sub-group 1", sg1.Title)
		assert.Len(t, sg1.Entries, 2)

		uncategorized := subGroups[1]
		assert.Equal(t, "Uncategorized", uncategorized.Title)
		assert.Len(t, uncategorized.Entries, 1)
		assert.Equal(t, int64(3), uncategorized.Entries[0].ID)
	})

	t.Run("summarization fails", func(t *testing.T) {
		subGroupingResp := AISubGroupingResponse{
			SubGroups: []struct {
				Title    string  `json:"title"`
				EntryIDs []int64 `json:"entry_ids"`
			}{
				{Title: "Sub-group 1", EntryIDs: []int64{1, 2, 3}},
			},
		}
		subGroupingJSON, _ := json.Marshal(subGroupingResp)

		llmService := &mockLLMService{
			GenerateContentFunc: func(ctx context.Context, prompt string, schema *genai.Schema) ([]byte, error) {
				if schema == SummaryResponseSchema {
					return nil, errors.New("summarization failed")
				}
				if schema == SubGroupingResponseSchema {
					return subGroupingJSON, nil
				}
				return nil, errors.New("unexpected schema")
			},
		}

		summary, subGroups, err := ProcessPrimaryGroupWithAI(context.Background(), pg, llmService)
		assert.NoError(t, err)
		assert.Empty(t, summary, "Summary should be empty when summarization fails")
		assert.Len(t, subGroups, 1, "Should still get sub-groups even if summarization fails")
		assert.Equal(t, "Sub-group 1", subGroups[0].Title)
	})

	t.Run("sub-grouping fails", func(t *testing.T) {
		summaryResp := AIGroupSummaryResponse{Summary: "This is a summary."}
		summaryJSON, _ := json.Marshal(summaryResp)

		llmService := &mockLLMService{
			GenerateContentFunc: func(ctx context.Context, prompt string, schema *genai.Schema) ([]byte, error) {
				if schema == SummaryResponseSchema {
					return summaryJSON, nil
				}
				if schema == SubGroupingResponseSchema {
					return nil, errors.New("sub-grouping failed")
				}
				return nil, errors.New("unexpected schema")
			},
		}

		summary, subGroups, err := ProcessPrimaryGroupWithAI(context.Background(), pg, llmService)
		assert.NoError(t, err)
		assert.Equal(t, "This is a summary.", summary)
		assert.Len(t, subGroups, 1, "Should have one fallback 'Uncategorized' group")
		assert.Equal(t, "Uncategorized", subGroups[0].Title)
		assert.Len(t, subGroups[0].Entries, 3, "All entries should be in the fallback group")
	})
}

func TestBuildDigestDataByAI(t *testing.T) {
	entries := []*models.Entry{
		{ID: 1, Title: "Entry 1", Content: "Content 1", FeedTitle: "Feed 1", GroupID: 101, GroupTitle: "Category 1", Date: time.Now()},
		{ID: 2, Title: "Entry 2", Content: "Content 2", FeedTitle: "Feed 2", GroupID: 101, GroupTitle: "Category 1", Date: time.Now().Add(-1 * time.Hour)},
		{ID: 3, Title: "Entry 3", Content: "Content 3", FeedTitle: "Feed 1", GroupID: 102, GroupTitle: "Category 2", Date: time.Now().Add(-2 * time.Hour)},
	}

	t.Run("happy path", func(t *testing.T) {
		groupingResp := GroupingResponse{
			Groups: []struct {
				Title    string  `json:"title"`
				EntryIDs []int64 `json:"entry_ids"`
			}{{Title: "AI Group 1", EntryIDs: []int64{1, 2}}},
		}
		groupingJSON, _ := json.Marshal(groupingResp)

		summaryResp := AIGroupSummaryResponse{Summary: "AI summary"}
		summaryJSON, _ := json.Marshal(summaryResp)

		subGroupingResp := AISubGroupingResponse{
			SubGroups: []struct {
				Title    string  `json:"title"`
				EntryIDs []int64 `json:"entry_ids"`
			}{{Title: "AI Sub-group", EntryIDs: []int64{1, 2}}},
		}
		subGroupingJSON, _ := json.Marshal(subGroupingResp)

		llmService := &mockLLMService{
			GenerateContentFunc: func(ctx context.Context, prompt string, schema *genai.Schema) ([]byte, error) {
				if schema == GroupingResponseSchema {
					return groupingJSON, nil
				}
				if schema == SummaryResponseSchema {
					return summaryJSON, nil
				}
				if schema == SubGroupingResponseSchema {
					return subGroupingJSON, nil
				}
				return nil, errors.New("unexpected schema")
			},
		}

		groups := BuildDigestDataByAI(entries, context.Background(), llmService)
		assert.Len(t, groups, 2)

		aiGroup := testutil.FindPrimaryGroup(groups, "AI Group 1")
		assert.NotNil(t, aiGroup)
		assert.Equal(t, "AI summary", aiGroup.Summary)
		assert.Len(t, aiGroup.SubGroups, 1)
		assert.Equal(t, "AI Sub-group", aiGroup.SubGroups[0].Title)
	})

	t.Run("fallback to category grouping", func(t *testing.T) {
		llmService := &mockLLMService{
			GenerateContentFunc: func(ctx context.Context, prompt string, schema *genai.Schema) ([]byte, error) {
				return nil, errors.New("llm error")
			},
		}

		groups := BuildDigestDataByAI(entries, context.Background(), llmService)
		assert.Len(t, groups, 2, "Should fall back to 2 categories")
		cat1 := testutil.FindPrimaryGroup(groups, "Category 1")
		assert.NotNil(t, cat1)
		cat2 := testutil.FindPrimaryGroup(groups, "Category 2")
		assert.NotNil(t, cat2)
	})

	t.Run("fallback to date grouping", func(t *testing.T) {
		singleCategoryEntries := []*models.Entry{
			{ID: 1, Title: "Entry 1", GroupID: 101, GroupTitle: "Category 1", Date: time.Now()},
			{ID: 2, Title: "Entry 2", GroupID: 101, GroupTitle: "Category 1", Date: time.Now().Add(-25 * time.Hour)},
		}
		llmService := &mockLLMService{
			GenerateContentFunc: func(ctx context.Context, prompt string, schema *genai.Schema) ([]byte, error) {
				return nil, errors.New("llm error")
			},
		}

		groups := BuildDigestDataByAI(singleCategoryEntries, context.Background(), llmService)
		assert.Len(t, groups, 2, "Should fall back to 2 date groups")
	})
}

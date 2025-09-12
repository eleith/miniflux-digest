package view

import (
	"context"
	"encoding/json"
	"errors"
	"miniflux-digest/internal/models"
	"miniflux-digest/internal/testutil"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
		mockResp := InitialGroupingResponse{
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

		rawGroups, groupedIDs, err := GroupAIEntries(context.Background(), entries, llmService)
		assert.NoError(t, err)
		assert.NotNil(t, rawGroups)
		assert.Len(t, rawGroups, 2, "Expected 2 grouped")

		assert.Contains(t, rawGroups, "Group 1")
		assert.Len(t, rawGroups["Group 1"], 2)
		assert.Contains(t, rawGroups, "Group 2")
		assert.Len(t, rawGroups["Group 2"], 1)

		assert.Len(t, groupedIDs, 3)
		assert.True(t, groupedIDs[1])
		assert.True(t, groupedIDs[2])
		assert.True(t, groupedIDs[3])
		assert.False(t, groupedIDs[4])
	})

	t.Run("llm service error", func(t *testing.T) {
		llmService := &mockLLMService{
			GenerateContentFunc: func(ctx context.Context, prompt string, schema *genai.Schema) ([]byte, error) {
				return nil, errors.New("llm error")
			},
		}

		_, _, err := GroupAIEntries(context.Background(), entries, llmService)
		assert.Error(t, err)
	})

	t.Run("llm returns invalid json", func(t *testing.T) {
		llmService := &mockLLMService{
			GenerateContentFunc: func(ctx context.Context, prompt string, schema *genai.Schema) ([]byte, error) {
				return []byte("invalid json"), nil
			},
		}

		_, _, err := GroupAIEntries(context.Background(), entries, llmService)
		assert.Error(t, err)
	})

	t.Run("handles chunking", func(t *testing.T) {
		var largeEntries []*models.Entry
		for i := 0; i < 205; i++ { // 2 chunks: 200 and 5
			largeEntries = append(largeEntries, &models.Entry{ID: int64(i)})
		}

		var callCount int
		var mu sync.Mutex

		llmService := &mockLLMService{
			GenerateContentFunc: func(ctx context.Context, prompt string, schema *genai.Schema) ([]byte, error) {
				mu.Lock()
				callCount++
				mu.Unlock()

				var resp InitialGroupingResponse
				if callCount == 1 { // First chunk
					resp.Groups = []struct {
						Title    string  `json:"title"`
						EntryIDs []int64 `json:"entry_ids"`
					}{
						{Title: "Chunk 1 Group", EntryIDs: []int64{0, 1}},
					}
				} else { // Second chunk
					resp.Groups = []struct {
						Title    string  `json:"title"`
						EntryIDs []int64 `json:"entry_ids"`
					}{
						{Title: "Chunk 2 Group", EntryIDs: []int64{200, 201}},
					}
				}
				mockJSON, _ := json.Marshal(resp)
				return mockJSON, nil
			},
		}

		rawGroups, groupedIDs, err := GroupAIEntries(context.Background(), largeEntries, llmService)
		assert.NoError(t, err)
		assert.Equal(t, 2, callCount, "Expected LLM to be called 2 times for 205 entries with chunk size 200")

		assert.Len(t, rawGroups, 2)
		assert.Contains(t, rawGroups, "Chunk 1 Group")
		assert.Len(t, rawGroups["Chunk 1 Group"], 2)
		assert.Contains(t, rawGroups, "Chunk 2 Group")
		assert.Len(t, rawGroups["Chunk 2 Group"], 2)
		assert.Len(t, groupedIDs, 4)
	})
}

func TestConsolidatePrimaryGroups(t *testing.T) {
	t.Run("happy path - consolidation occurs", func(t *testing.T) {
		rawGroups := map[string][]*models.Entry{
			"AI News":                  {{ID: 1}},
			"Machine Learning Updates": {{ID: 2}},
			"Go Programming":           {{ID: 3}},
			"Random Stuff":             {{ID: 4}},
			"Group 5":                  {{ID: 5}},
			"Group 6":                  {{ID: 6}},
			"Group 7":                  {{ID: 7}},
			"Group 8":                  {{ID: 8}},
			"Group 9":                  {{ID: 9}},
			"Group 10":                 {{ID: 10}},
			"Unrelated Group":          {{ID: 11}},
		}

		consolidationResp := ConsolidationResponse{
			ConsolidatedGroups: []struct {
				NewTitle  string   `json:"new_title"`
				OldTitles []string `json:"old_titles"`
			}{
				{NewTitle: "Artificial Intelligence", OldTitles: []string{"AI News", "Machine Learning Updates"}},
				{NewTitle: "Software Development", OldTitles: []string{"Go Programming"}},
				{NewTitle: "Miscellaneous", OldTitles: []string{"Random Stuff", "Group 5", "Group 6", "Group 7", "Group 8", "Group 9", "Group 10"}},
			},
		}
		mockJSON, _ := json.Marshal(consolidationResp)

		llmService := &mockLLMService{
			GenerateContentFunc: func(ctx context.Context, prompt string, schema *genai.Schema) ([]byte, error) {
				assert.Equal(t, ConsolidationResponseSchema, schema)
				return mockJSON, nil
			},
		}

		consolidated, err := consolidatePrimaryGroups(context.Background(), rawGroups, llmService)
		assert.NoError(t, err)
		assert.Len(t, consolidated, 4, "Expected 3 consolidated groups + 1 unmapped group")

		aiGroup := testutil.FindPrimaryGroup(consolidated, "Artificial Intelligence")
		require.NotNil(t, aiGroup)
		assert.Equal(t, 2, aiGroup.TotalEntries)

		devGroup := testutil.FindPrimaryGroup(consolidated, "Software Development")
		require.NotNil(t, devGroup)
		assert.Equal(t, 1, devGroup.TotalEntries)

		// Test that unmapped groups are preserved
		unrelatedGroup := testutil.FindPrimaryGroup(consolidated, "Unrelated Group")
		require.NotNil(t, unrelatedGroup)
		assert.Equal(t, 1, unrelatedGroup.TotalEntries)
	})

	t.Run("llm service fails - fallback to original groups", func(t *testing.T) {
		rawGroups := map[string][]*models.Entry{
			"Group 1": {{ID: 1}},
			"Group 2": {{ID: 2}},
		}
		llmService := &mockLLMService{
			GenerateContentFunc: func(ctx context.Context, prompt string, schema *genai.Schema) ([]byte, error) {
				return nil, errors.New("llm error")
			},
		}

		consolidated, err := consolidatePrimaryGroups(context.Background(), rawGroups, llmService)
		assert.NoError(t, err, "Fallback should not produce an error")
		assert.Len(t, consolidated, 2, "Should fall back to the original number of groups")
	})

	t.Run("too few groups - no consolidation needed", func(t *testing.T) {
		smallRawGroups := map[string][]*models.Entry{"Group 1": {{ID: 1}}}
		llmService := &mockLLMService{
			GenerateContentFunc: func(ctx context.Context, prompt string, schema *genai.Schema) ([]byte, error) {
				t.Fatal("LLM service should not be called when there are too few groups")
				return nil, nil
			},
		}

		consolidated, err := consolidatePrimaryGroups(context.Background(), smallRawGroups, llmService)
		assert.NoError(t, err)
		assert.Len(t, consolidated, 1)
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
		summaryResp := SummaryResponse{Summary: "This is a summary."}
		summaryJSON, _ := json.Marshal(summaryResp)

		subGroupingResp := SubGroupingResponse{
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

	// ... other tests for ProcessPrimaryGroupWithAI are unchanged ...
}

func TestBuildDigestDataByAI(t *testing.T) {
	entries := []*models.Entry{
		{ID: 1, Title: "E1", GroupID: 101, GroupTitle: "C1", Date: time.Now()},
		{ID: 2, Title: "E2", GroupID: 101, GroupTitle: "C1", Date: time.Now()},
		{ID: 3, Title: "E3", GroupID: 102, GroupTitle: "C2", Date: time.Now()},
		{ID: 4, Title: "E4", GroupID: 103, GroupTitle: "C3", Date: time.Now()},
	}

	t.Run("happy path with consolidation", func(t *testing.T) {
		// 1. Initial Grouping Response
		groupingResp := InitialGroupingResponse{
			Groups: []struct {
				Title    string  `json:"title"`
				EntryIDs []int64 `json:"entry_ids"`
			}{
				{Title: "AI News", EntryIDs: []int64{1, 2}},
				{Title: "Dev Stuff", EntryIDs: []int64{3}},
				{Title: "Group 3", EntryIDs: []int64{}},
				{Title: "Group 4", EntryIDs: []int64{}},
				{Title: "Group 5", EntryIDs: []int64{}},
				{Title: "Group 6", EntryIDs: []int64{}},
				{Title: "Group 7", EntryIDs: []int64{}},
				{Title: "Group 8", EntryIDs: []int64{}},
				{Title: "Group 9", EntryIDs: []int64{}},
				{Title: "Group 10", EntryIDs: []int64{}},
				{Title: "Group 11", EntryIDs: []int64{}},
			},
		}
		groupingJSON, _ := json.Marshal(groupingResp)

		// 2. Consolidation Response
		consolidationResp := ConsolidationResponse{
			ConsolidatedGroups: []struct {
				NewTitle  string   `json:"new_title"`
				OldTitles []string `json:"old_titles"`
			}{
				{NewTitle: "Tech", OldTitles: []string{"AI News", "Dev Stuff", "Group 3", "Group 4", "Group 5", "Group 6", "Group 7", "Group 8", "Group 9", "Group 10", "Group 11"}},
			},
		}
		consolidationJSON, _ := json.Marshal(consolidationResp)

		// 3. Summary & Sub-grouping Response
		summaryResp := SummaryResponse{Summary: "Tech summary"}
		summaryJSON, _ := json.Marshal(summaryResp)
		subGroupingResp := SubGroupingResponse{
			SubGroups: []struct {
				Title    string  `json:"title"`
				EntryIDs []int64 `json:"entry_ids"`
			}{{Title: "Tech Sub-group", EntryIDs: []int64{1, 2, 3}}},
		}
		subGroupingJSON, _ := json.Marshal(subGroupingResp)

		llmService := &mockLLMService{
			GenerateContentFunc: func(ctx context.Context, prompt string, schema *genai.Schema) ([]byte, error) {
				switch schema {
				case InitialGroupingResponseSchema:
					return groupingJSON, nil
				case ConsolidationResponseSchema:
					return consolidationJSON, nil
				case SummaryResponseSchema:
					return summaryJSON, nil
				case SubGroupingResponseSchema:
					return subGroupingJSON, nil
				default:
					return nil, errors.New("unexpected schema")
				}
			},
		}

		groups := BuildDigestDataByAI(entries, context.Background(), llmService)
		assert.Len(t, groups, 2, "Expected 1 consolidated group + 1 uncategorized")

		techGroup := testutil.FindPrimaryGroup(groups, "Tech")
		require.NotNil(t, techGroup)
		assert.Equal(t, "Tech summary", techGroup.Summary)
		assert.Equal(t, 3, techGroup.TotalEntries)
		assert.Len(t, techGroup.SubGroups, 1)
		assert.Equal(t, "Tech Sub-group", techGroup.SubGroups[0].Title)

		uncategorized := testutil.FindPrimaryGroup(groups, "Uncategorized")
		require.NotNil(t, uncategorized)
		assert.Equal(t, 1, uncategorized.TotalEntries)
		assert.Equal(t, int64(4), uncategorized.SubGroups[0].Entries[0].ID)
	})

	t.Run("fallback to category grouping", func(t *testing.T) {
		llmService := &mockLLMService{
			GenerateContentFunc: func(ctx context.Context, prompt string, schema *genai.Schema) ([]byte, error) {
				return nil, errors.New("llm error")
			},
		}

		groups := BuildDigestDataByAI(entries, context.Background(), llmService)
		assert.Len(t, groups, 3, "Should fall back to 3 categories")
	})

	// ... other fallback tests are similar and should still pass ...
}
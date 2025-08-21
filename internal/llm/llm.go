package llm

type LLMResponse struct {
	OverviewSummary string `json:"summary"`
	GroupSummaries  []struct {
		Title   string `json:"title"`
		Summary string `json:"summary"`
		EntryIDs []int64 `json:"entry_ids"`
	} `json:"groups"`
}

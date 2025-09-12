package view

import "google.golang.org/genai"

const initialGroupingPrompt = `You are creating high-level primary groups for a digest.

**Your Task:**
Group the entries below into broad, high-level primary groups.

The nature of a good primary group can vary:
- Sometimes it's a broad **topic** (e.g., "AI Development", "Global Economics").
- Sometimes it's based on the **source** (e.g., "Hacker News Discussions", "TechCrunch Articles").
- Sometimes it's about the **type of content** (e.g., "Software Development Blogs", "Video Game Reviews").

Follow these instructions:
- Create group titles that are concise and high-level (2-5 words).
- An entry can only belong to one group.

Below is the list of entries:
----------------------------
`

const consolidationPrompt = `You are an expert content curator responsible for organizing a digest. You have been given a list of proposed primary group titles that were automatically generated. Your task is to ensure that this list is broad and high level and does not contain potentially overlapping topics. Each topic will later on be further subdivided into sub-groups, so if two topics are better served as sub-topics, they should be merged into one high level topic. Our goal is to have no more than 10-12 primary groups.

**Your Goal:**
Generate high level primary groups by possibly merging similar or related group titles to produce a final list of no more than 10-12 primary groups.

**Instructions:**
1.  Review the list of "Original Group Titles" below.
2.  Identify titles that are duplicates, closely related, or could be merged into a broader, more meaningful category.
3.  Create a new, high-level title for each consolidated group.
4.  A group that is already distinct and high-level can be kept as is, with its original title becoming the "new_title".

**Example:**
If the original titles are:
- "AI Startups"
- "Machine Learning Research"
- "Tech Company Earnings"
- "Stock Market News"
- "Go Programming Language"

A good consolidation would be:
- new_title: "Artificial Intelligence", old_titles: ["AI Startups", "Machine Learning Research"]
- new_title: "Finance & Markets", old_titles: ["Tech Company Earnings", "Stock Market News"]
- new_title: "Software Development", old_titles: ["Go Programming Language"]

**Original Group Titles:**
{{range .}}
- {{.}}
{{end}}
`

const summaryPrompt = `You are an expert at summarizing content for busy readers. You will be given a list of entries from a primary group titled '%s'.

Your task is to write a concise, 2-3 sentence summary (under 150 words) that achieves two goals:
1.  **Provide a high-level overview** of the main themes in the group.
2.  **Highlight the most significant or surprising entries.** This could be a major announcement, a controversial opinion, or a particularly popular discussion.

The goal is to give the reader enough information to quickly decide if this group contains entries they want to explore further. Do not just list the topics; provide some insight into the content.
`

const subGroupingPrompt = `You are an expert at organizing content. You will be given a list of entries that are already part of a primary group titled '%s'.

Your task is to create smaller, more granular sub-groups to help a user navigate the content. A good sub-group contains entries that are very closely related.

- Create sub-groups that represent a specific **story, product, or discussion thread**.
- The sub-group titles should be very specific and descriptive (3-7 words). For example, instead of "AI News", a good sub-group title might be "Gemini 1.5 Pro Announcement" or "Discussion on AI Safety".
- Aim for smaller, tight-knit groups of 2 to 15 entries.
- An entry can only belong to one sub-group.
- If an entry doesn't fit into a specific sub-group, leave it out of your response.
`

// --- Prompt Response Types ---

type InitialGroupingResponse struct {
	Groups []struct {
		Title    string  `json:"title"`
		EntryIDs []int64 `json:"entry_ids"`
	} `json:"groups"`
}

var InitialGroupingResponseSchema = &genai.Schema{
	Type: genai.TypeObject,
	Properties: map[string]*genai.Schema{
		"groups": {
			Type: genai.TypeArray,
			Items: &genai.Schema{
				Type: genai.TypeObject,
				Properties: map[string]*genai.Schema{
					"title": {
						Type: genai.TypeString,
					},
					"entry_ids": {
						Type: genai.TypeArray,
						Items: &genai.Schema{
							Type: genai.TypeInteger,
						},
					},
				},
			},
		},
	},
}

type ConsolidationResponse struct {
	ConsolidatedGroups []struct {
		NewTitle  string   `json:"new_title"`
		OldTitles []string `json:"old_titles"`
	} `json:"consolidated_groups"`
}

var ConsolidationResponseSchema = &genai.Schema{
	Type: genai.TypeObject,
	Properties: map[string]*genai.Schema{
		"consolidated_groups": {
			Type: genai.TypeArray,
			Items: &genai.Schema{
				Type: genai.TypeObject,
				Properties: map[string]*genai.Schema{
					"new_title": {
						Type: genai.TypeString,
					},
					"old_titles": {
						Type: genai.TypeArray,
						Items: &genai.Schema{
							Type: genai.TypeString,
						},
					},
				},
			},
		},
	},
}

type SummaryResponse struct {
	Summary string `json:"summary"`
}

var SummaryResponseSchema = &genai.Schema{
	Type: genai.TypeObject,
	Properties: map[string]*genai.Schema{
		"summary": {
			Type: genai.TypeString,
		},
	},
}

type SubGroupingResponse struct {
	SubGroups []struct {
		Title    string  `json:"title"`
		EntryIDs []int64 `json:"entry_ids"`
	} `json:"sub_groups"`
}

var SubGroupingResponseSchema = &genai.Schema{
	Type: genai.TypeObject,
	Properties: map[string]*genai.Schema{
		"sub_groups": {
			Type: genai.TypeArray,
			Items: &genai.Schema{
				Type: genai.TypeObject,
				Properties: map[string]*genai.Schema{
					"title": {
						Type: genai.TypeString,
					},
					"entry_ids": {
						Type: genai.TypeArray,
						Items: &genai.Schema{
							Type: genai.TypeInteger,
						},
					},
				},
			},
		},
	},
}
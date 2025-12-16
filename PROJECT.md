# Project: Multi-Digest Architecture

## Overview
Transform `miniflux-digest` from a single-task reporter into a multi-digest engine. This allows users to configure specific digests (e.g., "Daily Tech", "Weekly Social") with unique schedules, content filters, and distinct LLM categorization.

**Note:** Changes to the Email Template are explicitly **out of scope** for this project.

## Core Architectural Principles

1.  **One Run, Many Outputs:** The system processes multiple defined digests.
2.  **Stack-Based Priority:** The order of digests in the configuration array determines priority. The first digest (index 0) has the highest priority.
    *   *Scenario:* Item A matches Digest X (Index 0) and Digest Y (Index 1).
    *   *Behavior:* Item A belongs to Digest X. Digest Y will never see it.
3.  **Deterministic Filtering:** Inclusion is based on rules (Feed Title, Category Title, Site URL, Entry URL).
    *   **Simple Mode:** Exact match (Titles) or Prefix match (URLs).
    *   **Power Mode:** Regex patterns (suffix `_patterns`).
4.  **Context-Aware Categorization:** The LLM receives digest-specific categories (with descriptions).
    *   If no categories are provided, **Default Categories** are injected.
    *   If categories are provided, they fully override the defaults.

## Configuration Schema (Target)

```yaml
# Global Settings
miniflux: ...
llm: ...

# New "Digests" Array (Order = Priority)
digests:
  - title: "Morning Tech" # Highest Priority
    schedule: "0 8 * * *"
    filters:
      # Simple Matchers (OR logic within list, AND logic between fields)
      feed_titles: ["Miniflux News"] # Exact Match
      category_titles: ["Tech", "Programming"] # Exact Match
      site_urls: ["https://github.com", "https://ycombinator.com"] # Prefix Match
      entry_urls: ["https://blog.miniflux.app/post/"] # Prefix Match
      
      # Power Matchers (Regex)
      feed_title_patterns: [".*Tech News.*"] 
      site_url_patterns: ["https://(www.)?reddit.com/r/.*"]
    categories: 
      - title: "AI & ML"
        description: "News about artificial intelligence"
  
  - title: "Weekly Social"
    schedule: "0 18 * * 5"
    filters:
      feed_titles: ["Twitter", "Mastodon"]
      # uses default LLM categories since none are provided
```

## Implementation Plan

### Phase 1: Configuration & Core Structures
*Goal: Update data structures on a new branch `gemini/feature/multi-digest-architecture`.*

- [x] **1.1 Config Refactor:** Update `internal/config`.
    -   Update `Config` struct to hold `[]DigestConfig`.
    -   Define `FilterConfig` with fields: `FeedTitles`, `CategoryTitles`, `SiteURLs`, `EntryURLs` (Simple) and `*_Patterns` (Regex).
    -   Implement `Load` logic: Check for legacy `digest` (singular) and migrate it to `digests[0]`.
    -   **Test:** Unit tests for `LoadConfig` verifying legacy, new array, and mixed filter formats.
- [x] **1.2 Slugification:** Ensure `utils` has a robust slug generator (for file/folder naming based on titles).
    -   **Test:** Unit tests for slug generation with edge cases.

### Phase 2: The Filter Engine
*Goal: Deterministic logic to assign items to digests.*

- [x] **2.1 Filter Logic:** Create `internal/filter`.
    -   `Matcher` struct.
    -   `Matches(item, digestConfig) bool`.
    -   Logic:
        -   `FeedTitles`, `CategoryTitles`: **Exact String Match**.
        -   `SiteURLs`, `EntryURLs`: **String Prefix Match** (`strings.HasPrefix`).
        -   `*_Patterns`: **Regex Match**.
    -   **Test:** Unit tests with various inputs (simple vs regex, matches vs non-matches).
- [x] **2.2 Assignment Logic:** Create `internal/manager`.
    -   `GetOwningDigest(item, allDigests) digestIndex`.
    -   Iterates through the stack. Returns the index of the first match.
    -   **Test:** Unit tests mimicking the "stealing" scenario.

### Phase 3: Processor & Scheduler Refactor
*Goal: Run specific jobs based on specific configs.*

- [x] **3.1 Processor Update:** Modify `internal/processor/processor.go`.
    -   Update `ProcessDigest` to accept `DigestConfig`.
    -   **Optimization:** If a digest *only* filters by specific Miniflux Category IDs (mapped from titles), pass that to `miniflux_client.UnreadEntries` to reduce fetch size. Otherwise, fetch all unread.
    -   **Crucial:** Inside the loop, check `GetOwningDigest`. If item owner != current digest, **skip** and **do not mark read**.
    -   **Test:** Mock Miniflux client to verify correct items are marked read/skipped.
- [x] **3.2 Main Loop:** Update `cmd/miniflux-digest/main.go`.
    -   Iterate `config.Digests`. Register a job for each unique schedule.
    -   **Test:** Verify (via log inspection or unit test) that multiple jobs are registered.

### Phase 4: LLM & Categorization
*Goal: Make the AI respect user-defined buckets.*

- [x] **4.1 Default Handling:** Define `DefaultCategories` constant.
    -   In `Config` loading, if `digest.Categories` is empty, apply defaults.
- [x] **4.2 Prompt Update:** Update `internal/llm/prompts.go`.
    -   Inject categories + descriptions into the system prompt.
    -   **Test:** Update `llm_test.go` to verify prompts are constructed correctly with custom categories.

### Phase 5: Output & Presentation
*Goal: New folder structure.*

- [x] **5.1 Folder Structure:** Update `internal/archive`.
    -   Path: `archives/<digest-slug>/<date>.html`.
    -   **Test:** Verify file creation in correct subfolders.
- [x] **5.2 Cleanup:** Remove legacy code.

### Phase 6: End-to-End Verification
- [ ] **6.1 Manual Dry Run:**
    -   Configure 2 digests locally.
    -   Run with a mock Miniflux or local instance.
    -   Verify logs: Item A assigned to Digest 1, Item B assigned to Digest 2.
    -   Verify outputs: 2 distinct emails (standard format), 2 distinct HTML files.


## Session Checkpoint (2025-12-16)

### Status
Phase 1 (Configuration & Core Structures) is mostly complete. The codebase has been refactored to support multiple digests.

### Key Changes
1.  **Configuration (internal/config):**
    -   Renamed Config.Digest to Config.Digests ([]ConfigDigest).
    -   Added strict validation tags (required, url, gocron) to ConfigDigest fields.
    -   Updated Load to use strict validation.
2.  **Execution (cmd/miniflux-digest):**
    -   Updated main.go to iterate through Digests and register a job for each.
3.  **Processing (internal/processor, internal/email):**
    -   Updated ProcessDigest and Send to accept a specific ConfigDigest.
4.  **Testing (internal/config/config_test.go):**
    -   Updated tests to reflect the new slice structure.

### Immediate Next Steps
1.  **Run Tests:** Execute 'go test -mod=vendor ./...' to verify that the stricter validation in config.go matches the test expectations in config_test.go. Fix any failures.
2.  **Phase 2:** Implement the Filter Engine as per the plan above.

# Configuration Reference

The `miniflux-digest` application is configured via a YAML file (default: `config.yaml`). This document details all available options, their default values, and validation rules.

## Root Object

| Field | Type | Required | Description |
| :--- | :--- | :--- | :--- |
| `miniflux` | Object | **Yes** | Connection settings for your Miniflux instance. |
| `smtp` | Object | No | SMTP server settings for sending emails. |
| `ai` | Object | Conditional | Required if any digest uses `view: ai`. |
| `digests` | Array | **Yes** | A list of digest configurations. At least one must be defined. |

---

## Miniflux (`miniflux`)

| Field | Type | Description |
| :--- | :--- | :--- |
| `host` | String | **Required.** The full URL to your Miniflux instance (e.g., `https://reader.miniflux.app`). |
| `api_token` | String | **Required.** Your Miniflux API Key. Create one in Miniflux under **Settings > API Keys**. |

## SMTP (`smtp`)

Required only if you want to receive digests via email.

| Field | Type | Default | Description |
| :--- | :--- | :--- | :--- |
| `host` | String | - | The hostname of your SMTP server (e.g., `smtp.gmail.com`). |
| `port` | Int | `587` | The SMTP port (1-65535). |
| `user` | String | - | SMTP username. |
| `password` | String | - | SMTP password. |

## AI (`ai`)

| Field | Type | Description |
| :--- | :--- | :--- |
| `api_key` | String | **Required if using AI View.** Your Google Gemini API Key. |

---

## Digests (`digests`)

You can define multiple digests. The order matters!

### Priority & Ownership
The application processes digests in the order they appear in the configuration list.
*   **First Match Wins:** The first digest that matches an entry "claims" it.
*   **No Duplicates:** An entry will never appear in more than one digest.
*   **Catch-All:** It is common to define specific digests first (e.g., "Tech News") and a broad, filter-less digest last to catch everything else.

### Digest Object

| Field | Type | Required | Description |
| :--- | :--- | :--- | :--- |
| `title` | String | **Yes** | The display title of the digest. Must be unique after "slugification" (e.g., "My Digest" and "My-Digest" conflict). |
| `schedule` | String | **Yes** | When to run the digest. Supports Cron syntax (`0 8 * * *`) or Go-cron descriptors (`@daily`, `@weekly`, `@every 2h`). |
| `host` | String | **Yes** | The base URL where the static HTML archives will be hosted (e.g., `https://digest.myserver.com`). Used for links in the email. |
| `view` | String | **Yes** | One of: `date`, `category`, `ai`. Controls how entries are grouped and presented. |
| `email` | Object | No | Email settings for this specific digest. |
| `filters` | Object | No | Rules for including entries in this digest. |
| `categories`| Array | No | **(AI View Only)** Custom categories to guide the LLM. |
| `compress` | Bool | No | If `true` (default), minifies the generated HTML. |
| `mark_as_read`| Bool | No | If `true`, marks entries as read in Miniflux *after* processing. Default: `false`. |
| `run_on_startup`| Bool | No | If `true`, runs the digest job immediately when the application starts. Default: `false`. |

### Email Object (`email`)

| Field | Type | Description |
| :--- | :--- | :--- |
| `to` | String | Recipient email address. |
| `from` | String | Sender email address. |

### Filters Object (`filters`)

All conditions within a list are **OR** (e.g., matching *any* feed title).
Conditions between different fields are **AND** (e.g., matching feed title *AND* site URL).

| Field | Type | Matching Logic | Description |
| :--- | :--- | :--- | :--- |
| `feed_titles` | String[] | Exact | Matches the exact title of the RSS feed. |
| `category_titles` | String[] | Exact | Matches the exact title of the Miniflux category. |
| `site_urls` | String[] | Prefix | Matches if the feed's site URL starts with this string. |
| `entry_urls` | String[] | Prefix | Matches if the entry's URL starts with this string. |
| `feed_title_patterns` | String[] | Regex | Go-flavor Regex match against feed title. |
| `category_title_patterns` | String[] | Regex | Go-flavor Regex match against category title. |
| `site_url_patterns` | String[] | Regex | Go-flavor Regex match against site URL. |
| `entry_url_patterns` | String[] | Regex | Go-flavor Regex match against entry URL. |

### Categories Object (`categories`)

Used only when `view: ai`. Provide a list of objects with `title` and `description` to help the AI categorize content.

**Default Categories:**
If you do not provide this block, the following defaults are injected:
1.  **Technology & Engineering:** Software, hardware, internet, gadgets, and engineering breakthroughs.
2.  **World News & Politics:** Global events, international relations, policy changes, and political discourse.
3.  **Business & Finance:** Markets, companies, startups, economy, and financial advice.
4.  **Science & Health:** Scientific discoveries, space, medicine, health tips, and environmental news.
5.  **Arts & Culture:** Books, music, movies, history, philosophy, and cultural analysis.
6.  **Entertainment & Lifestyle:** Celebrities, gaming, travel, food, fashion, and hobbies.
7.  **Sports:** Sports news, scores, teams, and athletes.
8.  **Other:** Anything that doesn't fit well into the other categories.

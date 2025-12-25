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

| Field | Type | Default | Description |
| :--- | :--- | :--- | :--- |
| `host` | String | - | **Required.** The full URL to your Miniflux instance (e.g., `https://reader.miniflux.app`). |
| `api_token` | String | - | **Required.** Your Miniflux API Key. Create one in Miniflux under **Settings > API Keys**. |

## SMTP (`smtp`)

Required only if you want to receive digests via email.

| Field | Type | Default | Description |
| :--- | :--- | :--- | :--- |
| `host` | String | - | The hostname of your SMTP server (e.g., `smtp.gmail.com`). **Required** if any digest has email configured. |
| `port` | Int | `587` | The SMTP port (1-65535). |
| `user` | String | - | SMTP username. |
| `password` | String | - | SMTP password. |

## AI (`ai`)

| Field | Type | Default | Description |
| :--- | :--- | :--- | :--- |
| `api_key` | String | - | **Required if using AI View.** Your Google Gemini API Key. |

---

## Digests (`digests`)

You can define multiple digests. The order matters!

### Priority & Ownership
The application processes digests in the order they appear in the configuration list.
*   **First Match Wins:** The first digest that matches an entry "claims" it.
*   **No Duplicates:** An entry will never appear in more than one digest.
*   **Catch-All:** It is common to define specific digests first (e.g., "Tech News") and a broad, filter-less digest last to catch everything else.

### Digest Object

| Field | Type | Default | Required | Description |
| :--- | :--- | :--- | :--- | :--- |
| `title` | String | - | **Yes** | The display title of the digest. Must be unique after "slugification" (e.g., "My Digest" and "My-Digest" conflict). |
| `schedule` | String | - | **Yes** | When to run the digest. Supports Cron syntax (`0 8 * * *`) or Go-cron descriptors (`@daily`, `@weekly`, `@every 2h`). |
| `host` | String | - | No | The base URL where the static HTML archives will be hosted (e.g., `https://digest.myserver.com`). **Required** for correct links in emails. |
| `view` | String | `date` | No | One of: `date`, `category`, `ai`. Controls how entries are grouped and presented. |
| `email` | Object | - | No | Email settings for this specific digest. |
| `filters` | Object | - | No | Rules for including entries in this digest. |
| `categories`| Array | - | No | **(AI View Only)** Custom categories to guide the LLM. |
| `compress` | Bool | `true` | No | If `true`, minifies the generated HTML. |
| `mark_as_read`| Bool | `false` | No | If `true`, marks entries as read in Miniflux *after* processing. |
| `run_on_startup`| Bool | `false` | No | If `true`, runs the digest job immediately when the application starts. |
| `send_if_empty` | Bool | `true` | No | If `true`, sends an email (saying "No new entries") even if the digest is empty. If `false`, skips email sending for empty digests. |

### Email Object (`digests[].email`)

| Field | Type | Default | Description |
| :--- | :--- | :--- | :--- |
| `to` | String | - | Recipient email address. |
| `from` | String | - | Sender email address. |

### Filters Object (`digests[].filters`)

This section defines which entries are included in the digest.

| Field | Type | Default | Description |
| :--- | :--- | :--- | :--- |
| `logic` | String | `OR` | Logic for combining different field conditions. <br> - `OR`: Match if *any* field condition is met.<br> - `AND`: Match only if *all* defined field conditions are met. |
| `feed_titles` | String[] | - | Exact matches for the RSS feed title. |
| `category_titles` | String[] | - | Exact matches for the Miniflux category title. |
| `site_urls` | String[] | - | Prefix matches for the site URL. (Legacy: `feed_urls`) |
| `entry_urls` | String[] | - | Prefix matches for the entry's URL. |
| `feed_title_patterns` | String[] | - | Regex matches for the feed title. |
| `category_title_patterns` | String[] | - | Regex matches for the category title. |
| `site_url_patterns` | String[] | - | Regex matches for the site URL. (Legacy: `feed_url_patterns`) |
| `entry_url_patterns` | String[] | - | Regex matches for the entry URL. |

**Note:** Conditions *within* a single list (e.g., multiple `feed_titles`) are always treated as **OR**.

### Categories Object (`digests[].categories`)

Used only when `view: ai`. Provide a list of objects with `title` and `description` to help the AI categorize content.

If you do not provide this block, the AI will automatically determine the best primary groups for your content.
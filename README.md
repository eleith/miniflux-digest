# Miniflux Digest

## Summary

HTML digests (email, web or both) for Miniflux Categories.

[![status-badge](https://ci.eleith.com/api/badges/26/status.svg)](https://ci.eleith.com/repos/26)

### Overview

This tool transforms Miniflux RSS consumption from a "pull" to a "push" model.

It fetches entries from your Miniflux categories, delivering them as personalized
HTML email digests. These digests are also stored as static HTML files that you
can make available on the web.

Entries are automatically marked as read, and the process re-runs on a
user-defined schedule.

## Features

* ⏰ Automated scheduling via cron syntax
* 📥 Fetches unread entries per Miniflux category
* 📧 Delivers personalized HTML digests via email
* 🛜 Archives HTML digests for static web serving
* ✅ Automatically marks entries as read in Miniflux
* 🧹 Manages storage by purging old archives
* ♻️ Wash, rinse, repeat

## Installation

You can get up and running in just a few steps:

### Prerequisites

* [Docker](https://docs.docker.com/get-docker/)
* An active [Miniflux](https://miniflux.app/) account
* A Miniflux API Key (Settings > API Keys > Create a new API key)

### Clone and Build

```bash
git clone https://git.eleith.com/eleith/miniflux-digest.git  
cd miniflux-digest
docker compose build
```

### Configuration Setup

A [config.yaml.example](config.yaml.example) file is provided in the project
root. Copy this file to `config.yaml`:

```bash
cp config.yaml.example config.yaml
```

Then, edit it as described in the [Configuration](#configuration) section.

### Start

Run the container:

```bash
docker-compose up -d
```

The service will now:

* Build the Docker image (if not already built).
* Mount your config file and archive folder.
* Start the main digest service.

Now have some ☕️, 🍵, 🧋 or a tall glass of water.

Let the feeds come to you.

Not the other way around.

### Stop

To stop the running service:

```bash
docker compose stop
```

## Configuration

### Config.yaml

A `config.yaml` file is required for operation.

Create this file in the root of the project directory and fill in the fields:

```yaml
miniflux:
  host: "YOUR_MINIFLUX_URL"
  api_token: "YOUR_MINIFLUX_API_API_KEY"

smtp:
  host: "YOUR_SMTP_HOST"
  port: 587
  user: "YOUR_SMTP_USERNAME"
  password: "YOUR_SMTP_PASSWORD"

digest:
  email:
    to: "RECIPIENT_EMAIL@example.com"
    from: "SENDER_EMAIL@example.com"
  schedule: "@every 1w" # cron syntax also supported
  host: "https://your-digest-host.com" # optional
```

### Docker Compose

Customize `docker-compose.yml` settings to your liking if you prefer a different
user or different locations to store your config file and archive folder.

```yaml
services:
  miniflux-digest:
    user: "1001:1001"
    volumes:
      - ./my-custom-config.yaml:/app/config.yaml:ro
      - ./my-custom-archive-folder:/app/web/miniflux-archive
  restart: unless-stopped
```

## License

This project is [licensed](LICENSE.md) under the [Apache License, Version
2.0](https://www.apache.org/licenses/LICENSE-2.0), aligning with the Miniflux
project's license.

## Contact

You can find me on:

* 🐘 [Mastodon](https://toot.eleith.com/@eleith)
* 🦋 [Bluesky](https://bsky.app/profile/eleith.com)

## Acknowledgements

* 🙏 The [Miniflux project](https://github.com/miniflux/v2) for showing us the
light after google reader's demise.

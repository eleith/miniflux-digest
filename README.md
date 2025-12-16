# Miniflux Digest

## Summary

Miniflux digests, an antidote to my doom scrolling.

[![ghcr build](https://github.com/eleith/miniflux-digest/actions/workflows/build.yml/badge.svg)](https://github.com/eleith/miniflux-digest/actions/workflows/build.yml)

### Overview

This tool creates digests from Miniflux RSS feed entries.

The goal is to give you more control of when and how you consume updates and
ongoings from the parts of the web you care about.

Unread entries are fetched on a user defined schedule and delivered to the web
and your inbox.

Customize the digest view, schedule and more. Oh my.

## Features

> [!NOTE]
> many of the following features are _optional_

* ⏰ Automated scheduling (ex: daily, weekly, ever third friday)
* 📚 **Multi-Digest Support:** Configure multiple digests (e.g., "Daily Tech", "Weekly Social") with unique schedules and content filters.
* 🌞 Dark and light themes
* 📥 Fetches all unread entries
* 📧 Delivers personalized HTML digests via email
* 🛜 Archives HTML digests for static web serving
* 🤖 Summarize and group entries for faster skimming (AI view)
* ✅ Automatically marks entries as read
* 🧹 Manages storage by purging older digests
* ♻️ Wash, rinse, repeat

## Installation

> [!NOTE]
> The following instructions focus on Docker-based deployment, as this is the
> most straight forward method.

### Prerequisites

* [Docker](https://docs.docker.com/get-docker/)
* A [Miniflux](https://miniflux.app/) API Key (Settings > API Keys > Create)
* An email account

### Setup

1. **Create a Project Directory**

   Create a directory on your system for the project.

   ```bash
   mkdir miniflux-digest
   cd miniflux-digest
   ```

2. **Create a `docker-compose.yml` File**

   Create a `docker-compose.yml` file with the following content. This example
   uses the `latest` tag, but you can pin to a specific version like `0.0.10`.

   ```yaml
   services:
     miniflux-digest:
       image: ghcr.io/eleith/miniflux-digest:latest
       container_name: miniflux-digest
       restart: unless-stopped
       user: "1001:1001" # Optional: Set to your user/group ID
       volumes:
         - ./config.yaml:/app/config.yaml:ro
         - ./archive:/app/web/archive
   ```

3. **Create a Configuration File**

   A `config.yaml` file is required for operation.

   Create this file in the root of the project directory. You can start by copying the [example](config.yaml.example):

   ```bash
   cp config.yaml.example config.yaml
   ```

   **📚 Documentation:**
   For a detailed reference of all available options, including how to configure multiple digests and filters, please read [docs/CONFIG.md](docs/CONFIG.md).

### Run

Run the container:

```bash
docker-compose up -d
```

The service will now pull the Docker image and start the main digest service.

Now have some ☕️, 🍵, 🧋 or a tall glass of water.

Let the feeds come to you.

Not the other way around.

### Stop

To stop the running service:

```bash
docker compose stop
```

## License

This project is [licensed](LICENSE.md) under the [Apache License, Version
2.0](https://www.apache.org/licenses/LICENSE-2.0), aligning with the Miniflux
project's license.

## Wishlist / Ideas / Roadmap / Todos

* support more LLM providers
* allow user to customize prompts
* support text/html email content
* fix web scrolling bug when selecting item content

## Contact

You can find me on:

* 🐘 [Mastodon](https://toot.eleith.com/@eleith)
* 🦋 [Bluesky](https://bsky.app/profile/eleith.com)

## Acknowledgements

* 🙏 The [Miniflux project](https://github.com/miniflux/v2) for showing us the
light after google reader's demise.

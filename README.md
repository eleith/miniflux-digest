# Miniflux Digest

## Summary

HTML digests (email, web or both) for Miniflux Categories.

[![build](https://ci.eleith.com/api/badges/26/status.svg)](https://ci.eleith.com/repos/26)
[![ghcr build](https://github.com/eleith/miniflux-digest/actions/workflows/build.yml/badge.svg)](https://github.com/eleith/miniflux-digest/actions/workflows/build.yml)

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

> [!NOTE]
> The following instructions focus on Docker-based deployment, as this is the
> most straight forward method.

### Prerequisites

* [Docker](https://docs.docker.com/get-docker/)
* An active [Miniflux](https://miniflux.app/) account
* A Miniflux API Key (Settings > API Keys > Create a new API key)

### Setup

1. **Create a Project Directory**

   Create a directory on your system for the project.

   ```bash
   mkdir miniflux-digest
   cd miniflux-digest
   ```

2. **Create a `docker-compose.yml` File**

   Create a `docker-compose.yml` file with the following content. This example
   uses the `latest` tag, but you can pin to a specific version like `0.0.8`.

   ```yaml
   services:
     miniflux-digest:
       image: ghcr.io/eleith/miniflux-digest:latest
       container_name: miniflux-digest
       restart: unless-stopped
       user: "1001:1001" # Optional: Set to your user/group ID
       volumes:
         - ./config.yaml:/app/config.yaml:ro
         - ./archive:/app/web/miniflux-archive
   ```

3. **Create a Configuration File**

   A `config.yaml` file is required for operation.

   Create this file in the root of the project directory and edit it.

   See the [config.yaml.example](config.yaml.example) to learn about
   requirements, defaults and other options.

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

## Contact

You can find me on:

* 🐘 [Mastodon](https://toot.eleith.com/@eleith)
* 🦋 [Bluesky](https://bsky.app/profile/eleith.com)

## Acknowledgements

* 🙏 The [Miniflux project](https://github.com/miniflux/v2) for showing us the
light after google reader's demise.

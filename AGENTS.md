# Agent Instructions

This document provides guidelines for AI agents interacting with the
`miniflux-digest` project.

## Project Overview

This tool creates digests from Miniflux RSS feed entries. The goal is to give
you more control of when and how you consume updates and ongoings from the part
of the web you care about. Unread entries are fetched on a user defined schedule
and delivered to the web and your inbox.

These entries can be sorted, summarized and group to enable skimming and to
decrease time spent catching up.

## Architecture

### Multi-Digest Engine

The application is designed to support multiple, distinct digests running in
parallel or on different schedules.

-   **One Run, Many Outputs:** The system processes multiple defined digests.
-   **Stack-Based Priority:** The order of digests in the configuration array
    determines priority. The first digest (index 0) has the highest priority. If
    an item matches Digest X (Index 0) and Digest Y (Index 1), it belongs to
    Digest X.
-   **Deterministic Filtering:** Inclusion is based on rules (Feed Title,
    Category Title, Site URL, Entry URL) using Exact, Prefix, or Regex matching.

## Development Workflow

### Writing Code

- Avoid writing extraneous comments. Code should be written to be readable. Only
leave comments to help explain what the code can not, or because a function is
complicated and the reader would want support in understanding it.

### Branching

- All work must be done on a branch.
- Branch names should follow the convention:
`gemini/(fix|refactor|feature)/<task-title>`.
- Never commit directly to the `main` branch.

### Environment

- The development environment is containerized and includes all necessary
tools.
- Use the `Makefile` for all project-specific commands.

### Quality Checks

- Run `make lint` to check for code style issues.
- Run `make test` to run the test suite.
- Run `make test-coverage` or `make test-coverage-full` to check test
coverage.
- During code writing, use `make lint && make test` since its faster

### Commits

- Commit messages should be concise and factual.
- The format should be a title, followed by two newlines, and then a description.
- Do not use backticks or embellishments in commit messages.

## Project Conventions

### Configuration

- The `config.yaml` file contains secrets and should not be read or modified
without explicit permission.
- Any file listed in `.gitignore` should be considered sensitive and not
accessed without permission.

### External Services

- The application interacts with external services (Miniflux, email).
- All network calls should be assumed to be fallible and must include robust
error handling.

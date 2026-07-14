# Gator - RSS Feed Aggregator CLI

Gator is a high-performance, concurrent command-line RSS feed aggregator written in Go and backed by a PostgreSQL database. It allows you to register users, follow multiple RSS feeds, and run a background worker thread that continuously scrapes and saves new blog posts right to your terminal.

---

## Prerequisites

Before running Gator, ensure you have the following software installed on your machine:

* **Go:** Version 1.22 or higher. [Download Go](https://go.dev/doc/install)
* **PostgreSQL:** An active Postgres server database instance. [Download PostgreSQL](https://www.postgresql.org/download/)
* **Goose (Optional but Recommended):** For managing database migrations cleanly.
  ```bash
  go install [github.com/pressly/goose/v3/cmd/goose@latest](https://github.com/pressly/goose/v3/cmd/goose@latest)

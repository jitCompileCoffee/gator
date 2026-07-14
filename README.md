# Gator - RSS Feed Aggregator CLI

Gator is a high-performance, concurrent command-line RSS feed aggregator written in Go and backed by a PostgreSQL database. It allows you to register users, follow multiple RSS feeds, and run a background worker thread that continuously scrapes and saves new blog posts right to your terminal.

---

## Prerequisites

Before running Gator, ensure you have the following software installed on your machine:

* **Go:** Version 1.22 or higher. [Download Go](https://go.dev/doc/install)
* **PostgreSQL:** An active Postgres server database instance. [Download PostgreSQL](https://www.postgresql.org/download/)
* **Goose:** For managing database migrations cleanly.

  ```bash
  go install github.com/pressly/goose/v3/cmd/goose@latest
  ```

* **SQLC (For Developers):** If you plan on changing SQL queries and regenerating Go code.

  ```bash
  go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest
  ```

---

## Installation & Local Development

### 1. Clone the Repository

```bash
git clone https://github.com/jitCompileCoffee/blog-agg.git
cd blog-agg
```

### 2. Install the CLI Globally

You can install the `gator` binary globally using the standard Go installation pipeline. Run the following command inside the root folder:

```bash
go install .
```

*(Make sure your `$GOPATH/bin` or `~/go/bin` is added to your system's global `PATH` variable so the `gator` command can be invoked from anywhere).*

---

## Configuration & Setup

### 1. Database Migrations

Create a new PostgreSQL database instance named `gator`:

```sql
CREATE DATABASE gator;
```

Navigate to your project's migration/schema folder and apply the Goose database schemas to build your tables:

```bash
goose postgres "postgres://your_username:your_password@localhost:5432/gator?sslmode=disable" up
```

### 2. Configuration File Creation

Gator looks for a JSON configuration file inside your home directory named `.gatorconfig.json` to store your connection string and current active user.

Create a `.gatorconfig.json` file in your home directory (`~/.gatorconfig.json`) with the following format:

```json
{
  "db_url": "postgres://your_username:your_password@localhost:5432/gator?sslmode=disable",
  "current_user_name": ""
}
```

---

## Architectural Layout

Gator uses a highly decoupled internal architecture to ensure fast network processing and strict type-safe database integrity:

* **State Object (`s *state`):** A shared, pointer-passed application context struct carrying runtime configurations and your PostgreSQL client thread pool.
* **Higher-Order Middleware (`middlewareLoggedIn`):** Gated decorators that validate active terminal user context, query database identities, and yield compile-time safety signatures down to specific sub-handlers.
* **Concurrency Protection:** The tracking infrastructure ensures duplicate web feeds gracefully step past unique index parameters via low-level `pgconn.PgError` state-handling rules.

---

## Usage & Core Commands

Gator operates through simple syntax rules: `gator <command> [arguments]`. Open your terminal and try a few of these core functionalities:

### Authentication & User Management

Register a new account:

```bash
gator register <username>
```

Switch or log in to an existing account:

```bash
gator login <username>
```

### Feed Management (Requires Authentication Middleware)

Add a new RSS feed to track:

```bash
gator addfeed <feed_name> <feed_url>
```

Follow an existing feed in the system:

```bash
gator follow <feed_url>
```

Unfollow an RSS feed safely:

```bash
gator unfollow <feed_url>
```

List all available system feeds and creators:

```bash
gator following
```

### Continuous Scraping & Reading

**Run the background worker loop (Never-ending process):**
Pass a time duration configuration parameter (e.g., `1s`, `1m`, `1h`) to determine how often Gator fetches data without overloading third-party servers.

```bash
gator agg 1m
```

> **Tip:** Leave this command running in the background inside one terminal tab, and use a second terminal tab to interact with your feeds.

**Browse downloaded feed posts:**
View latest articles downloaded by the background worker. Takes an optional limit parameter (defaults to 2).

```bash
gator browse 10
```

To gracefully exit running long-lived processes like `agg`, simply send a standard interrupt signal using `Ctrl+C`.

---

## Troubleshooting

### Unique Version Panic

If you run `goose up` and encounter a version panic:

```plaintext
panic: goose: duplicate version 3 detected
```

Ensure that no two migration filenames share identical leading sequential identifiers (e.g., `003_feed_follows.sql` and `003_feed_follow.sql`). Delete the stale migration and re-execute.

### Connection Refused Errors

Ensure your PostgreSQL daemon server is running locally on port 5432 and that the password matched in your `~/.gatorconfig.json` credentials explicitly aligns with your local system configuration.

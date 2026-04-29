<p align="center">
  <h1 align="center">⚡ inceptools</h1>
  <p align="center">
    <strong>A powerful, zero-dependency database migration CLI for Go</strong>
  </p>
  <p align="center">
    <a href="https://github.com/IncepTools/inceptools-cli/actions/workflows/ci.yml"><img src="https://github.com/IncepTools/inceptools-cli/actions/workflows/ci.yml/badge.svg" alt="CI"></a>
    <a href="https://github.com/IncepTools/inceptools-cli/releases/latest"><img src="https://img.shields.io/github/v/release/IncepTools/inceptools-cli?color=blue&label=latest" alt="Latest Release"></a>
    <a href="https://github.com/IncepTools/inceptools-cli/blob/main/LICENSE"><img src="https://img.shields.io/badge/license-GPL--3.0-green" alt="License"></a>
    <img src="https://img.shields.io/badge/go-%3E%3D1.22-00ADD8?logo=go" alt="Go Version">
  </p>
</p>

---

**inceptools** is a lightweight CLI tool that manages database schema migrations with automatic version tracking, rollback support, and multi-database compatibility — all powered by [GORM](https://gorm.io).

## ✨ Features

- 🗄️ **Multi-Database Support** — PostgreSQL, MySQL, SQLite, SQL Server, ClickHouse
- 📂 **File-Based Migrations** — Plain SQL up/down scripts, versioned and ordered
- 🔄 **Smart Sync** — Auto-detects removed migration files and rolls them back
- ⏪ **Granular Rollback** — Roll back all or `N` steps with a single command
- ⚙️ **Simple Configuration** — One `icpt.json` file, sensible defaults
- 🧾 **Full History** — Every migration run is tracked in the database

## 🗃️ Supported Databases

| Database   | Dialect Values       | Driver                       |
| ---------- | -------------------- | ---------------------------- |
| PostgreSQL | `postgres`, `pg`     | `gorm.io/driver/postgres`    |
| MySQL      | `mysql`              | `gorm.io/driver/mysql`       |
| SQLite     | `sqlite`             | `github.com/glebarez/sqlite` |
| SQL Server | `sqlserver`, `mssql` | `gorm.io/driver/sqlserver`   |
| ClickHouse | `clickhouse`         | `gorm.io/driver/clickhouse`  |

## 🏗️ Project Structure

The project follows a clean, modular architecture:

- `main.go`: Lightweight entry point and CLI dispatcher.
- `src/cmd/`: Individual command handlers (init, create, migrate, etc.).
- `src/db/`: Core database migration engine and dialect support.
- `src/config/`: Configuration management (`icpt.json`).
- `test/`: Comprehensive test suite for database and CLI logic.

## 📦 Installation

### Go Install

```bash
go install github.com/IncepTools/inceptools-cli@latest
```

### Homebrew (Recommended for macOS/Linux)

```bash
brew install inceptools/tap/inceptools
```

### APT (Ubuntu/Debian)

Download the `.deb` package from the [**Releases**](https://github.com/IncepTools/inceptools-cli/releases/latest) page and install:

```bash
sudo dpkg -i inceptools_*.deb
```

### Pre-built Binaries

Grab the latest release for your platform from the [**Releases**](https://github.com/IncepTools/inceptools-cli/releases/latest) page.

```bash
curl -fsSL https://raw.githubusercontent.com/IncepTools/inceptools-cli/main/install.sh
```

## 🛠️ Usage

### 1. Initialize

```bash
inceptools init
```

Creates a `migrations/` directory and `icpt.json` config.

### 2. Create Migration

```bash
inceptools create
```

Interactively prompts for a topic and generates `.up.sql` and `.down.sql` files.

### 3. Run Migrations

```bash
inceptools migrate [database_url]
```

If `database_url` is omitted, it uses the environment variable defined in `icpt.json`.

### 4. Rollback

```bash
inceptools down [database_url] --steps 1
```

### 5. Update CLI

```bash
inceptools update
```

# 000001_create_users_table.up.sql

# 000001_create_users_table.down.sql

````

Follow the prompt to enter a topic for your migration.

Create an `icpt.json` in your project root:

```json
{
  "env_variable": "DATABASE_URL",
  "migration_path": "./migrations",
  "dialect": "postgres"
}
````

| Key              | Default        | Description                                    |
| ---------------- | -------------- | ---------------------------------------------- |
| `env_variable`   | `DATABASE_URL` | Environment variable holding the database DSN  |
| `migration_path` | `./migrations` | Directory where migration SQL files are stored |
| `dialect`        | `postgres`     | Database dialect (see table above)             |

### 2. Create a Migration

```bash
inceptools create
# Enter migration topic (e.g., add_email_json): create_users_table
#
# Created:
#   000001_create_users_table.up.sql
#   000001_create_users_table.down.sql
```

Edit the generated files in `./migrations/`:

**`000001_create_users_table.up.sql`**

```sql
CREATE TABLE users (
    id    SERIAL PRIMARY KEY,
    name  VARCHAR(255) NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL
);
```

**`000001_create_users_table.down.sql`**

```sql
DROP TABLE IF EXISTS users;
```

### 3. Apply Migrations

```bash
# Using the environment variable from icpt.json
export DATABASE_URL="postgres://user:pass@localhost:5432/mydb?sslmode=disable"
inceptools migrate

# Or pass the DSN directly
inceptools migrate "postgres://user:pass@localhost:5432/mydb?sslmode=disable"
```

### 4. Rollback Migrations

```bash
# Roll back the last migration
inceptools down -steps 1

# Roll back all migrations
inceptools down

# With explicit DSN
inceptools down -steps 2 "postgres://user:pass@localhost:5432/mydb"
```

## 📁 Project Structure

```
your-project/
├── icpt.json                          # Configuration
├── migrations/
│   ├── 000001_create_users.up.sql     # Up migration
│   ├── 000001_create_users.down.sql   # Down migration
│   ├── 000002_add_email.up.sql
│   └── 000002_add_email.down.sql
└── ...
```

## 🔄 How Migration Sync Works

When you run `inceptools migrate`, the tool performs a **two-phase sync**:

1. **Phase A — Cleanup**: Scans for migrations that were previously applied but whose files have been **removed locally**. These are automatically rolled back using the stored down-script.
2. **Phase B — Apply**: Applies any **new or previously rolled-back** migrations in version order.

This ensures your database schema always mirrors the migration files in your project.

---

## 📂 Examples

Explore practical integration examples in the [examples/](./examples) directory:

- **[Go Project](./examples/go-project)**: Integration using `Makefile` and standard Go layout.
- **[Node.js Project](./examples/nodejs-project)**: Integration using `npm scripts` and `package.json`.

---

## 🤝 Contributing

Contributions are welcome! Please see [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines.

## 🔐 Security

To report a vulnerability, please see our [Security Policy](SECURITY.md).

## 📄 License

Copyright © 2025 [IncepTools](https://github.com/IncepTools)

This project is licensed under the **GNU General Public License v3.0** — see the [LICENSE](LICENSE) file for details.

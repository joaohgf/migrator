# Migrator

A Go package for managing PostgreSQL database migrations using SQL files and GORM.

## Features
- Migration file generator (ULID-based, up/down pairs)
- Transactional, ordered migrations
- Migration tracking with batch, timestamp, checksum
- TOML config (with Viper)
- Usable as a CLI or Go package

## Installation

```
go get github.com/joaohgf/migrator
```

## Configuration

Create a `config.toml`:
```toml
host = "localhost"
port = 5432
user = "youruser"
password = "yourpassword"
dbname = "yourdb"
sslmode = "disable"
migrationspath = "migrations"
```

## Usage as a Package

```go
import (
    "github.com/joaohgf/migrator/pkg/config"
    "github.com/joaohgf/migrator/pkg/db"
)

func main() {
    cfg, err := config.LoadConfig("config.toml")
    if err != nil {
        panic(err)
    }
    dbConn, err := db.ConnectDB(cfg)
    if err != nil {
        panic(err)
    }
    if err := db.RunMigrations(cfg, dbConn, false); err != nil {
        panic(err)
    }
}
```

## Migration Files

- Place your migrations in the directory specified by `migrationspath`.
- Name them as `{ULID}_{name}.up.sql` and `{ULID}_{name}.down.sql` (must be pairs).

## CLI

You can also use the CLI to generate migration files:

```
go run cmd/migrator/main.go create my_migration_name
```

---

MIT License 
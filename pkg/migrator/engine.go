package migrator

import (
	"crypto/sha256"
	"fmt"
	"os"
	"time"

	"github.com/joaohgf/migrator/pkg/config"
	"github.com/joaohgf/migrator/pkg/migrator/internal"
	"gorm.io/gorm"
)

// upMigrations applies all pending migrations. Used by Migrator.Up.
func upMigrations(cfg *config.Config, db *gorm.DB, dryRun bool) error {
	if err := internal.EnsureMigrationTable(db); err != nil {
		return fmt.Errorf("failed to migrate schema_migrations: %w", err)
	}

	migrations, err := internal.DiscoverMigrations(cfg.MigrationsPath)
	if err != nil {
		return err
	}

	batch := getNextBatch(db)
	var applied []string
	for _, m := range migrations {
		if isMigrationApplied(db, m.ULID) {
			continue
		}

		fmt.Printf("Preparing to apply migration: %s_%s\n", m.ULID, m.Name)
		if dryRun {
			fmt.Printf("[DRY RUN] Would apply: %s\n", m.Up)
			continue
		}

		content, err := os.ReadFile(m.Up)
		if err != nil {
			return fmt.Errorf("failed to read migration %s: %w", m.Up, err)
		}

		checksum := fmt.Sprintf("%x", sha256.Sum256(content))
		tx := db.Begin()
		if tx.Error != nil {
			return fmt.Errorf("failed to start transaction: %w", tx.Error)
		}

		if err := tx.Exec(string(content)).Error; err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to apply migration %s: %w", m.Up, err)
		}

		migrationRecord := internal.Migration{
			ID:        m.ULID,
			Name:      m.Name,
			AppliedAt: time.Now(),
			Batch:     batch,
			Direction: "up",
			Checksum:  checksum,
		}
		if err := tx.Create(&migrationRecord).Error; err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to record migration %s: %w", m.Up, err)
		}

		if err := tx.Commit().Error; err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to commit migration %s: %w", m.Up, err)
		}

		fmt.Printf("Applied migration: %s_%s\n", m.ULID, m.Name)
		applied = append(applied, fmt.Sprintf("%s_%s", m.ULID, m.Name))
	}

	fmt.Printf("\nMigration summary: %d applied, %d total\n", len(applied), len(migrations))
	if dryRun && len(applied) == 0 {
		fmt.Println("No migrations would be applied.")
	}
	return nil
}

// downMigrations rolls back the last batch of migrations using .down.sql files.
func downMigrations(cfg *config.Config, db *gorm.DB, steps int) error {
	if err := internal.EnsureMigrationTable(db); err != nil {
		return fmt.Errorf("failed to migrate schema_migrations: %w", err)
	}

	var lastBatch int
	db.Model(&internal.Migration{}).Select("MAX(batch)").Scan(&lastBatch)
	if lastBatch == 0 {
		fmt.Println("No migrations to rollback.")
		return nil
	}

	var applied []internal.Migration
	db.Where("batch = ? AND direction = ?", lastBatch, "up").Order("id DESC").Find(&applied)
	if len(applied) == 0 {
		fmt.Println("No migrations to rollback in last batch.")
		return nil
	}

	migrations, err := internal.DiscoverMigrations(cfg.MigrationsPath)
	if err != nil {
		return err
	}
	// Map for quick lookup
	migMap := make(map[string]internal.Migration)
	for _, m := range applied {
		migMap[m.ID] = m
	}
	// Rollback in reverse order
	count := 0
	for i := len(migrations) - 1; i >= 0 && count < steps; i-- {
		m := migrations[i]
		if _, ok := migMap[m.ULID]; !ok {
			continue
		}
		fmt.Printf("Rolling back migration: %s_%s\n", m.ULID, m.Name)
		content, err := os.ReadFile(m.Down)
		if err != nil {
			return fmt.Errorf("failed to read down migration %s: %w", m.Down, err)
		}
		tx := db.Begin()
		if tx.Error != nil {
			return fmt.Errorf("failed to start transaction: %w", tx.Error)
		}
		if err := tx.Exec(string(content)).Error; err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to rollback migration %s: %w", m.Down, err)
		}
		if err := tx.Where("id = ?", m.ULID).Delete(&internal.Migration{}).Error; err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to delete migration record %s: %w", m.ULID, err)
		}
		if err := tx.Commit().Error; err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to commit rollback %s: %w", m.Down, err)
		}
		fmt.Printf("Rolled back migration: %s_%s\n", m.ULID, m.Name)
		count++
	}
	fmt.Printf("\nRollback summary: %d rolled back from batch %d\n", count, lastBatch)
	return nil
}

// statusMigrations prints the current migration status.
func statusMigrations(cfg *config.Config, db *gorm.DB) error {
	if err := internal.EnsureMigrationTable(db); err != nil {
		return fmt.Errorf("failed to migrate schema_migrations: %w", err)
	}
	var applied []internal.Migration
	db.Order("applied_at").Find(&applied)
	migrations, err := internal.DiscoverMigrations(cfg.MigrationsPath)
	if err != nil {
		return err
	}
	appliedMap := make(map[string]internal.Migration)
	for _, m := range applied {
		appliedMap[m.ID] = m
	}
	fmt.Println("Applied migrations:")
	for _, m := range applied {
		fmt.Printf("- %s_%s (batch %d, at %s, checksum %s)\n", m.ID, m.Name, m.Batch, m.AppliedAt.Format(time.RFC3339), m.Checksum)
	}
	fmt.Println("\nPending migrations:")
	for _, m := range migrations {
		if _, ok := appliedMap[m.ULID]; !ok {
			fmt.Printf("- %s_%s\n", m.ULID, m.Name)
		}
	}
	return nil
}

func isMigrationApplied(db *gorm.DB, ulid string) bool {
	var count int64
	db.Model(&internal.Migration{}).Where("id = ?", ulid).Count(&count)
	return count > 0
}

func getNextBatch(db *gorm.DB) int {
	var maxBatch int
	db.Model(&internal.Migration{}).Select("MAX(batch)").Scan(&maxBatch)
	return maxBatch + 1
}

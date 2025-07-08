package internal

import (
	"time"

	"gorm.io/gorm"
)

// Migration is the migration record for internal use only.
type Migration struct {
	ID        string `gorm:"primaryKey"`
	Name      string
	AppliedAt time.Time `gorm:"autoCreateTime"`
	Batch     int
	Direction string
	Checksum  string
}

func EnsureMigrationTable(db *gorm.DB) error {
	return db.AutoMigrate(&Migration{})
}

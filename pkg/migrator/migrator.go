package migrator

import (
	"github.com/joaohgf/migrator/pkg/config"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// Migrator is the main struct for managing database migrations.
type Migrator struct {
	Cfg *config.Config
	DB  *gorm.DB
}

// New creates a new Migrator from a loaded config.
func New(cfg *config.Config) (*Migrator, error) {
	m := &Migrator{Cfg: cfg}
	if err := m.Connect(); err != nil {
		return nil, err
	}
	return m, nil
}

// ConnectDB connects the Migrator to the database using the resolved config secrets.
func (m *Migrator) Connect() error {
	dsn := "host=" + m.Cfg.Host.Value() +
		" port=" + m.Cfg.Port.Value() +
		" user=" + m.Cfg.User.Value() +
		" password=" + m.Cfg.Password.Value() +
		" dbname=" + m.Cfg.DBName.Value() +
		" sslmode=" + m.Cfg.SSLMode
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return err
	}
	m.DB = db
	return nil
}

// Up applies all pending migrations.
func (m *Migrator) Up(dryRun bool) error {
	return upMigrations(m.Cfg, m.DB, dryRun)
}

// Down rolls back the last batch of migrations (or up to 'steps' batches).
func (m *Migrator) Down(steps int) error {
	return downMigrations(m.Cfg, m.DB, steps)
}

// Status prints the current migration status.
func (m *Migrator) Status() error {
	return statusMigrations(m.Cfg, m.DB)
}

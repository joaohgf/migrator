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

// New creates a new Migrator from a config file path.
func New(cfgPath string) (*Migrator, error) {
	cfg, err := config.LoadConfig(cfgPath)
	if err != nil {
		return nil, err
	}
	conn, err := ConnectDB(cfg)
	if err != nil {
		return nil, err
	}
	return &Migrator{Cfg: cfg, DB: conn}, nil
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

// ConnectDB connects to the database using the resolved config secrets.
func ConnectDB(cfg *config.Config) (*gorm.DB, error) {
	dsn := "host=" + cfg.Host.Value() +
		" port=" + cfg.Port.Value() +
		" user=" + cfg.User.Value() +
		" password=" + cfg.Password.Value() +
		" dbname=" + cfg.DBName.Value() +
		" sslmode=" + cfg.SSLMode
	return gorm.Open(postgres.Open(dsn), &gorm.Config{})
}

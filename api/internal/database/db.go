package database

import (
	"time"

	"github.com/manuel/shopware-testenv-platform/api/internal/config"
	"github.com/manuel/shopware-testenv-platform/api/internal/logging"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func Connect(cfg config.DatabaseConfig) (*gorm.DB, error) {
	return gorm.Open(postgres.Open(cfg.DSN()), &gorm.Config{
		Logger: logging.NewGormLogger(),
		NowFunc: func() time.Time {
			return time.Now().UTC()
		},
	})
}

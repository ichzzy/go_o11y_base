package mysql

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/ichzzy/go_o11y_base/internal/config"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/plugin/dbresolver"
	"gorm.io/plugin/opentelemetry/tracing"
)

type MySQL struct {
	Orm *gorm.DB
	DB  *sql.DB
}

type MySqlConfig struct {
	EnableOtel bool
	GormLogger logger.Interface
	config.MySQLConfig
}

func NewMySQL(config MySqlConfig) (*MySQL, error) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		config.Master.Username,
		config.Master.Password,
		config.Master.Host,
		config.Master.Port,
		config.Master.Name,
	)

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: config.GormLogger,
	})
	if err != nil {
		return nil, fmt.Errorf("gorm.Open failed: %w", err)
	}

	// TODO: 目前 slaves 為空，預留擴展
	if err := db.Use(dbresolver.Register(dbresolver.Config{
		Sources:  []gorm.Dialector{mysql.Open(dsn)},
		Replicas: []gorm.Dialector{},
		Policy:   dbresolver.RandomPolicy{},
	})); err != nil {
		return nil, fmt.Errorf("register dbresolver failed: %w", err)
	}

	if config.EnableOtel {
		if err := db.Use(tracing.NewPlugin()); err != nil {
			return nil, fmt.Errorf("register otel failed: %w", err)
		}
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("db.DB failed: %w", err)
	}

	sqlDB.SetMaxIdleConns(config.Master.MaxIdleConns)
	sqlDB.SetMaxOpenConns(config.Master.MaxOpenConns)
	sqlDB.SetConnMaxLifetime(config.Master.ConnMaxLifetime)

	return &MySQL{Orm: db, DB: sqlDB}, nil
}

func (m *MySQL) Shutdown(ctx context.Context) error {
	return m.DB.Close()
}

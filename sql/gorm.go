package sql

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type GormConfig struct {
	Host                      string
	User                      string
	Password                  string
	Name                      string
	Port                      string
	SSLMode                   string
	Timezone                  string
	MaxIdleConns              int
	MaxOpenConns              int
	ConnMaxLifetime           time.Duration
	ConnMaxIdleTime           time.Duration
	TranslateErrors           bool
	LogLevel                  logger.LogLevel
	SlowThreshold             time.Duration
	IgnoreRecordNotFoundError bool
}

func (c GormConfig) dsn() string {
	ssl := c.SSLMode
	if ssl == "" {
		ssl = "disable"
	}
	tz := c.Timezone
	if tz == "" {
		tz = "UTC"
	}
	return fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=%s TimeZone=%s",
		c.Host, c.User, c.Password, c.Name, c.Port, ssl, tz,
	)
}

// DBManager wraps gorm.DB and provides additional functionality
type DBManager struct {
	*gorm.DB
	sqlDB *sql.DB
}

// Close closes the database connection
func (dm *DBManager) Close() error {
	if dm.sqlDB != nil {
		return dm.sqlDB.Close()
	}
	return nil
}

// Ping tests the database connection
func (dm *DBManager) Ping(ctx context.Context) error {
	if dm.sqlDB != nil {
		return dm.sqlDB.PingContext(ctx)
	}
	return fmt.Errorf("database connection not initialized")
}

// Stats returns database statistics
func (dm *DBManager) Stats() sql.DBStats {
	if dm.sqlDB != nil {
		return dm.sqlDB.Stats()
	}
	return sql.DBStats{}
}

func (c GormConfig) validate() error {
	if c.Host == "" {
		return fmt.Errorf("host is required")
	}
	if c.User == "" {
		return fmt.Errorf("user is required")
	}
	if c.Name == "" {
		return fmt.Errorf("database name is required")
	}
	if c.Port == "" {
		return fmt.Errorf("port is required")
	}
	return nil
}

func (c GormConfig) setDefaults() GormConfig {
	if c.SSLMode == "" {
		c.SSLMode = "disable"
	}
	if c.Timezone == "" {
		c.Timezone = "UTC"
	}
	if c.MaxIdleConns <= 0 {
		c.MaxIdleConns = 10
	}
	if c.MaxOpenConns <= 0 {
		c.MaxOpenConns = 100
	}
	if c.ConnMaxLifetime <= 0 {
		c.ConnMaxLifetime = 30 * time.Minute
	}
	if c.ConnMaxIdleTime <= 0 {
		c.ConnMaxIdleTime = 10 * time.Minute
	}
	if c.LogLevel == 0 {
		c.LogLevel = logger.Info
	}
	if c.SlowThreshold <= 0 {
		c.SlowThreshold = 200 * time.Millisecond
	}
	return c
}

func OpenGorm(c GormConfig, automigrateModels ...interface{}) (*DBManager, error) {
	if err := c.validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	c = c.setDefaults()

	gormConfig := &gorm.Config{
		TranslateError: c.TranslateErrors,
		Logger:         logger.Default.LogMode(c.LogLevel),
	}

	db, err := gorm.Open(postgres.Open(c.dsn()), gormConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}

	// Configure connection pool
	sqlDB.SetMaxIdleConns(c.MaxIdleConns)
	sqlDB.SetMaxOpenConns(c.MaxOpenConns)
	sqlDB.SetConnMaxLifetime(c.ConnMaxLifetime)
	sqlDB.SetConnMaxIdleTime(c.ConnMaxIdleTime)

	// Test the connection
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := sqlDB.PingContext(ctx); err != nil {
		sqlDB.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	// Auto-migrate models if provided
	if len(automigrateModels) > 0 {
		if err := db.AutoMigrate(automigrateModels...); err != nil {
			sqlDB.Close()
			return nil, fmt.Errorf("failed to auto-migrate: %w", err)
		}
	}

	return &DBManager{
		DB:    db,
		sqlDB: sqlDB,
	}, nil
}

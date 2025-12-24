package db

import (
	"fmt"
	"path/filepath"
	"sync"

	pathHelper "onx-screen-record/internal/pkg/path-file"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var (
	instance *gorm.DB
	once     sync.Once
)

type Database struct {
	DB         *gorm.DB
	pathHelper *pathHelper.PathHelper
}

// NewDatabase creates a new Database instance
func NewDatabase(appName string, ph *pathHelper.PathHelper) (*Database, error) {
	db, err := initDB(ph)
	if err != nil {
		return nil, err
	}

	return &Database{
		DB:         db,
		pathHelper: ph,
	}, nil
}

// initDB initializes the database connection
func initDB(ph *pathHelper.PathHelper) (*gorm.DB, error) {
	var initErr error

	once.Do(func() {
		appDataDir, err := ph.GetAppDataDir()
		fmt.Println("App Data Dir:", appDataDir)
		if err != nil {
			initErr = fmt.Errorf("failed to get app data directory: %w", err)
			return
		}

		dbPath := filepath.Join(appDataDir, "onx-screen-record.db")

		db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{
			Logger: logger.Default.LogMode(logger.Silent),
			// Logger: logger.Default.LogMode(logger.Info),
		})
		if err != nil {
			initErr = fmt.Errorf("failed to connect to database: %w", err)
			return
		}

		// Enable foreign keys for SQLite
		sqlDB, err := db.DB()
		if err != nil {
			initErr = fmt.Errorf("failed to get sql.DB: %w", err)
			return
		}

		_, err = sqlDB.Exec("PRAGMA foreign_keys = ON")
		if err != nil {
			initErr = fmt.Errorf("failed to enable foreign keys: %w", err)
			return
		}

		instance = db
	})

	if initErr != nil {
		return nil, initErr
	}

	return instance, nil
}

// GetDB returns the database instance
func (d *Database) GetDB() *gorm.DB {
	return d.DB
}

// Close closes the database connection
func (d *Database) Close() error {
	sqlDB, err := d.DB.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

// GetInstance returns the singleton database instance
func GetInstance() *gorm.DB {
	return instance
}

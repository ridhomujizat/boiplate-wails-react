package db

import (
	"embed"
	"fmt"
	"io/fs"
	"sort"
	"strings"

	"gorm.io/gorm"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

// Migration represents a database migration record
type Migration struct {
	ID      uint   `gorm:"primaryKey"`
	Name    string `gorm:"uniqueIndex;size:255"`
	Applied bool   `gorm:"default:false"`
}

// Migrator handles database migrations
type Migrator struct {
	db *gorm.DB
}

// NewMigrator creates a new Migrator instance
func NewMigrator(db *gorm.DB) *Migrator {
	return &Migrator{db: db}
}

// Run executes all pending migrations
func (m *Migrator) Run() error {
	// Create migrations table if not exists
	if err := m.db.AutoMigrate(&Migration{}); err != nil {
		return fmt.Errorf("failed to create migrations table: %w", err)
	}

	// Get all migration files
	files, err := m.getMigrationFiles()
	if err != nil {
		return fmt.Errorf("failed to get migration files: %w", err)
	}

	// Sort files by name to ensure order
	sort.Strings(files)

	// Run each migration
	for _, file := range files {
		if err := m.runMigration(file); err != nil {
			return fmt.Errorf("failed to run migration %s: %w", file, err)
		}
	}

	return nil
}

// getMigrationFiles returns list of migration file names
func (m *Migrator) getMigrationFiles() ([]string, error) {
	var files []string

	entries, err := fs.ReadDir(migrationsFS, "migrations")
	if err != nil {
		return nil, err
	}

	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".sql") {
			files = append(files, entry.Name())
		}
	}

	return files, nil
}

// runMigration executes a single migration
func (m *Migrator) runMigration(filename string) error {
	// Check if migration already applied
	var migration Migration
	result := m.db.Where("name = ?", filename).First(&migration)

	if result.Error == nil && migration.Applied {
		// Migration already applied, skip
		return nil
	}

	// Read migration file
	content, err := migrationsFS.ReadFile("migrations/" + filename)
	if err != nil {
		return fmt.Errorf("failed to read migration file: %w", err)
	}

	// Execute migration in transaction
	return m.db.Transaction(func(tx *gorm.DB) error {
		// Split SQL by semicolon and execute each statement
		statements := strings.Split(string(content), ";")
		for _, stmt := range statements {
			stmt = strings.TrimSpace(stmt)
			if stmt == "" {
				continue
			}

			if err := tx.Exec(stmt).Error; err != nil {
				return fmt.Errorf("failed to execute statement: %w", err)
			}
		}

		// Record migration
		if result.Error != nil {
			// Create new migration record
			return tx.Create(&Migration{
				Name:    filename,
				Applied: true,
			}).Error
		}

		// Update existing migration record
		return tx.Model(&migration).Update("applied", true).Error
	})
}

// Rollback rolls back the last migration (placeholder for future implementation)
func (m *Migrator) Rollback() error {
	// TODO: Implement rollback logic
	return nil
}

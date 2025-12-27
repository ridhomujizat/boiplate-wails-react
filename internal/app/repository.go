package app

import (
	"onx-screen-record/internal/pkg/db"
)

func (a *App) initializeDatabase() error {
	database, err := db.NewDatabase(a.AppName, a.path)
	if err != nil {
		return err
	}
	defer database.Close()

	// Run migrations
	migrator := db.NewMigrator(database.GetDB())
	if err := migrator.Run(); err != nil {
		return err

	}

	// // Use repository
	// settingsRepo := repository.NewSettingsRepository(database.GetDB())
	// settings, _ := settingsRepo.GetAll()
	return nil
}

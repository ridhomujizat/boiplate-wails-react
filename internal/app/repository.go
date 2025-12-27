package app

import (
	"onx-screen-record/internal/pkg/db"
	"onx-screen-record/internal/repository"
	"onx-screen-record/internal/repository/setting"
)

func (a *App) initializeDatabase() error {
	database, err := db.NewDatabase(a.AppName, a.path)
	if err != nil {
		return err
	}
	// Note: Do NOT close the database here - it needs to stay open for the app lifetime

	// Run migrations
	migrator := db.NewMigrator(database.GetDB())
	if err := migrator.Run(); err != nil {
		return err

	}

	// // Use repository
	// settingsRepo := repository.NewSettingsRepository(database.GetDB())
	// settings, _ := settingsRepo.GetAll()

	a.rp = repository.IRepository{
		Setting: *setting.NewRepository(database.GetDB()),
	}
	return nil
}

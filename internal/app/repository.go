package app

import (
	"onx-screen-record/internal/pkg/db"
	"onx-screen-record/internal/repository"
	"onx-screen-record/internal/repository/activity"
	"onx-screen-record/internal/repository/setting"
	"onx-screen-record/internal/repository/uploadjob"
)

func (a *App) initializeDatabase() error {
	database, err := db.NewDatabase(a.AppName, a.path)
	if err != nil {
		return err
	}

	migrator := db.NewMigrator(database.GetDB())
	if err := migrator.Run(); err != nil {
		return err
	}

	a.rp = repository.IRepository{
		Setting:   *setting.NewRepository(database.GetDB()),
		Activity:  *activity.NewRepository(database.GetDB()),
		UploadJob: *uploadjob.NewRepository(database.GetDB()),
	}
	return nil
}

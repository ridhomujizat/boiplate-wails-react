package repository

import (
	"onx-screen-record/internal/repository/activity"
	"onx-screen-record/internal/repository/setting"
	"onx-screen-record/internal/repository/uploadjob"
)

type IRepository struct {
	Setting   setting.Repository
	Activity  activity.Repository
	UploadJob uploadjob.Repository
}

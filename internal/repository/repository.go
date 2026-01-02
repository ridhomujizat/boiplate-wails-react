package repository

import (
	"onx-screen-record/internal/repository/activity"
	"onx-screen-record/internal/repository/setting"
)

type IRepository struct {
	Setting  setting.Repository
	Activity activity.Repository
}

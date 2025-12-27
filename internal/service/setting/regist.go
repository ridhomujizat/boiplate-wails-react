package setting

import (
	"context"
	"onx-screen-record/internal/repository"
	"onx-screen-record/internal/service/setting/dto"
)

type Service struct {
	ctx context.Context
	rp  repository.IRepository
}
type IService interface {
	GetSettings() (*dto.SettingResponse, error)
	SaveSettings(req dto.SettingRequest) (*dto.SaveSettingResponse, error)
	GetAudioSettings() (*dto.AudioSettingResponse, error)
	SaveAudioSettings(req dto.AudioSettingRequest) (*dto.SaveSettingResponse, error)
}

func NewService(ctx context.Context, rp repository.IRepository) IService {
	return &Service{
		ctx: ctx,
		rp:  rp,
	}
}

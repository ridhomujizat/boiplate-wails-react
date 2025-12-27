package permission

import (
	"context"
)

type Service struct {
	ctx context.Context
}

type IService interface {
}

func NewService(ctx context.Context) IService {
	return &Service{
		ctx: ctx,
	}
}

type PermissionConfig struct {
	ScreenRecording bool
	AudioRecording  bool
	InternetAccess  bool
}

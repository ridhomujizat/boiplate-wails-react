package mqtt

import (
	"context"
	"onx-screen-record/internal/repository"
	"onx-screen-record/internal/service/auth"
	"onx-screen-record/internal/service/setting"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

type Service struct {
	ctx         context.Context
	auth        auth.IService
	mqttClient  mqtt.Client
	rp          repository.IRepository
	setting     setting.IService
	isConnected bool
}

type IService interface {
	IsConnected() bool
	Connect(callback mqtt.MessageHandler) error
	Disconnect()
}

func NewService(ctx context.Context, auth auth.IService, rp repository.IRepository, setting setting.IService) IService {
	return &Service{
		ctx:     ctx,
		auth:    auth,
		rp:      rp,
		setting: setting,
	}
}

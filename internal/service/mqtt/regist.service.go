package mqtt

import (
	"context"
	"onx-screen-record/internal/repository"
	"onx-screen-record/internal/service/auth"
	"onx-screen-record/internal/service/setting"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// Connection state constants
const (
	StateDisconnected = "disconnected"
	StateConnecting   = "connecting"
	StateConnected    = "connected"
	StateReconnecting = "reconnecting"
)

type Service struct {
	ctx             context.Context
	auth            auth.IService
	mqttClient      mqtt.Client
	rp              repository.IRepository
	setting         setting.IService
	isConnected     bool
	connectionState string
}

type IService interface {
	IsConnected() bool
	GetConnectionState() string
	Connect(callback mqtt.MessageHandler) error
	Disconnect()
}

func NewService(ctx context.Context, auth auth.IService, rp repository.IRepository, setting setting.IService) IService {
	return &Service{
		ctx:             ctx,
		auth:            auth,
		rp:              rp,
		setting:         setting,
		connectionState: StateDisconnected,
	}
}

// setConnectionState updates the state and emits an event to the frontend
func (s *Service) setConnectionState(state string) {
	s.connectionState = state
	if s.ctx != nil {
		runtime.EventsEmit(s.ctx, "mqtt-status", map[string]interface{}{
			"state": state,
		})
	}
}

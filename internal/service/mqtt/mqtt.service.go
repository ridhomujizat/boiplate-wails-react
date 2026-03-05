package mqtt

import (
	"fmt"
	"log"
	"onx-screen-record/internal/pkg/helper"
	"onx-screen-record/internal/pkg/logger"
	"strings"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

func (s *Service) IsConnected() bool {
	return s.isConnected
}

func (s *Service) GetConnectionState() string {
	return s.connectionState
}

func (s *Service) Connect(callback mqtt.MessageHandler) error {
	// Guard: if already connected or connecting, skip
	if s.connectionState == StateConnecting || s.connectionState == StateConnected {
		log.Printf("[MQTT] Already %s, skipping Connect()", s.connectionState)
		return nil
	}

	// Disconnect existing client to prevent duplicate client ID loop
	if s.mqttClient != nil {
		s.mqttClient.Disconnect(250)
		s.mqttClient = nil
	}

	s.setConnectionState(StateConnecting)

	settings, err := s.setting.GetSettings()
	if err != nil {
		s.setConnectionState(StateDisconnected)
		return err
	}
	mqttURL := settings.MqttBroker
	tenant := settings.TenantCode
	deviceID, err := helper.GetDeviceID()
	if err != nil {
		logger.Error.Printf("Failed to get device ID: %v", err)
		s.setConnectionState(StateDisconnected)
		return err
	}

	if mqttURL == "" || tenant == "" || deviceID == "" {
		s.setConnectionState(StateDisconnected)
		log.Fatal("Environment variable VITE_APP_MQTT_URL, VITE_APP_TENANT, dan SESSION_ID harus diset")
	}

	userKey := fmt.Sprintf("%s:record:%s", tenant, deviceID)
	topic := strings.ReplaceAll(userKey, ":", "/")

	fmt.Printf("[MQTT DEBUG] Subscribing to topic: %s (Tenant: %s, DeviceID: %s, Broker: %s)\n", topic, tenant, deviceID, mqttURL)

	opts := mqtt.NewClientOptions()
	opts.AddBroker(mqttURL)
	opts.SetClientID(deviceID)
	opts.SetUsername(":" + userKey)
	opts.SetPassword(userKey + ":password")
	opts.SetKeepAlive(30 * time.Second)
	opts.SetAutoReconnect(true)
	opts.SetConnectRetry(true)
	opts.SetConnectRetryInterval(5 * time.Second)
	opts.SetCleanSession(true)
	
	// Add connection timeout
	opts.SetConnectTimeout(10 * time.Second)

	opts.SetDefaultPublishHandler(callback)

	// Callback when connection is established
	opts.OnConnect = func(client mqtt.Client) {
		log.Println("[MQTT] OnConnect callback triggered!")
		s.isConnected = true
		s.setConnectionState(StateConnected)
		log.Printf("[MQTT] Connection state set to: %s", StateConnected)

		if token := client.Subscribe(topic, 1, callback); token.Wait() && token.Error() != nil {
			log.Printf("Failed to subscribe to topic %s: %v", topic, token.Error())
		} else {
			log.Printf("Subscribed to topic: %s", topic)
		}
	}

	// Callback when connection is lost
	opts.OnConnectionLost = func(client mqtt.Client, err error) {
		s.isConnected = false
		s.setConnectionState(StateReconnecting)
		log.Printf("[MQTT] Connection lost: %v", err)
	}

	// Callback when reconnecting
	opts.SetReconnectingHandler(func(client mqtt.Client, opts *mqtt.ClientOptions) {
		s.setConnectionState(StateReconnecting)
		log.Println("[MQTT] Reconnecting...")
	})

	client := mqtt.NewClient(opts)
	s.mqttClient = client

	log.Printf("[MQTT] Calling Connect() to broker: %s", mqttURL)
	token := client.Connect()
	token.Wait()
	if token.Error() != nil {
		log.Printf("[MQTT] Connect() error: %v", token.Error())
		s.setConnectionState(StateDisconnected)
	} else {
		log.Println("[MQTT] Connect() initiated successfully, waiting for OnConnect callback...")
		log.Printf("[MQTT] Client IsConnected: %v", client.IsConnected())
	}

	return nil
}

func (s *Service) Disconnect() {
	if s.mqttClient != nil && s.mqttClient.IsConnected() {
		s.mqttClient.Disconnect(250) // 250 ms timeout
		s.isConnected = false
		s.setConnectionState(StateDisconnected)
	} else {
		log.Println("MQTT client not connected or already disconnected")
	}
}

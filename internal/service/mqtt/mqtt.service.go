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

func (s *Service) Connect(callback mqtt.MessageHandler) error {
	settings, err := s.setting.GetSettings()
	if err != nil {
		return err
	}
	mqttURL := settings.MqttBroker
	tenant := settings.TenantCode
	deviceID, err := helper.GetDeviceID()
	if err != nil {
		logger.Error.Printf("Failed to get device ID: %v", err)
		return err
	}

	if mqttURL == "" || tenant == "" || deviceID == "" {
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

	opts.SetDefaultPublishHandler(callback)

	// Callback ketika koneksi terputus
	opts.OnConnectionLost = func(client mqtt.Client, err error) {
		s.isConnected = false
		log.Printf("Koneksi MQTT terputus: %v", err)
		// logerror := fmt.Sprintf("MQTT connection lost: %v", err)
		// Log error to file
		// s.rp.Logger.CreateLogError(logerror, "mqtt_service_disconnect", err)
		// helper.LogErrorToFile("error.log", logerror, err)
	}

	// Callback ketika koneksi berhasil
	opts.OnConnect = func(client mqtt.Client) {
		log.Println("Koneksi MQTT berhasil!")
		s.isConnected = true

		// logSuccess := "MQTT connected successfully"
		// s.rp.Logger.CreateLogInfo(logSuccess, "mqtt_service_connect", nil)
		// helper.LogErrorToFile("app.log", logSuccess, nil)

		if token := client.Subscribe(topic, 1, callback); token.Wait() && token.Error() != nil {
			// logerror := fmt.Sprintf("Failed to subscribe to topic %s: %v", topic, token.Error())
			log.Printf("Gagal subscribe ke topic %s: %v", topic, token.Error())
			// Log error to file
			// s.rp.Logger.CreateLogError(logerror, "mqtt_service_connect", token.Error())
			// helper.LogErrorToFile("error.log", logerror, token.Error())
		} else {
			log.Printf("Berhasil subscribe ke topic: %s", topic)
		}
	}

	client := mqtt.NewClient(opts)
	s.mqttClient = client
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		// logerror := fmt.Sprintf("Failed to connect to MQTT broker: %v", token.Error())
		// Log error to file
		// s.rp.Logger.CreateLogError(logerror, "mqtt_service_connect", token.Error())
		// helper.LogErrorToFile("error.log", logerror, token.Error())
		// log.Fatalf("Gagal connect ke MQTT broker: %v", token.Error())

	}

	return nil
}

func (s *Service) Disconnect() {
	if s.mqttClient != nil && s.mqttClient.IsConnected() {
		s.mqttClient.Disconnect(250) // 250 ms timeout
		s.isConnected = false

	} else {
		log.Println("MQTT client not connected or already disconnected")
	}
}

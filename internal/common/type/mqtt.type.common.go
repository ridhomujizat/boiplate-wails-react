package types

// RecordMQTTPayload represents the MQTT message payload for recording commands
type RecordMQTTPayload struct {
	// ClientID identifies the client device
	ClientID string `json:"client_id"`

	// ClientEmail is the email associated with the client
	ClientEmail string `json:"client_email"`

	// SessionId uniquely identifies the recording session
	SessionId string `json:"session_id"`

	// Action indicates the recording action: "start", "stop"
	Action string `json:"action"`

	// ErrorMessage contains error details if the recording failed
	ErrorMessage *string `json:"error_message"`
}

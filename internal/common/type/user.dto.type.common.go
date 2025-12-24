package types

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
)

type UserRequest struct {
	DeviceID string `json:"device_id"`
	ClientID string `json:"client_id"`
	Email    string `json:"email"`
}

func (u UserRequest) Value() (driver.Value, error) {
	return json.Marshal(u)
}

func (u *UserRequest) Scan(value interface{}) error {
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}

	return json.Unmarshal(bytes, u)
}

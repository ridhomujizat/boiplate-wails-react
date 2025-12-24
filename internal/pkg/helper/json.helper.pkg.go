package helper

import (
	"encoding/json"
)

func JSONToString(payload any) (string, error) {
	jsonBytes, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}

	jsonString := string(jsonBytes)
	return jsonString, nil
}

func JSONToStruct[I any](payload any) (result *I, err error) {
	jsonBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(jsonBytes, &result)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func MapToStruct(data map[string]interface{}, target interface{}) error {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return err
	}
	return json.Unmarshal(jsonData, target)
}

func StrucToMap(payload any) (map[string]interface{}, error) {
	JsonMap := make(map[string]interface{})
	jsonBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(jsonBytes, &JsonMap)
	if err != nil {
		return nil, err
	}

	return JsonMap, nil
}

func JSONToByte(payload any) ([]byte, error) {
	jsonBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	return jsonBytes, nil
}

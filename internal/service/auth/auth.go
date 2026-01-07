package auth

import (
	"context"
	"fmt"
	"net/http"

	"onx-screen-record/internal/common/enum"
	helper "onx-screen-record/internal/pkg/helper"
	"onx-screen-record/internal/pkg/logger"
	"onx-screen-record/internal/service/auth/dto"
)

type Service struct {
	ctx        context.Context
	getBaseURL func() (string, error)
}

func (s *Service) getContext() context.Context {
	if s.ctx == nil {
		return context.Background()
	}
	return s.ctx
}

type IService interface {
	Login(email string, password string) (dto.LoginResponse, error)
}

func NewService(ctx context.Context, getBaseURL func() (string, error)) IService {
	return &Service{
		ctx:        ctx,
		getBaseURL: getBaseURL,
	}
}

// Login performs user authentication via API
func (s *Service) Login(email string, password string) (dto.LoginResponse, error) {
	// Get device ID
	deviceID, err := helper.GetDeviceID()
	if err != nil {
		logger.Error.Printf("Failed to get device ID: %v", err)
		return dto.LoginResponse{
			Message: "Failed to get device ID",
		}, err
	}

	// Get baseurl from settings
	baseURL, err := s.getBaseURL()
	if err != nil {
		logger.Error.Printf("Failed to get settings: %v", err)
		return dto.LoginResponse{
			Message: "Failed to get settings",
		}, err
	}

	if baseURL == "" {
		return dto.LoginResponse{
			Message: "BaseUrl not configured",
		}, nil
	}

	fmt.Println("Base URL:", baseURL)
	fmt.Println("Device ID:", deviceID)
	fmt.Println("Email:", email)
	fmt.Println("Password:", password)

	// Prepare login request
	loginReq := dto.LoginRequest{
		Email:    email,
		Password: password,
		DeviceID: deviceID,
	}

	// Build API URL
	apiURL := baseURL + "/api/auth/login"

	headers := http.Header{
		"Content-Type": []string{"application/json"},
	}

	response, err := helper.HTTPRequest(&helper.HTTPRequestPayload{
		Method: enum.POST,
		URL:    apiURL,
		Body:   loginReq,
	},
		&helper.HTTPRequestConfig{
			Headers: headers,
			Ctx:     s.getContext(),
		})
	if err != nil {
		return dto.LoginResponse{}, fmt.Errorf("failed post login: %w", err)
	}

	// Parse response - convert to JSON then unmarshal to struct
	jsonBytes, err := helper.JSONToByte(response.Data)
	if err != nil {
		logger.Error.Printf("Failed to parse response: %v", err)
		return dto.LoginResponse{
			Message: "Failed to parse response",
		}, err
	}

	var loginResp dto.LoginResponse
	if err := helper.JSONByteToStruct(jsonBytes, &loginResp); err != nil {
		logger.Error.Printf("Failed to unmarshal response: %v", err)
		return dto.LoginResponse{
			Message: "Failed to parse response data",
		}, err
	}

	return loginResp, nil
}

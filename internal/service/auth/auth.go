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
	DeepLinkAuth(email string, token string) (dto.DeepLinkAuthResponse, error)
	Logout(token string) (dto.LogoutResponse, error)
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

// DeepLinkAuth performs authentication via deep link token
func (s *Service) DeepLinkAuth(email string, token string) (dto.DeepLinkAuthResponse, error) {
	// Get device ID
	deviceID, err := helper.GetDeviceID()
	if err != nil {
		logger.Error.Printf("Failed to get device ID: %v", err)
		return dto.DeepLinkAuthResponse{
			Message: "Failed to get device ID",
		}, err
	}

	// Get baseurl from settings
	baseURL, err := s.getBaseURL()
	if err != nil {
		logger.Error.Printf("Failed to get settings: %v", err)
		return dto.DeepLinkAuthResponse{
			Message: "Failed to get settings",
		}, err
	}

	if baseURL == "" {
		return dto.DeepLinkAuthResponse{
			Message: "BaseUrl not configured",
		}, nil
	}

	// Prepare deep link auth request
	authReq := dto.DeepLinkAuthRequest{
		Email:    email,
		Token:    token,
		DeviceID: deviceID,
	}

	// Build API URL - assuming the endpoint is /api/auth/deeplink
	apiURL := baseURL + "/api/auth/login"

	headers := http.Header{
		"Content-Type": []string{"application/json"},
	}

	response, err := helper.HTTPRequest(&helper.HTTPRequestPayload{
		Method: enum.POST,
		URL:    apiURL,
		Body:   authReq,
	},
		&helper.HTTPRequestConfig{
			Headers: headers,
			Ctx:     s.getContext(),
		})
	if err != nil {
		return dto.DeepLinkAuthResponse{}, fmt.Errorf("failed post deep link auth: %w", err)
	}

	// Parse response - convert to JSON then unmarshal to struct
	jsonBytes, err := helper.JSONToByte(response.Data)
	if err != nil {
		logger.Error.Printf("Failed to parse response: %v", err)
		return dto.DeepLinkAuthResponse{
			Message: "Failed to parse response",
		}, err
	}

	var authResp dto.DeepLinkAuthResponse
	if err := helper.JSONByteToStruct(jsonBytes, &authResp); err != nil {
		logger.Error.Printf("Failed to unmarshal response: %v", err)
		return dto.DeepLinkAuthResponse{
			Message: "Failed to parse response data",
		}, err
	}

	return authResp, nil
}

// Logout performs user logout via API
func (s *Service) Logout(token string) (dto.LogoutResponse, error) {
	// Get device ID
	deviceID, err := helper.GetDeviceID()
	if err != nil {
		logger.Error.Printf("Failed to get device ID: %v", err)
		return dto.LogoutResponse{
			Message: "Failed to get device ID",
			Success: false,
		}, err
	}

	// Get baseurl from settings
	baseURL, err := s.getBaseURL()
	if err != nil {
		logger.Error.Printf("Failed to get settings: %v", err)
		return dto.LogoutResponse{
			Message: "Failed to get settings",
			Success: false,
		}, err
	}

	if baseURL == "" {
		return dto.LogoutResponse{
			Message: "BaseUrl not configured",
			Success: false,
		}, nil
	}

	// Prepare logout request
	logoutReq := dto.LogoutRequest{
		Token:    token,
		DeviceID: deviceID,
	}

	// Build API URL
	apiURL := baseURL + "/api/auth/logout"

	headers := http.Header{
		"Content-Type":  []string{"application/json"},
		"Authorization": []string{"Bearer " + token},
	}

	response, err := helper.HTTPRequest(&helper.HTTPRequestPayload{
		Method: enum.POST,
		URL:    apiURL,
		Body:   logoutReq,
	},
		&helper.HTTPRequestConfig{
			Headers: headers,
			Ctx:     s.getContext(),
		})
	if err != nil {
		return dto.LogoutResponse{
			Message: "Failed to logout",
			Success: false,
		}, fmt.Errorf("failed post logout: %w", err)
	}

	// Parse response
	jsonBytes, err := helper.JSONToByte(response.Data)
	if err != nil {
		logger.Error.Printf("Failed to parse response: %v", err)
		return dto.LogoutResponse{
			Message: "Failed to parse response",
			Success: false,
		}, err
	}

	var logoutResp dto.LogoutResponse
	if err := helper.JSONByteToStruct(jsonBytes, &logoutResp); err != nil {
		logger.Error.Printf("Failed to unmarshal response: %v", err)
		return dto.LogoutResponse{
			Message: "Failed to parse response data",
			Success: false,
		}, err
	}

	logger.Info.Printf("User logged out successfully")
	return logoutResp, nil
}

package helper

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"net/url"
	"onx-screen-record/internal/common/enum"
	types "onx-screen-record/internal/common/type"
	"onx-screen-record/internal/pkg/logger"
	"reflect"
	"strings"
	"time"
)

type HTTPAPIResponse struct {
	StatusCode int         `json:"status_code"`
	Headers    http.Header `json:"headers"`
	Data       interface{} `json:"data"`
}

type HTTPRequestPayload struct {
	Method enum.HTTPMethodEnum
	URL    string
	Body   interface{}
	Params map[string]string
}

type HTTPRequestConfig struct {
	Ctx       context.Context
	Headers   http.Header
	Auth      *BasicAuthConfig
	HTTPAgent *http.Transport
}

type BasicAuthConfig struct {
	Username string
	Password string
}

func HTTPRequest(
	payload *HTTPRequestPayload,
	config *HTTPRequestConfig,
) (*HTTPAPIResponse, error) {
	requestBody, err := handleRequestBody(payload, config)
	if err != nil {
		logger.Error.Println("Error handling request body:", err.Error())
		return nil, err
	}

	req, client, err := prepareRequest(payload, requestBody, config)
	if err != nil {
		logger.Error.Println("Error preparing request:", err.Error())
		return nil, err
	}
	return executeRequest(req, client)
}

func SanitizeURL(rawURL string) (*url.URL, error) {
	rawURL = strings.TrimSpace(rawURL)

	if rawURL == "" {
		return nil, fmt.Errorf("URL cannot be empty")
	}

	if !strings.Contains(rawURL, "://") {
		rawURL = "https://" + rawURL
	}

	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse URL: %w", err)
	}

	if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
		return nil, fmt.Errorf("invalid URL scheme: only http and https are allowed, got '%s'", parsedURL.Scheme)
	}

	if parsedURL.Host == "" {
		return nil, fmt.Errorf("URL must contain a valid host")
	}

	return parsedURL, nil
}

func handleRequestBody(payload *HTTPRequestPayload, config *HTTPRequestConfig) (io.Reader, error) {
	var requestBody io.Reader
	var err error

	if payload.Method == enum.GET {
		return nil, nil
	} else {
		switch config.Headers.Get("Content-Type") {
		case enum.ApplicationXform.ToString():
			if payload.Body != nil {
				requestBody, err = createFormURLEncodedBody(payload.Body)
			}
		case enum.MultipartForm.ToString():
			if payload.Body != nil {
				var ct string
				requestBody, ct, err = createMultipartBody(payload.Body)
				config.Headers.Set("Content-Type", ct)
			}
		case enum.ApplicationJSON.ToString():
			requestBody, err = createJSONBody(payload.Body)
		case "":
			config.Headers.Set("Content-Type", enum.ApplicationJSON.ToString())
			requestBody, err = createJSONBody(payload.Body)
		default:
			return nil, errors.New("unsupported content type")
		}
	}

	return requestBody, err
}

func prepareRequest(payload *HTTPRequestPayload, body io.Reader, config *HTTPRequestConfig) (*http.Request, *http.Client, error) {
	sanitizedURL, err := SanitizeURL(payload.URL)
	if err != nil {
		return nil, nil, fmt.Errorf("invalid URL: %w", err)
	}

	req, err := http.NewRequestWithContext(config.Ctx, payload.Method.ToString(), sanitizedURL.String(), body)
	if err != nil {
		return nil, nil, err
	}

	for key, values := range config.Headers {
		req.Header[key] = append(req.Header[key], values...)
	}

	if config.Auth != nil {
		req.SetBasicAuth(config.Auth.Username, config.Auth.Password)
	}

	client := &http.Client{Timeout: 60 * time.Second}

	if config.HTTPAgent != nil {
		client.Transport = config.HTTPAgent
	}

	return req, client, nil
}

func executeRequest(req *http.Request, client *http.Client) (*HTTPAPIResponse, error) {
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	result, err := parseResponseBody(resp)
	if err != nil {
		return nil, err
	}

	return &HTTPAPIResponse{
		StatusCode: resp.StatusCode,
		Headers:    resp.Header,
		Data:       result,
	}, nil
}

func parseResponseBody(resp *http.Response) (interface{}, error) {
	contentType := resp.Header.Get("Content-Type")
	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var result interface{}
	switch {
	case strings.Contains(contentType, "application/json"):
		if err := json.Unmarshal(responseBody, &result); err != nil {
			return nil, err
		}

	case strings.Contains(contentType, "text/plain"), strings.Contains(contentType, "text/html"):
		result = string(responseBody)

	case strings.Contains(contentType, "application/xml"), strings.Contains(contentType, "text/xml"):
		var xmlResult interface{}
		if err := xml.Unmarshal(responseBody, &xmlResult); err != nil {
			return nil, err
		} else {
			result = xmlResult
		}

	case strings.Contains(contentType, "application/octet-stream"), strings.Contains(contentType, "image/"):
		result = responseBody

	case strings.Contains(contentType, "application/x-www-form-urlencoded"):
		parsedForm, err := url.ParseQuery(string(responseBody))
		if err != nil {
			return nil, err
		} else {
			result = parsedForm
		}

	default:
		result = responseBody
	}

	return result, nil
}

func createJSONBody(body interface{}) (io.Reader, error) {
	actualBody := dereferencePointer(body)
	jsonData, err := json.Marshal(actualBody)
	if err != nil {
		return nil, err
	}
	logger.Debug.Println("JSON Request Body:", string(jsonData))
	return bytes.NewReader(jsonData), nil
}

func createFormURLEncodedBody(body interface{}) (io.Reader, error) {
	actualBody := dereferencePointer(body)
	formData, ok := actualBody.(map[string]string)
	if !ok {
		form, err := JSONToStruct[map[string]string](actualBody)
		if err != nil || form == nil {
			return nil, errors.New("body must be a map[string]string for form-urlencoded content type")
		}
		formData = *form
	}
	values := url.Values{}
	for key, value := range formData {
		values.Set(key, value)
	}
	return strings.NewReader(values.Encode()), nil
}

func createMultipartBody(body interface{}) (io.Reader, string, error) {
	actualBody := dereferencePointer(body)
	formData, ok := actualBody.(map[string]interface{})
	if !ok {
		form, err := JSONToStruct[map[string]interface{}](actualBody)
		if err != nil || form == nil {
			return nil, "", errors.New("body must be a map[string]interface{} for multipart/form-data content type")
		}
		formData = *form

	}
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	for key, value := range formData {
		switch v := value.(type) {
		case string:
			_ = writer.WriteField(key, v)
		case []byte:
			part, err := writer.CreateFormFile(key, key)
			if err != nil {
				return nil, "", err
			}
			_, err = part.Write(v)
			if err != nil {
				return nil, "", err
			}
		case []types.BufferedFile:
			for _, v := range v {
				fileField, err := writer.CreatePart(textproto.MIMEHeader{
					"Content-Disposition": []string{fmt.Sprintf(`form-data; name=%q; filename=%q`, key, v.OriginalName)},
					"Content-Type":        []string{v.MimeType},
					"Content-Encoding":    []string{v.Encoding},
					"Content-Length":      []string{fmt.Sprintf("%d", v.Size)},
				})
				if err != nil {
					return nil, "", err
				}
				_, err = fileField.Write(v.Buffer)
				if err != nil {
					return nil, "", err
				}
			}
		case types.BufferedFile:
			fileField, err := writer.CreatePart(textproto.MIMEHeader{
				"Content-Disposition": []string{fmt.Sprintf(`form-data; name=%q; filename=%q`, key, v.OriginalName)},
				"Content-Type":        []string{v.MimeType},
				"Content-Encoding":    []string{v.Encoding},
				"Content-Length":      []string{fmt.Sprintf("%d", v.Size)},
			})
			if err != nil {
				return nil, "", err
			}
			_, err = fileField.Write(v.Buffer)
			if err != nil {
				return nil, "", err
			}
		default:
			return nil, "", errors.New("unsupported multipart data type")
		}
	}

	err := writer.Close()
	if err != nil {
		return nil, "", err
	}
	return &buf, writer.FormDataContentType(), nil
}

func dereferencePointer(value interface{}) interface{} {
	if value == nil {
		return nil
	}

	v := reflect.ValueOf(value)

	for v.Kind() == reflect.Ptr {
		if v.IsNil() {
			return nil
		}
		v = v.Elem()
	}

	return v.Interface()
}

func FileToBase64(rawURL string) (string, error) {
	sanitizedURL, err := SanitizeURL(rawURL)
	if err != nil {
		return "", fmt.Errorf("invalid URL: %w", err)
	}

	// Get the file from URL
	resp, err := http.Get(sanitizedURL.String())
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	// Read response body
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	// Convert to base64
	base64String := base64.StdEncoding.EncodeToString(data)
	return base64String, nil
}

// Package builtin provides built-in tools for the tool execution framework.
package builtin

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/ainative/ainative-code/internal/tools"
)

// HTTPRequestTool implements a tool for making HTTP requests with security restrictions.
type HTTPRequestTool struct {
	allowedHosts []string
	client       *http.Client
}

// NewHTTPRequestTool creates a new HTTPRequestTool instance.
// If allowedHosts is empty, all hosts are allowed (use with caution).
// A default HTTP client with 30-second timeout is used.
func NewHTTPRequestTool(allowedHosts []string) *HTTPRequestTool {
	return &HTTPRequestTool{
		allowedHosts: allowedHosts,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// Name returns the unique name of the tool.
func (t *HTTPRequestTool) Name() string {
	return "http_request"
}

// Description returns a human-readable description of what the tool does.
func (t *HTTPRequestTool) Description() string {
	return "Makes HTTP/HTTPS requests with method validation, URL restrictions, and response handling"
}

// Schema returns the JSON schema for the tool's input parameters.
func (t *HTTPRequestTool) Schema() tools.ToolSchema {
	maxURLLength := 8192
	maxBodyLength := 1024 * 1024 // 1MB
	return tools.ToolSchema{
		Type: "object",
		Properties: map[string]tools.PropertyDef{
			"url": {
				Type:        "string",
				Description: "The URL to make the request to (must start with http:// or https://)",
				MaxLength:   &maxURLLength,
			},
			"method": {
				Type:        "string",
				Description: "HTTP method to use (default: GET)",
				Enum:        []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS"},
				Default:     "GET",
			},
			"headers": {
				Type:        "object",
				Description: "HTTP headers to send as key-value pairs",
			},
			"body": {
				Type:        "string",
				Description: "Request body content (for POST, PUT, PATCH methods)",
				MaxLength:   &maxBodyLength,
			},
			"timeout_seconds": {
				Type:        "integer",
				Description: "Request timeout in seconds (default: 30, max: 300)",
			},
			"follow_redirects": {
				Type:        "boolean",
				Description: "Whether to follow HTTP redirects (default: true)",
				Default:     true,
			},
			"max_response_size": {
				Type:        "integer",
				Description: "Maximum response body size in bytes (default: 10MB, max: 100MB)",
			},
		},
		Required: []string{"url"},
	}
}

// Execute runs the tool with the provided input and returns the result.
func (t *HTTPRequestTool) Execute(ctx context.Context, input map[string]interface{}) (*tools.Result, error) {
	// Extract and validate URL
	urlRaw, ok := input["url"]
	if !ok {
		return nil, &tools.ErrInvalidInput{
			ToolName: t.Name(),
			Field:    "url",
			Reason:   "url is required",
		}
	}

	urlStr, ok := urlRaw.(string)
	if !ok {
		return nil, &tools.ErrInvalidInput{
			ToolName: t.Name(),
			Field:    "url",
			Reason:   fmt.Sprintf("url must be a string, got %T", urlRaw),
		}
	}

	if urlStr == "" {
		return nil, &tools.ErrInvalidInput{
			ToolName: t.Name(),
			Field:    "url",
			Reason:   "url cannot be empty",
		}
	}

	// Parse and validate URL
	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		return nil, &tools.ErrInvalidInput{
			ToolName: t.Name(),
			Field:    "url",
			Reason:   fmt.Sprintf("invalid URL format: %v", err),
		}
	}

	// Validate URL scheme
	if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
		return nil, &tools.ErrInvalidInput{
			ToolName: t.Name(),
			Field:    "url",
			Reason:   fmt.Sprintf("url scheme must be http or https, got '%s'", parsedURL.Scheme),
		}
	}

	// Validate host against allowed list
	if len(t.allowedHosts) > 0 {
		if err := t.validateHost(parsedURL.Host); err != nil {
			return nil, err
		}
	}

	// Extract method parameter with default
	method := "GET"
	if methodRaw, exists := input["method"]; exists {
		var ok bool
		method, ok = methodRaw.(string)
		if !ok {
			return nil, &tools.ErrInvalidInput{
				ToolName: t.Name(),
				Field:    "method",
				Reason:   fmt.Sprintf("method must be a string, got %T", methodRaw),
			}
		}
		// Normalize to uppercase
		method = strings.ToUpper(method)
	}

	// Extract body parameter
	var body string
	if bodyRaw, exists := input["body"]; exists {
		var ok bool
		body, ok = bodyRaw.(string)
		if !ok {
			return nil, &tools.ErrInvalidInput{
				ToolName: t.Name(),
				Field:    "body",
				Reason:   fmt.Sprintf("body must be a string, got %T", bodyRaw),
			}
		}
	}

	// Extract headers parameter
	headers := make(map[string]string)
	if headersRaw, exists := input["headers"]; exists {
		headersMap, ok := headersRaw.(map[string]interface{})
		if !ok {
			return nil, &tools.ErrInvalidInput{
				ToolName: t.Name(),
				Field:    "headers",
				Reason:   fmt.Sprintf("headers must be an object, got %T", headersRaw),
			}
		}

		// Convert map[string]interface{} to map[string]string
		for key, valueRaw := range headersMap {
			value, ok := valueRaw.(string)
			if !ok {
				return nil, &tools.ErrInvalidInput{
					ToolName: t.Name(),
					Field:    "headers",
					Reason:   fmt.Sprintf("header '%s' must have a string value, got %T", key, valueRaw),
				}
			}
			headers[key] = value
		}
	}

	// Extract timeout_seconds parameter with default
	timeoutSeconds := 30
	if timeoutRaw, exists := input["timeout_seconds"]; exists {
		switch v := timeoutRaw.(type) {
		case float64:
			timeoutSeconds = int(v)
		case int:
			timeoutSeconds = v
		case int64:
			timeoutSeconds = int(v)
		default:
			return nil, &tools.ErrInvalidInput{
				ToolName: t.Name(),
				Field:    "timeout_seconds",
				Reason:   fmt.Sprintf("timeout_seconds must be an integer, got %T", timeoutRaw),
			}
		}

		// Validate timeout bounds
		if timeoutSeconds <= 0 {
			return nil, &tools.ErrInvalidInput{
				ToolName: t.Name(),
				Field:    "timeout_seconds",
				Reason:   "timeout_seconds must be positive",
			}
		}
		if timeoutSeconds > 300 { // 5 minutes max
			return nil, &tools.ErrInvalidInput{
				ToolName: t.Name(),
				Field:    "timeout_seconds",
				Reason:   "timeout_seconds cannot exceed 300 seconds (5 minutes)",
			}
		}
	}

	// Extract follow_redirects parameter with default
	followRedirects := true
	if followRedirectsRaw, exists := input["follow_redirects"]; exists {
		var ok bool
		followRedirects, ok = followRedirectsRaw.(bool)
		if !ok {
			return nil, &tools.ErrInvalidInput{
				ToolName: t.Name(),
				Field:    "follow_redirects",
				Reason:   fmt.Sprintf("follow_redirects must be a boolean, got %T", followRedirectsRaw),
			}
		}
	}

	// Extract max_response_size parameter with default
	maxResponseSize := int64(10 * 1024 * 1024) // 10MB default
	if maxSizeRaw, exists := input["max_response_size"]; exists {
		switch v := maxSizeRaw.(type) {
		case float64:
			maxResponseSize = int64(v)
		case int:
			maxResponseSize = int64(v)
		case int64:
			maxResponseSize = v
		default:
			return nil, &tools.ErrInvalidInput{
				ToolName: t.Name(),
				Field:    "max_response_size",
				Reason:   fmt.Sprintf("max_response_size must be an integer, got %T", maxSizeRaw),
			}
		}

		// Validate max_response_size bounds
		if maxResponseSize <= 0 {
			return nil, &tools.ErrInvalidInput{
				ToolName: t.Name(),
				Field:    "max_response_size",
				Reason:   "max_response_size must be positive",
			}
		}
		if maxResponseSize > 100*1024*1024 { // 100MB absolute maximum
			return nil, &tools.ErrInvalidInput{
				ToolName: t.Name(),
				Field:    "max_response_size",
				Reason:   "max_response_size cannot exceed 100MB",
			}
		}
	}

	// Create HTTP client with timeout and redirect policy
	client := &http.Client{
		Timeout: time.Duration(timeoutSeconds) * time.Second,
	}

	if !followRedirects {
		client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		}
	}

	// Create request
	var reqBody io.Reader
	if body != "" {
		reqBody = bytes.NewBufferString(body)
	}

	req, err := http.NewRequestWithContext(ctx, method, urlStr, reqBody)
	if err != nil {
		return nil, &tools.ErrExecutionFailed{
			ToolName: t.Name(),
			Reason:   fmt.Sprintf("failed to create HTTP request: %v", err),
			Cause:    err,
		}
	}

	// Set headers
	for key, value := range headers {
		req.Header.Set(key, value)
	}

	// Set Content-Length if body is provided
	if body != "" && req.Header.Get("Content-Type") == "" {
		req.Header.Set("Content-Type", "application/json")
	}

	// Record start time
	startTime := time.Now()

	// Execute request
	resp, err := client.Do(req)
	if err != nil {
		// Check if it was a timeout
		if ctx.Err() == context.DeadlineExceeded {
			return nil, &tools.ErrTimeout{
				ToolName: t.Name(),
				Duration: fmt.Sprintf("%d seconds", timeoutSeconds),
			}
		}

		// Check if it was a cancellation
		if ctx.Err() == context.Canceled {
			return nil, &tools.ErrExecutionFailed{
				ToolName: t.Name(),
				Reason:   "request was cancelled",
				Cause:    ctx.Err(),
			}
		}

		return nil, &tools.ErrExecutionFailed{
			ToolName: t.Name(),
			Reason:   fmt.Sprintf("HTTP request failed: %v", err),
			Cause:    err,
		}
	}
	defer resp.Body.Close()

	// Calculate request duration
	duration := time.Since(startTime)

	// Read response body with size limit
	limitedReader := io.LimitReader(resp.Body, maxResponseSize+1)
	respBody, err := io.ReadAll(limitedReader)
	if err != nil {
		return nil, &tools.ErrExecutionFailed{
			ToolName: t.Name(),
			Reason:   "failed to read response body",
			Cause:    err,
		}
	}

	// Check if response exceeded max size
	if int64(len(respBody)) > maxResponseSize {
		return nil, &tools.ErrOutputTooLarge{
			ToolName:   t.Name(),
			OutputSize: int64(len(respBody)),
			MaxSize:    maxResponseSize,
		}
	}

	// Build response headers map
	responseHeaders := make(map[string]string)
	for key, values := range resp.Header {
		// Join multiple header values with comma
		responseHeaders[key] = strings.Join(values, ", ")
	}

	// Build output string
	var output strings.Builder
	output.WriteString(fmt.Sprintf("HTTP %s %s\n", method, urlStr))
	output.WriteString(fmt.Sprintf("Status: %d %s\n", resp.StatusCode, resp.Status))
	output.WriteString(fmt.Sprintf("Duration: %s\n", duration))
	output.WriteString(fmt.Sprintf("Response Size: %d bytes\n", len(respBody)))
	output.WriteString("\n--- RESPONSE HEADERS ---\n")
	for key, value := range responseHeaders {
		output.WriteString(fmt.Sprintf("%s: %s\n", key, value))
	}
	output.WriteString("\n--- RESPONSE BODY ---\n")
	output.WriteString(string(respBody))

	// Build metadata
	metadata := map[string]interface{}{
		"url":              urlStr,
		"method":           method,
		"status_code":      resp.StatusCode,
		"status":           resp.Status,
		"duration_ms":      duration.Milliseconds(),
		"response_size":    len(respBody),
		"response_headers": responseHeaders,
		"content_type":     resp.Header.Get("Content-Type"),
	}

	if len(headers) > 0 {
		metadata["request_headers"] = headers
	}

	if body != "" {
		metadata["request_body_size"] = len(body)
	}

	// Determine success based on status code (2xx is success)
	success := resp.StatusCode >= 200 && resp.StatusCode < 300

	result := &tools.Result{
		Success:  success,
		Output:   output.String(),
		Metadata: metadata,
	}

	return result, nil
}

// Category returns the category this tool belongs to.
func (t *HTTPRequestTool) Category() tools.Category {
	return tools.CategoryNetwork
}

// RequiresConfirmation returns true if this tool requires user confirmation before execution.
func (t *HTTPRequestTool) RequiresConfirmation() bool {
	return true // HTTP requests can modify state and should require confirmation
}

// validateHost checks if the given host is in the allowed list.
func (t *HTTPRequestTool) validateHost(host string) error {
	// If no allowed hosts configured, deny all access
	if len(t.allowedHosts) == 0 {
		return &tools.ErrPermissionDenied{
			ToolName:  t.Name(),
			Operation: "request",
			Resource:  host,
			Reason:    "no allowed hosts configured, HTTP requests denied",
		}
	}

	// Check if host is in allowed list
	for _, allowedHost := range t.allowedHosts {
		if host == allowedHost {
			return nil // Host allowed
		}
	}

	// Host not in allowed list
	return &tools.ErrPermissionDenied{
		ToolName:  t.Name(),
		Operation: "request",
		Resource:  host,
		Reason:    fmt.Sprintf("host '%s' is not in allowed hosts: %v", host, t.allowedHosts),
	}
}

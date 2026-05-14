// Package http provides HTTP client functionality for Goal extensions
// This enables Goal to make HTTP requests to external APIs (e.g., Strava)
//
// Implements requirements from:
// - specs/strava_dashboard_spec_v0.3.md §5.1 (Strava API)
// - specs/adrs/ADR-2_http_extension.md
package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// Request represents an HTTP request
type Request struct {
	Method  string            `json:"method"`
	URL     string            `json:"url"`
	Headers map[string]string `json:"headers"`
	Body    string            `json:"body"`
}

// Response represents an HTTP response
type Response struct {
	Status  int    `json:"status"`
	Body    string `json:"body"`
	Error   string `json:"error"`
}

// Default timeout for HTTP requests
const defaultTimeout = 30 * time.Second

// Default retry configuration
const (
	maxRetries    = 3
	retryDelay    = 1 * time.Second
)

// client is the HTTP client with timeout
var client = &http.Client{
	Timeout: defaultTimeout,
}

// Get makes an HTTP GET request
// Input: JSON string with url and headers
// Output: JSON string with status, body, error
func Get(requestJSON string) string {
	var req Request
	if err := json.Unmarshal([]byte(requestJSON), &req); err != nil {
		return fmtResponse(0, "", fmt.Sprintf("Invalid request JSON: %v", err))
	}

	// Set default method
	if req.Method == "" {
		req.Method = "GET"
	}

	resp, err := doRequest(req)
	if err != nil {
		return fmtResponse(0, "", err.Error())
	}
	return fmtResponse(resp.Status, resp.Body, resp.Error)
}

// Post makes an HTTP POST request
// Input: JSON string with url, headers, and body
// Output: JSON string with status, body, error
func Post(requestJSON string) string {
	var req Request
	if err := json.Unmarshal([]byte(requestJSON), &req); err != nil {
		return fmtResponse(0, "", fmt.Sprintf("Invalid request JSON: %v", err))
	}

	// Set default method
	if req.Method == "" {
		req.Method = "POST"
	}

	resp, err := doRequest(req)
	if err != nil {
		return fmtResponse(0, "", err.Error())
	}
	return fmtResponse(resp.Status, resp.Body, resp.Error)
}

// doRequest executes the HTTP request with retry logic
func doRequest(req Request) (*Response, error) {
	var lastErr error
	var lastResp *http.Response

	for i := 0; i < maxRetries; i++ {
		var bodyReader io.Reader
		if req.Body != "" {
			bodyReader = bytes.NewBufferString(req.Body)
		}

		httpReq, err := http.NewRequest(req.Method, req.URL, bodyReader)
		if err != nil {
			lastErr = err
			continue
		}

		// Add headers
		for key, value := range req.Headers {
			httpReq.Header.Set(key, value)
		}

		lastResp, err = client.Do(httpReq)
		if err != nil {
			lastErr = err
			// Check if we should retry
			if i < maxRetries-1 {
				time.Sleep(retryDelay * time.Duration(i+1))
				continue
			}
			break
		}

		// Success
		body, _ := io.ReadAll(lastResp.Body)
		lastResp.Body.Close()

		return &Response{
			Status: lastResp.StatusCode,
			Body:   string(body),
		}, nil
	}

	return nil, lastErr
}

// fmtResponse formats a Response as JSON string
func fmtResponse(status int, body, err string) string {
	resp := Response{
		Status: status,
		Body:   body,
		Error:  err,
	}
	jsonBytes, _ := json.Marshal(resp)
	return string(jsonBytes)
}

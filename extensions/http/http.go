// Package http provides HTTP client functionality for Goal.
// This enables Goal to make HTTP requests to external APIs (e.g., Strava).
//
// Implements requirements from:
// - specs/strava_dashboard_spec_v0.3.md §5.1 (Strava API)
// - specs/adrs/ADR-2_http_extension.md
//
// To use this extension, create a custom Goal build that imports this package
// and calls http.Import(ctx, ""). See cmd/strava_goal/main.go for example.
package http

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"codeberg.org/anaseto/goal"
)

// Default timeout for HTTP requests
const defaultTimeout = 30 * time.Second

// Default retry configuration
const (
	maxRetries = 3
	retryDelay = 1 * time.Second
)

// client is the HTTP client with timeout
var client = &http.Client{
	Timeout: defaultTimeout,
}

// Request represents an HTTP request structure
type Request struct {
	Method  string            `json:"method"`
	URL     string            `json:"url"`
	Headers map[string]string `json:"headers"`
	Body    string            `json:"body"`
}

// Response represents an HTTP response structure
type Response struct {
	Status  int    `json:"status"`
	Body    string `json:"body"`
	Error   string `json:"error"`
}

// Import registers the HTTP extension functions with the Goal context.
// The following functions are registered:
//
//   http.get[request] : make HTTP GET request, returns response dict
//   http.post[request] : make HTTP POST request, returns response dict
//
// Where request is a dict with keys: url, headers (optional), body (optional)
// Response is a dict with keys: status, body, error
func Import(ctx *goal.Context, pfx string) {
	if pfx != "" {
		pfx += "."
	}

	// Register http.get as a monad (takes one argument: request dict)
	ctx.AssignGlobal(pfx+"http.get", ctx.RegisterMonad("."+pfx+"http.get", vfGet))
	
	// Register http.post as a monad (takes one argument: request dict)
	ctx.AssignGlobal(pfx+"http.post", ctx.RegisterMonad("."+pfx+"http.post", vfPost))
}

// HelpFunc returns help text for the HTTP extension
func HelpFunc() func(string) string {
	return func(s string) string {
		if s == "http" || s == "http.get" || s == "http.post" {
			return `http.get[request]  : make HTTP GET request (request is dict with url, headers, body)
  http.post[request] : make HTTP POST request (request is dict with url, headers, body)
  
  Request dict keys: url (string, required), headers (dict, optional), body (string, optional)
  Response dict keys: status (int), body (string), error (string)
  
  Example:
    request:=(,("url")!"https://api.example.com/data";("headers")!((,"Authorization")!"Bearer token"))
    response:=http.get[request]
    status:=(response["status"])
    body:=(response["body"])
    error:=(response["error"])`
		}
		return ""
	}
}

// vfGet implements the http.get monadic function
func vfGet(ctx *goal.Context, args []goal.V) goal.V {
	if len(args) != 1 {
		return goal.Panicf("http.get: expected 1 argument, got %d", len(args))
	}

	req, ok := args[0].BV().(goal.D)
	if !ok {
		return goal.Panicf("http.get: expected dict argument, got %s", args[0].Type())
	}

	// Parse request dict
	request, err := parseRequest(req)
	if err != nil {
		return goal.Panicf("http.get: %v", err)
	}

	// Set default method
	if request.Method == "" {
		request.Method = "GET"
	}

	// Execute request with retries
	resp, err := doRequest(request)
	if err != nil {
		return makeResponse(0, "", err.Error())
	}

	return makeResponse(resp.Status, resp.Body, resp.Error)
}

// vfPost implements the http.post monadic function
func vfPost(ctx *goal.Context, args []goal.V) goal.V {
	if len(args) != 1 {
		return goal.Panicf("http.post: expected 1 argument, got %d", len(args))
	}

	req, ok := args[0].BV().(goal.D)
	if !ok {
		return goal.Panicf("http.post: expected dict argument, got %s", args[0].Type())
	}

	// Parse request dict
	request, err := parseRequest(req)
	if err != nil {
		return goal.Panicf("http.post: %v", err)
	}

	// Set default method
	if request.Method == "" {
		request.Method = "POST"
	}

	// Execute request with retries
	resp, err := doRequest(request)
	if err != nil {
		return makeResponse(0, "", err.Error())
	}

	return makeResponse(resp.Status, resp.Body, resp.Error)
}

// parseRequest converts a Goal dict to a Request struct
func parseRequest(d goal.D) (*Request, error) {
	req := &Request{}
	
	// Get URL (required)
	urlVal, ok := d.Get("url")
	if !ok {
		return nil, fmt.Errorf("missing required field: url")
	}
	urlStr, ok := urlVal.BV().(goal.S)
	if !ok {
		return nil, fmt.Errorf("url must be a string, got %s", urlVal.Type())
	}
	req.URL = string(urlStr)

	// Get method (optional)
	if methodVal, ok := d.Get("method"); ok {
		if methodStr, ok := methodVal.BV().(goal.S); ok {
			req.Method = string(methodStr)
		}
	}

	// Get headers (optional)
	if headersVal, ok := d.Get("headers"); ok {
		if headersDict, ok := headersVal.BV().(goal.D); ok {
			req.Headers = make(map[string]string)
			for _, k := range headersDict.Keys() {
				if v, ok := headersDict.Get(k); ok {
					if vStr, ok := v.BV().(goal.S); ok {
						req.Headers[k] = string(vStr)
					}
				}
			}
		}
	}

	// Get body (optional)
	if bodyVal, ok := d.Get("body"); ok {
		if bodyStr, ok := bodyVal.BV().(goal.S); ok {
			req.Body = string(bodyStr)
		}
	}

	return req, nil
}

// doRequest executes the HTTP request with retry logic
func doRequest(req *Request) (*Response, error) {
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

// makeResponse creates a Goal dict response
func makeResponse(status int, body, err string) goal.V {
	result := goal.NewD(
		goal.NewAS([]string{"status", "body", "error"}),
		goal.NewAS([]string{fmt.Sprintf("%d", status), body, err}),
	)
	return result
}

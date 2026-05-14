// Package json provides JSON parsing and encoding functionality for Goal extensions
// This enables Goal to work with JSON data (e.g., Strava API responses)
//
// Implements requirements from:
// - specs/strava_dashboard_spec_v0.3.md §5.1 (Strava API - JSON responses)
// - specs/adrs/ADR-2_http_extension.md (needs JSON parsing)
package main

import (
	"encoding/json"
	"fmt"
)

// ParseRequest represents a request to parse JSON
type ParseRequest struct {
	JSON string `json:"json"`
}

// ParseResponse represents the response from parsing JSON
type ParseResponse struct {
	Result map[string]interface{} `json:"result"`
	Error  string                `json:"error"`
}

// EncodeRequest represents a request to encode data to JSON
type EncodeRequest struct {
	Data interface{} `json:"data"`
}

// EncodeResponse represents the response from encoding to JSON
type EncodeResponse struct {
	JSON  string `json:"json"`
	Error string `json:"error"`
}

// Parse parses a JSON string into a Goal-accessible structure
// Input: JSON string with "json" field containing the JSON to parse
// Output: JSON string with "result" (parsed object) and "error" fields
func Parse(requestJSON string) string {
	var req ParseRequest
	if err := json.Unmarshal([]byte(requestJSON), &req); err != nil {
		return fmtResponse(ParseResponse{Error: fmt.Sprintf("Invalid request JSON: %v", err)})
	}

	if req.JSON == "" {
		return fmtResponse(ParseResponse{Error: "JSON field is required"})
	}

	var result map[string]interface{}
	if err := json.Unmarshal([]byte(req.JSON), &result); err != nil {
		return fmtResponse(ParseResponse{Error: fmt.Sprintf("Invalid JSON: %v", err)})
	}

	return fmtResponse(ParseResponse{Result: result})
}

// Encode encodes a Goal data structure to JSON
// Input: JSON string with "data" field containing the data to encode
// Output: JSON string with "json" and "error" fields
func Encode(requestJSON string) string {
	var req EncodeRequest
	if err := json.Unmarshal([]byte(requestJSON), &req); err != nil {
		return fmtResponse(EncodeResponse{Error: fmt.Sprintf("Invalid request JSON: %v", err)})
	}

	jsonBytes, err := json.Marshal(req.Data)
	if err != nil {
		return fmtResponse(EncodeResponse{Error: fmt.Sprintf("Encode error: %v", err)})
	}

	return fmtResponse(EncodeResponse{JSON: string(jsonBytes)})
}

// fmtResponse formats a response as JSON string
func fmtResponse[T any](resp T) string {
	jsonBytes, err := json.Marshal(resp)
	if err != nil {
		return `{"error": "Internal error"}`
	}
	return string(jsonBytes)
}

func main() {
	// Test parsing
	testJSON := `{"name": "test", "value": 123}`
	parseReq := ParseRequest{JSON: testJSON}
	parseReqJSON, _ := json.Marshal(parseReq)
	parseResp := Parse(string(parseReqJSON))
	fmt.Println("Parse:", parseResp)

	// Test encoding
	encodeReq := EncodeRequest{Data: map[string]interface{}{"key": "value"}}
	encodeReqJSON, _ := json.Marshal(encodeReq)
	encodeResp := Encode(string(encodeReqJSON))
	fmt.Println("Encode:", encodeResp)
}

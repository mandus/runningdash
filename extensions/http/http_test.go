package http_test

import (
	"fmt"
	"strings"
	"testing"

	"codeberg.org/anaseto/goal"
	"github.com/mandus/runningdash/extensions/http"
)

func TestHTTPGet(t *testing.T) {
	ctx := goal.NewContext()
	http.Import(ctx, "")

	// Test simple GET request to httpbin
	// Note: This test requires network access
	// In a real test environment, you might want to use a local test server
	
	// For now, just test that the function is registered
	_, err := ctx.Eval(`http.get[(,("url")!"https://httpbin.org/get")]`)
	if err != nil {
		// Network error is expected in sandbox, but function should be callable
		if !strings.Contains(err.Error(), "connection refused") && 
		   !strings.Contains(err.Error(), "no such host") &&
		   !strings.Contains(err.Error(), "i/o timeout") {
			t.Fatalf("Unexpected error: %v", err)
		}
		// Network error is OK for this test
		return
	}
	
	// If we got here, the request succeeded
	x, err := ctx.Eval(`http.get[(,("url")!"https://httpbin.org/get")]`)
	if err != nil {
		t.Fatalf("http.get failed: %v", err)
	}
	
	// Check that response is a dict
	if x.Type() != goal.DT {
		t.Fatalf("Expected dict, got %s", x.Type())
	}
	
	// Check for status key
	status, ok := x.BV().(goal.D).Get("status")
	if !ok {
		t.Fatal("Response missing 'status' key")
	}
	
	// Status should be 200
	if statusStr, ok := status.BV().(goal.S); !ok {
		t.Fatalf("Status is not a string: %s", status.Type())
	} else if string(statusStr) != "200" {
		t.Fatalf("Expected status 200, got %s", statusStr)
	}
}

func TestHTTPPost(t *testing.T) {
	ctx := goal.NewContext()
	http.Import(ctx, "")

	// Test POST request
	// Note: This test requires network access
	
	_, err := ctx.Eval(`http.post[(,("url")!"https://httpbin.org/post";("body")!"test data")]`)
	if err != nil {
		// Network error is expected in sandbox
		if !strings.Contains(err.Error(), "connection refused") && 
		   !strings.Contains(err.Error(), "no such host") &&
		   !strings.Contains(err.Error(), "i/o timeout") {
			t.Fatalf("Unexpected error: %v", err)
		}
		return
	}
	
	// If we got here, the request succeeded
	x, err := ctx.Eval(`http.post[(,("url")!"https://httpbin.org/post";("body")!"test data")]`)
	if err != nil {
		t.Fatalf("http.post failed: %v", err)
	}
	
	// Check that response is a dict
	if x.Type() != goal.DT {
		t.Fatalf("Expected dict, got %s", x.Type())
	}
}

func TestHTTPErrorHandling(t *testing.T) {
	ctx := goal.NewContext()
	http.Import(ctx, "")

	// Test with invalid URL - should return error in response
	x, err := ctx.Eval(`http.get[(,("url")!"invalid-url")]`)
	if err != nil {
		// The function itself should not panic, it should return an error in the response
		t.Fatalf("http.get should not panic on invalid URL: %v", err)
	}
	
	// Check that response contains error
	resp, ok := x.BV().(goal.D)
	if !ok {
		t.Fatalf("Expected dict response, got %s", x.Type())
	}
	
	errVal, ok := resp.Get("error")
	if !ok {
		t.Fatal("Response missing 'error' key for invalid URL")
	}
	
	errStr, ok := errVal.BV().(goal.S)
	if !ok {
		t.Fatalf("Error is not a string: %s", errVal.Type())
	}
	
	if string(errStr) == "" {
		t.Fatal("Expected non-empty error for invalid URL")
	}
	
	fmt.Printf("Got expected error: %s\n", errStr)
}

func TestMissingURL(t *testing.T) {
	ctx := goal.NewContext()
	http.Import(ctx, "")

	// Test with missing URL - should panic
	_, err := ctx.Eval(`http.get[((,"headers")!((,"test")!"value"))]`)
	if err == nil {
		t.Fatal("Expected error for missing URL")
	}
	
	if !strings.Contains(err.Error(), "missing required field: url") {
		t.Fatalf("Expected 'missing required field: url' error, got: %v", err)
	}
}

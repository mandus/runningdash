package http_test

import (
	"strings"
	"testing"

	"codeberg.org/anaseto/goal"
	"github.com/mandus/runningdash/extensions/http"
)

// newCtx builds a fresh Goal context with the http extension registered.
func newCtx(t *testing.T) *goal.Context {
	t.Helper()
	ctx := goal.NewContext()
	http.Import(ctx, "")
	return ctx
}

// isNetworkErr returns true for transport errors that are expected in a
// sandboxed test environment.
func isNetworkErr(s string) bool {
	return strings.Contains(s, "connection refused") ||
		strings.Contains(s, "no such host") ||
		strings.Contains(s, "i/o timeout") ||
		strings.Contains(s, "dial tcp") ||
		strings.Contains(s, "timeout")
}

// getStringField extracts a string field from a Goal dict response.
func getStringField(t *testing.T, v goal.V, name string) (string, bool) {
	t.Helper()
	d, ok := v.BV().(*goal.D)
	if !ok {
		t.Fatalf("expected dict response, got %q", v.Type())
	}
	f, ok := d.GetS(name)
	if !ok {
		return "", false
	}
	s, ok := f.BV().(goal.S)
	if !ok {
		return "", false
	}
	return string(s), true
}

// getIntField extracts an int field from a Goal dict response.
func getIntField(t *testing.T, v goal.V, name string) (int64, bool) {
	t.Helper()
	d, ok := v.BV().(*goal.D)
	if !ok {
		t.Fatalf("expected dict response, got %q", v.Type())
	}
	f, ok := d.GetS(name)
	if !ok {
		return 0, false
	}
	if !f.IsI() {
		return 0, false
	}
	return f.I(), true
}

func TestHTTPGetRegistered(t *testing.T) {
	ctx := newCtx(t)

	v, err := ctx.Eval(`http.get[(,"url")!,"https://httpbin.org/get"]`)
	if err != nil {
		t.Fatalf("http.get eval failed: %v", err)
	}
	if v.Type() != "d" {
		t.Fatalf("expected dict response, got %q", v.Type())
	}

	// If we have network, status should be 200 and error empty.
	// Otherwise, status should be 0 and error non-empty.
	status, ok := getIntField(t, v, "status")
	if !ok {
		t.Fatal("response missing 'status' key")
	}
	errStr, _ := getStringField(t, v, "error")

	if errStr != "" {
		if !isNetworkErr(errStr) {
			t.Fatalf("unexpected error: %s", errStr)
		}
		if status != 0 {
			t.Fatalf("expected status 0 on network error, got %d", status)
		}
		t.Logf("skipping live-network assertions: %s", errStr)
		return
	}
	if status != 200 {
		t.Fatalf("expected status 200, got %d", status)
	}
}

func TestHTTPPostRegistered(t *testing.T) {
	ctx := newCtx(t)

	v, err := ctx.Eval(`http.post[("url";"body")!("https://httpbin.org/post";"test data")]`)
	if err != nil {
		t.Fatalf("http.post eval failed: %v", err)
	}
	if v.Type() != "d" {
		t.Fatalf("expected dict response, got %q", v.Type())
	}
}

func TestHTTPErrorHandling(t *testing.T) {
	ctx := newCtx(t)

	// Invalid URL — should return an error in the response, not panic.
	v, err := ctx.Eval(`http.get[(,"url")!,"invalid-url"]`)
	if err != nil {
		t.Fatalf("http.get must not panic on invalid URL: %v", err)
	}
	errStr, ok := getStringField(t, v, "error")
	if !ok {
		t.Fatal("response missing 'error' key for invalid URL")
	}
	if errStr == "" {
		t.Fatal("expected non-empty error for invalid URL")
	}
	t.Logf("got expected error: %s", errStr)
}

func TestMissingURL(t *testing.T) {
	ctx := newCtx(t)

	// Missing URL — should produce a Goal panic surfaced as an error.
	_, err := ctx.Eval(`http.get[(,"headers")!,((,"test")!,"value")]`)
	if err == nil {
		t.Fatal("expected error for missing URL")
	}
	if !strings.Contains(err.Error(), "missing required field: url") {
		t.Fatalf("expected 'missing required field: url' error, got: %v", err)
	}
}

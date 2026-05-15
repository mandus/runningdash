// Package http provides HTTP client functionality for Goal.
// This enables Goal to make HTTP requests to external APIs (e.g., Strava).
//
// Implements requirements from:
// - specs/strava_dashboard_spec_v0.3.md §5.1 (Strava API)
// - specs/adrs/ADR-2_http_extension.md
//
// To use this extension, create a custom Goal build that imports this package
// and calls http.Import(ctx, ""). See cmd/strava_goal/main.go for an example.
package http

import (
	"bytes"
	"fmt"
	"io"
	nethttp "net/http"
	"strings"
	"time"

	"codeberg.org/anaseto/goal"
)

// Default timeout for HTTP requests.
const defaultTimeout = 30 * time.Second

// Default retry configuration.
const (
	maxRetries = 3
	retryDelay = 1 * time.Second
)

// client is the HTTP client with timeout.
var client = &nethttp.Client{
	Timeout: defaultTimeout,
}

// request represents an HTTP request structure (internal).
type request struct {
	Method  string
	URL     string
	Headers map[string]string
	Body    string
}

// response represents an HTTP response structure (internal).
type response struct {
	Status int
	Body   string
	Error  string
}

// Import registers the HTTP extension functions with the Goal context.
//
// The following functions are registered:
//
//	http.get[request]  : make HTTP GET request, returns response dict
//	http.post[request] : make HTTP POST request, returns response dict
//
// Where request is a dict with keys: url (required), method (optional),
// headers (optional dict of string->string), body (optional string).
//
// Response is a dict with keys:
//
//	status (int)    : HTTP status code (0 on transport error)
//	body   (string) : response body
//	error  (string) : error message, empty on success
//
// The pfx argument is an optional prefix prepended to the registered names
// (e.g., pfx="x" registers "x.http.get"). Pass "" for the default names.
func Import(ctx *goal.Context, pfx string) {
	ctx.RegisterExtension("net/http", "")
	if pfx != "" {
		pfx += "."
	}

	ctx.AssignGlobal(pfx+"http.get", ctx.RegisterMonad("."+pfx+"http.get", vfGet))
	ctx.AssignGlobal(pfx+"http.post", ctx.RegisterMonad("."+pfx+"http.post", vfPost))
}

// HelpFunc returns a help function suitable for help.Wrap.
func HelpFunc() func(string) string {
	return func(s string) string {
		if strings.HasPrefix(s, "http.") || s == "http" || s == "net/http" {
			return strings.TrimSpace(`
http.get[d]   perform HTTP GET; d is a dict with keys url, headers, body
http.post[d]  perform HTTP POST; d is a dict with keys url, headers, body

  Response dict keys:
    status (int)    HTTP status code, 0 on transport error
    body   (string) response body
    error  (string) error message, empty on success`)
		}
		return ""
	}
}

// vfGet implements the http.get monadic function.
func vfGet(_ *goal.Context, args []goal.V) goal.V {
	if len(args) != 1 {
		return goal.Panicf("http.get : expected 1 argument, got %d", len(args))
	}
	return doVerb("http.get", "GET", args[0])
}

// vfPost implements the http.post monadic function.
func vfPost(_ *goal.Context, args []goal.V) goal.V {
	if len(args) != 1 {
		return goal.Panicf("http.post : expected 1 argument, got %d", len(args))
	}
	return doVerb("http.post", "POST", args[0])
}

// doVerb is the shared implementation for http.get and http.post.
func doVerb(verbName, defaultMethod string, arg goal.V) goal.V {
	d, ok := arg.BV().(*goal.D)
	if !ok {
		return goal.Panicf("%s : expected dict argument, got %q", verbName, arg.Type())
	}

	req, err := parseRequest(d)
	if err != nil {
		return goal.Panicf("%s : %v", verbName, err)
	}
	if req.Method == "" {
		req.Method = defaultMethod
	}

	resp := doRequest(req)
	return makeResponse(resp)
}

// parseRequest converts a Goal dict to a request struct.
func parseRequest(d *goal.D) (*request, error) {
	req := &request{}

	// URL (required).
	urlVal, ok := d.GetS("url")
	if !ok {
		return nil, fmt.Errorf("missing required field: url")
	}
	urlStr, ok := urlVal.BV().(goal.S)
	if !ok {
		return nil, fmt.Errorf("url must be a string, got %q", urlVal.Type())
	}
	req.URL = string(urlStr)

	// Method (optional).
	if methodVal, ok := d.GetS("method"); ok {
		if methodStr, ok := methodVal.BV().(goal.S); ok {
			req.Method = string(methodStr)
		}
	}

	// Headers (optional dict).
	if headersVal, ok := d.GetS("headers"); ok {
		if hd, ok := headersVal.BV().(*goal.D); ok {
			keys := hd.KeyArray()
			vals := hd.ValueArray()
			req.Headers = make(map[string]string, keys.Len())
			for i := 0; i < keys.Len(); i++ {
				kS, kok := keys.At(i).BV().(goal.S)
				vS, vok := vals.At(i).BV().(goal.S)
				if kok && vok {
					req.Headers[string(kS)] = string(vS)
				}
			}
		}
	}

	// Body (optional string).
	if bodyVal, ok := d.GetS("body"); ok {
		if bodyStr, ok := bodyVal.BV().(goal.S); ok {
			req.Body = string(bodyStr)
		}
	}

	return req, nil
}

// doRequest executes the HTTP request with retry logic. It always returns a
// non-nil response; transport failures are reported via response.Error with
// status 0.
func doRequest(req *request) *response {
	var lastErr error

	for i := 0; i < maxRetries; i++ {
		var bodyReader io.Reader
		if req.Body != "" {
			bodyReader = bytes.NewBufferString(req.Body)
		}

		httpReq, err := nethttp.NewRequest(req.Method, req.URL, bodyReader)
		if err != nil {
			// Bad request construction is not retryable.
			return &response{Status: 0, Error: err.Error()}
		}

		for key, value := range req.Headers {
			httpReq.Header.Set(key, value)
		}

		httpResp, err := client.Do(httpReq)
		if err != nil {
			lastErr = err
			if i < maxRetries-1 {
				time.Sleep(time.Duration(1<<i) * retryDelay)
				continue
			}
			break
		}

		body, _ := io.ReadAll(httpResp.Body)
		httpResp.Body.Close()

		return &response{
			Status: httpResp.StatusCode,
			Body:   string(body),
		}
	}

	msg := "unknown error"
	if lastErr != nil {
		msg = lastErr.Error()
	}
	return &response{Status: 0, Error: msg}
}

// makeResponse creates a Goal dict response with mixed-typed values.
func makeResponse(r *response) goal.V {
	keys := goal.NewAS([]string{"status", "body", "error"})
	values := goal.NewAV([]goal.V{
		goal.NewI(int64(r.Status)),
		goal.NewS(r.Body),
		goal.NewS(r.Error),
	})
	return goal.NewD(keys, values)
}

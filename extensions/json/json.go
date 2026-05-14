// Package json provides JSON encoding and decoding functionality for Goal.
// This enables Goal to work with JSON data (e.g., Strava API responses).
//
// Implements requirements from:
// - specs/strava_dashboard_spec_v0.3.md §5.1 (Strava API - JSON responses)
// - specs/adrs/ADR-2_http_extension.md (needs JSON parsing)
//
// To use this extension, create a custom Goal build that imports this package
// and calls json.Import(ctx, ""). See cmd/strava_goal/main.go for example.
package json

import (
	"encoding/json"
	"fmt"
	"strings"

	"codeberg.org/anaseto/goal"
)

// Import registers the JSON extension functions with the Goal context.
// The following functions are registered:
//
//   json.parse[string] : parse JSON string, returns dict or array
//   json.encode[value] : encode Goal value to JSON string
//
// json.parse takes a string and returns the parsed JSON as a Goal value:
//   - JSON objects become Goal dicts
//   - JSON arrays become Goal arrays
//   - JSON strings become Goal strings
//   - JSON numbers become Goal numbers
//   - JSON booleans become Goal booleans
//   - JSON null becomes Goal nil
//
// json.encode takes any Goal value and returns its JSON representation as a string.
func Import(ctx *goal.Context, pfx string) {
	if pfx != "" {
		pfx += "."
	}

	// Register json.parse as a monad (takes JSON string, returns parsed value)
	ctx.AssignGlobal(pfx+"json.parse", ctx.RegisterMonad("."+pfx+"json.parse", vfParse))
	
	// Register json.encode as a monad (takes any value, returns JSON string)
	ctx.AssignGlobal(pfx+"json.encode", ctx.RegisterMonad("."+pfx+"json.encode", vfEncode))
}

// HelpFunc returns help text for the JSON extension
func HelpFunc() func(string) string {
	return func(s string) string {
		if strings.HasPrefix(s, "json.") || s == "json" {
			return `json.parse[string] : parse JSON string, returns dict/array/value
  json.encode[value] : encode Goal value to JSON string
  
  Example:
    data:=json.parse["{\"name\":\"test\",\"value\":123}"]
    name:=(data["name"])
    
    encoded:=json.encode[(,("a")!1;("b")!2)]
    / encoded is "{"a":1,"b":2}"`
		}
		return ""
	}
}

// vfParse implements the json.parse monadic function
func vfParse(ctx *goal.Context, args []goal.V) goal.V {
	if len(args) != 1 {
		return goal.Panicf("json.parse: expected 1 argument (JSON string), got %d", len(args))
	}

	jsonStr, ok := args[0].BV().(goal.S)
	if !ok {
		return goal.Panicf("json.parse: expected string, got %s", args[0].Type())
	}

	// Parse the JSON
	var result interface{}
	if err := json.Unmarshal([]byte(jsonStr), &result); err != nil {
		return goal.Panicf("json.parse: invalid JSON: %v", err)
	}

	// Convert to Goal value
	return goValueToGoalValue(result)
}

// vfEncode implements the json.encode monadic function
func vfEncode(ctx *goal.Context, args []goal.V) goal.V {
	if len(args) != 1 {
		return goal.Panicf("json.encode: expected 1 argument, got %d", len(args))
	}

	// Convert Goal value to Go value
	goVal := goalValueToGoValue(args[0])

	// Marshal to JSON
	jsonBytes, err := json.Marshal(goVal)
	if err != nil {
		return goal.Panicf("json.encode: %v", err)
	}

	return goal.NewS(string(jsonBytes))
}

// goValueToGoalValue converts a Go value (from json.Unmarshal) to a Goal value
func goValueToGoalValue(v interface{}) goal.V {
	switch val := v.(type) {
	case nil:
		return goal.NV()
	case bool:
		if val {
			return goal.TW()
		}
		return goal.FW()
	case float64:
		// JSON numbers are float64
		// Check if it's actually an integer
		if float64(int64(val)) == val {
			return goal.NewI(int64(val))
		}
		return goal.NewF(val)
	case string:
		return goal.NewS(val)
	case []interface{}:
		// JSON array -> Goal array
		arr := make([]goal.V, len(val))
		for i, item := range val {
			arr[i] = goValueToGoalValue(item)
		}
		return goal.NewAS(arr)
	case map[string]interface{}:
		// JSON object -> Goal dict
		keys := make([]string, 0, len(val))
		values := make([]goal.V, 0, len(val))
		for k, v := range val {
			keys = append(keys, k)
			values = append(values, goValueToGoalValue(v))
		}
		return goal.NewD(goal.NewAS(keys), goal.NewAS(values))
	default:
		return goal.Panicf("json.parse: unsupported type %T", v)
	}
}

// goalValueToGoValue converts a Goal value to a Go value (for json.Marshal)
func goalValueToGoValue(v goal.V) interface{} {
	switch v.Type() {
	case goal.NT:
		return nil
	case goal.TT:
		return true
	case goal.FT:
		return false
	case goal.IT:
		return v.BV().(goal.I)
	case goal.FT:
		return v.BV().(goal.F)
	case goal.ST:
		return v.BV().(goal.S)
	case goal.AT:
		arr := v.BV().(*goal.AS)
		result := make([]interface{}, arr.Len())
		for i, item := range arr.Slice {
			result[i] = goalValueToGoValue(item)
		}
		return result
	case goal.DT:
		dict := v.BV().(goal.D)
		result := make(map[string]interface{})
		for _, k := range dict.Keys() {
			if val, ok := dict.Get(k); ok {
				result[k] = goalValueToGoValue(val)
			}
		}
		return result
	default:
		// For other types, convert to string
		return v.Sprint(nil, false)
	}
}

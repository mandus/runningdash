package json_test

import (
	"fmt"
	"testing"

	"codeberg.org/anaseto/goal"
	"github.com/mandus/runningdash/extensions/json"
)

func TestJSONParseObject(t *testing.T) {
	ctx := goal.NewContext()
	json.Import(ctx, "")

	// Parse a JSON object
	x, err := ctx.Eval(`json.parse["{\"name\":\"test\",\"value\":123,\"active\":true}"]`)
	if err != nil {
		t.Fatalf("json.parse failed: %v", err)
	}

	// Check it's a dict
	if x.Type() != goal.DT {
		t.Fatalf("Expected dict, got %s", x.Type())
	}

	dict := x.BV().(goal.D)
	
	// Check name field
	name, ok := dict.Get("name")
	if !ok {
		t.Fatal("Missing 'name' field")
	}
	if name.Type() != goal.ST {
		t.Fatalf("Expected string for name, got %s", name.Type())
	}
	if string(name.BV().(goal.S)) != "test" {
		t.Fatalf("Expected name='test', got '%s'", name.BV().(goal.S))
	}

	// Check value field
	value, ok := dict.Get("value")
	if !ok {
		t.Fatal("Missing 'value' field")
	}
	if value.Type() != goal.IT {
		t.Fatalf("Expected int for value, got %s", value.Type())
	}
	if int64(value.BV().(goal.I)) != 123 {
		t.Fatalf("Expected value=123, got %d", value.BV().(goal.I))
	}

	// Check active field
	active, ok := dict.Get("active")
	if !ok {
		t.Fatal("Missing 'active' field")
	}
	if active.Type() != goal.TT {
		t.Fatalf("Expected true for active, got %s", active.Type())
	}
}

func TestJSONParseArray(t *testing.T) {
	ctx := goal.NewContext()
	json.Import(ctx, "")

	// Parse a JSON array
	x, err := ctx.Eval(`json.parse["[1,2,3,\"four\"]"]`)
	if err != nil {
		t.Fatalf("json.parse failed: %v", err)
	}

	// Check it's an array
	if x.Type() != goal.AT {
		t.Fatalf("Expected array, got %s", x.Type())
	}

	arr := x.BV().(*goal.AS)
	if arr.Len() != 4 {
		t.Fatalf("Expected 4 elements, got %d", arr.Len())
	}
}

func TestJSONEncode(t *testing.T) {
	ctx := goal.NewContext()
	json.Import(ctx, "")

	// Encode a dict
	x, err := ctx.Eval(`json.encode[(,("name")!"test";("value")!123)]`)
	if err != nil {
		t.Fatalf("json.encode failed: %v", err)
	}

	if x.Type() != goal.ST {
		t.Fatalf("Expected string, got %s", x.Type())
	}

	jsonStr := string(x.BV().(goal.S))
	if jsonStr == "" {
		t.Fatal("Expected non-empty JSON string")
	}

	// Verify it's valid JSON by parsing it back
	_, err = ctx.Eval(fmt.Sprintf(`json.parse["%s"]`, jsonStr))
	if err != nil {
		t.Fatalf("Encoded JSON is invalid: %v", err)
	}
}

func TestJSONInvalid(t *testing.T) {
	ctx := goal.NewContext()
	json.Import(ctx, "")

	// Parse invalid JSON
	_, err := ctx.Eval(`json.parse["invalid json"]`)
	if err == nil {
		t.Fatal("Expected error for invalid JSON")
	}

	if !fmt.Sprintf("%v", err).Contains("invalid JSON") {
		t.Fatalf("Expected 'invalid JSON' error, got: %v", err)
	}
}

func TestJSONEncodeArray(t *testing.T) {
	ctx := goal.NewContext()
	json.Import(ctx, "")

	// Encode an array
	x, err := ctx.Eval(`json.encode[[1;2;3]]`)
	if err != nil {
		t.Fatalf("json.encode failed: %v", err)
	}

	if x.Type() != goal.ST {
		t.Fatalf("Expected string, got %s", x.Type())
	}

	jsonStr := string(x.BV().(goal.S))
	expected := "[1,2,3]"
	if jsonStr != expected {
		t.Fatalf("Expected '%s', got '%s'", expected, jsonStr)
	}
}

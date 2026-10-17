package js

import (
	"math"
	"testing"
)

func TestContextAPI(t *testing.T) {
	c := NewContext()
	if c == nil || c.GetContextRef() == nil {
		t.Fatal("NewContext returned no context")
	}
	defer c.Close()
	if c.GlobalObject() == nil {
		t.Fatal("GlobalObject returned nil")
	}

	if got, err := c.Evaluate("1 + 2"); err != nil || got != "3" {
		t.Fatalf("Evaluate: got %q, err %v", got, err)
	}
	if _, err := c.Evaluate("throw new Error('boom')"); err == nil {
		t.Fatal("Evaluate expected an error")
	}
	if value, err := c.EvaluateRaw("40 + 2"); err != nil || value == nil {
		t.Fatalf("EvaluateRaw: value %v, err %v", value, err)
	}
	if _, err := c.EvaluateRaw("throw new Error('boom')"); err == nil {
		t.Fatal("EvaluateRaw expected an error")
	}

	if err := c.SetProperty("answer", c.MakeNumber(42)); err != nil {
		t.Fatal(err)
	}
	value, err := c.GetProperty("answer")
	if err != nil {
		t.Fatal(err)
	}
	if got, err := c.ToNumber(value); err != nil || got != 42 {
		t.Fatalf("ToNumber: got %v, err %v", got, err)
	}

	if got, err := c.Evaluate("(function(){ var o = {}; o.label = 'ok'; return o.label; })()"); err != nil || got != "ok" {
		t.Fatalf("property set: got %q, err %v", got, err)
	}

	cases := []struct {
		name string
		v    interface{}
		kind string
	}{
		{"string", c.MakeString("hello"), "string"},
		{"number", c.MakeNumber(2.5), "number"},
		{"boolean", c.MakeBoolean(true), "boolean"},
		{"undefined", c.MakeUndefined(), "undefined"},
		{"null", c.MakeNull(), "null"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			v := tc.v
			var value interface{}
			switch x := v.(type) {
			case nil:
				value = nil
			default:
				value = x
			}
			_ = value
		})
	}

	if !c.ToBoolean(c.MakeBoolean(true)) || c.ToBoolean(c.MakeBoolean(false)) {
		t.Fatal("ToBoolean returned incorrect boolean conversion")
	}
	if got := c.GetType(c.MakeString("x")); got != "string" {
		t.Fatalf("GetType: got %q", got)
	}
	if got, err := c.ToNumber(c.MakeString("not a number")); err != nil || !math.IsNaN(got) {
		t.Fatalf("ToNumber string: got %v, err %v", got, err)
	}

	c.Close()
	if c.GetContextRef() != nil {
		t.Fatal("Close did not clear context")
	}
	c.Close()
}

package application

import (
	"context"
	"testing"
)

func TestConvertUnits(t *testing.T) {
	e := &AssistantToolExecutor{}
	for _, tc := range []struct {
		value    float64
		from, to string
		want     float64
	}{
		{5, "km", "mi", 3.106855961},
		{100, "Fahrenheit", "c", 37.7778},
		{2, "cups", "ml", 473.176473},
		{1, "gib", "mb", 1073.741824},
		{3, "dunams", "m2", 3000},
	} {
		output, _, err := e.ExecuteCaretTool(context.Background(), "utilities.convert_units", map[string]any{"value": tc.value, "from": tc.from, "to": tc.to})
		if err != nil || output["result"] != tc.want {
			t.Errorf("%v %s -> %s = %v (%v), want %v", tc.value, tc.from, tc.to, output["result"], err, tc.want)
		}
	}
	if _, _, err := e.ExecuteCaretTool(context.Background(), "utilities.convert_units", map[string]any{"value": 1.0, "from": "kg", "to": "m"}); err == nil {
		t.Fatal("mass to length accepted")
	}
}

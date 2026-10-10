package application

import (
	"context"
	"errors"
	"fmt"
	"strings"
)

func init() {
	registerGeneralTool("utilities.convert_units", (*AssistantToolExecutor).convertUnits)
}

type unitDef struct {
	kind   string
	factor float64 // to the kind's base unit
}

// units maps names and symbols to a factor of the base unit: metre,
// kilogram, litre, square metre, metre per second, byte, second, joule.
// Temperature is converted separately because it has offsets.
var units = map[string]unitDef{}

func addUnits(kind string, factor float64, names ...string) {
	for _, name := range names {
		units[name] = unitDef{kind: kind, factor: factor}
	}
}

func init() {
	addUnits("length", 1, "m", "meter", "meters", "metre", "metres")
	addUnits("length", 1000, "km", "kilometer", "kilometers", "kilometre", "kilometres")
	addUnits("length", 0.01, "cm", "centimeter", "centimeters", "centimetre", "centimetres")
	addUnits("length", 0.001, "mm", "millimeter", "millimeters", "millimetre", "millimetres")
	addUnits("length", 0.0254, "in", "inch", "inches")
	addUnits("length", 0.3048, "ft", "foot", "feet")
	addUnits("length", 0.9144, "yd", "yard", "yards")
	addUnits("length", 1609.344, "mi", "mile", "miles")
	addUnits("length", 1852, "nmi", "nautical mile", "nautical miles")
	addUnits("mass", 1, "kg", "kilogram", "kilograms", "kilo", "kilos")
	addUnits("mass", 0.001, "g", "gram", "grams")
	addUnits("mass", 0.000001, "mg", "milligram", "milligrams")
	addUnits("mass", 1000, "t", "tonne", "tonnes", "metric ton", "metric tons")
	addUnits("mass", 0.45359237, "lb", "lbs", "pound", "pounds")
	addUnits("mass", 0.028349523125, "oz", "ounce", "ounces")
	addUnits("mass", 6.35029318, "st", "stone", "stones")
	addUnits("volume", 1, "l", "liter", "liters", "litre", "litres")
	addUnits("volume", 0.001, "ml", "milliliter", "milliliters", "millilitre", "millilitres")
	addUnits("volume", 1000, "m3", "cubic meter", "cubic meters", "cubic metre", "cubic metres")
	addUnits("volume", 3.785411784, "gal", "gallon", "gallons", "us gallon", "us gallons")
	addUnits("volume", 4.54609, "imperial gallon", "imperial gallons", "uk gallon", "uk gallons")
	addUnits("volume", 0.946352946, "qt", "quart", "quarts")
	addUnits("volume", 0.473176473, "pt", "pint", "pints")
	addUnits("volume", 0.2365882365, "cup", "cups")
	addUnits("volume", 0.0295735295625, "fl oz", "fluid ounce", "fluid ounces")
	addUnits("volume", 0.01478676478125, "tbsp", "tablespoon", "tablespoons")
	addUnits("volume", 0.00492892159375, "tsp", "teaspoon", "teaspoons")
	addUnits("area", 1, "m2", "sqm", "square meter", "square meters", "square metre", "square metres")
	addUnits("area", 0.09290304, "ft2", "sqft", "sq ft", "square foot", "square feet")
	addUnits("area", 10000, "ha", "hectare", "hectares")
	addUnits("area", 4046.8564224, "acre", "acres")
	addUnits("area", 1e6, "km2", "square kilometer", "square kilometers", "square kilometre", "square kilometres")
	addUnits("area", 1000, "dunam", "dunams", "dönüm")
	addUnits("speed", 1, "m/s", "mps", "meters per second", "metres per second")
	addUnits("speed", 1/3.6, "km/h", "kph", "kmh", "kilometers per hour", "kilometres per hour")
	addUnits("speed", 0.44704, "mph", "miles per hour")
	addUnits("speed", 0.514444, "kn", "knot", "knots")
	addUnits("data", 1, "b", "byte", "bytes")
	addUnits("data", 1e3, "kb", "kilobyte", "kilobytes")
	addUnits("data", 1e6, "mb", "megabyte", "megabytes")
	addUnits("data", 1e9, "gb", "gigabyte", "gigabytes")
	addUnits("data", 1e12, "tb", "terabyte", "terabytes")
	addUnits("data", 1024, "kib", "kibibyte", "kibibytes")
	addUnits("data", 1<<20, "mib", "mebibyte", "mebibytes")
	addUnits("data", 1<<30, "gib", "gibibyte", "gibibytes")
	addUnits("data", 1<<40, "tib", "tebibyte", "tebibytes")
	addUnits("time", 1, "s", "sec", "second", "seconds")
	addUnits("time", 60, "min", "minute", "minutes")
	addUnits("time", 3600, "h", "hr", "hour", "hours")
	addUnits("time", 86400, "d", "day", "days")
	addUnits("time", 604800, "wk", "week", "weeks")
	addUnits("energy", 1, "j", "joule", "joules")
	addUnits("energy", 1000, "kj", "kilojoule", "kilojoules")
	addUnits("energy", 4184, "kcal", "calorie", "calories", "kilocalorie", "kilocalories")
	addUnits("energy", 3.6e6, "kwh", "kilowatt hour", "kilowatt hours")
}

var temperatureUnits = map[string]string{
	"c": "c", "°c": "c", "celsius": "c", "f": "f", "°f": "f", "fahrenheit": "f", "k": "k", "kelvin": "k",
}

func (e *AssistantToolExecutor) convertUnits(_ context.Context, input map[string]any) (map[string]any, error) {
	value, ok := inputNumber(input, "value", "amount", "quantity")
	if !ok {
		return nil, errors.New("give the value to convert")
	}
	from := strings.ToLower(inputText(input, "from", "from_unit", "unit"))
	to := strings.ToLower(inputText(input, "to", "to_unit"))
	if from == "" || to == "" {
		return nil, errors.New("give both units, e.g. {\"value\":5,\"from\":\"km\",\"to\":\"mi\"}")
	}
	if tf, ok := temperatureUnits[from]; ok {
		tt, ok := temperatureUnits[to]
		if !ok {
			return nil, fmt.Errorf("can't convert a temperature to %q", to)
		}
		result := fromKelvin(toKelvin(value, tf), tt)
		return map[string]any{"value": value, "from": from, "to": to, "result": roundTo(result, 4)}, nil
	}
	source, ok := units[from]
	if !ok {
		return nil, fmt.Errorf("unknown unit %q", from)
	}
	target, ok := units[to]
	if !ok {
		return nil, fmt.Errorf("unknown unit %q", to)
	}
	if source.kind != target.kind {
		return nil, fmt.Errorf("can't convert %s (%s) to %s (%s)", from, source.kind, to, target.kind)
	}
	result := value * source.factor / target.factor
	return map[string]any{"value": value, "from": from, "to": to, "result": roundSignificant(result, 10), "kind": source.kind}, nil
}

func toKelvin(value float64, unit string) float64 {
	switch unit {
	case "c":
		return value + 273.15
	case "f":
		return (value-32)*5/9 + 273.15
	default:
		return value
	}
}

func fromKelvin(value float64, unit string) float64 {
	switch unit {
	case "c":
		return value - 273.15
	case "f":
		return (value-273.15)*9/5 + 32
	default:
		return value
	}
}

// roundSignificant drops float noise (0.30000000000000004) while keeping
// tiny and huge results meaningful.
func roundSignificant(value float64, digits int) float64 {
	var parsed float64
	fmt.Sscanf(fmt.Sprintf("%.*g", digits, value), "%g", &parsed)
	return parsed
}

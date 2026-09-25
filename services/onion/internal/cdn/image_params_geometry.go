package cdn

import (
	"fmt"
	"math"
	"strconv"
	"strings"
)

func cropGeometryStep(raw string, width, height int) (string, int, int, error) {
	parts := splitParamList(raw)
	if len(parts) == 0 {
		return "", width, height, nil
	}

	var cropW, cropH int
	var err error
	if strings.Contains(parts[0], ":") {
		cropW, cropH, err = ratioDimensions(parts[0], width, height)
		if err != nil {
			return "", width, height, err
		}
	} else {
		cropW, cropH, err = pixelDimensions(parts[0], parts)
		if err != nil {
			return "", width, height, err
		}
	}
	if cropW == 0 && width > 0 {
		cropW = width
	}
	if cropH == 0 && height > 0 {
		cropH = height
	}
	if cropW <= 0 || cropH <= 0 {
		return "", width, height, fmt.Errorf("crop requires width and height")
	}

	mode := "fill"
	gravity := ""
	extras := []string{}
	for _, part := range parts[1:] {
		switch {
		case part == "smart":
			gravity = "auto"
		case part == "safe":
			// Our crop op clamps to the source bounds, so safe is the default.
		case strings.HasPrefix(part, "x"):
			extras = append(extras, "x_"+strings.TrimPrefix(part, "x"))
			mode = "crop"
		case strings.HasPrefix(part, "y"):
			extras = append(extras, "y_"+strings.TrimPrefix(part, "y"))
			mode = "crop"
		case strings.HasPrefix(part, "offset-x"), strings.HasPrefix(part, "offset-y"):
			return "", width, height, fmt.Errorf("crop offset positioning is not supported")
		}
	}
	values := []string{"w_" + strconv.Itoa(cropW), "h_" + strconv.Itoa(cropH), "c_" + mode}
	if gravity != "" {
		values = append(values, "g_"+gravity)
	}
	values = append(values, extras...)
	return strings.Join(values, ","), cropW, cropH, nil
}

func cropStep(raw string, width, height int, mode string) (string, error) {
	if strings.TrimSpace(raw) == "" {
		return "", nil
	}
	step, _, _, err := cropGeometryStep(raw, width, height)
	if err != nil {
		return "", err
	}
	if mode == "crop" {
		step = strings.Replace(step, "c_fill", "c_crop", 1)
	}
	return step, nil
}

func ratioDimensions(raw string, width, height int) (int, int, error) {
	parts := strings.SplitN(raw, ":", 2)
	if len(parts) != 2 {
		return 0, 0, fmt.Errorf("invalid crop ratio: %s", raw)
	}
	a, err := strconv.ParseFloat(parts[0], 64)
	if err != nil || a <= 0 {
		return 0, 0, fmt.Errorf("invalid crop ratio: %s", raw)
	}
	b, err := strconv.ParseFloat(parts[1], 64)
	if err != nil || b <= 0 {
		return 0, 0, fmt.Errorf("invalid crop ratio: %s", raw)
	}
	if width <= 0 && height <= 0 {
		return 0, 0, fmt.Errorf("ratio crop requires width or height")
	}
	if width > 0 && height <= 0 {
		height = int(math.Round(float64(width) * b / a))
	}
	if height > 0 && width <= 0 {
		width = int(math.Round(float64(height) * a / b))
	}
	return width, height, nil
}

func pixelDimensions(first string, parts []string) (int, int, error) {
	width, err := parseDimension(first)
	if err != nil {
		return 0, 0, err
	}
	height := 0
	if len(parts) > 1 {
		if n, parseErr := parseDimension(parts[1]); parseErr == nil {
			height = n
		}
	}
	return width, height, nil
}

func parseDimension(raw string) (int, error) {
	raw = strings.TrimSuffix(strings.TrimSpace(raw), "p")
	n, err := strconv.ParseFloat(raw, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid dimension: %s", raw)
	}
	if n < 0 {
		return 0, fmt.Errorf("dimension must not be negative")
	}
	return int(math.Round(n)), nil
}

// orientStep maps the request's orient/rotate value onto a chain step. Bare
// h/v agree with the hflip/vflip aliases: fx_flop is a left-to-right mirror
// and fx_flip a top-to-bottom one (media-engine apply.cpp).
func orientStep(raw string) string {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "", "auto":
		return ""
	case "90", "r90":
		return "a_90"
	case "180", "r180":
		return "a_180"
	case "270", "r270", "-90":
		return "a_270"
	case "v", "flip", "vflip":
		return "fx_flip"
	case "h", "flop", "hflip":
		return "fx_flop"
	default:
		return "a_" + sanitizeChainValue(raw)
	}
}

func canvasStep(raw, bg string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return ""
	}
	raw = strings.NewReplacer("x", ",", "X", ",").Replace(raw)
	parts := splitParamList(raw)
	if len(parts) < 2 {
		return ""
	}
	w, errW := parseDimension(parts[0])
	h, errH := parseDimension(parts[1])
	if errW != nil || errH != nil || w <= 0 || h <= 0 {
		return ""
	}
	step := fmt.Sprintf("w_%d,h_%d,c_pad", w, h)
	if bg != "" {
		step += ",b_" + bg
	}
	return step
}

func padStep(raw, bg string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" || isFalsey(raw) {
		return ""
	}
	n, err := parseDimension(splitParamList(raw)[0])
	if err != nil || n <= 0 || n > maxTransformBorder {
		return ""
	}
	if bg == "" {
		bg = "transparent"
	}
	return "bo_" + strconv.Itoa(n) + ":" + bg
}

func borderStep(raw, bg string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" || isFalsey(raw) {
		return ""
	}
	parts := splitParamList(raw)
	n, err := parseDimension(parts[0])
	if err != nil || n <= 0 || n > maxTransformBorder {
		return ""
	}
	color := bg
	if len(parts) > 1 {
		color = sanitizeChainValue(parts[1])
	}
	if color == "" {
		color = "black"
	}
	return "bo_" + strconv.Itoa(n) + ":" + color
}

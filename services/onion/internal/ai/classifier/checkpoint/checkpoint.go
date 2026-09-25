// Package checkpoint loads and verifies cua-s1 checkpoints: a safetensors
// file plus a JSON document with the model config and a SHA-256 signature.
// It ports the reference serialization helpers (trycua/cua, libs/cua-s1,
// MIT) without pickle support.
package checkpoint

import (
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

// Config mirrors the validated reference model fields; dropout is accepted
// and ignored at inference.
type Config struct {
	Encoder       string
	Width         int
	Rank          int
	ContextTokens int
	OptionTokens  int
	Layers        int
	Heads         int
}

// Tensor is one F32 safetensors entry with its decoded shape.
type Tensor struct {
	Shape []int
	Data  []float32
}

// File is a verified checkpoint with decoded weights and hosting metadata.
type File struct {
	Config       Config
	Tensors      map[string]Tensor
	Params       int
	WeightsBytes int64
}

var (
	// ErrSignatureMismatch reports a checkpoint whose state digest does not match.
	ErrSignatureMismatch = errors.New("checkpoint state signature mismatch")
	// ErrUnsupportedFormat reports a checkpoint format this package cannot load.
	ErrUnsupportedFormat = errors.New("unsupported checkpoint format")
)

const (
	checkpointFormat = "cua-s1"
	maxConfigBytes   = 1 << 20
	maxWeightsBytes  = 64 << 20
	maxHeaderBytes   = 8 << 20
)

// Load reads a checkpoint from a directory, .safetensors, or .json path,
// verifies its format markers and state signature, and returns the weights.
// Pickle-based checkpoints are rejected.
func Load(path string) (*File, error) {
	wPath, cPath, err := resolvePaths(path)
	if err != nil {
		return nil, err
	}
	raw, err := readFileLimited(cPath, maxConfigBytes)
	if err != nil {
		return nil, fmt.Errorf("s1: read config: %w", err)
	}
	docRaw, err := decodeJSON(raw)
	if err != nil {
		return nil, fmt.Errorf("s1: invalid checkpoint JSON: %w", err)
	}
	doc, ok := docRaw.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("s1: checkpoint JSON must contain an object")
	}
	cfg, rawCfg, sig, err := parseConfig(doc)
	if err != nil {
		return nil, err
	}
	tensors, dtypes, err := readSafetensors(wPath)
	if err != nil {
		return nil, err
	}
	if err := verifySignature(rawCfg, tensors, dtypes, sig); err != nil {
		return nil, err
	}
	info, err := os.Stat(wPath)
	if err != nil {
		return nil, fmt.Errorf("s1: stat weights: %w", err)
	}
	params := 0
	for _, t := range tensors {
		params += len(t.Data)
	}
	return &File{Config: cfg, Tensors: tensors, Params: params, WeightsBytes: info.Size()}, nil
}

func resolvePaths(p string) (weights, config string, err error) {
	switch ext := strings.ToLower(filepath.Ext(p)); ext {
	case ".pt", ".pth", ".bin", ".pkl", ".pickle":
		return "", "", fmt.Errorf("s1: legacy pickle checkpoints are not supported: %s", p)
	case ".safetensors":
		return p, strings.TrimSuffix(p, ext) + ".json", nil
	case ".json":
		return strings.TrimSuffix(p, ext) + ".safetensors", p, nil
	case "":
		fi, statErr := os.Stat(p)
		if statErr != nil {
			return "", "", fmt.Errorf("s1: checkpoint dir: %w", statErr)
		}
		if !fi.IsDir() {
			return "", "", fmt.Errorf("s1: not a directory: %s", p)
		}
		return filepath.Join(p, "model.safetensors"), filepath.Join(p, "config.json"), nil
	default:
		return "", "", fmt.Errorf("s1: checkpoint path must be a directory, .safetensors or .json: %s", p)
	}
}

func canonical(v any) (string, error) {
	switch t := v.(type) {
	case nil:
		return "null", nil
	case bool:
		if t {
			return "true", nil
		}
		return "false", nil
	case json.Number:
		s := t.String()
		if !strings.ContainsAny(s, ".eE") {
			return s, nil
		}
		f, err := parseJSONFloat(s)
		if err != nil {
			return "", err
		}
		return pyFloat(f), nil
	case string:
		return canonicalString(t), nil
	case []any:
		var sb strings.Builder
		sb.WriteByte('[')
		for i, e := range t {
			if i > 0 {
				sb.WriteByte(',')
			}
			s, err := canonical(e)
			if err != nil {
				return "", err
			}
			sb.WriteString(s)
		}
		sb.WriteByte(']')
		return sb.String(), nil
	case map[string]any:
		keys := make([]string, 0, len(t))
		for k := range t {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		var sb strings.Builder
		sb.WriteByte('{')
		for i, k := range keys {
			if i > 0 {
				sb.WriteByte(',')
			}
			sb.WriteString(canonicalString(k))
			sb.WriteByte(':')
			s, err := canonical(t[k])
			if err != nil {
				return "", err
			}
			sb.WriteString(s)
		}
		sb.WriteByte('}')
		return sb.String(), nil
	default:
		return "", fmt.Errorf("s1: unsupported checkpoint JSON value %T", v)
	}
}

func parseJSONFloat(s string) (float64, error) {
	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0, fmt.Errorf("s1: invalid number in checkpoint JSON: %s", s)
	}
	return f, nil
}

// pyFloat formats f like Python's json.dumps: shortest round-trip digits,
// fixed notation for -4 < decpt <= 16, scientific otherwise.
func pyFloat(f float64) string {
	if math.IsNaN(f) {
		return "NaN"
	}
	if math.IsInf(f, 1) {
		return "Infinity"
	}
	if math.IsInf(f, -1) {
		return "-Infinity"
	}
	sign := ""
	if math.Signbit(f) {
		sign = "-"
		f = -f
	}
	if f == 0 {
		return sign + "0.0"
	}
	raw := strconv.FormatFloat(f, 'e', -1, 64)
	mant, expStr, _ := strings.Cut(raw, "e")
	exp := 0
	fmt.Sscanf(expStr, "%d", &exp)
	digits := strings.Replace(mant, ".", "", 1)
	digits = strings.TrimRight(digits, "0")
	if digits == "" {
		digits = "0"
	}
	decpt := exp + 1
	if decpt > -4 && decpt <= 16 {
		var sb strings.Builder
		sb.WriteString(sign)
		switch {
		case decpt <= 0:
			sb.WriteString("0.")
			sb.WriteString(strings.Repeat("0", -decpt))
			sb.WriteString(digits)
		case decpt >= len(digits):
			sb.WriteString(digits)
			sb.WriteString(strings.Repeat("0", decpt-len(digits)))
			sb.WriteString(".0")
		default:
			sb.WriteString(digits[:decpt])
			sb.WriteByte('.')
			sb.WriteString(digits[decpt:])
		}
		return sb.String()
	}
	var sb strings.Builder
	sb.WriteString(sign)
	sb.WriteByte(digits[0])
	if len(digits) > 1 {
		sb.WriteByte('.')
		sb.WriteString(digits[1:])
	}
	sb.WriteByte('e')
	if exp < 0 {
		sb.WriteByte('-')
		exp = -exp
	} else {
		sb.WriteByte('+')
	}
	if exp < 10 {
		sb.WriteByte('0')
	}
	fmt.Fprintf(&sb, "%d", exp)
	return sb.String()
}

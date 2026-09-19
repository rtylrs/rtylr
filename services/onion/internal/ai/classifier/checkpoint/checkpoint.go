// Package checkpoint loads and verifies cua-s1 checkpoints: a safetensors
// file plus a JSON document with the model config and a SHA-256 signature.
// It ports the reference serialization helpers (trycua/cua, libs/cua-s1,
// MIT) without pickle support.
package checkpoint

import (
	"bufio"
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
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

func canonicalString(s string) string {
	var sb strings.Builder
	sb.WriteByte('"')
	for _, r := range s {
		switch r {
		case '"':
			sb.WriteString("\\\"")
		case '\\':
			sb.WriteString("\\\\")
		case '\n':
			sb.WriteString("\\n")
		case '\r':
			sb.WriteString("\\r")
		case '\t':
			sb.WriteString("\\t")
		case '\b':
			sb.WriteString("\\b")
		case '\f':
			sb.WriteString("\\f")
		default:
			switch {
			case r < 0x20:
				fmt.Fprintf(&sb, "\\u%04x", r)
			case r < 0x7f:
				sb.WriteRune(r)
			case r <= 0xffff:
				fmt.Fprintf(&sb, "\\u%04x", r)
			default:
				r -= 0x10000
				fmt.Fprintf(&sb, "\\u%04x\\u%04x", 0xd800+r>>10, 0xdc00+r&0x3ff)
			}
		}
	}
	sb.WriteByte('"')
	return sb.String()
}

// Kind classifies a checkpoint load failure for operator-facing output
// without exposing filesystem paths.
func Kind(err error) string {
	switch {
	case err == nil:
		return ""
	case errors.Is(err, ErrSignatureMismatch):
		return "signature_mismatch"
	case errors.Is(err, ErrUnsupportedFormat):
		return "unsupported_format"
	case errors.Is(err, os.ErrNotExist):
		return "not_found"
	default:
		return "load_error"
	}
}

func readFileLimited(path string, limit int64) ([]byte, error) {
	info, err := os.Stat(path)
	if err != nil {
		return nil, err
	}
	if info.Size() > limit {
		return nil, fmt.Errorf("file exceeds %d bytes: %s", limit, path)
	}
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	raw, err := io.ReadAll(bufio.NewReader(io.LimitReader(f, limit+1)))
	if err != nil {
		return nil, err
	}
	if int64(len(raw)) > limit {
		return nil, fmt.Errorf("file exceeds %d bytes: %s", limit, path)
	}
	return raw, nil
}

func strictAtoi(s string) (int, error) {
	if strings.HasPrefix(s, "+") || strings.HasPrefix(s, "-") {
		return 0, fmt.Errorf("signed integer %q", s)
	}
	return strconv.Atoi(s)
}

func decodeJSON(data []byte) (any, error) {
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.UseNumber()
	var v any
	if err := dec.Decode(&v); err != nil {
		return nil, err
	}
	return v, nil
}

func positiveInt(m map[string]any, key string) (int, error) {
	n, ok := m[key].(json.Number)
	if !ok {
		return 0, fmt.Errorf("s1: config field %q must be an integer", key)
	}
	i, err := strictAtoi(n.String())
	if err != nil || i <= 0 {
		return 0, fmt.Errorf("s1: config field %q must be a positive integer", key)
	}
	return i, nil
}

func parseConfig(doc map[string]any) (Config, map[string]any, string, error) {
	var cfg Config
	if doc["format"] != checkpointFormat {
		return cfg, nil, "", fmt.Errorf("s1: %w %v", ErrUnsupportedFormat, doc["format"])
	}
	if doc["format_version"] != json.Number("1") {
		return cfg, nil, "", fmt.Errorf("s1: %w version %v", ErrUnsupportedFormat, doc["format_version"])
	}
	rawCfg, ok := doc["config"].(map[string]any)
	if !ok {
		return cfg, nil, "", fmt.Errorf("s1: checkpoint JSON field 'config' must be an object")
	}
	sig, ok := doc["state_signature"].(string)
	if !ok || len(sig) != 64 {
		return cfg, nil, "", fmt.Errorf("s1: checkpoint JSON field 'state_signature' must be a SHA-256 digest")
	}
	enc, _ := rawCfg["encoder"].(string)
	if enc != "tinyx" {
		return cfg, nil, "", fmt.Errorf("s1: %w encoder %q (want tinyx)", ErrUnsupportedFormat, enc)
	}
	cfg.Encoder = enc
	var err error
	if cfg.Width, err = positiveInt(rawCfg, "width"); err != nil {
		return cfg, nil, "", err
	}
	if cfg.Rank, err = positiveInt(rawCfg, "rank"); err != nil {
		return cfg, nil, "", err
	}
	if cfg.ContextTokens, err = positiveInt(rawCfg, "context_tokens"); err != nil {
		return cfg, nil, "", err
	}
	if cfg.OptionTokens, err = positiveInt(rawCfg, "option_tokens"); err != nil {
		return cfg, nil, "", err
	}
	if cfg.Layers, err = positiveInt(rawCfg, "layers"); err != nil {
		return cfg, nil, "", err
	}
	if cfg.Heads, err = positiveInt(rawCfg, "heads"); err != nil {
		return cfg, nil, "", err
	}
	if cfg.Width%cfg.Heads != 0 {
		return cfg, nil, "", fmt.Errorf("s1: model width must be divisible by attention heads")
	}
	return cfg, rawCfg, sig, nil
}

func readSafetensors(path string) (map[string]Tensor, map[string]string, error) {
	raw, err := readFileLimited(path, maxWeightsBytes)
	if err != nil {
		return nil, nil, fmt.Errorf("s1: read weights: %w", err)
	}
	if len(raw) < 8 {
		return nil, nil, fmt.Errorf("s1: truncated safetensors header: %s", path)
	}
	headerLen := binary.LittleEndian.Uint64(raw[:8])
	if headerLen > maxHeaderBytes {
		return nil, nil, fmt.Errorf("s1: safetensors header exceeds %d bytes: %s", maxHeaderBytes, path)
	}
	if uint64(len(raw)-8) < headerLen {
		return nil, nil, fmt.Errorf("s1: truncated safetensors header: %s", path)
	}
	headerRaw := raw[8 : 8+headerLen]
	data := raw[8+headerLen:]
	hv, err := decodeJSON(headerRaw)
	if err != nil {
		return nil, nil, fmt.Errorf("s1: invalid safetensors header: %w", err)
	}
	header, ok := hv.(map[string]any)
	if !ok {
		return nil, nil, fmt.Errorf("s1: invalid safetensors header in %s", path)
	}
	out := map[string]Tensor{}
	dtypes := map[string]string{}
	names := make([]string, 0, len(header))
	for name := range header {
		if name == "__metadata__" {
			continue
		}
		names = append(names, name)
	}
	for _, name := range names {
		entry, ok := header[name].(map[string]any)
		if !ok {
			return nil, nil, fmt.Errorf("s1: invalid header entry %q", name)
		}
		dtype, _ := entry["dtype"].(string)
		if dtype != "F32" {
			return nil, nil, fmt.Errorf("s1: unsupported dtype %q for %q (reads F32 only)", dtype, name)
		}
		shapeRaw, ok := entry["shape"].([]any)
		if !ok {
			return nil, nil, fmt.Errorf("s1: invalid shape for %q", name)
		}
		shape := make([]int, len(shapeRaw))
		n := 1
		for i, dim := range shapeRaw {
			d, ok := dim.(json.Number)
			if !ok {
				return nil, nil, fmt.Errorf("s1: invalid shape for %q", name)
			}
			di, err := strictAtoi(d.String())
			if err != nil || di < 0 {
				return nil, nil, fmt.Errorf("s1: invalid shape for %q", name)
			}
			shape[i] = di
			n *= di
		}
		offsRaw, ok := entry["data_offsets"].([]any)
		if !ok || len(offsRaw) != 2 {
			return nil, nil, fmt.Errorf("s1: invalid data offsets for %q", name)
		}
		start, ok1 := offsRaw[0].(json.Number)
		end, ok2 := offsRaw[1].(json.Number)
		if !ok1 || !ok2 {
			return nil, nil, fmt.Errorf("s1: invalid data offsets for %q", name)
		}
		si, err := strictAtoi(start.String())
		if err != nil {
			return nil, nil, fmt.Errorf("s1: invalid data offsets for %q", name)
		}
		ei, err := strictAtoi(end.String())
		if err != nil || si < 0 || ei < si || ei > len(data) {
			return nil, nil, fmt.Errorf("s1: data offsets out of range for %q", name)
		}
		if ei-si != n*4 {
			return nil, nil, fmt.Errorf("s1: size mismatch for %q", name)
		}
		vals := make([]float32, n)
		for i := range vals {
			vals[i] = math.Float32frombits(binary.LittleEndian.Uint32(data[si+4*i:]))
		}
		out[name] = Tensor{Shape: shape, Data: vals}
		dtypes[name] = "torch.float32"
	}
	return out, dtypes, nil
}

func verifySignature(rawCfg map[string]any, tensors map[string]Tensor, dtypes map[string]string, want string) error {
	canonCfg, err := canonical(rawCfg)
	if err != nil {
		return err
	}
	digest := sha256.New()
	digest.Write([]byte(canonCfg))
	names := make([]string, 0, len(tensors))
	for name := range tensors {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		t := tensors[name]
		shape := make([]any, len(t.Shape))
		for i, d := range t.Shape {
			shape[i] = json.Number(fmt.Sprint(d))
		}
		row, err := canonical([]any{name, dtypes[name], shape})
		if err != nil {
			return err
		}
		digest.Write([]byte(row))
		buf := make([]byte, 4)
		for _, v := range t.Data {
			binary.LittleEndian.PutUint32(buf, math.Float32bits(v))
			digest.Write(buf)
		}
	}
	if got := hex.EncodeToString(digest.Sum(nil)); got != want {
		return fmt.Errorf("s1: %w", ErrSignatureMismatch)
	}
	return nil
}

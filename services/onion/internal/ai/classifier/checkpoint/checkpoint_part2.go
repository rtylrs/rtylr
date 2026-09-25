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
	"sort"
	"strconv"
	"strings"
)

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
			if di != 0 && n > math.MaxInt/di {
				return nil, nil, fmt.Errorf("s1: invalid shape for %q", name)
			}
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
		if n > len(data)/4 {
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

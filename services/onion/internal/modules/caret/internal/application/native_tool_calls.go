package application

import (
	"encoding/json"
	"regexp"
	"strings"
)

// Some models fall back to their own function-call markup instead of the
// <tool> marker the prompt asks for, most often deep into a long chat.
// DeepSeek's looks like this (the bars are U+FF5C):
//
//	<｜DSML｜function_calls>
//	<｜DSML｜invoke name="portal.customize">
//	<｜DSML｜parameter name="area" string="true">rail</｜DSML｜parameter>
//	<｜DSML｜parameter name="moves" string="false">[{"id":"profile"}]</｜DSML｜parameter>
//	</｜DSML｜invoke>
//	</｜DSML｜function_calls>
//
// It is read as the same tool calls, so the call still runs and the markup
// never reaches the person.

var (
	nativeCallStartRe = regexp.MustCompile(`<\s*/?\s*[｜|]\s*DSML\s*[｜|]`)
	nativeCallTagRe   = regexp.MustCompile(`<\s*(/?)\s*[｜|]\s*DSML\s*[｜|]\s*([^\s>/]+)([^<>]*)>`)
	nativeCallAttrRe  = regexp.MustCompile(`([A-Za-z_]+)\s*=\s*"([^"]*)"`)
)

// NativeToolCall is one call read from native markup.
type NativeToolCall struct {
	Name  string         `json:"name"`
	Input map[string]any `json:"input"`
}

// NativeToolCallStart returns where native call markup starts in text, or -1.
func NativeToolCallStart(text string) int {
	if loc := nativeCallStartRe.FindStringIndex(text); loc != nil {
		return loc[0]
	}
	return -1
}

// ParseNativeToolCalls reads the native call block text starts with. It
// returns the calls and the block's length, or ok false while the block is
// still arriving. At the end of the reply (final) an unclosed block runs to
// the end of text, since the model may stop before closing it.
func ParseNativeToolCalls(text string, final bool) (calls []NativeToolCall, end int, ok bool) {
	open := nativeCallTagRe.FindStringSubmatchIndex(text)
	if open == nil || open[0] != 0 {
		if final {
			return nil, len(text), true
		}
		return nil, 0, false
	}
	if text[open[2]:open[3]] == "/" {
		// A stray closing tag: drop it.
		return nil, open[1], true
	}
	name := text[open[4]:open[5]]
	end = -1
	for _, tag := range nativeCallTagRe.FindAllStringSubmatchIndex(text[open[1]:], -1) {
		if text[open[1]+tag[2]:open[1]+tag[3]] == "/" && text[open[1]+tag[4]:open[1]+tag[5]] == name {
			end = open[1] + tag[1]
			break
		}
	}
	if end < 0 {
		if !final {
			return nil, 0, false
		}
		end = len(text)
	}
	return readNativeCalls(text[:end]), end, true
}

// StripNativeToolCalls removes native call markup from text, keeping what
// is around it.
func StripNativeToolCalls(text string) string {
	var out strings.Builder
	for {
		start := NativeToolCallStart(text)
		if start < 0 {
			out.WriteString(text)
			return out.String()
		}
		out.WriteString(text[:start])
		_, end, _ := ParseNativeToolCalls(text[start:], true)
		text = text[start+end:]
	}
}

// readNativeCalls walks the tags of one block: each invoke is a call and
// each parameter inside it an input.
func readNativeCalls(block string) []NativeToolCall {
	var calls []NativeToolCall
	var current *NativeToolCall
	tags := nativeCallTagRe.FindAllStringSubmatchIndex(block, -1)
	for i, tag := range tags {
		closing := block[tag[2]:tag[3]] == "/"
		kind := strings.ToLower(block[tag[4]:tag[5]])
		attrs := nativeCallAttrs(block[tag[6]:tag[7]])
		switch {
		case kind == "invoke" && !closing:
			if name := strings.TrimSpace(attrs["name"]); name != "" {
				calls = append(calls, NativeToolCall{Name: name, Input: map[string]any{}})
				current = &calls[len(calls)-1]
			} else {
				current = nil
			}
		case kind == "invoke":
			current = nil
		case kind == "parameter" && !closing && current != nil && attrs["name"] != "":
			valueEnd := len(block)
			if i+1 < len(tags) {
				valueEnd = tags[i+1][0]
			}
			current.Input[attrs["name"]] = nativeParameterValue(block[tag[1]:valueEnd], attrs["string"])
		}
	}
	return calls
}

func nativeCallAttrs(raw string) map[string]string {
	attrs := map[string]string{}
	for _, match := range nativeCallAttrRe.FindAllStringSubmatch(raw, -1) {
		attrs[strings.ToLower(match[1])] = match[2]
	}
	return attrs
}

// nativeParameterValue reads a parameter: string="true" is text as written,
// anything else is JSON when it parses.
func nativeParameterValue(raw, isString string) any {
	if strings.EqualFold(isString, "true") {
		return raw
	}
	var value any
	if json.Unmarshal([]byte(strings.TrimSpace(raw)), &value) == nil {
		return value
	}
	return strings.TrimSpace(raw)
}

package handlers

import "regexp"

// ThinkingBlock represents a model "scratch pad" reasoning chunk. The
// providers we use either expose this natively or we coerce it from
// fenced <thinking> sections in the streamed text.
type ThinkingBlock struct {
	Content string
}

var thinkingRe = regexp.MustCompile(`(?s)<thinking>(.*?)</thinking>`)

// ExtractThinkingBlocks returns parsed blocks and the remainder text with
// the fenced sections stripped — so the client only sees the public text
// in the message body.
func ExtractThinkingBlocks(s string) (blocks []ThinkingBlock, remainder string) {
	matches := thinkingRe.FindAllStringSubmatchIndex(s, -1)
	if len(matches) == 0 {
		return nil, s
	}
	blocks = make([]ThinkingBlock, 0, len(matches))
	out := make([]byte, 0, len(s))
	last := 0
	for _, m := range matches {
		out = append(out, s[last:m[0]]...)
		last = m[1]
		blocks = append(blocks, ThinkingBlock{Content: s[m[2]:m[3]]})
	}
	out = append(out, s[last:]...)
	remainder = string(out)
	return blocks, remainder
}

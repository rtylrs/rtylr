package application

import (
	"encoding/base64"
	"errors"
	"fmt"
	"strings"
	"unicode/utf8"
)

// AttachedFile is a text file the person attached to a message: a CSV
// export, notes, a JSON payload, code. The client sends its text.
type AttachedFile struct {
	Name string `json:"name"`
	Text string `json:"text"`
	// DataURL carries a binary document (a PDF) whose text the server
	// extracts; Text is empty then.
	DataURL string `json:"data_url,omitempty"`
}

// maxAttachedDocumentBytes bounds a binary attachment before decoding.
const maxAttachedDocumentBytes = 3 << 20

// ExtractAttachedFiles turns binary attachments into text, so the rest of
// the turn only deals with text files.
func ExtractAttachedFiles(files []AttachedFile) ([]AttachedFile, error) {
	out := make([]AttachedFile, len(files))
	for i, file := range files {
		out[i] = AttachedFile{Name: file.Name, Text: file.Text}
		if file.DataURL == "" {
			continue
		}
		mediaType, encoded, ok := strings.Cut(strings.TrimPrefix(file.DataURL, "data:"), ";base64,")
		if !ok || mediaType != "application/pdf" {
			return nil, fmt.Errorf("%s: only PDF documents can be sent as data", file.Name)
		}
		raw, err := base64.StdEncoding.DecodeString(encoded)
		if err != nil {
			return nil, fmt.Errorf("%s could not be decoded", file.Name)
		}
		text, err := extractPDFText(raw)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", file.Name, err)
		}
		if len(text) > maxAttachedFileBytes {
			text = TruncateUTF8ToBytes(text, maxAttachedFileBytes)
		}
		out[i].Text = text
	}
	return out, nil
}

const (
	MaxAttachedFiles     = 3
	maxAttachedFileBytes = 400 << 10
	// attachedFilesPromptBytes is how much file text one prompt carries;
	// files share it evenly.
	attachedFilesPromptBytes = 90000
	// attachedFileHistoryBytes is how much of each file the chat history
	// keeps for follow-up questions.
	attachedFileHistoryBytes = 5000
)

// ValidateAttachedFiles bounds the attachments of one message.
func ValidateAttachedFiles(files []AttachedFile) error {
	if len(files) > MaxAttachedFiles {
		return fmt.Errorf("attach up to %d files per message", MaxAttachedFiles)
	}
	for _, file := range files {
		name := strings.TrimSpace(file.Name)
		if name == "" || utf8.RuneCountInString(name) > 200 {
			return errors.New("each attached file needs a name under 200 characters")
		}
		if len(file.DataURL) > maxAttachedDocumentBytes {
			return fmt.Errorf("%s is larger than 2 MB", name)
		}
		if len(file.Text) > maxAttachedFileBytes {
			return fmt.Errorf("%s is larger than 400 KB; attach a smaller extract", name)
		}
		if !utf8.ValidString(file.Text) {
			return fmt.Errorf("%s is not a text file", name)
		}
	}
	return nil
}

// RenderAttachedFiles is the prompt region for a message's files. Long files
// keep their start and end, where headers and totals usually are.
func RenderAttachedFiles(files []AttachedFile) string {
	if len(files) == 0 {
		return ""
	}
	budget := attachedFilesPromptBytes / len(files)
	var b strings.Builder
	for _, file := range files {
		lines := strings.Count(file.Text, "\n") + 1
		text := file.Text
		clipped := ""
		if len(text) > budget {
			head := TruncateUTF8ToBytes(text, budget*3/4)
			tail := text[len(text)-budget/4:]
			for len(tail) > 0 && !utf8.RuneStart(tail[0]) {
				tail = tail[1:]
			}
			text = head + "\n[... middle of the file omitted ...]\n" + tail
			clipped = ` clipped="true"`
		}
		fmt.Fprintf(&b, "<file name=\"%s\" lines=\"%d\"%s>\n%s\n</file>\n",
			escapeCaretPromptData(strings.TrimSpace(file.Name)), lines, clipped, escapeCaretPromptData(text))
	}
	return strings.TrimSpace(b.String())
}

// AttachedFilesHistoryNote is appended to the stored message so later turns
// know which files were shared and what they began with.
func AttachedFilesHistoryNote(files []AttachedFile) string {
	var b strings.Builder
	for _, file := range files {
		lines := strings.Count(file.Text, "\n") + 1
		fmt.Fprintf(&b, "\n\n[Attached file %s, %d lines; start of the file:]\n%s", strings.TrimSpace(file.Name), lines, TruncateUTF8ToBytes(file.Text, attachedFileHistoryBytes))
		if len(file.Text) > attachedFileHistoryBytes {
			b.WriteString("\n[...]")
		}
	}
	return b.String()
}

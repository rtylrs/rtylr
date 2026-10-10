package application

import (
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

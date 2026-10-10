package application

import (
	"strings"
	"unicode/utf8"
)

// correctionOpenings start a message that pushes back on Caret's last reply.
var correctionOpenings = []string{
	"no,", "no.", "no ", "nope", "wrong", "that's wrong", "thats wrong", "that is wrong", "that's not", "thats not",
	"that is not", "not what i", "i said", "i told you", "i asked", "you didn't", "you didnt", "you did not",
	"you forgot", "you missed", "you ignored", "why did you", "why are you", "stop ", "please stop", "don't ", "dont ",
	"do not ", "please don't", "please dont", "that's incorrect", "incorrect", "not quite", "still wrong",
	"لا،", "لا ", "غلط", "خطأ", "مش هيك", "مو هيك", "non,", "non ", "faux", "c'est faux",
}

// correctionPhrases anywhere in a short message are feedback on style.
var correctionPhrases = []string{
	"too long", "too short", "too wordy", "too verbose", "shorter please", "be shorter", "be concise", "more concise",
	"less verbose", "stop using", "stop adding", "no need to", "i didn't ask", "i didnt ask", "didn't ask for",
	"i don't want", "i dont want", "not helpful", "you keep", "every time you", "again you", "in english please",
	"in arabic please", "answer in", "reply in", "طويل", "باختصار",
}

// friendlyOpenings start with a correction word but aren't pushback.
var friendlyOpenings = []string{
	"no worries", "no problem", "no thanks", "no thank you", "no rush", "no, thanks", "no, that's all", "no that's all",
	"no, that's it", "no that's it", "no, all good", "no all good", "nope, all good", "nope that's all", "no, perfect",
	"don't worry", "dont worry", "لا شكرا", "لا مشكلة",
}

// LooksLikeCorrection reports whether a message is the person pushing back on
// Caret's previous reply, so Caret can learn from it. It is deliberately
// narrow; the learning pass also discards feedback without a clear lesson.
func LooksLikeCorrection(message string) bool {
	text := strings.ToLower(strings.TrimSpace(message))
	if text == "" || utf8.RuneCountInString(text) > 600 {
		return false
	}
	text = strings.NewReplacer("’", "'", "‘", "'").Replace(text)
	for _, friendly := range friendlyOpenings {
		if strings.HasPrefix(text, friendly) {
			return false
		}
	}
	for _, opening := range correctionOpenings {
		if strings.HasPrefix(text, opening) {
			return true
		}
	}
	for _, phrase := range correctionPhrases {
		if strings.Contains(text, phrase) {
			return true
		}
	}
	return false
}

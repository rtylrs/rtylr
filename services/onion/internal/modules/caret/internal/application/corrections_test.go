package application

import "testing"

func TestLooksLikeCorrection(t *testing.T) {
	for _, message := range []string{"No, I meant last month", "That’s wrong, it's 12%", "too long, give me the number", "I told you to use USD", "stop adding emojis", "غلط، المبلغ أكبر", "why did you search the web?"} {
		if !LooksLikeCorrection(message) {
			t.Errorf("missed correction %q", message)
		}
	}
	for _, message := range []string{"What were sales yesterday?", "Notes for the meeting", "Nobody showed up today", "Thanks!", "Do you know a good supplier?", "No worries, thanks!", "no, that's all for now", "Don't worry about it"} {
		if LooksLikeCorrection(message) {
			t.Errorf("false correction %q", message)
		}
	}
}

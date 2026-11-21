package domain

import (
	"slices"
	"strconv"
	"strings"
)

// MatchWords are the keys of the words a scanned phone can show, such as
// "panda". The portal holds the same keys with their emoji and their English,
// French and Arabic words (clients/app/src/modules/pager/shared/lib/matchWords.ts);
// a test there fails when the two lists differ. A key's position picks its
// colour.
var MatchWords = []string{
	"aura", "slay", "yeet", "bestie", "vibe", "rizz", "delulu", "goat",
	"sidequest", "drip", "bussin", "sus", "glowup", "chefskiss", "nocap", "bet",
	"iconic", "lowkey", "mood", "hype", "legend", "chill", "tea", "cringe",
	"yap", "era", "lockedin", "npc",
	"panda", "lion", "tiger", "koala", "rabbit", "fox", "penguin", "frog",
	"turtle", "octopus", "whale", "dolphin", "butterfly", "horse", "giraffe", "elephant",
	"camel", "hedgehog", "crab", "shark", "kangaroo", "wolf", "cat", "crocodile",
	"hamster", "dinosaur",
	"banana", "grapes", "strawberry", "cherry", "lemon", "pineapple", "mango", "coconut",
	"avocado", "carrot", "broccoli", "donut", "pizza", "burger", "popcorn", "cookie",
	"cake", "croissant", "icecream", "chocolate",
	"balloon", "gift", "guitar", "umbrella", "trophy", "bicycle",
}

// MatchWordCode is how a word is stored: the key alone, or the key followed by
// its number once every key is taken ("panda2").
func MatchWordCode(key string, number int) string {
	if number < 2 {
		return key
	}
	return key + strconv.Itoa(number)
}

// ParseMatchWord splits a stored word into its key and number, which is 0 for
// the key alone. A code from before the words, such as "A7", is not a word.
func ParseMatchWord(code string) (key string, number int, ok bool) {
	key = strings.TrimRightFunc(code, func(r rune) bool { return r >= '0' && r <= '9' })
	if !slices.Contains(MatchWords, key) {
		return "", 0, false
	}
	if digits := code[len(key):]; digits != "" {
		n, err := strconv.Atoi(digits)
		if err != nil || n < 2 || digits[0] == '0' {
			return "", 0, false
		}
		number = n
	}
	return key, number, true
}

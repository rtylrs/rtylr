package application

import (
	"context"
	"crypto/rand"
	"errors"
	"math/big"
)

const maxRandomItems = 5000

// randomPick draws fairly with crypto randomness: giveaway winners, a
// shuffled order, or a number in a range. Models can't draw at random.
func (e *AssistantToolExecutor) randomPick(_ context.Context, input map[string]any) (map[string]any, error) {
	items := []string{}
	if list, ok := input["from"].([]any); ok {
		for _, item := range list {
			if text, ok := item.(string); ok && text != "" {
				items = append(items, text)
			}
		}
	}
	if len(items) > maxRandomItems {
		return nil, errors.New("give at most 5,000 items")
	}
	if len(items) == 0 {
		low, hasLow := inputNumber(input, "min")
		high, hasHigh := inputNumber(input, "max")
		if !hasLow || !hasHigh || high < low || high-low > 1e12 {
			return nil, errors.New(`give a list to pick from, e.g. {"from":["Ali","Sara"],"count":1}, or a range, e.g. {"min":1,"max":100}`)
		}
		n, err := cryptoInt(int64(high-low) + 1)
		if err != nil {
			return nil, err
		}
		return map[string]any{"number": int64(low) + n, "min": low, "max": high}, nil
	}
	count := 1
	if n, ok := inputNumber(input, "count", "pick", "winners"); ok {
		count = int(n)
	}
	if inputBool(input, "shuffle") {
		count = len(items)
	}
	if count < 1 || count > len(items) {
		return nil, errors.New("count must be between 1 and the number of items")
	}
	pool := append([]string(nil), items...)
	// Fisher-Yates over the first count positions.
	for i := 0; i < count; i++ {
		j, err := cryptoInt(int64(len(pool) - i))
		if err != nil {
			return nil, err
		}
		pool[i], pool[i+int(j)] = pool[i+int(j)], pool[i]
	}
	return map[string]any{"picked": pool[:count], "from_count": len(items), "method": "uniform draw with a cryptographic random source"}, nil
}

func cryptoInt(n int64) (int64, error) {
	value, err := rand.Int(rand.Reader, big.NewInt(n))
	if err != nil {
		return 0, errors.New("could not draw a random number")
	}
	return value.Int64(), nil
}

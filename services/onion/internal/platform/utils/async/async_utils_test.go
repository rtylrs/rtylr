package async

import (
	"testing"

	"github.com/google/uuid"
)

func getSomeMap() map[string]string {
	bla := 2
	results := make(map[string]string, bla)
	for i := 0; i < bla; i++ {
		u, err := uuid.NewUUID()
		if err != nil {
			panic("Failed to generate UUID: " + err.Error())
		}
		theStr := u.String()
		results[theStr] = theStr
	}
	return results
}

func TestFutureResults(t *testing.T) {
	megaFuture := NewFutureResult(CombineMaps[string, string])
	megaFuture.AddTasks(
		getSomeMap,
		getSomeMap,
		getSomeMap,
		getSomeMap,
	)

	allResults := megaFuture.Execute()

	if len(allResults) != 8 {
		t.Errorf("Expected 8 elements in the map got %d instead", len(allResults))
	}
}

func TestFuture(t *testing.T) {
	var firstResult int
	expectedFirstResult := 5

	var secondResult int
	expectedSecondResult := 10

	var thirdResult int
	expectedThirdResult := 15

	future := NewFuture()

	future.AddTasks(
		func() {
			firstResult = 5
		},
		func() {
			secondResult = 10
		},
		func() {
			thirdResult = 15
		},
	)

	future.Execute()

	if firstResult != expectedFirstResult || secondResult != expectedSecondResult || thirdResult != expectedThirdResult {
		t.Errorf("Results did not match expected values")
	}
}

func TestFutureWithPanic(t *testing.T) {
	var firstResult int
	expectedFirstResult := 5

	future := NewFuture()

	future.AddTasks(
		func() {
			firstResult = 5
		},
		func() {
			panic("I'm dead")
		},
	)

	future.Execute()

	if firstResult != expectedFirstResult {
		t.Errorf("Results did not match expected values")
	}
}

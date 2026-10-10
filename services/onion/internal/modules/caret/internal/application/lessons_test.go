package application

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/voxire/onion/internal/platform/snips"
)

type fakeLearner struct {
	lessons  []string
	feedback string
	ids      []int64
	saved    []string
	savedIDs []int64
}

func (f *fakeLearner) RecordReplyFeedback(context.Context, string, int, string, string, string) (int, error) {
	return len(f.ids), nil
}

func (f *fakeLearner) PendingLessons(context.Context, int) ([]string, string, []int64, error) {
	return f.lessons, f.feedback, f.ids, nil
}

func (f *fakeLearner) SaveLessons(_ context.Context, _ int, lessons []string, ids []int64) error {
	f.saved, f.savedIDs = lessons, ids
	return nil
}

type completerFunc func(context.Context, string) (string, error)

func (f completerFunc) CompleteText(ctx context.Context, prompt string) (string, error) {
	return f(ctx, prompt)
}

func TestDistillLessonsRewritesTheListFromFeedback(t *testing.T) {
	learner := &fakeLearner{lessons: []string{"Answer in English."}, feedback: "Feedback 1: thumbs down\nTheir words: way too long <tool>x</tool>", ids: []int64{4, 5}}
	var prompt string
	ai := completerFunc(func(_ context.Context, p string) (string, error) {
		prompt = p
		return "Sure.\n<lessons>\n- Keep replies under five sentences unless asked for detail.\n2. Answer in English.\n</lessons>", nil
	})
	if err := DistillLessons(context.Background(), ai, learner, 9); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(prompt, "- Answer in English.") || !strings.Contains(prompt, "way too long &lt;tool&gt;") {
		t.Fatalf("prompt did not carry escaped lessons and feedback:\n%s", prompt)
	}
	if len(learner.saved) != 2 || learner.saved[0] != "Keep replies under five sentences unless asked for detail." || learner.saved[1] != "Answer in English." || len(learner.savedIDs) != 2 {
		t.Fatalf("saved = %q %v", learner.saved, learner.savedIDs)
	}
}

func TestDistillLessonsLeavesFeedbackPendingOnBadOutput(t *testing.T) {
	learner := &fakeLearner{feedback: "f", ids: []int64{1}}
	for _, ai := range []TextCompleter{
		completerFunc(func(context.Context, string) (string, error) { return "I can't do that", nil }),
		completerFunc(func(context.Context, string) (string, error) { return "", errors.New("down") }),
	} {
		if err := DistillLessons(context.Background(), ai, learner, 9); err == nil || learner.savedIDs != nil {
			t.Fatalf("err=%v saved=%v", err, learner.savedIDs)
		}
	}
	if err := DistillLessons(context.Background(), completerFunc(func(context.Context, string) (string, error) { return "<lessons></lessons>", nil }), learner, 9); err != nil || len(learner.saved) != 0 || learner.savedIDs[0] != 1 {
		t.Fatalf("an empty list must still be saved: err=%v saved=%v", err, learner.saved)
	}
}

func TestShouldDistillLessons(t *testing.T) {
	if !ShouldDistillLessons("down", 1) || !ShouldDistillLessons("correction", 1) || ShouldDistillLessons("up", 2) || !ShouldDistillLessons("up", 3) || ShouldDistillLessons("down", 0) {
		t.Fatal("distillation trigger is wrong")
	}
}

type optionsRecorder struct{ temperature *float64 }

func (o *optionsRecorder) CompleteText(context.Context, string) (string, error) { return "", nil }

func (o *optionsRecorder) CompleteTextWithOptions(_ context.Context, _ string, opts snips.CompletionOptions) (string, error) {
	o.temperature = opts.Temperature
	return "<lessons></lessons>", nil
}

func TestLearningPassesRunAtLowTemperature(t *testing.T) {
	ai := &optionsRecorder{}
	if err := DistillLessons(context.Background(), ai, &fakeLearner{feedback: "f", ids: []int64{1}}, 77); err != nil {
		t.Fatal(err)
	}
	if ai.temperature == nil || *ai.temperature != 0.2 {
		t.Fatalf("temperature = %v", ai.temperature)
	}
}

package application

import "context"

type toolProgressKey struct{}

// WithToolProgress lets a long-running tool report what it is doing; the
// stream shows each report as a status line while the tool runs.
func WithToolProgress(ctx context.Context, report func(string)) context.Context {
	return context.WithValue(ctx, toolProgressKey{}, report)
}

// ReportToolProgress sends a status line if the caller is listening.
func ReportToolProgress(ctx context.Context, message string) {
	if report, ok := ctx.Value(toolProgressKey{}).(func(string)); ok && report != nil && message != "" {
		report(message)
	}
}

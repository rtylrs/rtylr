package application

import (
	"context"
	"strings"
)

// LearnToolRoute remembers that a request with these words was served by a
// tool Caret had to search for, so the next similar request offers it
// straight away.
func (s *Service) LearnToolRoute(ctx context.Context, userID int, tool string, words []string) error {
	tool = strings.TrimSpace(tool)
	if userID <= 0 || tool == "" || len(tool) > 80 || len(words) == 0 {
		return nil
	}
	if len(words) > 12 {
		words = words[:12]
	}
	return s.repository.SaveToolRoute(ctx, userID, tool, strings.Join(words, " "))
}

// ToolRoutes returns learned routes as tool name and request words.
func (s *Service) ToolRoutes(ctx context.Context, userID int) (map[string][][]string, error) {
	rows, err := s.repository.ToolRoutes(ctx, userID)
	if err != nil {
		return nil, err
	}
	routes := map[string][][]string{}
	for _, row := range rows {
		tool, words, ok := strings.Cut(row, "\t")
		if ok && tool != "" {
			routes[tool] = append(routes[tool], strings.Fields(words))
		}
	}
	return routes, nil
}

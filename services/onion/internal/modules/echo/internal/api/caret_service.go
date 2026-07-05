package api

import (
	"context"

	echohandlers "github.com/voxire/onion/internal/modules/echo/internal/handlers"
	hubcaretv1 "github.com/voxire/voxie/gen/go/hub/caret/v1"
)

type CaretAPI struct{}

func NewCaretAPI() *CaretAPI {
	return &CaretAPI{}
}

func (a *CaretAPI) ListCaretTools(_ context.Context, req *hubcaretv1.ListCaretToolsRequest) (*hubcaretv1.ListCaretToolsResponse, error) {
	mode := echohandlers.ParseMode(req.GetMode())
	defs := echohandlers.ToolsForMode(mode)
	tools := make([]*hubcaretv1.CaretTool, 0, len(defs))
	for _, def := range defs {
		tools = append(tools, &hubcaretv1.CaretTool{
			Name:             def.Name,
			Label:            def.Name,
			Description:      def.Description,
			Category:         def.Module,
			RequiresApproval: def.RequiresApproval,
		})
	}
	return &hubcaretv1.ListCaretToolsResponse{Tools: tools}, nil
}

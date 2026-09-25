package api

import (
	"context"

	echoapp "github.com/voxire/onion/internal/modules/echo/internal/application"
	hubcaretv1 "github.com/voxire/voxie/gen/go/hub/caret/v1"
)

type CaretAPI struct{}

func NewCaretAPI() *CaretAPI {
	return &CaretAPI{}
}

func (a *CaretAPI) ListCaretTools(_ context.Context, req *hubcaretv1.ListCaretToolsRequest) (*hubcaretv1.ListCaretToolsResponse, error) {
	mode := echoapp.ParseMode(req.GetMode())
	defs := echoapp.ToolsForMode(mode)
	tools := make([]*hubcaretv1.CaretTool, 0, len(defs))
	for _, def := range defs {
		tools = append(tools, &hubcaretv1.CaretTool{
			Name:             def.Name,
			Label:            def.Name,
			Description:      def.Description,
			Category:         def.Module,
			RequiresApproval: def.RequiresApproval,
			Scope:            def.Scope,
		})
	}
	return &hubcaretv1.ListCaretToolsResponse{Tools: tools}, nil
}

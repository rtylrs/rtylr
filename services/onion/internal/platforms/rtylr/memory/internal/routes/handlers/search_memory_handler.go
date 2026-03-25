package handlers

import (
	"context"
	"net/http"

	errorModels "github.com/voxire/onion/internal/platform/errors"
	"github.com/voxire/onion/internal/platforms/rtylr/memory/internal/application"
)

func SearchRoutes() []string {
	return []string{
		"/memory/search",
	}
}

type SearchRequest struct {
	UserID int    `json:"user_id"`
	Query  string `json:"q"`
}

type SearchResponse struct {
	Success  bool        `json:"success"`
	Memories interface{} `json:"memories,omitempty"`
	Error    string      `json:"error,omitempty"`
}

type SearchHandler struct {
	MemoryService *application.Service
}

func (h *SearchHandler) Serve(ctx context.Context, req *SearchRequest) (SearchResponse, *errorModels.ServerError) {
	if req.UserID <= 0 {
		return SearchResponse{}, &errorModels.ServerError{
			Code:    http.StatusBadRequest,
			Message: "User ID is required",
		}
	}

	if req.Query == "" {
		return SearchResponse{}, &errorModels.ServerError{
			Code:    http.StatusBadRequest,
			Message: "Search query is required",
		}
	}

	memories, err := h.MemoryService.SearchMemories(ctx, req.UserID, req.Query)
	if err != nil {
		return SearchResponse{}, &errorModels.ServerError{
			Code:    http.StatusInternalServerError,
			Message: "Failed to search memories",
		}
	}

	return SearchResponse{
		Success:  true,
		Memories: memories,
	}, nil
}

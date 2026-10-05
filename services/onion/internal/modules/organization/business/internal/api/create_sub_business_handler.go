package api

import (
	"github.com/voxire/onion/internal/modules/organization/business/internal/application"
)

type (
	CreateSubBusinessHandler struct{ Service *application.Service }
	CreateSubBusinessRequest struct {
		Name             string `json:"name"`
		ParentBusinessID int64  `json:"parent_business_id"`
	}
	CreateSubBusinessResponse struct {
		ID int64 `json:"id"`
	}
)

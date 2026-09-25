package api

import (
	"github.com/voxire/onion/internal/modules/apps/internal/domain"
	hubappv1 "github.com/voxire/voxie/gen/go/hub/app/v1"
)

func toProtoApp(a domain.App) *hubappv1.App {
	return &hubappv1.App{
		Id:          a.ID,
		Name:        a.Name,
		Description: a.Description,
		Icon:        a.Icon,
		IsEnabled:   a.IsEnabled,
	}
}

package infrastructure

import (
	"github.com/voxire/onion/internal/modules/recruit/careers/internal/domain"
	sqlc "github.com/voxire/onion/internal/modules/recruit/careers/internal/infrastructure/persistence/sqlc"
	"github.com/voxire/onion/internal/platform/utils/convert"
)

func careerFromRow(row sqlc.GetAllCareersRow) domain.Career {
	return domain.Career{
		Id:             row.ID,
		Title:          row.Title,
		Department:     convert.ToString(row.Department),
		Location:       convert.ToString(row.Location),
		EmploymentType: convert.ToString(row.EmploymentType),
		WorkMode:       convert.ToString(row.WorkMode),
		Seniority:      convert.ToString(row.Seniority),
		Details:        row.Details,
		IsActive:       row.IsActive,
	}
}

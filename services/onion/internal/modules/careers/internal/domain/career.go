package domain

import (
	sqlc "github.com/voxire/onion/internal/modules/careers/internal/infrastructure/persistence/sqlc"
	"github.com/voxire/onion/internal/platform/utils/convert"
)

type Career struct {
	Id             int64
	Title          string
	Department     string
	Location       string
	EmploymentType string
	WorkMode       string
	Seniority      string
	Details        string
	IsActive       bool
}

func (c *Career) FromRows(row sqlc.GetAllCareersRow) {
	c.Id = row.ID
	c.Title = row.Title
	c.Department = convert.ToString(row.Department)
	c.Location = convert.ToString(row.Location)
	c.EmploymentType = convert.ToString(row.EmploymentType)
	c.WorkMode = convert.ToString(row.WorkMode)
	c.Seniority = convert.ToString(row.Seniority)
	c.Details = row.Details
	c.IsActive = row.IsActive
}

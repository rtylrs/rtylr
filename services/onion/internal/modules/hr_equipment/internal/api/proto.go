package api

import (
	"github.com/voxire/onion/internal/modules/hr_equipment/internal/domain"
	hrequipmentv1 "github.com/voxire/voxie/gen/go/hr/equipment/v1"
)

func equipmentRequestToProto(r domain.EquipmentRequestRecord) *hrequipmentv1.EquipmentRequestPayload {
	var neededByDate *string
	if r.NeededByDate != "" {
		neededByDate = &r.NeededByDate
	}

	var approvedBy *int64
	if r.ApprovedBy != 0 {
		approvedBy = &r.ApprovedBy
	}

	return &hrequipmentv1.EquipmentRequestPayload{
		Id:                   r.ID,
		OwnerId:              r.OwnerID,
		EmployeeId:           r.EmployeeID,
		EquipmentType:        r.ItemName,
		Description:          r.Reason,
		Urgency:              "",
		NeededByDate:         neededByDate,
		Status:               r.Status,
		CurrentApprovalLevel: r.CurrentApprovalLevel,
		ApprovedBy:           approvedBy,
		CreatedAt:            r.CreatedAt.Unix(),
		UpdatedAt:            r.UpdatedAt.Unix(),
	}
}

func assetToProto(a domain.AssetRecord) *hrequipmentv1.AssetPayload {
	var serialNumber *string
	if a.SerialNumber != "" {
		serialNumber = &a.SerialNumber
	}

	var returnedDate *string

	// hr_assets stores only purchase_date, so there is no assignment timestamp
	// to report. Leaving assigned_date unset keeps it from advertising the
	// purchase date as an assignment that never happened.
	return &hrequipmentv1.AssetPayload{
		Id:           a.ID,
		OwnerId:      a.OwnerID,
		EmployeeId:   a.AssignedTo,
		Name:         a.Name,
		SerialNumber: serialNumber,
		ReturnedDate: returnedDate,
		Condition:    a.Status,
		CreatedAt:    a.CreatedAt.Unix(),
		UpdatedAt:    a.UpdatedAt.Unix(),
	}
}

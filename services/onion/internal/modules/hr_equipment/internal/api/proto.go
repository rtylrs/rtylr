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

	return &hrequipmentv1.AssetPayload{
		Id:           a.ID,
		OwnerId:      a.OwnerID,
		EmployeeId:   a.AssignedTo,
		Name:         a.Name,
		SerialNumber: serialNumber,
		AssignedDate: a.PurchaseDate,
		ReturnedDate: returnedDate,
		Condition:    a.Status,
		CreatedAt:    a.CreatedAt.Unix(),
		UpdatedAt:    a.UpdatedAt.Unix(),
	}
}

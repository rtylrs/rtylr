package api

import (
	"github.com/voxire/onion/internal/modules/hr_onboarding/internal/domain"
	hronboardingv1 "github.com/voxire/voxie/gen/go/hr/onboarding/v1"
)

func templateToProto(t domain.ChecklistTemplateRecord) *hronboardingv1.ChecklistTemplatePayload {
	var roleID *int64
	if t.RoleID != 0 {
		roleID = &t.RoleID
	}

	items := make([]*hronboardingv1.ChecklistItemPayload, 0, len(t.Items))
	for _, item := range t.Items {
		items = append(items, checklistItemToProto(item))
	}

	return &hronboardingv1.ChecklistTemplatePayload{
		Id:          t.ID,
		OwnerId:     t.OwnerID,
		Name:        t.Name,
		TriggerType: t.TriggerType,
		RoleId:      roleID,
		Items:       items,
		CreatedAt:   t.CreatedAt.Unix(),
		UpdatedAt:   t.UpdatedAt.Unix(),
	}
}

func checklistItemToProto(i domain.ChecklistItemRecord) *hronboardingv1.ChecklistItemPayload {
	return &hronboardingv1.ChecklistItemPayload{
		Id:                 i.ID,
		OwnerId:            i.OwnerID,
		TemplateId:         i.TemplateID,
		Title:              i.Title,
		Description:        i.Description,
		AssigneeRole:       i.AssigneeRole,
		DueOffsetDays:      i.DueOffsetDays,
		RequiresAttachment: i.RequiresAttachment,
		CreatedAt:          i.CreatedAt.Unix(),
		UpdatedAt:          i.UpdatedAt.Unix(),
	}
}

func employeeChecklistToProto(c domain.EmployeeChecklistRecord) *hronboardingv1.EmployeeChecklistPayload {
	items := make([]*hronboardingv1.EmployeeChecklistItemPayload, 0, len(c.Items))
	for _, item := range c.Items {
		items = append(items, employeeChecklistItemToProto(item))
	}

	return &hronboardingv1.EmployeeChecklistPayload{
		Id:          c.ID,
		OwnerId:     c.OwnerID,
		EmployeeId:  c.EmployeeID,
		TemplateId:  c.TemplateID,
		TriggeredAt: c.TriggeredAt.Unix(),
		Items:       items,
		CreatedAt:   c.CreatedAt.Unix(),
		UpdatedAt:   c.UpdatedAt.Unix(),
	}
}

func employeeChecklistItemToProto(i domain.EmployeeChecklistItemRecord) *hronboardingv1.EmployeeChecklistItemPayload {
	var completedAt *int64
	if !i.CompletedAt.IsZero() {
		v := i.CompletedAt.Unix()
		completedAt = &v
	}

	var attachmentURL *string
	if i.AttachmentURL != "" {
		attachmentURL = &i.AttachmentURL
	}

	return &hronboardingv1.EmployeeChecklistItemPayload{
		Id:                   i.ID,
		OwnerId:              i.OwnerID,
		EmployeeChecklistId:  i.EmployeeChecklistID,
		ChecklistItemId:      i.ChecklistItemID,
		AssignedToEmployeeId: i.AssignedToEmployeeID,
		DueDate:              i.DueDate,
		Status:               i.Status,
		CompletedAt:          completedAt,
		AttachmentUrl:        attachmentURL,
		CreatedAt:            i.CreatedAt.Unix(),
		UpdatedAt:            i.UpdatedAt.Unix(),
	}
}

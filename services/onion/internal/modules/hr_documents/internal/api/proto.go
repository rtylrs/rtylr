package api

import (
	"github.com/voxire/onion/internal/modules/hr_documents/internal/domain"
	hrdocumentv1 "github.com/voxire/voxie/gen/go/hr/document/v1"
)

func documentToProto(d domain.EmployeeDocumentRecord) *hrdocumentv1.EmployeeDocumentPayload {
	var expiryDate *string
	if d.ExpiryDate != nil {
		s := d.ExpiryDate.Format("2006-01-02")
		expiryDate = &s
	}

	fileURL := d.FileURL
	if d.StorageMode == "linked" {
		fileURL = d.ExternalURL
	}

	return &hrdocumentv1.EmployeeDocumentPayload{
		Id:           d.ID,
		OwnerId:      d.OwnerID,
		EmployeeId:   d.EmployeeID,
		DocumentType: d.DocumentType,
		Name:         d.Title,
		StorageMode:  d.StorageMode,
		FileUrl:      fileURL,
		ExpiryDate:   expiryDate,
		UploadedBy:   d.UploadedBy,
		UploadedAt:   d.CreatedAt.Unix(),
		CreatedAt:    d.CreatedAt.Unix(),
		UpdatedAt:    d.UpdatedAt.Unix(),
	}
}

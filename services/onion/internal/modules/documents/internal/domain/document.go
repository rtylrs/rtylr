package domain

import (
	"database/sql"
	"errors"
	"time"
)

var ErrDocumentNotFound = errors.New("document not found")

type Document struct {
	ID         int64     `json:"id"`
	OwnerID    int64     `json:"ownerId"`
	StaffID    int64     `json:"staffId"`
	Name       string    `json:"name"`
	Type       string    `json:"type"`
	FileURL    string    `json:"fileUrl"`
	ExpiryDate time.Time `json:"expiryDate"`
	CreatedAt  time.Time `json:"createdAt"`
}

func (d *Document) FromRow(row *sql.Rows) error {
	var (
		staffID    sql.NullInt64
		docType    sql.NullString
		fileURL    sql.NullString
		expiryDate sql.NullTime
	)

	err := row.Scan(&d.ID, &d.OwnerID, &staffID, &d.Name, &docType, &fileURL, &expiryDate, &d.CreatedAt)
	if err != nil {
		return err
	}

	if staffID.Valid {
		d.StaffID = staffID.Int64
	}
	if docType.Valid {
		d.Type = docType.String
	}
	if fileURL.Valid {
		d.FileURL = fileURL.String
	}
	if expiryDate.Valid {
		d.ExpiryDate = expiryDate.Time
	}

	return nil
}

type DocumentInput struct {
	StaffID    int64  `json:"staffId"`
	Name       string `json:"name"`
	Type       string `json:"type"`
	FileURL    string `json:"fileUrl"`
	ExpiryDate string `json:"expiryDate"`
}

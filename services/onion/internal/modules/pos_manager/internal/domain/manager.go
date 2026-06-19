package domain

import "errors"

var ErrInvalidPIN = errors.New("invalid manager PIN")

type VerifyResult struct {
	Valid      bool   `json:"valid"`
	StaffID    int64  `json:"staff_id"`
	OwnerID    int64  `json:"owner_id"`
	BusinessID int64  `json:"business_id"`
	Name       string `json:"name"`
}

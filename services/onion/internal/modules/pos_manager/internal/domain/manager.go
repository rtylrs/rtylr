package domain

import "errors"

var ErrInvalidPIN = errors.New("invalid manager PIN")

type VerifyResult struct {
	Valid   bool   `json:"valid"`
	StaffID int64  `json:"staff_id"`
	Name    string `json:"name"`
}

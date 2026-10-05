package domain

import "errors"

var (
	ErrEquipmentRequestNotFound   = errors.New("equipment request not found")
	ErrAssetNotFound              = errors.New("asset not found")
	ErrAssetAlreadyAssigned       = errors.New("asset is already assigned")
	ErrEquipmentRequestNotPending = errors.New("equipment request is no longer pending")
	ErrApprovalChainNotConfigured = errors.New("approval chain not configured for equipment requests")
)

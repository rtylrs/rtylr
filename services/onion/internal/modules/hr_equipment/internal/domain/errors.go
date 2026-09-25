package domain

import "errors"

var (
	ErrEquipmentRequestNotFound   = errors.New("equipment request not found")
	ErrAssetNotFound              = errors.New("asset not found")
	ErrAssetAlreadyAssigned       = errors.New("asset is already assigned")
	ErrApprovalChainNotConfigured = errors.New("approval chain not configured for equipment requests")
)

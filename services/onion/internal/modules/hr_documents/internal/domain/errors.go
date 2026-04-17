package domain

import "errors"

var (
	ErrDocumentNotFound   = errors.New("document not found")
	ErrStorageModeInvalid = errors.New("storage_mode must be 'stored' or 'linked'")
)

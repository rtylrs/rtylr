package domain

import "errors"

var (
	ErrCareerNotFound        = errors.New("career not found")
	ErrSubmissionNotFound    = errors.New("career submission not found")
	ErrCustomizationNotFound = errors.New("customization not found")
	ErrFormFieldNotFound     = errors.New("form field not found")
	ErrSectionNotFound       = errors.New("section not found")
)

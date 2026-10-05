package domain

import "errors"

var (
	ErrRunNotClaimed    = errors.New("automation run already claimed")
	ErrCustomerNotFound = errors.New("automation customer not found")
)

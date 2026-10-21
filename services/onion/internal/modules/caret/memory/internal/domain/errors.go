package domain

import "errors"

var ErrChatSessionNotFound = errors.New("chat session not found")

// ErrMemoryNotFound means the memory does not exist or is not visible to the
// signed-in user in the current business.
var ErrMemoryNotFound = errors.New("memory not found")

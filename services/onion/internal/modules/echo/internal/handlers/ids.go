package handlers

import (
	"fmt"
	"sync/atomic"
	"time"
)

var caretIDCounter atomic.Uint64

func newAssistantMessageID() string {
	return fmt.Sprintf("assistant-%d-%d", time.Now().UnixNano(), caretIDCounter.Add(1))
}

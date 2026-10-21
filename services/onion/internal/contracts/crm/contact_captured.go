package crm

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
)

const CustomerContactCapturedType = "customer_contact_captured"
const CustomerLinkedType = "crm_customer_linked"

// CustomerContactCaptured describes a contact fact produced by another
// business capability. CRM consumes it into its customer directory.
type CustomerContactCaptured struct {
	Name       string  `json:"name"`
	Email      string  `json:"email"`
	Phone      string  `json:"phone"`
	TotalSpent float64 `json:"total_spent"`
	CountVisit bool    `json:"count_visit"`
}

type CustomerLinked struct {
	SourceType string `json:"source_type"`
	SourceID   int64  `json:"source_id"`
	CustomerID int64  `json:"customer_id"`
}

func CustomerLinkDedupeKey(sourceType string, sourceID, customerID int64) string {
	return fmt.Sprintf("crm_customer_linked:%s:%d:%d", sourceType, sourceID, customerID)
}

func CustomerContactCaptureDedupeKey(sourceType string, subjectID int64, contact CustomerContactCaptured) string {
	payload, _ := json.Marshal(contact)
	digest := sha256.Sum256(payload)
	return fmt.Sprintf("customer-contact:%s:%d:%s", sourceType, subjectID, hex.EncodeToString(digest[:12]))
}

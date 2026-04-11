package domain

import "time"

type MemberLocation struct {
	ID                 int64
	BusinessID         int64
	UserID             int64
	LocationID         int64
	CanManagePOSUsers  bool
	CanManageTerminals bool
	CanViewReports     bool
	CreatedAt          time.Time
}

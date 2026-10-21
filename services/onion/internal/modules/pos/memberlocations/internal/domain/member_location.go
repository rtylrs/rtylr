package domain

import "time"

type MyLocation struct {
	LocationID int64
	Name       string
	Address    string
	Phone      string
}

type RosterEntry struct {
	UserID    int64
	Name      string
	HasPIN    bool
	IsManager bool
	POSRole   string
}

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

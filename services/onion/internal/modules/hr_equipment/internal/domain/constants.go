package domain

// Equipment request status values.
const (
	EquipmentStatusPending   = "pending"
	EquipmentStatusApproved  = "approved"
	EquipmentStatusRejected  = "rejected"
	EquipmentStatusFulfilled = "fulfilled"
)

// Asset status values.
const (
	AssetStatusAvailable = "available"
	AssetStatusAssigned  = "assigned"
	AssetStatusRetired   = "retired"
	AssetStatusRepair    = "repair"
)

// Asset condition is independent of the lifecycle status above. An asset can
// be assigned while being in fair condition, or available while damaged and
// awaiting repair.
const (
	AssetConditionNew     = "new"
	AssetConditionGood    = "good"
	AssetConditionFair    = "fair"
	AssetConditionDamaged = "damaged"
)

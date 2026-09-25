package domain

// Review cycle status values.
const (
	CycleStatusDraft    = "draft"
	CycleStatusOpen     = "open"
	CycleStatusClosed   = "closed"
	CycleStatusArchived = "archived"
)

// Review cycle type values.
const (
	CycleTypeQuarterly = "quarterly"
	CycleTypeAnnual    = "annual"
	CycleTypeMidYear   = "mid_year"
	CycleTypeCustom    = "custom"
	CycleTypeProbation = "probation"
	CycleTypeAdHoc     = "adhoc"
)

// Review type values.
const (
	ReviewTypeSelf    = "self"
	ReviewTypeManager = "manager"
	ReviewTypePeer    = "peer"
)

// Goal status values.
const (
	GoalStatusActive    = "active"
	GoalStatusCompleted = "completed"
	GoalStatusCanceled  = "canceled"
)

// Rating scale type values.
const (
	RatingScaleNumeric = "numeric"
	RatingScaleStar    = "star"
	RatingScaleLabel   = "label"
)

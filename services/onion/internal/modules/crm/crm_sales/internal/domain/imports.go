package domain

import "time"

// MaxImportRows bounds one import request; files go up in chunks.
const MaxImportRows = 500

// ImportBatch is one uploaded file's progress and outcome.
type ImportBatch struct {
	ID          int64
	FileName    string
	TotalRows   int
	Created     int
	Updated     int
	Skipped     int
	Failed      int
	Status      string
	CreatedBy   int64
	CreatedAt   time.Time
	CompletedAt *time.Time
}

// DuplicatePolicy is what an import does with a row matching an existing
// lead by email, phone, or company and name.
type DuplicatePolicy string

const (
	DuplicatesSkip DuplicatePolicy = "skip"
	// DuplicatesFill fills only the fields the existing lead is missing, so a
	// rep's edits are never overwritten.
	DuplicatesFill DuplicatePolicy = "fill"
)

// ImportOptions apply to every row of a chunk.
type ImportOptions struct {
	Duplicates DuplicatePolicy
	AssignedTo *int64
	Source     string
}

// ImportRow is one spreadsheet row; Row is its line in the file, for
// reporting.
type ImportRow struct {
	Row  int
	Lead LeadInput
}

// ImportOutcome is what happened to a row.
type ImportOutcome string

const (
	ImportCreated ImportOutcome = "created"
	ImportUpdated ImportOutcome = "updated"
	ImportSkipped ImportOutcome = "skipped"
	ImportFailed  ImportOutcome = "failed"
)

// ImportRowResult reports one row back to the file it came from.
type ImportRowResult struct {
	Row     int
	Outcome ImportOutcome
	LeadID  int64
	Error   string
}

// UndoImportResult counts what undoing an import removed and what it kept
// because someone has worked those leads since.
type UndoImportResult struct {
	Removed int64
	Kept    int64
}

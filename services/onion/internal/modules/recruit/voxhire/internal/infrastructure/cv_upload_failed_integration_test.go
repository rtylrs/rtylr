package infrastructure_test

import (
	"context"
	"database/sql"
	"testing"

	"github.com/voxire/onion/internal/modules/recruit/voxhire/internal/domain"
	"github.com/voxire/onion/internal/modules/recruit/voxhire/internal/infrastructure"
	"github.com/voxire/onion/internal/testutil"
)

// An application whose CV could not be stored is still saved, marked
// upload_failed with the reason, and with no cv_key so the CV worker skips it.
func TestVoxhire_CreateTalentApplication_CVUploadFailedIsPersisted(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id, err := repo.CreateTalentApplication(ctx, domain.TalentApplicationInput{
		FullName:       "Upload Failed",
		Email:          "upload-failed@example.com",
		Consent:        true,
		CVUploadFailed: true,
		CVUploadError:  "AccessDenied",
	})
	if err != nil {
		t.Fatalf("CreateTalentApplication: %v", err)
	}
	t.Cleanup(func() { _, _ = db.ExecContext(ctx, `DELETE FROM voxhire_talent_applications WHERE id = ?`, id) })

	assertCVColumns(t, db, `SELECT cv_processing_status, cv_processing_error, cv_key FROM voxhire_talent_applications WHERE id = ?`, id)

	targets, err := repo.ListPendingCVProcessingTargets(ctx, 100)
	if err != nil {
		t.Fatalf("ListPendingCVProcessingTargets: %v", err)
	}
	for _, target := range targets {
		if target.Source == "talent" && target.ID == id {
			t.Fatal("an application without a stored CV was queued for CV processing")
		}
	}
}

func TestVoxhire_CreateCareerApplication_CVUploadFailedIsPersisted(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id, err := repo.CreateCareerApplication(ctx, domain.CareerApplicationInput{
		JobID:          1,
		Name:           "Upload Failed",
		Email:          "upload-failed@example.com",
		CVUploadFailed: true,
		CVUploadError:  "AccessDenied",
	})
	if err != nil {
		t.Fatalf("CreateCareerApplication: %v", err)
	}
	t.Cleanup(func() { _, _ = db.ExecContext(ctx, `DELETE FROM voxire_career_applications WHERE id = ?`, id) })

	assertCVColumns(t, db, `SELECT cv_processing_status, cv_processing_error, cv_key FROM voxire_career_applications WHERE id = ?`, id)
}

func TestVoxhire_FindRecentTalentApplicationID(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()
	email := "recent-duplicate@example.com"

	if id, err := repo.FindRecentTalentApplicationID(ctx, email); err != nil || id != 0 {
		t.Fatalf("before insert: id=%d err=%v, want 0 and nil", id, err)
	}
	id, err := repo.CreateTalentApplication(ctx, domain.TalentApplicationInput{FullName: "Recent", Email: email, Consent: true})
	if err != nil {
		t.Fatalf("CreateTalentApplication: %v", err)
	}
	t.Cleanup(func() { _, _ = db.ExecContext(ctx, `DELETE FROM voxhire_talent_applications WHERE id = ?`, id) })

	// The column collation is case-insensitive, so a retry typed differently still matches.
	got, err := repo.FindRecentTalentApplicationID(ctx, "Recent-Duplicate@example.com")
	if err != nil || got != id {
		t.Fatalf("FindRecentTalentApplicationID = %d, %v; want %d", got, err, id)
	}
}

func assertCVColumns(t *testing.T, db *sql.DB, query string, id int64) {
	t.Helper()
	var status string
	var cvErr, cvKey sql.NullString
	if err := db.QueryRowContext(context.Background(), query, id).Scan(&status, &cvErr, &cvKey); err != nil {
		t.Fatalf("read back row %d: %v", id, err)
	}
	if status != domain.CVStatusUploadFailed {
		t.Fatalf("cv_processing_status = %q, want %q", status, domain.CVStatusUploadFailed)
	}
	if cvErr.String != "AccessDenied" {
		t.Fatalf("cv_processing_error = %q, want AccessDenied", cvErr.String)
	}
	if cvKey.Valid && cvKey.String != "" {
		t.Fatalf("cv_key = %q, want empty", cvKey.String)
	}
}

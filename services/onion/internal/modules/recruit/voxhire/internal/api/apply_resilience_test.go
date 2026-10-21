package api

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/voxire/onion/internal/modules/recruit/voxhire/internal/domain"
)

// ── fakes ───────────────────────────────────────────────────────────────────

type fakeCVStore struct {
	err  error
	keys []string
}

func (f *fakeCVStore) Upload(_ context.Context, key, _ string, body io.Reader) error {
	if _, err := io.Copy(io.Discard, body); err != nil {
		return err
	}
	if f.err != nil {
		return f.err
	}
	f.keys = append(f.keys, key)
	return nil
}

// stubCVStore swaps the storage opener for the duration of the test. A
// non-nil openErr simulates storage.Default failing (e.g. missing bucket).
func stubCVStore(t *testing.T, store *fakeCVStore, openErr error) {
	t.Helper()
	original := openCVStore
	t.Cleanup(func() { openCVStore = original })
	openCVStore = func() (cvObjectStore, error) {
		if openErr != nil {
			return nil, openErr
		}
		return store, nil
	}
}

type fakeTalentSubmitter struct {
	id        int64
	err       error
	recentID  int64
	recentErr error
	calls     int
	got       domain.TalentApplicationInput
}

func (f *fakeTalentSubmitter) RecentTalentApplication(_ context.Context, _ string) (int64, error) {
	return f.recentID, f.recentErr
}

func (f *fakeTalentSubmitter) SubmitTalentApplication(_ context.Context, in domain.TalentApplicationInput) (int64, error) {
	f.calls++
	f.got = in
	return f.id, f.err
}

type fakeCareerSubmitter struct {
	id        int64
	err       error
	recentID  int64
	recentErr error
	calls     int
	got       domain.CareerApplicationInput
}

func (f *fakeCareerSubmitter) RecentCareerApplication(_ context.Context, _ int64, _ string) (int64, error) {
	return f.recentID, f.recentErr
}

func (f *fakeCareerSubmitter) SubmitCareerApplication(_ context.Context, in domain.CareerApplicationInput) (int64, error) {
	f.calls++
	f.got = in
	return f.id, f.err
}

// ── helpers ─────────────────────────────────────────────────────────────────

func multipartRequest(t *testing.T, path string, fields map[string]string, filename string, cv []byte) *http.Request {
	t.Helper()
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	for key, value := range fields {
		if err := writer.WriteField(key, value); err != nil {
			t.Fatalf("WriteField(%s): %v", key, err)
		}
	}
	if filename != "" {
		part, err := writer.CreateFormFile("cv", filename)
		if err != nil {
			t.Fatalf("CreateFormFile: %v", err)
		}
		if _, err := part.Write(cv); err != nil {
			t.Fatalf("write cv: %v", err)
		}
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}
	req := httptest.NewRequest(http.MethodPost, path, body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	return req
}

func talentFields() map[string]string {
	return map[string]string{
		"fullName": "Khadija Example",
		"email":    "khadija@example.com",
		"consent":  "true",
		"track":    "Backend Engineering",
	}
}

func careerFields() map[string]string {
	return map[string]string{
		"job_id":  "7",
		"name":    "Khadija Example",
		"email":   "khadija@example.com",
		"consent": "true",
	}
}

func decodeBody(t *testing.T, rec *httptest.ResponseRecorder) map[string]any {
	t.Helper()
	var out map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &out); err != nil {
		t.Fatalf("decode body %q: %v", rec.Body.String(), err)
	}
	return out
}

func serveTalent(t *testing.T, svc TalentApplicationSubmitter, req *http.Request) *httptest.ResponseRecorder {
	t.Helper()
	rec := httptest.NewRecorder()
	ApplicationsHandler(svc).ServeHTTP(rec, req)
	return rec
}

func serveCareer(t *testing.T, svc CareerApplicationSubmitter, req *http.Request) *httptest.ResponseRecorder {
	t.Helper()
	rec := httptest.NewRecorder()
	CareerApplyHandler(svc).ServeHTTP(rec, req)
	return rec
}

func assertCode(t *testing.T, rec *httptest.ResponseRecorder, status int, code string) map[string]any {
	t.Helper()
	if rec.Code != status {
		t.Fatalf("status = %d, want %d (body %s)", rec.Code, status, rec.Body.String())
	}
	body := decodeBody(t, rec)
	if got, _ := body["code"].(string); got != code {
		t.Fatalf("code = %q, want %q (body %s)", got, code, rec.Body.String())
	}
	return body
}

var pdf = []byte("%PDF-1.4 test")

// ── talent pool: POST /voxhire/applications ────────────────────────────────

func TestApplicationsHandler_StoresCVAndSaves(t *testing.T) {
	store := &fakeCVStore{}
	stubCVStore(t, store, nil)
	svc := &fakeTalentSubmitter{id: 41}

	rec := serveTalent(t, svc, multipartRequest(t, "/voxhire/applications", talentFields(), "Khadija CV.pdf", pdf))

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200 (body %s)", rec.Code, rec.Body.String())
	}
	body := decodeBody(t, rec)
	if body["cv_stored"] != true {
		t.Fatalf("cv_stored = %v, want true", body["cv_stored"])
	}
	if _, ok := body["warning_code"]; ok {
		t.Fatalf("unexpected warning_code on a clean submission: %v", body)
	}
	if len(store.keys) != 1 || !strings.HasPrefix(store.keys[0], "voxhire/applications/backend-engineering/") {
		t.Fatalf("uploaded keys = %v", store.keys)
	}
	if svc.got.CVKey != store.keys[0] || svc.got.CVUploadFailed {
		t.Fatalf("submitted CVKey=%q CVUploadFailed=%v", svc.got.CVKey, svc.got.CVUploadFailed)
	}
}

func TestApplicationsHandler_StorageUnavailable_SavesWithoutCV(t *testing.T) {
	stubCVStore(t, nil, errors.New("s3 storage is not configured"))
	svc := &fakeTalentSubmitter{id: 42}

	rec := serveTalent(t, svc, multipartRequest(t, "/voxhire/applications", talentFields(), "cv.pdf", pdf))

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200 (body %s)", rec.Code, rec.Body.String())
	}
	body := decodeBody(t, rec)
	if body["warning_code"] != applyCodeCVUploadFailed || body["cv_stored"] != false {
		t.Fatalf("body = %v, want cv_upload_failed warning and cv_stored=false", body)
	}
	if svc.calls != 1 {
		t.Fatalf("application not saved: %d submit calls", svc.calls)
	}
	if !svc.got.CVUploadFailed || svc.got.CVKey != "" {
		t.Fatalf("submitted CVUploadFailed=%v CVKey=%q", svc.got.CVUploadFailed, svc.got.CVKey)
	}
	if !strings.Contains(svc.got.CVUploadError, "not configured") {
		t.Fatalf("CVUploadError = %q", svc.got.CVUploadError)
	}
}

func TestApplicationsHandler_UploadFails_SavesWithoutCV(t *testing.T) {
	stubCVStore(t, &fakeCVStore{err: errors.New("AccessDenied: s3:PutObject")}, nil)
	svc := &fakeTalentSubmitter{id: 43}

	rec := serveTalent(t, svc, multipartRequest(t, "/voxhire/applications", talentFields(), "cv.pdf", pdf))

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200 (body %s)", rec.Code, rec.Body.String())
	}
	if body := decodeBody(t, rec); body["warning_code"] != applyCodeCVUploadFailed {
		t.Fatalf("body = %v, want cv_upload_failed warning", body)
	}
	if !svc.got.CVUploadFailed || !strings.Contains(svc.got.CVUploadError, "AccessDenied") {
		t.Fatalf("submitted CVUploadFailed=%v CVUploadError=%q", svc.got.CVUploadFailed, svc.got.CVUploadError)
	}
}

func TestApplicationsHandler_SaveFails_Returns500WithCode(t *testing.T) {
	stubCVStore(t, &fakeCVStore{}, nil)
	svc := &fakeTalentSubmitter{err: errors.New("Error 1054: Unknown column 'current_role'")}

	rec := serveTalent(t, svc, multipartRequest(t, "/voxhire/applications", talentFields(), "cv.pdf", pdf))

	assertCode(t, rec, http.StatusInternalServerError, applyCodeSaveFailed)
}

func TestApplicationsHandler_FieldTooLong_Returns400NamingField(t *testing.T) {
	store := &fakeCVStore{}
	stubCVStore(t, store, nil)
	svc := &fakeTalentSubmitter{id: 1}
	fields := talentFields()
	fields["track"] = strings.Repeat("x", 65) // track is VARCHAR(64)

	rec := serveTalent(t, svc, multipartRequest(t, "/voxhire/applications", fields, "cv.pdf", pdf))

	body := assertCode(t, rec, http.StatusBadRequest, applyCodeFieldTooLong)
	if body["field"] != "track" || body["max_length"] != float64(64) {
		t.Fatalf("body = %v, want field=track max_length=64", body)
	}
	if svc.calls != 0 || len(store.keys) != 0 {
		t.Fatalf("rejected submission still reached storage (%d) or the database (%d)", len(store.keys), svc.calls)
	}
}

func TestApplicationsHandler_TextFieldOverByteLimit_Returns400(t *testing.T) {
	stubCVStore(t, &fakeCVStore{}, nil)
	fields := talentFields()
	fields["pitch"] = strings.Repeat("x", mysqlTextMaxBytes+1)

	rec := serveTalent(t, &fakeTalentSubmitter{}, multipartRequest(t, "/voxhire/applications", fields, "cv.pdf", pdf))

	if body := assertCode(t, rec, http.StatusBadRequest, applyCodeFieldTooLong); body["field"] != "pitch" {
		t.Fatalf("field = %v, want pitch", body["field"])
	}
}

// VARCHAR limits count characters, not bytes: a 255-letter Arabic name is
// 510 bytes and must still be accepted.
func TestApplicationsHandler_MultibyteAtVarcharLimit_Accepted(t *testing.T) {
	stubCVStore(t, &fakeCVStore{}, nil)
	svc := &fakeTalentSubmitter{id: 5}
	fields := talentFields()
	fields["fullName"] = strings.Repeat("خ", 255)

	rec := serveTalent(t, svc, multipartRequest(t, "/voxhire/applications", fields, "cv.pdf", pdf))

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200 (body %s)", rec.Code, rec.Body.String())
	}
}

func TestApplicationsHandler_ValidationErrorsCarryCodes(t *testing.T) {
	stubCVStore(t, &fakeCVStore{}, nil)

	missingName := talentFields()
	delete(missingName, "fullName")
	noConsent := talentFields()
	noConsent["consent"] = "false"

	tests := []struct {
		name   string
		req    *http.Request
		status int
		code   string
		field  string
	}{
		{"missing name", multipartRequest(t, "/voxhire/applications", missingName, "cv.pdf", pdf), http.StatusBadRequest, applyCodeMissingField, "fullName"},
		{"no consent", multipartRequest(t, "/voxhire/applications", noConsent, "cv.pdf", pdf), http.StatusBadRequest, applyCodeConsentRequired, "consent"},
		{"no cv", multipartRequest(t, "/voxhire/applications", talentFields(), "", nil), http.StatusBadRequest, applyCodeCVRequired, "cv"},
		{"wrong type", multipartRequest(t, "/voxhire/applications", talentFields(), "notes.txt", []byte("hi")), http.StatusBadRequest, applyCodeCVUnsupportedType, "cv"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body := assertCode(t, serveTalent(t, &fakeTalentSubmitter{}, tt.req), tt.status, tt.code)
			if body["field"] != tt.field {
				t.Fatalf("field = %v, want %s", body["field"], tt.field)
			}
		})
	}
}

func TestApplicationsHandler_OversizedBodyCarriesCode(t *testing.T) {
	body, contentType := newVoxhireMultipartBody(t, maxApplicationMultipartBytes+1)
	req := httptest.NewRequest(http.MethodPost, "/voxhire/applications", body)
	req.Header.Set("Content-Type", contentType)

	assertCode(t, serveTalent(t, nil, req), http.StatusRequestEntityTooLarge, applyCodeUploadTooLarge)
}

// ── specific role: POST /voxhire/careers/apply ──────────────────────────────

func TestCareerApplyHandler_StoresCVAndSaves(t *testing.T) {
	store := &fakeCVStore{}
	stubCVStore(t, store, nil)
	svc := &fakeCareerSubmitter{id: 9}

	rec := serveCareer(t, svc, multipartRequest(t, "/voxhire/careers/apply", careerFields(), "cv.pdf", pdf))

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200 (body %s)", rec.Code, rec.Body.String())
	}
	if len(store.keys) != 1 || !strings.HasPrefix(store.keys[0], "voxire/applications/7/") {
		t.Fatalf("uploaded keys = %v", store.keys)
	}
	if svc.got.CVKey != store.keys[0] || svc.got.JobID != 7 {
		t.Fatalf("submitted %+v", svc.got)
	}
}

func TestCareerApplyHandler_StorageUnavailable_SavesWithoutCV(t *testing.T) {
	stubCVStore(t, nil, errors.New("s3 storage is not configured"))
	svc := &fakeCareerSubmitter{id: 10}

	rec := serveCareer(t, svc, multipartRequest(t, "/voxhire/careers/apply", careerFields(), "cv.pdf", pdf))

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200 (body %s)", rec.Code, rec.Body.String())
	}
	if body := decodeBody(t, rec); body["warning_code"] != applyCodeCVUploadFailed {
		t.Fatalf("body = %v, want cv_upload_failed warning", body)
	}
	if svc.calls != 1 || !svc.got.CVUploadFailed || svc.got.CVKey != "" {
		t.Fatalf("calls=%d submitted %+v", svc.calls, svc.got)
	}
}

func TestCareerApplyHandler_UploadFails_SavesWithoutCV(t *testing.T) {
	stubCVStore(t, &fakeCVStore{err: errors.New("RequestTimeout")}, nil)
	svc := &fakeCareerSubmitter{id: 11}

	rec := serveCareer(t, svc, multipartRequest(t, "/voxhire/careers/apply", careerFields(), "cv.pdf", pdf))

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200 (body %s)", rec.Code, rec.Body.String())
	}
	if !svc.got.CVUploadFailed || svc.got.CVUploadError != "RequestTimeout" {
		t.Fatalf("submitted %+v", svc.got)
	}
}

func TestCareerApplyHandler_SaveFails_Returns500WithCode(t *testing.T) {
	stubCVStore(t, &fakeCVStore{}, nil)

	rec := serveCareer(t, &fakeCareerSubmitter{err: errors.New("db down")},
		multipartRequest(t, "/voxhire/careers/apply", careerFields(), "cv.pdf", pdf))

	assertCode(t, rec, http.StatusInternalServerError, applyCodeSaveFailed)
}

func TestCareerApplyHandler_FieldTooLong_Returns400NamingField(t *testing.T) {
	stubCVStore(t, &fakeCVStore{}, nil)
	svc := &fakeCareerSubmitter{}
	fields := careerFields()
	fields["location"] = strings.Repeat("x", 161) // location is VARCHAR(160)

	rec := serveCareer(t, svc, multipartRequest(t, "/voxhire/careers/apply", fields, "cv.pdf", pdf))

	body := assertCode(t, rec, http.StatusBadRequest, applyCodeFieldTooLong)
	if body["field"] != "location" || body["max_length"] != float64(160) {
		t.Fatalf("body = %v, want field=location max_length=160", body)
	}
	if svc.calls != 0 {
		t.Fatal("rejected submission reached the database")
	}
}

func TestCareerApplyHandler_Salary(t *testing.T) {
	tests := []struct {
		raw    string
		status int
		want   float64
	}{
		{"", http.StatusOK, 0},
		{"2500", http.StatusOK, 2500},
		{"not a number", http.StatusOK, 0}, // stays lenient, as before
		{"1e20", http.StatusBadRequest, 0}, // does not fit DECIMAL(12,2)
		{"NaN", http.StatusBadRequest, 0},
	}
	for _, tt := range tests {
		t.Run(tt.raw, func(t *testing.T) {
			stubCVStore(t, &fakeCVStore{}, nil)
			svc := &fakeCareerSubmitter{id: 1}
			fields := careerFields()
			fields["salary_expectation"] = tt.raw

			rec := serveCareer(t, svc, multipartRequest(t, "/voxhire/careers/apply", fields, "cv.pdf", pdf))

			if tt.status == http.StatusBadRequest {
				assertCode(t, rec, http.StatusBadRequest, applyCodeInvalidSalary)
				return
			}
			if rec.Code != tt.status || svc.got.SalaryExpectation != tt.want {
				t.Fatalf("status=%d salary=%v, want %d and %v", rec.Code, svc.got.SalaryExpectation, tt.status, tt.want)
			}
		})
	}
}

func TestCareerApplyHandler_MissingJobIDCarriesCode(t *testing.T) {
	fields := careerFields()
	delete(fields, "job_id")

	rec := serveCareer(t, &fakeCareerSubmitter{}, multipartRequest(t, "/voxhire/careers/apply", fields, "cv.pdf", pdf))

	assertCode(t, rec, http.StatusBadRequest, applyCodeInvalidJobID)
}

// ── duplicate guard ─────────────────────────────────────────────────────────

// An applicant who resubmits after seeing an error must not create a second
// row, a second CV object or a second team email.
func TestApplicationsHandler_RecentDuplicate_ReturnsExistingApplication(t *testing.T) {
	store := &fakeCVStore{}
	stubCVStore(t, store, nil)
	svc := &fakeTalentSubmitter{id: 99, recentID: 41}

	rec := serveTalent(t, svc, multipartRequest(t, "/voxhire/applications", talentFields(), "cv.pdf", pdf))

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200 (body %s)", rec.Code, rec.Body.String())
	}
	body := decodeBody(t, rec)
	if body["id"] != float64(41) || body["duplicate"] != true {
		t.Fatalf("body = %v, want the existing id 41 flagged duplicate", body)
	}
	if svc.calls != 0 || len(store.keys) != 0 {
		t.Fatalf("duplicate still saved (%d) or uploaded (%d)", svc.calls, len(store.keys))
	}
}

func TestApplicationsHandler_DuplicateCheckFails_StillSaves(t *testing.T) {
	stubCVStore(t, &fakeCVStore{}, nil)
	svc := &fakeTalentSubmitter{id: 7, recentErr: errors.New("db timeout")}

	rec := serveTalent(t, svc, multipartRequest(t, "/voxhire/applications", talentFields(), "cv.pdf", pdf))

	if rec.Code != http.StatusOK || svc.calls != 1 {
		t.Fatalf("status=%d calls=%d, want the application saved despite the failed lookup", rec.Code, svc.calls)
	}
}

func TestCareerApplyHandler_RecentDuplicate_ReturnsExistingApplication(t *testing.T) {
	store := &fakeCVStore{}
	stubCVStore(t, store, nil)
	svc := &fakeCareerSubmitter{id: 99, recentID: 12}

	rec := serveCareer(t, svc, multipartRequest(t, "/voxhire/careers/apply", careerFields(), "cv.pdf", pdf))

	body := decodeBody(t, rec)
	if rec.Code != http.StatusOK || body["id"] != float64(12) || body["duplicate"] != true {
		t.Fatalf("status=%d body=%v, want existing id 12 flagged duplicate", rec.Code, body)
	}
	if svc.calls != 0 || len(store.keys) != 0 {
		t.Fatalf("duplicate still saved (%d) or uploaded (%d)", svc.calls, len(store.keys))
	}
}

// ── helpers under test ─────────────────────────────────────────────────────

func TestCVObjectName_CapsLengthAndKeepsExtension(t *testing.T) {
	name := cvObjectName(strings.Repeat("a", 300) + ".docx")
	if len(name) != maxCVFileNameLength {
		t.Fatalf("len = %d, want %d", len(name), maxCVFileNameLength)
	}
	if !strings.HasSuffix(name, ".docx") {
		t.Fatalf("extension lost: %q", name)
	}
	if got := cvObjectName("My CV.pdf"); got != "my-cv.pdf" {
		t.Fatalf("short name changed: %q", got)
	}
}

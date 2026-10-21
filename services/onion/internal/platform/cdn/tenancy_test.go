package cdn

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/voxire/onion/internal/platform/utils/jwtutils"
	"github.com/voxire/onion/internal/testutil"
)

// cdn_files carries no tenant of its own, so these tests seed one row per
// business directly. The two ids are deliberately unrelated so a missing
// filter shows up as a cross-tenant hit rather than as a self-read.
const (
	tenantA int64 = 700001
	tenantB int64 = 700002
)

// seedFile inserts a cdn_files row owned by businessID and removes it
// afterwards. businessID 0 models a row that predates the tenant column.
func seedFile(tb testing.TB, db *sql.DB, businessID int64, public bool) string {
	tb.Helper()

	fileUUID := uuid.NewString()
	if _, err := db.ExecContext(context.Background(),
		`INSERT INTO cdn_files (file_uuid, business_id, is_public, storage_key, mime_type, size, original_filename)
		 VALUES (?, ?, ?, ?, 'image/png', 3, 'x.png')`,
		fileUUID, businessID, public, "originals/"+fileUUID+".png",
	); err != nil {
		tb.Fatalf("seed cdn_files row: %v", err)
	}
	tb.Cleanup(func() {
		_, _ = db.ExecContext(context.Background(), `DELETE FROM cdn_files WHERE file_uuid = ?`, fileUUID)
	})
	return fileUUID
}

// seedDerivative records one cdn_transformations row for a file, so a denied
// purge can be shown to have destroyed nothing.
func seedDerivative(tb testing.TB, db *sql.DB, fileUUID string) {
	tb.Helper()
	_, err := db.ExecContext(context.Background(),
		`INSERT IGNORE INTO cdn_transformations (file_id, cache_key, width, height, format, quality, fit, storage_key)
		 SELECT id, ?, 10, 10, 'webp', 80, 'cover', ? FROM cdn_files WHERE file_uuid = ?`,
		uuid.NewString(), "transformed/"+fileUUID+"/deadbeef.webp", fileUUID)
	if err != nil {
		tb.Fatalf("seed cdn_transformations row: %v", err)
	}
}

func countDerivatives(tb testing.TB, db *sql.DB, fileUUID string) int64 {
	tb.Helper()
	var n int64
	err := db.QueryRowContext(context.Background(),
		`SELECT COUNT(*) FROM cdn_transformations t
		   JOIN cdn_files f ON f.id = t.file_id
		  WHERE f.file_uuid = ?`, fileUUID).Scan(&n)
	if err != nil {
		tb.Fatalf("count derivatives: %v", err)
	}
	return n
}

func fileIDOf(tb testing.TB, db *sql.DB, fileUUID string) int64 {
	tb.Helper()
	var id int64
	if err := db.QueryRowContext(context.Background(),
		`SELECT id FROM cdn_files WHERE file_uuid = ?`, fileUUID).Scan(&id); err != nil {
		tb.Fatalf("read file id: %v", err)
	}
	return id
}

// TestReadScopeIsTheTenantBoundary pins the delivery rule itself: the
// predicate is the whole authorization, so every combination of caller and
// file ownership gets a case. A row that belongs to no business is never
// anonymously readable, which is what keeps unattributable legacy rows off the
// public surface.
func TestReadScopeIsTheTenantBoundary(t *testing.T) {
	db := testutil.OpenDB(t)
	c := &CDN{db: db, baseURL: "https://cdn.voxire.com"}
	ctx := context.Background()

	aPrivate := seedFile(t, db, tenantA, false)
	aPublic := seedFile(t, db, tenantA, true)
	bPrivate := seedFile(t, db, tenantB, false)
	bPublic := seedFile(t, db, tenantB, true)
	orphan := seedFile(t, db, 0, false)

	cases := []struct {
		name      string
		scope     readScope
		fileUUID  string
		wantFound bool
	}{
		{"anonymous cannot read a private file", anonymousRead, aPrivate, false},
		{"anonymous cannot read another tenant's private file", anonymousRead, bPrivate, false},
		{"anonymous cannot read an unattributable file", anonymousRead, orphan, false},
		{"anonymous reads a public file", anonymousRead, aPublic, true},
		{"anonymous reads another tenant's public file", anonymousRead, bPublic, true},
		{"tenant reads its own private file", tenantRead(tenantA), aPrivate, true},
		{"tenant cannot read another tenant's private file", tenantRead(tenantA), bPrivate, false},
		{"tenant reads its own public file", tenantRead(tenantA), aPublic, true},
		{"tenant cannot read an unattributable file", tenantRead(tenantA), orphan, false},
		{"a validated signature reads a private file", signedRead(), aPrivate, true},
		{"a validated signature reads an unattributable file", signedRead(), orphan, true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := c.fileForRead(ctx, tc.scope, tc.fileUUID)
			if tc.wantFound {
				if err != nil {
					t.Fatalf("expected the file to be readable, got %v", err)
				}
				return
			}
			if !errors.Is(err, sql.ErrNoRows) {
				t.Fatalf("expected sql.ErrNoRows for an out-of-scope file, got %v", err)
			}
			if got != nil {
				t.Fatal("a denied read must not return a row")
			}
		})
	}
}

// TestServeFileDeniesCrossTenantAndPrivateReads drives the delivery handler so
// the status a caller actually observes is covered. A request that gets past
// the lookup falls through to the nil storage client, which panics, so the
// cases are expressed as "was the row resolved" by watching for the 404 the
// denied path returns.
func TestServeFileDeniesCrossTenantAndPrivateReads(t *testing.T) {
	db := testutil.OpenDB(t)
	c := &CDN{db: db, baseURL: "https://cdn.voxire.com"}

	other := seedFile(t, db, tenantB, false)
	mine := seedFile(t, db, tenantA, false)

	denied := []struct {
		name     string
		scope    readScope
		fileUUID string
	}{
		{"anonymous private file", anonymousRead, mine},
		{"cross-tenant private file", tenantRead(tenantA), other},
	}
	for _, tc := range denied {
		t.Run(tc.name+" is 404", func(t *testing.T) {
			rec := httptest.NewRecorder()
			c.serveFile(rec, withFileID(httptest.NewRequest(http.MethodGet, "/files/x", nil), tc.fileUUID, tc.scope))
			if rec.Code != http.StatusNotFound {
				t.Fatalf("status = %d, want %d (body %q)", rec.Code, http.StatusNotFound, rec.Body.String())
			}
			if rec.Header().Get("ETag") != "" {
				t.Fatal("a denied read must not disclose the file ETag")
			}
		})
	}

	// The positive side is covered by TestReadScopeIsTheTenantBoundary, which
	// asserts the row resolves; reaching the storage client from here would
	// need a live object store.
	t.Run("own private file passes the lookup", func(t *testing.T) {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/files/x", nil)
		if _, err := c.fileForRead(req.Context(), tenantRead(tenantA), mine); err != nil {
			t.Fatalf("own private file should resolve: %v", err)
		}
		if rec.Code != http.StatusOK {
			t.Fatalf("sanity: unexpected recorder status %d", rec.Code)
		}
	})
}

// TestSignDeliveryRejectsCrossTenantFile covers the sub-finding that POST /sign
// minted a delivery credential for any tenant's file_uuid.
func TestSignDeliveryRejectsCrossTenantFile(t *testing.T) {
	db := testutil.OpenDB(t)
	c := &CDN{db: db, baseURL: "https://cdn.voxire.com", apiKey: "signing-secret"}

	other := seedFile(t, db, tenantB, false)
	mine := seedFile(t, db, tenantA, false)

	t.Run("another tenant's uuid yields no signature", func(t *testing.T) {
		rec := postSign(t, c, signDeliveryRequest{ID: other}, tenantA)
		if rec.Code != http.StatusNotFound {
			t.Fatalf("status = %d, want %d (body %q)", rec.Code, http.StatusNotFound, rec.Body.String())
		}
		if bytes.Contains(rec.Body.Bytes(), []byte("sig")) {
			t.Fatalf("response leaked a signature: %q", rec.Body.String())
		}
	})

	t.Run("another tenant's canonical url form is refused", func(t *testing.T) {
		rec := postSign(t, c, signDeliveryRequest{URL: "https://cdn.voxire.com/files/" + other}, tenantA)
		if rec.Code != http.StatusNotFound {
			t.Fatalf("status = %d, want %d (body %q)", rec.Code, http.StatusNotFound, rec.Body.String())
		}
	})

	t.Run("another tenant's storage key form is refused", func(t *testing.T) {
		rec := postSign(t, c, signDeliveryRequest{URL: "https://cdn.voxire.com/upload?url=/originals/" + other + ".png"}, tenantA)
		if rec.Code != http.StatusNotFound {
			t.Fatalf("status = %d, want %d (body %q)", rec.Code, http.StatusNotFound, rec.Body.String())
		}
	})

	t.Run("the caller's own uuid is signed", func(t *testing.T) {
		rec := postSign(t, c, signDeliveryRequest{ID: mine}, tenantA)
		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, want 200 (body %q)", rec.Code, rec.Body.String())
		}
		var resp signDeliveryResponse
		if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
			t.Fatalf("decode: %v", err)
		}
		if resp.Signature == "" || resp.ExpiresAt <= time.Now().Unix() {
			t.Fatalf("expected a usable signed url, got %+v", resp)
		}
	})

	t.Run("no principal at all is refused", func(t *testing.T) {
		body, _ := json.Marshal(signDeliveryRequest{ID: mine})
		req := httptest.NewRequest(http.MethodPost, "/sign", bytes.NewReader(body))
		rec := httptest.NewRecorder()
		c.signDelivery(rec, req)
		if rec.Code != http.StatusForbidden {
			t.Fatalf("status = %d, want %d", rec.Code, http.StatusForbidden)
		}
	})
}

// TestPurgeRejectsCrossTenantFile covers the destructive sub-finding: a purge
// aimed at another tenant's uuid must delete nothing at all.
func TestPurgeRejectsCrossTenantFile(t *testing.T) {
	db := testutil.OpenDB(t)
	c := &CDN{db: db, baseURL: "https://cdn.voxire.com"}
	router := http.NewServeMux()
	router.HandleFunc("PURGE /files/{id}", c.purgeFileVariants)

	other := seedFile(t, db, tenantB, false)
	seedDerivative(t, db, other)
	mine := seedFile(t, db, tenantA, false)
	seedDerivative(t, db, mine)

	t.Run("cross-tenant purge is refused and destroys nothing", func(t *testing.T) {
		req := httptest.NewRequest("PURGE", "/files/"+other, nil)
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, asTenant(req, tenantA))

		if rec.Code != http.StatusNotFound {
			t.Fatalf("status = %d, want %d (body %q)", rec.Code, http.StatusNotFound, rec.Body.String())
		}
		if got := countDerivatives(t, db, other); got != 1 {
			t.Fatalf("another tenant's derivative rows were destroyed: count = %d, want 1", got)
		}
	})

	t.Run("the same tenant still purges its own file", func(t *testing.T) {
		req := httptest.NewRequest("PURGE", "/files/"+mine, nil)
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, asTenant(req, tenantA))

		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, want 200 (body %q)", rec.Code, rec.Body.String())
		}
		if got := countDerivatives(t, db, mine); got != 0 {
			t.Fatalf("own derivative rows survived the purge: count = %d", got)
		}
	})

	t.Run("another tenant's public file is not purgable either", func(t *testing.T) {
		otherPublic := seedFile(t, db, tenantB, true)
		seedDerivative(t, db, otherPublic)

		req := httptest.NewRequest("PURGE", "/files/"+otherPublic, nil)
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, asTenant(req, tenantA))

		if rec.Code != http.StatusNotFound {
			t.Fatalf("status = %d, want %d", rec.Code, http.StatusNotFound)
		}
		if got := countDerivatives(t, db, otherPublic); got != 1 {
			t.Fatalf("a public asset's cache was evicted cross-tenant: count = %d", got)
		}
	})

	t.Run("the ?url= form cannot reach another tenant's original", func(t *testing.T) {
		target := "https://cdn.voxire.com/originals/" + other + ".png"
		req := httptest.NewRequest("PURGE", "/upload?url="+target, nil)
		rec := httptest.NewRecorder()

		c.purgeUploadURL(rec, asTenant(req, tenantA))

		if rec.Code != http.StatusNotFound {
			t.Fatalf("status = %d, want %d (body %q)", rec.Code, http.StatusNotFound, rec.Body.String())
		}
	})
}

// TestInsertFileScopesReadsByBusiness proves the write path stamps the
// resolved tenant and that fileByID is not a way around the filter.
func TestInsertFileScopesReadsByBusiness(t *testing.T) {
	db := testutil.OpenDB(t)
	c := &CDN{db: db, baseURL: "https://cdn.voxire.com"}
	ctx := context.Background()

	fileUUID := uuid.NewString()
	t.Cleanup(func() {
		_, _ = db.ExecContext(context.Background(), `DELETE FROM cdn_files WHERE file_uuid = ?`, fileUUID)
	})

	id, err := c.insertFile(ctx, tenantA, fileUUID, "originals/"+fileUUID+".png", "image/png", "x.png", 3, false)
	if err != nil {
		t.Fatalf("insertFile: %v", err)
	}
	if _, err := c.fileByID(ctx, tenantA, id); err != nil {
		t.Fatalf("the owning tenant should read its own row: %v", err)
	}
	if _, err := c.fileByID(ctx, tenantB, id); !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("another tenant read the row by numeric id: %v", err)
	}
	if _, err := c.fileByID(ctx, 0, id); !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("an unresolved caller read the row: %v", err)
	}
}

// TestInsertTransformationRequiresOwnership proves a derivative cannot be
// attached to a file the caller does not own.
func TestInsertTransformationRequiresOwnership(t *testing.T) {
	db := testutil.OpenDB(t)
	c := &CDN{db: db, baseURL: "https://cdn.voxire.com"}
	ctx := context.Background()

	other := seedFile(t, db, tenantB, false)
	own := seedFile(t, db, tenantA, false)
	p := Params{Width: 10, Height: 10, Format: "webp", Quality: 80, Fit: "cover"}

	if err := c.insertTransformation(ctx, tenantA, fileIDOf(t, db, other), "cross-tenant-key", "transformed/x.webp", p); err != nil {
		t.Fatalf("insertTransformation: %v", err)
	}
	if n := countDerivatives(t, db, other); n != 0 {
		t.Fatalf("a derivative was attached to another tenant's file: count = %d", n)
	}

	if err := c.insertTransformation(ctx, tenantA, fileIDOf(t, db, own), "own-key", "transformed/own.webp", p); err != nil {
		t.Fatalf("insertTransformation: %v", err)
	}
	if n := countDerivatives(t, db, own); n != 1 {
		t.Fatalf("own derivative was not recorded: count = %d", n)
	}
}

// TestTransformationCacheLookupIsTenantPinned proves the derivative cache is
// not reachable by cache key alone.
func TestTransformationCacheLookupIsTenantPinned(t *testing.T) {
	db := testutil.OpenDB(t)
	c := &CDN{db: db, baseURL: "https://cdn.voxire.com"}
	ctx := context.Background()

	other := seedFile(t, db, tenantB, false)
	seedDerivative(t, db, other)

	var cacheKey string
	if err := db.QueryRowContext(ctx,
		`SELECT t.cache_key FROM cdn_transformations t
		   JOIN cdn_files fi ON fi.id = t.file_id
		  WHERE fi.file_uuid = ?`, other).Scan(&cacheKey); err != nil {
		t.Fatalf("read cache key: %v", err)
	}

	if _, err := c.transformationByCacheKey(ctx, tenantA, cacheKey); !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("another tenant hit the derivative cache: %v", err)
	}
	if _, err := c.transformationByCacheKey(ctx, 0, cacheKey); !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("an unresolved caller hit the derivative cache: %v", err)
	}
	if _, err := c.transformationByCacheKey(ctx, tenantB, cacheKey); err != nil {
		t.Fatalf("the owning tenant missed its own derivative cache: %v", err)
	}
}

// TestAPIKeyAuthRejectsUnresolvableTenant is the sub-finding that any valid JWT
// was enough. A signature that does not name a reachable business must be a
// denial, not a pass-through carrying a zero scope.
func TestAPIKeyAuthRejectsUnresolvableTenant(t *testing.T) {
	db := testutil.OpenDB(t)
	t.Setenv("JWT_SECRET", "unit-test-secret")

	// next answers 418 when the middleware passed a real principal through.
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		p, ok := principalFrom(r.Context())
		if ok && p.businessID > 0 {
			w.WriteHeader(http.StatusTeapot)
			return
		}
		w.WriteHeader(http.StatusUnauthorized)
	})

	t.Run("a token naming an unreachable business is refused", func(t *testing.T) {
		c := &CDN{db: db, apiKey: "s3cret"}
		// Seed a real business so the denial is about the token's id, not an
		// empty database.
		testutil.EnsureBusiness(t, db, 710001, 710001)

		req := httptest.NewRequest(http.MethodPost, "/upload", nil)
		req.Header.Set("Authorization", "Bearer "+testJWT(t, "unit-test-secret", 710099, 710099))
		rec := httptest.NewRecorder()

		c.apiKeyAuth(next).ServeHTTP(rec, req)

		if rec.Code != http.StatusForbidden {
			t.Fatalf("status = %d, want %d (body %q)", rec.Code, http.StatusForbidden, rec.Body.String())
		}
	})

	t.Run("a token resolving to a real business passes with that scope", func(t *testing.T) {
		c := &CDN{db: db, apiKey: "s3cret"}
		testutil.GrantBusinessAccess(t, db, 710002, 710002)

		req := httptest.NewRequest(http.MethodPost, "/upload", nil)
		req.Header.Set("Authorization", "Bearer "+testJWT(t, "unit-test-secret", 710002, 710002))
		rec := httptest.NewRecorder()

		c.apiKeyAuth(next).ServeHTTP(rec, req)

		if rec.Code != http.StatusTeapot {
			t.Fatalf("status = %d, want the middleware to pass the request through", rec.Code)
		}
	})
}

// TestStorageKeyFileUUID covers the mapping that lets a purge or sign target
// supplied as a storage path be held to the same ownership check as
// /files/{uuid}.
func TestStorageKeyFileUUID(t *testing.T) {
	id := "4c85c397-9fc3-4bd8-adf5-9d1ef64f5e8b"
	cases := []struct {
		key  string
		want string
		ok   bool
	}{
		{"originals/" + id + ".png", id, true},
		{"transformed/" + id + "/abc123.webp", id, true},
		{"/originals/" + id + ".jpg", id, true},
		{"upload-url-transformed/abc/originals/" + id + ".png", "", false},
		{"somewhere/else.png", "", false},
		{"originals/", "", false},
	}
	for _, tc := range cases {
		got, ok := storageKeyFileUUID(tc.key)
		if got != tc.want || ok != tc.ok {
			t.Errorf("storageKeyFileUUID(%q) = (%q,%v), want (%q,%v)", tc.key, got, ok, tc.want, tc.ok)
		}
	}
}

func postSign(t *testing.T, c *CDN, req signDeliveryRequest, businessID int64) *httptest.ResponseRecorder {
	t.Helper()
	body, err := json.Marshal(req)
	if err != nil {
		t.Fatal(err)
	}
	r := httptest.NewRequest(http.MethodPost, "/sign", bytes.NewReader(body))
	rec := httptest.NewRecorder()
	c.signDelivery(rec, asTenant(r, businessID))
	return rec
}

// withFileID sets the path value and delivery scope a router would normally
// have set before calling a delivery handler directly.
func withFileID(r *http.Request, id string, scope readScope) *http.Request {
	r = r.WithContext(withDeliveryScope(r.Context(), scope))
	r.SetPathValue("id", id)
	return r
}

func testJWT(t *testing.T, secret string, userID, businessID int64) string {
	t.Helper()
	claims := &uploadJWTClaims{UserID: userID, BusinessID: businessID}
	token, err := jwtutils.NewJWTHelperWithDuration(secret, "onion-platform", time.Hour).CreateJWT(claims)
	if err != nil {
		t.Fatalf("CreateJWT: %v", err)
	}
	return token
}

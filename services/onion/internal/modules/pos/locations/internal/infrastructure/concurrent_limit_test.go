package infrastructure_test

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"testing"

	"github.com/voxire/onion/internal/modules/pos/locations/internal/domain"
	"github.com/voxire/onion/internal/modules/pos/locations/internal/infrastructure"
	"github.com/voxire/onion/internal/testutil"
)

func TestConcurrentLocationsRespectPlanLimit(t *testing.T) {
	db := testutil.OpenDB(t)
	const businessID = int64(9956001)
	t.Cleanup(func() {
		_, _ = db.Exec("DELETE FROM pos_locations WHERE business_id = ?", businessID)
		_, _ = db.Exec("DELETE FROM pos_location_creation_locks WHERE business_id = ?", businessID)
	})
	repo := infrastructure.NewRepository(db).WithLocationLimitReader(func(context.Context, int64) (int64, bool, error) { return 3, true, nil })
	var wait sync.WaitGroup
	results := make(chan error, 8)
	for i := 0; i < 8; i++ {
		wait.Add(1)
		go func(i int) {
			defer wait.Done()
			_, err := repo.Create(context.Background(), domain.LocationInput{BusinessID: businessID, OwnerID: testutil.OwnerID, Name: fmt.Sprintf("Synthetic location %d", i), IsActive: true})
			results <- err
		}(i)
	}
	wait.Wait()
	close(results)
	var created, limited int
	for err := range results {
		if err == nil {
			created++
		} else if errors.Is(err, domain.ErrLocationLimitExceeded) {
			limited++
		} else {
			t.Fatalf("unexpected create failure: %v", err)
		}
	}
	if created != 3 || limited != 5 {
		t.Fatalf("created=%d limited=%d, want 3 and 5", created, limited)
	}
}

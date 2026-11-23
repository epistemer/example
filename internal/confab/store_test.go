package confab_test

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/epistemer/errors"
	"github.com/shopspring/decimal"

	"github.com/epistemer/example/internal/confab"
	"github.com/epistemer/example/internal/ids"
	"github.com/epistemer/example/internal/models"
)

// open connects to the database CONFAB_DB_URL names, or skips.
func open(t *testing.T) *confab.Store {
	t.Helper()
	url := os.Getenv("CONFAB_DB_URL")
	if url == "" {
		t.Skip("CONFAB_DB_URL isn't set: run `make db migrate` and `make test`")
	}
	s, err := confab.Open(t.Context(), url)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(s.Close)
	return s
}

func TestUpdateEventVersion(t *testing.T) {
	s := open(t)
	ctx := context.Background()
	slug := ids.Slug("test-" + time.Now().Format("150405.000000000")[7:])
	in := confab.EventInput{
		Name: "Versioned", Slug: slug, Status: models.EventStatusDraft, Online: true,
		StartsOn: time.Date(2030, 1, 1, 0, 0, 0, 0, time.UTC), EndsOn: time.Date(2030, 1, 2, 0, 0, 0, 0, time.UTC),
		Capacity: 10, Currency: "EUR", Tiers: []confab.Tier{{Name: "Free", Price: decimal.Zero, Quantity: 10}},
	}
	id, err := s.CreateEvent(ctx, in)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = s.DeleteEvent(context.Background(), int64(id)) })
	if err := s.UpdateEvent(ctx, id, 1, in); err != nil {
		t.Fatalf("first update: %v", err)
	}
	if err := s.UpdateEvent(ctx, id, 1, in); !errors.Is(err, confab.ErrStale) {
		t.Fatalf("second update with the old version: %v, want ErrStale", err)
	}
	if err := s.UpdateEvent(ctx, id, 2, in); err != nil {
		t.Fatalf("update with the new version: %v", err)
	}
}

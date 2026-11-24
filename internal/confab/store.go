// Package confab is the back office's domain: events, their ticket
// tiers, speakers and talks, the rules between them, and where files go.
// Handlers bind forms into its inputs; it answers with errors people can
// read (errors.go).
package confab

import (
	"context"
	"fmt"
	"time"

	"github.com/epistemer/errors"
	"github.com/epistemer/models/listing"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shopspring/decimal"

	"github.com/epistemer/example/internal/ids"
	"github.com/epistemer/example/internal/models"
)

// Q is everything sqlc and models generate on the queries.
type Q interface {
	models.Querier
	models.Listings
}

// Store is the database. Its embedded Queries answer the generated
// queries and listings; its own methods keep the rules.
type Store struct {
	*models.Queries
	pool *pgxpool.Pool
}

func Open(ctx context.Context, url string) (*Store, error) {
	cfg, err := pgxpool.ParseConfig(url)
	if err != nil {
		return nil, fmt.Errorf("database url: %w", err)
	}
	// pgx encodes an array of an enum only once it knows the types
	// (findings F15): talk.topics is talk_topic[].
	cfg.AfterConnect = models.RegisterTypes // enums and enum arrays
	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("reach database: %w", err)
	}
	return &Store{Queries: models.New(pool), pool: pool}, nil
}

func (s *Store) Close() { s.pool.Close() }

// Exec runs SQL as it is, for migrations.
func (s *Store) Exec(ctx context.Context, sql string) error {
	_, err := s.pool.Exec(ctx, sql)
	return err
}

func (s *Store) tx(ctx context.Context, fn func(q *models.Queries) error) error {
	return pgx.BeginFunc(ctx, s.pool, func(tx pgx.Tx) error { return fn(s.Queries.WithTx(tx)) })
}

// unique reports whether err is a unique violation of constraint.
func unique(err error, constraint string) bool {
	var pe *pgconn.PgError
	return errors.As(err, &pe) && pe.Code == "23505" && pe.ConstraintName == constraint
}

func missing(err error) error {
	if errors.Is(err, pgx.ErrNoRows) {
		return ErrNotFound
	}
	return err
}

// Events

type Venue struct {
	Name, Street, City, Country, PostCode string
	Latitude, Longitude                   *float64
}

type Tier struct {
	Name       string
	Price      decimal.Decimal
	Quantity   int
	SaleEndsOn *time.Time
}

type EventInput struct {
	Name     string
	Slug     ids.Slug
	Summary  string
	Status   models.EventStatus
	StartsOn time.Time
	EndsOn   time.Time
	Online   bool
	Capacity int32
	Venue    *Venue // nil when online
	Currency string
	Tiers    []Tier
}

func (in EventInput) check() error {
	if in.EndsOn.Before(in.StartsOn) {
		return ErrEndsFirst
	}
	if !in.Online && in.Venue == nil {
		return ErrNeedsVenue
	}
	return nil
}

func (in EventInput) venue() (n, st, c, co, pc *string, lat, lng *float64) {
	if in.Venue == nil {
		return
	}
	v := in.Venue
	opt := func(s string) *string {
		if s == "" {
			return nil
		}
		return &s
	}
	return opt(v.Name), opt(v.Street), opt(v.City), opt(v.Country), opt(v.PostCode), v.Latitude, v.Longitude
}

func (s *Store) CreateEvent(ctx context.Context, in EventInput) (ids.EventID, error) {
	if err := in.check(); err != nil {
		return 0, err
	}
	var id int64
	err := s.tx(ctx, func(q *models.Queries) error {
		n, st, c, co, pc, lat, lng := in.venue()
		var err error
		id, err = q.CreateEvent(ctx, models.CreateEventParams{
			Name: in.Name, Slug: string(in.Slug), Summary: in.Summary, Status: in.Status,
			StartsOn: in.StartsOn, EndsOn: in.EndsOn, Online: in.Online,
			Capacity: in.Capacity, VenueName: n, VenueStreet: st,
			VenueCity: c, VenueCountry: co, VenuePostCode: pc,
			Latitude: lat, Longitude: lng, Currency: in.Currency,
		})
		if err != nil {
			return err
		}
		return saveTiers(ctx, q, id, in.Tiers)
	})
	if unique(err, "event_slug_key") {
		return 0, ErrSlugTaken.With(string(in.Slug))
	}
	if err != nil {
		return 0, ErrNotSaved.Wrap(err)
	}
	return ids.EventID(id), nil
}

// UpdateEvent saves an edit made to version; ErrStale when someone saved
// another edit first.
func (s *Store) UpdateEvent(ctx context.Context, id ids.EventID, version int32, in EventInput) error {
	if err := in.check(); err != nil {
		return err
	}
	err := s.tx(ctx, func(q *models.Queries) error {
		n, st, c, co, pc, lat, lng := in.venue()
		rows, err := q.UpdateEvent(ctx, models.UpdateEventParams{
			ID: int64(id), Version: version,
			Name: in.Name, Slug: string(in.Slug), Summary: in.Summary, Status: in.Status,
			StartsOn: in.StartsOn, EndsOn: in.EndsOn, Online: in.Online,
			Capacity: in.Capacity, VenueName: n, VenueStreet: st,
			VenueCity: c, VenueCountry: co, VenuePostCode: pc,
			Latitude: lat, Longitude: lng, Currency: in.Currency,
		})
		if err != nil {
			return err
		}
		if rows == 0 {
			return ErrStale
		}
		if err := q.DeleteTicketTiers(ctx, int64(id)); err != nil {
			return err
		}
		return saveTiers(ctx, q, int64(id), in.Tiers)
	})
	switch {
	case unique(err, "event_slug_key"):
		return ErrSlugTaken.With(string(in.Slug))
	case errors.Is(err, ErrStale), err == nil:
		return err
	}
	return ErrNotSaved.Wrap(err)
}

func saveTiers(ctx context.Context, q *models.Queries, event int64, tiers []Tier) error {
	for i, t := range tiers {
		err := q.CreateTicketTier(ctx, models.CreateTicketTierParams{
			EventID: event, Position: int32(i), Name: t.Name, Price: t.Price,
			Quantity: int32(t.Quantity), SaleEndsOn: t.SaleEndsOn,
		})
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *Store) Event(ctx context.Context, id ids.EventID) (models.Event, error) {
	e, err := s.GetEvent(ctx, int64(id))
	return e, missing(err)
}

func (s *Store) EventBySlug(ctx context.Context, slug ids.Slug) (models.Event, error) {
	e, err := s.GetEventBySlug(ctx, string(slug))
	return e, missing(err)
}

func (s *Store) Tiers(ctx context.Context, id ids.EventID) ([]Tier, error) {
	rows, err := s.ListTicketTiers(ctx, int64(id))
	if err != nil {
		return nil, err
	}
	out := make([]Tier, len(rows))
	for i, r := range rows {
		out[i] = Tier{Name: r.Name, Price: r.Price, Quantity: int(r.Quantity), SaleEndsOn: r.SaleEndsOn}
	}
	return out, nil
}

// Speakers

type SpeakerInput struct {
	Name, Email, Country, Bio string
	Phone, Website            *string
	Photo                     *StoredFile // nil: none
}

type StoredFile struct{ Key, Name string }

func (s *Store) CreateSpeaker(ctx context.Context, in SpeakerInput) (ids.SpeakerID, error) {
	key, name := in.photo()
	id, err := s.Queries.CreateSpeaker(ctx, models.CreateSpeakerParams{
		Name: in.Name, Email: in.Email, Phone: in.Phone, Country: in.Country,
		Website: in.Website, Bio: in.Bio, PhotoKey: key, PhotoName: name,
	})
	if unique(err, "speaker_email_key") {
		return 0, ErrEmailTaken
	}
	if err != nil {
		return 0, ErrNotSaved.Wrap(err)
	}
	return ids.SpeakerID(id), nil
}

func (s *Store) UpdateSpeaker(ctx context.Context, id ids.SpeakerID, in SpeakerInput) error {
	key, name := in.photo()
	err := s.Queries.UpdateSpeaker(ctx, models.UpdateSpeakerParams{
		ID: int64(id), Name: in.Name, Email: in.Email, Phone: in.Phone, Country: in.Country,
		Website: in.Website, Bio: in.Bio, PhotoKey: key, PhotoName: name,
	})
	if unique(err, "speaker_email_key") {
		return ErrEmailTaken
	}
	if err != nil {
		return ErrNotSaved.Wrap(err)
	}
	return nil
}

func (in SpeakerInput) photo() (key, name *string) {
	if in.Photo == nil {
		return nil, nil
	}
	return &in.Photo.Key, &in.Photo.Name
}

func (s *Store) Speaker(ctx context.Context, id ids.SpeakerID) (models.Speaker, error) {
	sp, err := s.GetSpeaker(ctx, int64(id))
	return sp, missing(err)
}

func (s *Store) DeleteSpeaker(ctx context.Context, id ids.SpeakerID) error {
	_, total, err := s.ListTalkView(ctx, listing.Query{
		Filters: []listing.Filter{{Column: models.TalkViewColumns.SpeakerID.Name, Op: listing.Eq, Value: int64(id)}},
		Limit:   1,
	})
	if err != nil {
		return err
	}
	if total > 0 {
		return ErrHasTalks.With(int(total))
	}
	return s.Queries.DeleteSpeaker(ctx, int64(id))
}

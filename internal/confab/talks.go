package confab

import (
	"context"
	"time"

	"github.com/epistemer/errors"
	"github.com/epistemer/i18n"
	"github.com/epistemer/models/listing"
	"github.com/jackc/pgx/v5"

	"github.com/epistemer/example/internal/ids"
	"github.com/epistemer/example/internal/models"
	"github.com/epistemer/example/internal/pg"
)

// Rules are the call for papers' settings the talk rules read: the
// generated settings.CFP satisfies it.
type Rules interface {
	CFPClosed() bool
	CFPMaxTalks() int
}

type TalkInput struct {
	Event    ids.EventID
	Speaker  ids.SpeakerID
	Title    string
	Abstract string
	Level    models.TalkLevel
	Format   models.TalkFormat
	Topics   []models.TalkTopic
	StartsAt *time.Time
	Duration time.Duration
	Seats    *int  // workshops only
	Recorded *bool // nil: not decided yet
}

// CheckTalk runs the rules between a talk, its event and its speaker's
// other talks. talk is 0 for a new one.
func (s *Store) CheckTalk(ctx context.Context, rules Rules, talk ids.TalkID, in TalkInput) error {
	if talk == 0 && rules.CFPClosed() {
		return ErrCFPClosed
	}
	e, err := s.Event(ctx, in.Event)
	if err != nil {
		return err
	}
	_, n, err := s.ListTalkView(ctx, listing.Query{Filters: []listing.Filter{
		{Column: models.TalkViewColumns.EventID.Name, Op: listing.Eq, Value: int64(in.Event)},
		{Column: models.TalkViewColumns.SpeakerID.Name, Op: listing.Eq, Value: int64(in.Speaker)},
		{Column: models.TalkViewColumns.ID.Name, Op: listing.Ne, Value: int64(talk)},
	}, Limit: 1})
	if err != nil {
		return err
	}
	if max := rules.CFPMaxTalks(); int(n) >= max {
		return ErrTooManyTalks.With(max)
	}
	if in.StartsAt == nil {
		return nil
	}
	first, last := e.StartsOn.Time, e.EndsOn.Time.AddDate(0, 0, 1)
	if in.StartsAt.Before(first) || !in.StartsAt.Before(last) {
		return ErrOutsideEvent.With(i18n.StaticDate(first), i18n.StaticDate(e.EndsOn.Time))
	}
	o, err := s.OverlappingTalk(ctx, models.OverlappingTalkParams{
		EventID: int64(in.Event), SpeakerID: int64(in.Speaker), ID: int64(talk),
		StartsAt: pg.Time(in.StartsAt), EndsAt: pg.Time(new(in.StartsAt.Add(in.Duration))),
	})
	switch {
	case err == nil:
		return ErrOverlap.With(o.Title)
	case errors.Is(err, pgx.ErrNoRows):
		return nil
	}
	return err
}

func (in TalkInput) params() models.CreateTalkParams {
	return models.CreateTalkParams{
		EventID: int64(in.Event), SpeakerID: int64(in.Speaker), Title: in.Title, Abstract: in.Abstract,
		Level: in.Level, Format: in.Format, Topics: in.Topics, StartsAt: pg.Time(in.StartsAt),
		DurationMinutes: int32(in.Duration / time.Minute), Seats: pg.Int4(in.Seats), Recorded: pg.Bool(in.Recorded),
	}
}

// SaveTalk creates (talk 0) or updates a talk, with its slides: kept
// lists the stored keys to keep, added the new files, already stored.
func (s *Store) SaveTalk(ctx context.Context, rules Rules, talk ids.TalkID, in TalkInput, kept []string, added []File) (ids.TalkID, []string, error) {
	if err := s.CheckTalk(ctx, rules, talk, in); err != nil {
		return 0, nil, err
	}
	var dropped []string
	err := s.tx(ctx, func(q *models.Queries) error {
		p := in.params()
		if talk == 0 {
			id, err := q.CreateTalk(ctx, p)
			if err != nil {
				return err
			}
			talk = ids.TalkID(id)
		} else {
			err := q.UpdateTalk(ctx, models.UpdateTalkParams{
				ID: int64(talk), EventID: p.EventID, SpeakerID: p.SpeakerID, Title: p.Title,
				Abstract: p.Abstract, Level: p.Level, Format: p.Format, Topics: p.Topics,
				StartsAt: p.StartsAt, DurationMinutes: p.DurationMinutes, Seats: p.Seats, Recorded: p.Recorded,
			})
			if err != nil {
				return err
			}
		}
		stored, err := q.ListTalkFiles(ctx, int64(talk))
		if err != nil {
			return err
		}
		keep := map[string]bool{}
		for _, k := range kept {
			keep[k] = true
		}
		for _, f := range stored {
			if !keep[f.Key] {
				if err := q.DeleteTalkFile(ctx, models.DeleteTalkFileParams{TalkID: int64(talk), Key: f.Key}); err != nil {
					return err
				}
				dropped = append(dropped, f.Key)
			}
		}
		for _, f := range added {
			err := q.CreateTalkFile(ctx, models.CreateTalkFileParams{
				TalkID: int64(talk), Key: f.Key, Name: f.Name, ContentType: f.ContentType, Size: f.Size,
			})
			if err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return 0, nil, ErrNotSaved.Wrap(err)
	}
	return talk, dropped, nil
}

func (s *Store) Talk(ctx context.Context, id ids.TalkID) (models.Talk, error) {
	t, err := s.GetTalk(ctx, int64(id))
	return t, missing(err)
}

func (s *Store) TalkView(ctx context.Context, id ids.TalkID) (models.TalkView, error) {
	t, err := s.GetTalkView(ctx, int64(id))
	return t, missing(err)
}

// Schedule moves an accepted talk with a start time to scheduled.
func CanSchedule(t models.TalkView) error {
	if t.State != models.TalkStateAccepted || !t.StartsAt.Valid {
		return ErrNotScheduled
	}
	return nil
}

func (s *Store) SetState(ctx context.Context, ids []int64, state models.TalkState) (int64, error) {
	return s.SetTalksState(ctx, models.SetTalksStateParams{State: state, Ids: ids})
}

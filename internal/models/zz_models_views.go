// Code generated from sqlc's request by github.com/epistemer/models. DO NOT EDIT.

package models

import (
	"context"

	"github.com/epistemer/models/column"
	"github.com/epistemer/models/listing"
)

var _ column.Kind // the import stays when no view has a column of a kind

// Listings is the listing methods of Queries, one per view.
type Listings interface {
	ListEventView(ctx context.Context, src listing.Source) ([]EventView, int64, error)
	ListSpeakerView(ctx context.Context, src listing.Source) ([]SpeakerView, int64, error)
	ListTalkView(ctx context.Context, src listing.Source) ([]TalkView, int64, error)
}

var _ Listings = (*Queries)(nil)

// ListEventView returns the rows of event_view
// that src selects, and how many rows its search and filters match.
func (q *Queries) ListEventView(ctx context.Context, src listing.Source) ([]EventView, int64, error) {
	lq := src.ListingQuery()
	rowsSQL, countSQL, args, err := listing.Build(lq, "event_view", []string{
		"id",
		"name",
		"slug",
		"status",
		"starts_on",
		"ends_on",
		"online",
		"capacity",
		"venue_city",
		"currency",
		"talks",
		"from_price",
		"created_at",
	})
	if err != nil {
		return nil, 0, err
	}
	rows, err := q.db.Query(ctx, rowsSQL, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var items []EventView
	for rows.Next() {
		var i EventView
		if err := rows.Scan(
			&i.ID,
			&i.Name,
			&i.Slug,
			&i.Status,
			&i.StartsOn,
			&i.EndsOn,
			&i.Online,
			&i.Capacity,
			&i.VenueCity,
			&i.Currency,
			&i.Talks,
			&i.FromPrice,
			&i.CreatedAt,
		); err != nil {
			return nil, 0, err
		}
		items = append(items, i)
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}
	// The rows tell the total when they end the listing.
	n := len(items)
	if (lq.Limit == 0 || n < lq.Limit) && (lq.Offset == 0 || n > 0) {
		return items, int64(lq.Offset + n), nil
	}
	var total int64
	if err := q.db.QueryRow(ctx, countSQL, args...).Scan(&total); err != nil {
		return nil, 0, err
	}
	if n > 0 && total < int64(lq.Offset+n) {
		total = int64(lq.Offset + n)
	}
	return items, total, nil
}

// ListSpeakerView returns the rows of speaker_view
// that src selects, and how many rows its search and filters match.
func (q *Queries) ListSpeakerView(ctx context.Context, src listing.Source) ([]SpeakerView, int64, error) {
	lq := src.ListingQuery()
	rowsSQL, countSQL, args, err := listing.Build(lq, "speaker_view", []string{
		"id",
		"name",
		"email",
		"country",
		"website",
		"talks",
		"created_at",
	})
	if err != nil {
		return nil, 0, err
	}
	rows, err := q.db.Query(ctx, rowsSQL, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var items []SpeakerView
	for rows.Next() {
		var i SpeakerView
		if err := rows.Scan(
			&i.ID,
			&i.Name,
			&i.Email,
			&i.Country,
			&i.Website,
			&i.Talks,
			&i.CreatedAt,
		); err != nil {
			return nil, 0, err
		}
		items = append(items, i)
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}
	// The rows tell the total when they end the listing.
	n := len(items)
	if (lq.Limit == 0 || n < lq.Limit) && (lq.Offset == 0 || n > 0) {
		return items, int64(lq.Offset + n), nil
	}
	var total int64
	if err := q.db.QueryRow(ctx, countSQL, args...).Scan(&total); err != nil {
		return nil, 0, err
	}
	if n > 0 && total < int64(lq.Offset+n) {
		total = int64(lq.Offset + n)
	}
	return items, total, nil
}

// ListTalkView returns the rows of talk_view
// that src selects, and how many rows its search and filters match.
func (q *Queries) ListTalkView(ctx context.Context, src listing.Source) ([]TalkView, int64, error) {
	lq := src.ListingQuery()
	rowsSQL, countSQL, args, err := listing.Build(lq, "talk_view", []string{
		"id",
		"title",
		"level",
		"format",
		"state",
		"starts_at",
		"duration_minutes",
		"recorded",
		"speaker_id",
		"speaker_name",
		"event_id",
		"event_name",
		"created_at",
	})
	if err != nil {
		return nil, 0, err
	}
	rows, err := q.db.Query(ctx, rowsSQL, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var items []TalkView
	for rows.Next() {
		var i TalkView
		if err := rows.Scan(
			&i.ID,
			&i.Title,
			&i.Level,
			&i.Format,
			&i.State,
			&i.StartsAt,
			&i.DurationMinutes,
			&i.Recorded,
			&i.SpeakerID,
			&i.SpeakerName,
			&i.EventID,
			&i.EventName,
			&i.CreatedAt,
		); err != nil {
			return nil, 0, err
		}
		items = append(items, i)
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}
	// The rows tell the total when they end the listing.
	n := len(items)
	if (lq.Limit == 0 || n < lq.Limit) && (lq.Offset == 0 || n > 0) {
		return items, int64(lq.Offset + n), nil
	}
	var total int64
	if err := q.db.QueryRow(ctx, countSQL, args...).Scan(&total); err != nil {
		return nil, 0, err
	}
	if n > 0 && total < int64(lq.Offset+n) {
		total = int64(lq.Offset + n)
	}
	return items, total, nil
}

// EventViewColumns are event_view's columns, for a table package (A4).
var EventViewColumns = struct {
	ID        column.Column[EventView]
	Name      column.Column[EventView]
	Slug      column.Column[EventView]
	Status    column.Column[EventView]
	StartsOn  column.Column[EventView]
	EndsOn    column.Column[EventView]
	Online    column.Column[EventView]
	Capacity  column.Column[EventView]
	VenueCity column.Column[EventView]
	Currency  column.Column[EventView]
	Talks     column.Column[EventView]
	FromPrice column.Column[EventView]
	CreatedAt column.Column[EventView]
}{
	ID: column.Column[EventView]{
		Name:  "id",
		Label: L10nEventView_Id,
		Kind:  column.Integer,
		Get: func(r EventView) column.Value {
			return column.Value{Int: r.ID}
		},
	},
	Name: column.Column[EventView]{
		Name:  "name",
		Label: L10nEventView_Name,
		Kind:  column.Text,
		Get: func(r EventView) column.Value {
			return column.Value{Text: r.Name}
		},
	},
	Slug: column.Column[EventView]{
		Name:  "slug",
		Label: L10nEventView_Slug,
		Kind:  column.Text,
		Get: func(r EventView) column.Value {
			return column.Value{Text: r.Slug}
		},
	},
	Status: column.Column[EventView]{
		Name:    "status",
		Label:   L10nEventView_Status,
		Kind:    column.Enum,
		Options: AllEventStatusOptions(),
		Get: func(r EventView) column.Value {
			return column.Value{Enum: r.Status}
		},
	},
	StartsOn: column.Column[EventView]{
		Name:  "starts_on",
		Label: L10nEventView_StartsOn,
		Kind:  column.Date,
		Get: func(r EventView) column.Value {
			if !r.StartsOn.Valid {
				return column.Value{Null: true}
			}
			if m := r.StartsOn.InfinityModifier; m != 0 {
				return column.Infinite(int8(m))
			}
			return column.Value{Time: r.StartsOn.Time}
		},
	},
	EndsOn: column.Column[EventView]{
		Name:  "ends_on",
		Label: L10nEventView_EndsOn,
		Kind:  column.Date,
		Get: func(r EventView) column.Value {
			if !r.EndsOn.Valid {
				return column.Value{Null: true}
			}
			if m := r.EndsOn.InfinityModifier; m != 0 {
				return column.Infinite(int8(m))
			}
			return column.Value{Time: r.EndsOn.Time}
		},
	},
	Online: column.Column[EventView]{
		Name:  "online",
		Label: L10nEventView_Online,
		Kind:  column.Bool,
		Get: func(r EventView) column.Value {
			return column.Value{Bool: r.Online}
		},
	},
	Capacity: column.Column[EventView]{
		Name:  "capacity",
		Label: L10nEventView_Capacity,
		Kind:  column.Integer,
		Get: func(r EventView) column.Value {
			return column.Value{Int: int64(r.Capacity)}
		},
	},
	VenueCity: column.Column[EventView]{
		Name:  "venue_city",
		Label: L10nEventView_VenueCity,
		Kind:  column.Text,
		Get: func(r EventView) column.Value {
			if !r.VenueCity.Valid {
				return column.Value{Null: true}
			}
			return column.Value{Text: r.VenueCity.String}
		},
	},
	Currency: column.Column[EventView]{
		Name:  "currency",
		Label: L10nEventView_Currency,
		Kind:  column.Text,
		Get: func(r EventView) column.Value {
			return column.Value{Text: r.Currency}
		},
	},
	Talks: column.Column[EventView]{
		Name:  "talks",
		Label: L10nEventView_Talks,
		Kind:  column.Integer,
		Get: func(r EventView) column.Value {
			return column.Value{Int: int64(r.Talks)}
		},
	},
	FromPrice: column.Column[EventView]{
		Name:  "from_price",
		Label: L10nEventView_FromPrice,
		Kind:  column.Number,
		Get: func(r EventView) column.Value {
			if !r.FromPrice.Valid {
				return column.Value{Null: true}
			}
			v, _ := r.FromPrice.Value()
			digits, _ := v.(string)
			return column.Numeric(digits)
		},
	},
	CreatedAt: column.Column[EventView]{
		Name:  "created_at",
		Label: L10nEventView_CreatedAt,
		Kind:  column.Time,
		Get: func(r EventView) column.Value {
			if !r.CreatedAt.Valid {
				return column.Value{Null: true}
			}
			if m := r.CreatedAt.InfinityModifier; m != 0 {
				return column.Infinite(int8(m))
			}
			return column.Value{Time: r.CreatedAt.Time}
		},
	},
}

// SpeakerViewColumns are speaker_view's columns, for a table package (A4).
var SpeakerViewColumns = struct {
	ID        column.Column[SpeakerView]
	Name      column.Column[SpeakerView]
	Email     column.Column[SpeakerView]
	Country   column.Column[SpeakerView]
	Website   column.Column[SpeakerView]
	Talks     column.Column[SpeakerView]
	CreatedAt column.Column[SpeakerView]
}{
	ID: column.Column[SpeakerView]{
		Name:  "id",
		Label: L10nSpeakerView_Id,
		Kind:  column.Integer,
		Get: func(r SpeakerView) column.Value {
			return column.Value{Int: r.ID}
		},
	},
	Name: column.Column[SpeakerView]{
		Name:  "name",
		Label: L10nSpeakerView_Name,
		Kind:  column.Text,
		Get: func(r SpeakerView) column.Value {
			return column.Value{Text: r.Name}
		},
	},
	Email: column.Column[SpeakerView]{
		Name:  "email",
		Label: L10nSpeakerView_Email,
		Kind:  column.Text,
		Get: func(r SpeakerView) column.Value {
			return column.Value{Text: r.Email}
		},
	},
	Country: column.Column[SpeakerView]{
		Name:  "country",
		Label: L10nSpeakerView_Country,
		Kind:  column.Text,
		Get: func(r SpeakerView) column.Value {
			return column.Value{Text: r.Country}
		},
	},
	Website: column.Column[SpeakerView]{
		Name:  "website",
		Label: L10nSpeakerView_Website,
		Kind:  column.Text,
		Get: func(r SpeakerView) column.Value {
			if !r.Website.Valid {
				return column.Value{Null: true}
			}
			return column.Value{Text: r.Website.String}
		},
	},
	Talks: column.Column[SpeakerView]{
		Name:  "talks",
		Label: L10nSpeakerView_Talks,
		Kind:  column.Integer,
		Get: func(r SpeakerView) column.Value {
			return column.Value{Int: int64(r.Talks)}
		},
	},
	CreatedAt: column.Column[SpeakerView]{
		Name:  "created_at",
		Label: L10nSpeakerView_CreatedAt,
		Kind:  column.Time,
		Get: func(r SpeakerView) column.Value {
			if !r.CreatedAt.Valid {
				return column.Value{Null: true}
			}
			if m := r.CreatedAt.InfinityModifier; m != 0 {
				return column.Infinite(int8(m))
			}
			return column.Value{Time: r.CreatedAt.Time}
		},
	},
}

// TalkViewColumns are talk_view's columns, for a table package (A4).
var TalkViewColumns = struct {
	ID              column.Column[TalkView]
	Title           column.Column[TalkView]
	Level           column.Column[TalkView]
	Format          column.Column[TalkView]
	State           column.Column[TalkView]
	StartsAt        column.Column[TalkView]
	DurationMinutes column.Column[TalkView]
	Recorded        column.Column[TalkView]
	SpeakerID       column.Column[TalkView]
	SpeakerName     column.Column[TalkView]
	EventID         column.Column[TalkView]
	EventName       column.Column[TalkView]
	CreatedAt       column.Column[TalkView]
}{
	ID: column.Column[TalkView]{
		Name:  "id",
		Label: L10nTalkView_Id,
		Kind:  column.Integer,
		Get: func(r TalkView) column.Value {
			return column.Value{Int: r.ID}
		},
	},
	Title: column.Column[TalkView]{
		Name:  "title",
		Label: L10nTalkView_Title,
		Kind:  column.Text,
		Get: func(r TalkView) column.Value {
			return column.Value{Text: r.Title}
		},
	},
	Level: column.Column[TalkView]{
		Name:    "level",
		Label:   L10nTalkView_Level,
		Kind:    column.Enum,
		Options: AllTalkLevelOptions(),
		Get: func(r TalkView) column.Value {
			return column.Value{Enum: r.Level}
		},
	},
	Format: column.Column[TalkView]{
		Name:    "format",
		Label:   L10nTalkView_Format,
		Kind:    column.Enum,
		Options: AllTalkFormatOptions(),
		Get: func(r TalkView) column.Value {
			return column.Value{Enum: r.Format}
		},
	},
	State: column.Column[TalkView]{
		Name:    "state",
		Label:   L10nTalkView_State,
		Kind:    column.Enum,
		Options: AllTalkStateOptions(),
		Get: func(r TalkView) column.Value {
			return column.Value{Enum: r.State}
		},
	},
	StartsAt: column.Column[TalkView]{
		Name:  "starts_at",
		Label: L10nTalkView_StartsAt,
		Kind:  column.Time,
		Get: func(r TalkView) column.Value {
			if !r.StartsAt.Valid {
				return column.Value{Null: true}
			}
			if m := r.StartsAt.InfinityModifier; m != 0 {
				return column.Infinite(int8(m))
			}
			return column.Value{Time: r.StartsAt.Time}
		},
	},
	DurationMinutes: column.Column[TalkView]{
		Name:  "duration_minutes",
		Label: L10nTalkView_DurationMinutes,
		Kind:  column.Integer,
		Get: func(r TalkView) column.Value {
			return column.Value{Int: int64(r.DurationMinutes)}
		},
	},
	Recorded: column.Column[TalkView]{
		Name:  "recorded",
		Label: L10nTalkView_Recorded,
		Kind:  column.Bool,
		Get: func(r TalkView) column.Value {
			if !r.Recorded.Valid {
				return column.Value{Null: true}
			}
			return column.Value{Bool: r.Recorded.Bool}
		},
	},
	SpeakerID: column.Column[TalkView]{
		Name:  "speaker_id",
		Label: L10nTalkView_SpeakerId,
		Kind:  column.Integer,
		Get: func(r TalkView) column.Value {
			return column.Value{Int: r.SpeakerID}
		},
	},
	SpeakerName: column.Column[TalkView]{
		Name:  "speaker_name",
		Label: L10nTalkView_SpeakerName,
		Kind:  column.Text,
		Get: func(r TalkView) column.Value {
			return column.Value{Text: r.SpeakerName}
		},
	},
	EventID: column.Column[TalkView]{
		Name:  "event_id",
		Label: L10nTalkView_EventId,
		Kind:  column.Integer,
		Get: func(r TalkView) column.Value {
			return column.Value{Int: r.EventID}
		},
	},
	EventName: column.Column[TalkView]{
		Name:  "event_name",
		Label: L10nTalkView_EventName,
		Kind:  column.Text,
		Get: func(r TalkView) column.Value {
			return column.Value{Text: r.EventName}
		},
	},
	CreatedAt: column.Column[TalkView]{
		Name:  "created_at",
		Label: L10nTalkView_CreatedAt,
		Kind:  column.Time,
		Get: func(r TalkView) column.Value {
			if !r.CreatedAt.Valid {
				return column.Value{Null: true}
			}
			if m := r.CreatedAt.InfinityModifier; m != 0 {
				return column.Infinite(int8(m))
			}
			return column.Value{Time: r.CreatedAt.Time}
		},
	},
}

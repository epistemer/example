// Package api is Confab's JSON API: api/openapi.yaml is the contract.
// oapi-codegen writes the server's types (zz_apigen.go, go:generate);
// handlers' openapi producer writes apiroutes (routes, labels and the
// permissions of the x-permission extensions), through genx.
package api

//go:generate go tool oapi-codegen -config ../../api/oapi-codegen.yaml ../../api/openapi.yaml

import (
	"context"
	"errors"
	"io"
	"slices"

	"github.com/epistemer/auth"
	"github.com/epistemer/models/listing"

	"github.com/epistemer/example/internal/confab"
	"github.com/epistemer/example/internal/ids"
	"github.com/epistemer/example/internal/models"
)

// Server implements the generated StrictServerInterface.
type Server struct {
	Auth  *auth.Auth
	Store *confab.Store
	Files *confab.Files
}

var _ StrictServerInterface = (*Server)(nil)

func (s *Server) SignIn(ctx context.Context, req SignInRequestObject) (SignInResponseObject, error) {
	b := req.Body
	var f auth.Flow
	var err error
	switch {
	case b.Flow != nil && b.Code != nil:
		f, err = s.Auth.Prove(ctx, *b.Flow, auth.EmailCode(*b.Code))
	case b.Login != nil && b.Password != nil:
		f, err = s.Auth.SignIn(ctx, string(*b.Login), auth.Bearer())
		if err == nil {
			f, err = s.Auth.Prove(ctx, f.Token, auth.Password(*b.Password))
		}
		if err == nil && !f.Verified && slices.Contains(f.Offers, auth.MethodEmailCode) {
			if _, err := s.Auth.Challenge(ctx, f.Token, auth.MethodEmailCode); err != nil {
				return nil, err
			}
			next := SignInResultNext("code")
			return SignIn200JSONResponse{Next: &next, Flow: &f.Token}, nil
		}
	default:
		return nil, ErrSignInBody
	}
	if err != nil {
		return nil, err
	}
	res, err := s.Auth.Finish(ctx, f.Token)
	if err != nil {
		return nil, err
	}
	return SignIn200JSONResponse{Access: &res.Access, Refresh: &res.Refresh, Verified: &res.Session.Verified}, nil
}

func (s *Server) Refresh(ctx context.Context, req RefreshRequestObject) (RefreshResponseObject, error) {
	res, err := s.Auth.Refresh(ctx, req.Body.Refresh)
	if err != nil {
		return nil, err
	}
	return Refresh200JSONResponse{Access: &res.Access, Refresh: &res.Refresh}, nil
}

func (s *Server) GetMe(ctx context.Context, _ GetMeRequestObject) (GetMeResponseObject, error) {
	sess, _ := auth.From(ctx)
	return GetMe200JSONResponse{User: string(sess.User), Verified: sess.Verified}, nil
}

func (s *Server) GetSchedule(ctx context.Context, req GetScheduleRequestObject) (GetScheduleResponseObject, error) {
	e, err := s.Store.EventBySlug(ctx, ids.Slug(req.Slug))
	if err != nil {
		return nil, err
	}
	if e.Status != models.EventStatusPublished {
		return nil, confab.ErrPublishedOnly
	}
	rows, err := s.Store.PublicSchedule(ctx, e.ID)
	if err != nil {
		return nil, err
	}
	out := make(GetSchedule200JSONResponse, len(rows))
	for i, r := range rows {
		out[i] = ScheduledTalk{Id: r.ID, Title: r.Title, StartsAt: *r.StartsAt, Minutes: int(r.DurationMinutes),
			Speaker: r.Speaker, Level: string(r.Level), Format: string(r.Format)}
	}
	return out, nil
}

// ListTalks reads the same view, through the same listing, as the admin's
// talks table.
func (s *Server) ListTalks(ctx context.Context, req ListTalksRequestObject) (ListTalksResponseObject, error) {
	q := listing.Query{
		Sort:  []listing.Sort{{Column: models.TalkViewColumns.CreatedAt.Name, Desc: true}, {Column: models.TalkViewColumns.ID.Name}},
		Limit: deref(req.Params.Limit, 20), Offset: deref(req.Params.Offset, 0),
	}
	if st := req.Params.State; st != nil {
		q.Filters = []listing.Filter{{Column: models.TalkViewColumns.State.Name, Op: listing.Eq, Value: *st}}
	}
	rows, total, err := s.Store.ListTalkView(ctx, q)
	if err != nil {
		return nil, err
	}
	out := ListTalks200JSONResponse{Total: total, Talks: make([]Talk, len(rows))}
	for i, r := range rows {
		t := Talk{Id: r.ID, Title: r.Title, State: r.State, Speaker: r.SpeakerName, Event: r.EventName}
		t.StartsAt = r.StartsAt
		out.Talks[i] = t
	}
	return out, nil
}

func (s *Server) SetTalkState(ctx context.Context, req SetTalkStateRequestObject) (SetTalkStateResponseObject, error) {
	t, err := s.Store.TalkView(ctx, ids.TalkID(req.Id))
	if err != nil {
		return nil, err
	}
	if req.Body.State == models.TalkStateScheduled {
		if err := confab.CanSchedule(t); err != nil {
			return nil, err
		}
	}
	if _, err := s.Store.SetState(ctx, []int64{req.Id}, req.Body.State); err != nil {
		return nil, err
	}
	return SetTalkState204Response{}, nil
}

func (s *Server) UploadSlides(ctx context.Context, req UploadSlidesRequestObject) (UploadSlidesResponseObject, error) {
	if _, err := s.Store.Talk(ctx, ids.TalkID(req.Id)); err != nil {
		return nil, err
	}
	for {
		part, err := req.Body.NextPart()
		if errors.Is(err, io.EOF) {
			return nil, ErrNoFile
		}
		if err != nil {
			return nil, err
		}
		if part.FormName() != "file" {
			continue
		}
		f, err := s.Files.PutReader(part.FileName(), part)
		if err != nil {
			return nil, err
		}
		if err := s.Store.AddSlides(ctx, ids.TalkID(req.Id), f); err != nil {
			_ = s.Files.Delete(f.Key)
			return nil, err
		}
		return UploadSlides201JSONResponse{Key: f.Key}, nil
	}
}

func deref[T any](p *T, def T) T {
	if p == nil {
		return def
	}
	return *p
}

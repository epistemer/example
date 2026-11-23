package main

import (
	"context"
	"time"

	"github.com/epistemer/auth"
	"github.com/epistemer/auth/access"
	"github.com/epistemer/errors"
	"github.com/shopspring/decimal"
	"github.com/sirupsen/logrus"

	"github.com/epistemer/example/internal/admin/adminroutes"
	"github.com/epistemer/example/internal/api/apiroutes"
	"github.com/epistemer/example/internal/confab"
	"github.com/epistemer/example/internal/ids"
	"github.com/epistemer/example/internal/models"
	"github.com/epistemer/example/internal/settings"
)

// seed adds the organiser, who holds every permission (access.Root), a
// reviewer group, and a programme to look at. Run it once, on an empty
// database, while serve isn't running (bbolt has one writer).
func seed(ctx context.Context, s *settings.Seed, log *logrus.Logger) error {
	db := s.Store().DB()
	store, err := confab.Open(ctx, s.DBURL().Reveal())
	if err != nil {
		return err
	}
	defer store.Close()
	a, err := auth.New(db, s, confab.Directory{S: store}, confab.Outbox{Dir: s.MailOutbox()})
	if err != nil {
		return err
	}
	ac, err := access.New(db)
	if err != nil {
		return err
	}

	const ada = "u-ada"
	if err := store.UpsertUser(ctx, models.UpsertUserParams{ID: ada, Email: s.DemoEmail(), Name: s.DemoName()}); err != nil {
		return err
	}
	if err := a.SetPassword(ctx, ada, s.DemoPassword().Reveal()); err != nil {
		return err
	}
	if err := ac.Join(ctx, ada, access.Root); err != nil {
		return err
	}
	g, err := ac.CreateGroup(ctx, "Reviewers")
	if err == nil {
		for _, p := range []*access.Perm{adminroutes.PermStaff, adminroutes.PermTalksManage, adminroutes.PermTalksReview, apiroutes.PermApiTalksRead, apiroutes.PermApiTalksReview} {
			if err := ac.Grant(ctx, g.ID, p); err != nil {
				return err
			}
		}
	}
	log.WithField("email", s.DemoEmail()).Info("organiser ready")

	if _, err := store.EventBySlug(ctx, "gophercon-delta"); err == nil {
		log.Info("demo programme already there")
		return nil
	} else if !errors.Is(err, confab.ErrNotFound) {
		return err
	}
	day := func(m time.Month, d int) time.Time { return time.Date(2027, m, d, 0, 0, 0, 0, time.UTC) }
	delta, err := store.CreateEvent(ctx, confab.EventInput{
		Name: "GopherCon Delta", Slug: "gophercon-delta", Status: models.EventStatusPublished,
		Summary:  "Three days of Go by the Nile: talks, workshops and a boat trip.",
		StartsOn: day(4, 12), EndsOn: day(4, 14), Capacity: 600, Currency: "EGP",
		Venue: &confab.Venue{Name: "Cairo Conference Centre", Street: "Salah Salem", City: "Cairo", Country: "EG", PostCode: "11371"},
		Tiers: []confab.Tier{
			{Name: "Early bird", Price: decimal.RequireFromString("1500.00"), Quantity: 150},
			{Name: "Regular", Price: decimal.RequireFromString("2400.00"), Quantity: 400},
		},
	})
	if err != nil {
		return err
	}
	if _, err := store.CreateEvent(ctx, confab.EventInput{
		Name: "Web Days Online", Slug: "web-days-online", Status: models.EventStatusDraft, Online: true,
		StartsOn: day(9, 3), EndsOn: day(9, 3), Capacity: 5000, Currency: "EUR",
		Tiers: []confab.Tier{{Name: "Free", Price: decimal.Zero, Quantity: 5000}},
	}); err != nil {
		return err
	}
	people := []confab.SpeakerInput{
		{Name: "Layla Hassan", Email: "layla@example.org", Country: "EG", Bio: "Builds payment systems in Go."},
		{Name: "Tomás Ruiz", Email: "tomas@example.org", Country: "ES", Bio: "Database person; likes query plans."},
		{Name: "Mei Chen", Email: "mei@example.org", Country: "SG", Bio: "Designs interfaces that read right to left and left to right."},
		{Name: "Ola Nordmann", Email: "ola@example.org", Country: "NO", Bio: "Runs the platform team."},
	}
	var sp []ids.SpeakerID
	for _, p := range people {
		id, err := store.CreateSpeaker(ctx, p)
		if err != nil {
			return err
		}
		sp = append(sp, id)
	}
	at := func(d, h int) *time.Time { t := time.Date(2027, 4, d, h, 0, 0, 0, time.UTC); return &t }
	yes := true
	seats := 30
	talks := []struct {
		in    confab.TalkInput
		state models.TalkState
	}{
		{confab.TalkInput{Speaker: sp[0], Title: "Idempotent payments with Go", Level: models.TalkLevelIntermediate, Format: models.TalkFormatTalk, Topics: []models.TalkTopic{models.TalkTopicGo, models.TalkTopicData}, StartsAt: at(12, 10), Duration: 40 * time.Minute, Recorded: &yes}, models.TalkStateScheduled},
		{confab.TalkInput{Speaker: sp[1], Title: "Reading query plans together", Level: models.TalkLevelAdvanced, Format: models.TalkFormatWorkshop, Topics: []models.TalkTopic{models.TalkTopicData}, StartsAt: at(13, 14), Duration: 120 * time.Minute, Seats: &seats}, models.TalkStateAccepted},
		{confab.TalkInput{Speaker: sp[2], Title: "Right to left, by design", Level: models.TalkLevelBeginner, Format: models.TalkFormatTalk, Topics: []models.TalkTopic{models.TalkTopicDesign, models.TalkTopicWeb}, Duration: 30 * time.Minute}, models.TalkStateSubmitted},
		{confab.TalkInput{Speaker: sp[3], Title: "On call without tears", Level: models.TalkLevelBeginner, Format: models.TalkFormatLightning, Topics: []models.TalkTopic{models.TalkTopicOperations}, Duration: 10 * time.Minute}, models.TalkStateSubmitted},
		{confab.TalkInput{Speaker: sp[0], Title: "Secrets in config, done right", Level: models.TalkLevelIntermediate, Format: models.TalkFormatPanel, Topics: []models.TalkTopic{models.TalkTopicSecurity, models.TalkTopicGo}, Duration: 45 * time.Minute}, models.TalkStateRejected},
	}
	rules := seedRules{}
	for _, t := range talks {
		t.in.Event = delta
		t.in.Abstract = "A talk for people who build and run back offices: what we tried, what broke, and what we kept. " + t.in.Title + "."
		id, _, err := store.SaveTalk(ctx, rules, 0, t.in, nil, nil)
		if err != nil {
			return err
		}
		if _, err := store.SetState(ctx, []int64{int64(id)}, t.state); err != nil {
			return err
		}
	}
	log.Info("demo programme added")
	return nil
}

type seedRules struct{}

func (seedRules) CFPClosed() bool  { return false }
func (seedRules) CFPMaxTalks() int { return 3 }

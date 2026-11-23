package admin

import (
	"slices"
	"testing"
	"time"

	"github.com/epistemer/ui/forms/formstest"

	"github.com/epistemer/example/internal/models"
)

// Every form's fields, groups and lists carry a label of their own.
func TestFormsAreLabelled(t *testing.T) {
	formstest.AssertLabelled(t, NewSignInForm().Form)
	formstest.AssertLabelled(t, NewCodeForm().Form)
	formstest.AssertLabelled(t, NewEventForm().Form)
	formstest.AssertLabelled(t, NewSpeakerForm(nil).Form)
	formstest.AssertLabelled(t, NewTalkForm(nil, speakerLabels{}, nil).Form)
	formstest.AssertLabelled(t, NewImportForm().Form)
	formstest.AssertLabelled(t, NewSettingsForm().Form)
	formstest.AssertLabelled(t, NewMoveForm().Form)
}

const event = "name=GopherCon&slug=gophercon-x&status=draft&starts_on=2027-04-12&ends_on=2027-04-14" +
	"&capacity=300&currency=EUR&version=1" +
	"&tiers[0].name=Regular&tiers[0].price=120.50&tiers[0].quantity=100"

func TestEventFormOnlineDropsTheVenue(t *testing.T) {
	f := NewEventForm()
	if !formstest.Submit(t, t.Context(), f.Form, event+"&online=true&venue.name=ignored") {
		t.Fatalf("errors: %v", formstest.Errors(f.Form))
	}
	if in := f.Input(); in.Venue != nil {
		t.Errorf("an online event kept a venue: %+v", in.Venue)
	}
}

func TestEventFormNeedsAVenueOffline(t *testing.T) {
	f := NewEventForm()
	if formstest.Submit(t, t.Context(), f.Form, event) {
		t.Fatal("an offline event without a venue was accepted")
	}
	got := formstest.Errors(f.Form)
	for _, want := range []string{"venue.name: errors.forms.required", "venue.country: errors.forms.required"} {
		if !slices.Contains(got, want) {
			t.Errorf("errors %v lack %q", got, want)
		}
	}
}

func TestEventFormRules(t *testing.T) {
	f := NewEventForm()
	formstest.Submit(t, t.Context(), f.Form,
		"name=Go&slug=Not_A_Slug&status=nope&starts_on=2027-04-14&ends_on=2027-04-12&capacity=0&currency=EURO&online=true")
	got := formstest.Errors(f.Form)
	for _, want := range []string{
		"name: errors.forms.rule.min_len",
		"slug: errors.confab.bad_slug", // ids.Slug's own error, through Parsed
		"status: errors.forms.not_an_option",
		"capacity: errors.forms.rule.between",
		"currency: errors.forms.rule.currency_code",
		"tiers: errors.forms.min_items",
	} {
		if !slices.Contains(got, want) {
			t.Errorf("errors %v lack %q", got, want)
		}
	}
}

func TestEventFormEndsBeforeItStarts(t *testing.T) {
	f := NewEventForm()
	formstest.Submit(t, t.Context(), f.Form, event+"&online=true&ends_on=2027-04-01")
	if got := formstest.Errors(f.Form); !slices.Contains(got, "ends_on: errors.confab.ends_first") {
		t.Errorf("errors %v lack the form check's", got)
	}
}

func TestEventFormPostCodeFollowsTheCountry(t *testing.T) {
	venue := "&venue.name=Hall&venue.street=Main&venue.city=Amsterdam&venue.country=NL"
	f := NewEventForm()
	formstest.Submit(t, t.Context(), f.Form, event+venue+"&venue.post_code=ABC")
	if got := formstest.Errors(f.Form); !slices.Contains(got, "venue.post_code: errors.forms.rule.post_code") {
		t.Errorf("errors %v lack the post code's", got)
	}
	f = NewEventForm()
	if !formstest.Submit(t, t.Context(), f.Form, event+venue+"&venue.post_code=1012 AB") {
		t.Errorf("a Dutch post code was refused: %v", formstest.Errors(f.Form))
	}
}

func TestTalkFormSeatsOnlyForWorkshops(t *testing.T) {
	base := "event=1&speaker=2&title=Hello there&level=beginner&duration=30" +
		"&abstract=" + "An abstract long enough to pass the rule of fifty characters."
	f := NewTalkForm(eventOptions([]models.ListEventOptionsRow{{ID: 1, Name: "E"}}), speakerLabels{}, nil)
	if !formstest.Submit(t, t.Context(), f.Form, base+"&format=talk&seats=7") {
		t.Fatalf("errors: %v", formstest.Errors(f.Form))
	}
	if f.Seats.Ptr() != nil {
		t.Error("a talk kept seats")
	}
	if f.Duration.Value() != 30*time.Minute {
		t.Errorf("duration = %v, want 30m (As(minutes))", f.Duration.Value())
	}

	f = NewTalkForm(eventOptions([]models.ListEventOptionsRow{{ID: 1, Name: "E"}}), speakerLabels{}, nil)
	formstest.Submit(t, t.Context(), f.Form, base+"&format=workshop")
	if got := formstest.Errors(f.Form); !slices.Contains(got, "seats: errors.forms.required") {
		t.Errorf("errors %v lack seats", got)
	}
}

func TestTalkFormTopics(t *testing.T) {
	base := "event=1&speaker=2&title=Hello there&level=beginner&format=talk&duration=30" +
		"&abstract=" + "An abstract long enough to pass the rule of fifty characters."
	f := NewTalkForm(eventOptions([]models.ListEventOptionsRow{{ID: 1, Name: "E"}}), speakerLabels{}, nil)
	if !formstest.Submit(t, t.Context(), f.Form, base+"&topics=go&topics=web") {
		t.Fatalf("errors: %v", formstest.Errors(f.Form))
	}
	if got := f.Topics.Value(); !slices.Equal(got, []models.TalkTopic{models.TalkTopicGo, models.TalkTopicWeb}) {
		t.Errorf("topics = %v", got)
	}
	f = NewTalkForm(eventOptions([]models.ListEventOptionsRow{{ID: 1, Name: "E"}}), speakerLabels{}, nil)
	formstest.Submit(t, t.Context(), f.Form, base+"&topics=go&topics=web&topics=data&topics=design")
	if got := formstest.Errors(f.Form); !slices.Contains(got, "topics: errors.forms.rule.max_items") {
		t.Errorf("errors %v lack max_items", got)
	}
	f = NewTalkForm(eventOptions([]models.ListEventOptionsRow{{ID: 1, Name: "E"}}), speakerLabels{}, nil)
	formstest.Submit(t, t.Context(), f.Form, base+"&topics=cobol")
	if got := formstest.Errors(f.Form); !slices.Contains(got, "topics: errors.forms.not_an_option") {
		t.Errorf("errors %v lack not_an_option", got)
	}
}

package admin

import (
	"testing"

	"github.com/epistemer/handlers"
	"github.com/epistemer/handlers/handlerstest"
	"github.com/epistemer/ui/tables/tablestest"

	"github.com/epistemer/example/internal/admin/adminroutes"
	"github.com/epistemer/example/internal/api/apiroutes"
)

// The routes anyone may reach without signing in. A new public route
// fails this test until it's added here on purpose.
func TestPublicRoutes(t *testing.T) {
	handlerstest.AssertPublic(t, adminroutes.Routes,
		"sign_in", "sign_in.password", "sign_in.code", "sign_in.verify", "event_page", "locale")
	handlerstest.AssertPublic(t, apiroutes.Routes, "api.sign_in", "api.refresh", "api.get_schedule")
}

func TestTables(t *testing.T) {
	tablestest.Check(t, EventsTable)
	tablestest.Check(t, SpeakersTable)
	tablestest.Check(t, TranslationsTable)
	h := &H{}
	tablestest.Check(t, h.talksTable())
}

func TestEventsTableQuery(t *testing.T) {
	st := tablestest.Bind(t, EventsTable, "q=delta&status=published&sort=-starts_on")
	q := tablestest.Query(st)
	if q.Search != "delta" || len(q.Filters) != 1 || q.Filters[0].Column != "status" {
		t.Errorf("query = %+v", q)
	}
	st = tablestest.Bind(t, EventsTable, "status=bogus")
	if got := tablestest.Errors(st); len(got) == 0 {
		t.Error("a bogus status filter was accepted")
	}
}

// A reviewer's one grant, talks.review, opens the review routes of both the
// admin and the API (findings F22): the API takes the admin's permission.
func TestReviewPermissionShared(t *testing.T) {
	for _, routes := range [][]handlers.Route{adminroutes.Routes, apiroutes.Routes} {
		for _, r := range routes {
			if (r.Key == "talks.review" || r.Key == "api.talks.review") && r.Permission != "talks.review" {
				t.Errorf("%s needs %q, want talks.review", r.Key, r.Permission)
			}
		}
	}
}

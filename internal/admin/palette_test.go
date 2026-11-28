package admin

import (
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/epistemer/handlers"
	"github.com/epistemer/handlers/handlerstest"
	"github.com/epistemer/i18n"
	"github.com/epistemer/i18n/i18ntest"
	"github.com/epistemer/ui"
	"github.com/epistemer/ui/actions"

	"github.com/epistemer/example/internal/admin/adminroutes"
)

// The palette's sources (neozp ADR 0002, findings F13): each entity route
// names its suggestions route, and the palette offers it as a chip.
func TestPaletteSources(t *testing.T) {
	for key, from := range map[string]string{
		"events.edit":   "events.search",
		"speakers.edit": "speakers.search",
		"translations":  "locales.search",
	} {
		rt, ok := adminroutes.ByKey(key)
		if !ok || len(rt.Params) != 1 || rt.Params[0].From != from || !rt.Searchable {
			t.Errorf("%s: %+v, want a searchable route whose parameter is from %s", key, rt.Params, from)
		}
	}

	tr := i18ntest.New(t, i18n.Locale{Code: "en", NativeName: "English"})
	h := &H{
		Routes: handlers.NewTable(adminroutes.Routes),
		Guard:  handlerstest.Guard{Session: true, Granted: map[string]bool{"events.manage": true, "translations.manage": true}},
		Tr:     tr,
	}
	p := h.NewPalette()
	r := httptest.NewRequest("GET", "/admin/palette?q=@", nil)
	r = r.WithContext(actions.Gate.With(ui.WithIDs(i18ntest.Context(t, tr, "en"), ui.Sequence("t")), actions.AllowAll))
	r.Header.Set("HX-Request", "true")
	w := httptest.NewRecorder()
	if err := p.Search(w, r); err != nil {
		t.Fatal(err)
	}
	body := w.Body.String()
	if !strings.Contains(body, `data-ui-chip-source="/admin/events/search"`) {
		t.Errorf("@ doesn't offer events.edit as a chip:\n%s", body)
	}
	// speakers.search needs talks.manage, which this person hasn't: the
	// record can't be searched for, so @ leaves speakers.edit out.
	if strings.Contains(body, "/admin/speakers/search") {
		t.Errorf("@ offers a source the guard refuses:\n%s", body)
	}
	// translations is a view: its form, not a chip.
	r2 := httptest.NewRequest("GET", "/admin/palette?q=transl", nil)
	r2 = r2.WithContext(r.Context())
	r2.Header.Set("HX-Request", "true")
	w = httptest.NewRecorder()
	if err := p.Search(w, r2); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(w.Body.String(), `data-ui-form="/admin/palette/translations/form"`) {
		t.Errorf("translations should open its form:\n%s", w.Body.String())
	}
}

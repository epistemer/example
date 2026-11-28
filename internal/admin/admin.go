// Package admin is Confab's back office: HTML pages whose routes are
// declared by //h: directives on H's methods. `go tool genx` reads them
// into internal/admin/adminroutes, which holds the typed URL builders,
// the route labels and the permissions.
package admin

import (
	"bytes"
	"context"
	"embed"
	"io"
	"log/slog"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/a-h/templ"
	"github.com/epistemer/auth"
	"github.com/epistemer/auth/access"
	"github.com/epistemer/errors"
	"github.com/epistemer/handlers"
	"github.com/epistemer/handlers/middleware"
	"github.com/epistemer/i18n"
	"github.com/epistemer/ui"
	"github.com/epistemer/ui/actions"
	"github.com/epistemer/ui/components"
	"github.com/epistemer/ui/inputs"
	"github.com/epistemer/ui/palette"
	"github.com/epistemer/ui/tables"
	"github.com/epistemer/ui/theme"

	"github.com/epistemer/example/internal/admin/adminroutes"
	"github.com/epistemer/example/internal/confab"
	"github.com/epistemer/example/internal/models"
	"github.com/epistemer/example/internal/settings"
)

// H holds what the admin's handlers need. Its exported methods with //h:
// directives are the routes.
type H struct {
	Store    *confab.Store
	Files    *confab.Files
	Auth     *auth.Auth
	Access   *access.Access
	Guard    handlers.Guard
	Theme    *theme.Theme
	Tr       *i18n.Translator
	Settings *settings.Serve
	Routes   *handlers.Table   // every route, for the palette
	Palette  *palette.Palette  // the top bar's search (ui/palette)
	Flash    *middleware.Flash // carries a toast across a redirect

	// The talks table filters by speaker through a Labeler that reads the
	// database, so Check builds it at startup, once the store exists, as
	// ui documents for tables with remote filters (findings F11).
	talksOnce sync.Once
	talks     *tables.Table[models.TalkView]
	talksView *components.TableView[models.TalkView]
}

// Check builds the tables that need the store and checks every table, so
// a mistake in a declaration stops the start, not a request.
func (h *H) Check() error {
	h.talksTable()
	for _, err := range []error{h.talks.Check(), EventsTable.Check(), SpeakersTable.Check(), TranslationsTable.Check()} {
		if err != nil {
			return err
		}
	}
	return nil
}

//go:embed assets
var assets embed.FS

// Assets serves the files ui doesn't: htmx (see findings).
func Assets() http.Handler { return http.StripPrefix(adminroutes.Prefix, http.FileServerFS(assets)) }

// Middleware runs on every admin route after the guard: it gives the
// request ui's actions gate (from the guard, so an action or a nav item
// the user can't reach is absent) and a label cache for remote selects.
func (h *H) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(w, h.request(r))
	})
}

// request adds what pages read from the context. The error writer calls
// it too: handlers runs it before the middleware, or without it when the
// guard refuses.
func (h *H) request(r *http.Request) *http.Request {
	admit, err := h.Guard.Admit(r)
	if err != nil {
		slog.WarnContext(r.Context(), "admin: admit", "err", err)
	}
	ctx := actions.Gate.With(r.Context(), handlers.Gate(admit))
	ctx = inputs.WithLabelCache(ctx)
	ctx = CurrentTheme.With(ctx, h.Theme)
	ctx = components.Flash.With(ctx, h.Flash)
	r = r.WithContext(ctx)
	return r.WithContext(pageRequest.With(r.Context(), r))
}

// pageRequest is the request a page renders for: palette.Page gives the
// shell only a title and a body, and Confab's shell reads the request
// (the current nav item, the colour scheme) (findings F36).
var pageRequest = ui.NewProvider[*http.Request](nil)

// NewPalette is the top bar's palette over h.Routes: its three routes,
// the labelers of the sources it offers, and pages in the shell.
func (h *H) NewPalette() *palette.Palette {
	return palette.New(h.Routes, h.Guard, h.Tr, palette.Endpoints{
		Search: adminroutes.PaletteSearch(),
		Form:   func(key string) handlers.Link { return adminroutes.PaletteForm(key) },
		Go:     func(key string) handlers.Link { return adminroutes.PaletteGo(key) },
	},
		palette.Labels("events.search", eventLabels{h.Store}),
		palette.Labels("speakers.search", speakerLabels{h.Store}),
		palette.Labels("locales.search", localeLabels{h.Tr}),
		palette.Page(func(title i18n.Ter, body templ.Component) templ.Component {
			return templ.ComponentFunc(func(ctx context.Context, w io.Writer) error {
				return shell(h, pageRequest.Value(ctx), title, body).Render(ctx, w)
			})
		}),
	)
}

// render writes a whole page, buffered so a render error can still become
// an error page.
func render(w http.ResponseWriter, r *http.Request, status int, c templ.Component) error {
	var b bytes.Buffer
	if err := c.Render(r.Context(), &b); err != nil {
		return err
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(status)
	_, err := w.Write(b.Bytes())
	return err
}

// page renders body in the app shell.
func (h *H) page(w http.ResponseWriter, r *http.Request, status int, title i18n.Ter, body templ.Component) error {
	return render(w, r, status, shell(h, r, title, body))
}

// ErrorPage answers every error of the admin: the guard's refusals, auth's,
// handlers' (a recovered panic included) and the handlers' own.
// handlers/middleware logs it and picks the status; errorPage writes it.
func (h *H) ErrorPage(w http.ResponseWriter, r *http.Request, err error) {
	middleware.ErrorPage(h.errorPage)(w, h.request(r), err)
}

// errorPage sends a person without a session to sign in, keeps the page
// under htmx with a toast, and otherwise renders the error in the shell.
func (h *H) errorPage(w http.ResponseWriter, r *http.Request, status int, err error) error {
	if errors.Is(err, auth.ErrNoSession) && r.Method == http.MethodGet {
		return middleware.Redirect(w, r, adminroutes.SignIn())
	}
	if d, ok := auth.RetryAfter(err); ok {
		w.Header().Set("Retry-After", strconv.Itoa(max(1, int((d+time.Second-1)/time.Second))))
	}
	if middleware.HTMX(r) {
		w.Header().Set("HX-Reswap", "none")
		return render(w, r, status, errorToast(err))
	}
	return render(w, r, status, shell(h, r, L10nErrorTitle, errorBody(status, err)))
}

// CurrentTheme carries the theme to every page, sign-in pages included,
// through ui's typed context values.
var CurrentTheme = ui.NewProvider[*theme.Theme](nil)

func themeHead(ctx context.Context) templ.Component { return CurrentTheme.Value(ctx).Head() }

func errorsT(ctx context.Context, err error) string { return errors.T(ctx, err) }

func handlersStatus(err error) int { return handlers.Status(err) }

// deref reads a nullable column: the zero value when it is NULL.
func deref[T any](p *T) T {
	if p == nil {
		var z T
		return z
	}
	return *p
}

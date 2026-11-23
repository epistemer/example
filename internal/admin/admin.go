// Package admin is Confab's back office: HTML pages whose routes are
// declared by //h: directives on H's methods. `go tool genx` reads them
// into internal/admin/adminroutes, which holds the typed URL builders,
// the route labels and the permissions.
package admin

import (
	"bytes"
	"context"
	"embed"
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
	"github.com/epistemer/i18n"
	"github.com/epistemer/ui"
	"github.com/epistemer/ui/actions"
	"github.com/epistemer/ui/components"
	"github.com/epistemer/ui/inputs"
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
	Routes   *handlers.Table // every route, for the search page

	// The talks table filters by speaker through a Labeler that reads the
	// database, so it's built once the store exists (findings F11).
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
	return r.WithContext(ctx)
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
	return render(w, r, status, shell(h, r, title, body, takeFlash(w, r)))
}

// ErrorPage answers every error of the admin: the guard's refusals, auth's,
// handlers' and the handlers' own. Without a session it sends the person
// to sign in.
func (h *H) ErrorPage(w http.ResponseWriter, r *http.Request, err error) {
	r = h.request(r)
	if errors.Is(err, auth.ErrNoSession) && r.Method == http.MethodGet {
		http.Redirect(w, r, adminroutes.SignIn().URL(), http.StatusSeeOther)
		return
	}
	if d, ok := auth.RetryAfter(err); ok {
		w.Header().Set("Retry-After", strconv.Itoa(max(1, int((d+time.Second-1)/time.Second))))
	}
	status := handlers.Status(err)
	if status >= 500 {
		slog.ErrorContext(r.Context(), "admin", "err", err, "path", r.URL.Path)
	}
	if r.Header.Get("HX-Request") == "true" {
		// An htmx request: a toast, out of band, keeps the page.
		w.Header().Set("HX-Reswap", "none")
		_ = render(w, r, status, errorToast(err))
		return
	}
	if rerr := render(w, r, status, shell(h, r, L10nErrorTitle, errorBody(status, err), nil)); rerr != nil {
		handlers.WriteText(w, r, err)
	}
}

// CurrentTheme carries the theme to every page, sign-in pages included,
// through ui's typed context values.
var CurrentTheme = ui.NewProvider[*theme.Theme](nil)

func themeHead(ctx context.Context) templ.Component { return CurrentTheme.Value(ctx).Head() }

func errorsT(ctx context.Context, err error) string { return errors.T(ctx, err) }

func handlersStatus(err error) int { return handlers.Status(err) }

// ErrPanic is what people see when a handler panicked.
var ErrPanic = errors.Key("errors.confab.panic", "the page failed; the error was logged")

// Recover turns a handler's panic into a 500 page. handlers has none yet
// (findings F16): without it the connection drops, and a browser may send
// the POST again.
func (h *H) Recover(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if v := recover(); v != nil {
				if v == http.ErrAbortHandler {
					panic(v)
				}
				slog.ErrorContext(r.Context(), "admin: panic", "panic", v, "path", r.URL.Path)
				h.ErrorPage(w, r, ErrPanic)
			}
		}()
		next.ServeHTTP(w, r)
	})
}

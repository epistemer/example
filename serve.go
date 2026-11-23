package main

import (
	"context"
	"errors"
	"log/slog"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/epistemer/auth"
	"github.com/epistemer/auth/access"
	"github.com/epistemer/handlers"
	"github.com/epistemer/handlers/authguard"
	"github.com/epistemer/handlers/openapi"
	"github.com/epistemer/i18n"
	"github.com/epistemer/ui/theme"
	"github.com/sirupsen/logrus"

	"github.com/epistemer/example/internal/admin"
	"github.com/epistemer/example/internal/admin/adminroutes"
	"github.com/epistemer/example/internal/api"
	"github.com/epistemer/example/internal/api/apiroutes"
	"github.com/epistemer/example/internal/confab"
	"github.com/epistemer/example/internal/settings"
)

// The libraries read the generated settings through their own getter
// interfaces: the compiler checks the schema provides them.
var (
	_ auth.Config  = (*settings.Serve)(nil)
	_ theme.Config = (*settings.Serve)(nil)
	_ i18n.Config  = (*settings.Serve)(nil)
	_ confab.Rules = (*settings.Serve)(nil)
)

var (
	english = i18n.Locale{Code: "en", NativeName: "English", FlagCode: "gb"}
	arabic  = i18n.Locale{Code: "ar", NativeName: "العربية", FlagCode: "eg", RTL: true}
)

// serve runs the back office until ctx is canceled.
func serve(ctx context.Context, s *settings.Serve, log *logrus.Logger) error {
	db := s.Store().DB() // cfgx's bbolt file: settings, translations, sign-ins

	tr := i18n.New(db, english, s, i18n.WithLogger(slog.Default()))
	if err := tr.Load(); err != nil {
		return err
	}
	if err := seedArabic(tr); err != nil {
		return err
	}

	store, err := confab.Open(ctx, s.DBURL().Reveal())
	if err != nil {
		return err
	}
	defer store.Close()
	files, err := confab.NewFiles(s.StorageFiles())
	if err != nil {
		return err
	}

	ac, err := access.New(db)
	if err != nil {
		return err
	}
	th, err := theme.New(adminroutes.Prefix+"/ui/", s)
	if err != nil {
		return err
	}

	h := &admin.H{Store: store, Files: files, Access: ac, Theme: th, Tr: tr, Settings: s}
	a, err := auth.New(db, s, confab.Directory{S: store}, confab.Outbox{Dir: s.MailOutbox()},
		// Staff may sign in with the password alone to read the dashboard;
		// everything else needs the emailed code too, unless the browser is
		// a remembered device.
		auth.WithPolicy(func(context.Context, auth.User) auth.Policy {
			return auth.Policy{AllowUnverified: true, TrustDevices: true}
		}),
		auth.WithReject(h.ErrorPage),
	)
	if err != nil {
		return err
	}
	guard := authguard.New(a, ac)
	h.Auth, h.Guard = a, guard
	if err := h.Check(); err != nil {
		return err
	}

	mux := http.NewServeMux()
	adminroutes.Register(mux, h, guard,
		handlers.WithErrors(h.ErrorPage),
		handlers.WithMiddleware(h.Recover, h.Middleware),
	)

	// The JSON API: its own mux and the bearer carrier, so a cookie never
	// reaches it and a token never reaches the admin.
	spec, err := api.GetSpec()
	if err != nil {
		return err
	}
	apiMux := http.NewServeMux()
	apiSrv := &api.Server{Auth: a, Store: store, Files: files}
	apiroutes.Mount(apiMux, guard, spec, func(rt *openapi.Router) {
		api.HandlerWithOptions(
			api.NewStrictHandlerWithOptions(apiSrv, nil, api.StrictHTTPServerOptions{
				RequestErrorHandlerFunc:  rt.RequestError,
				ResponseErrorHandlerFunc: rt.WriteError,
			}),
			api.StdHTTPServerOptions{BaseURL: apiroutes.Prefix, BaseRouter: rt, ErrorHandlerFunc: rt.RequestError},
		)
	})
	h.Routes = handlers.NewTable(adminroutes.Routes, apiroutes.Routes)

	root := http.NewServeMux()
	root.Handle(apiroutes.Prefix+"/", a.Load(auth.Bearer())(apiMux))
	root.Handle(adminroutes.Prefix+"/", a.Load(auth.Cookie())(mux))
	root.Handle(adminroutes.Prefix+"/ui/", th)
	root.Handle(adminroutes.Prefix+"/assets/", admin.Assets())
	root.Handle("GET /{$}", http.RedirectHandler(adminroutes.Dashboard().URL(), http.StatusSeeOther))

	var wg sync.WaitGroup
	defer wg.Wait()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	wg.Go(func() {
		t := time.NewTicker(time.Hour)
		defer t.Stop()
		for {
			if err := a.Sweep(ctx); err != nil && ctx.Err() == nil {
				log.WithError(err).Error("auth sweep")
			}
			select {
			case <-ctx.Done():
				return
			case <-t.C:
			}
		}
	})

	ln, err := net.Listen("tcp", s.ServerAddr())
	if err != nil {
		return err
	}
	srv := &http.Server{Handler: tr.Middleware()(root), ReadHeaderTimeout: 10 * time.Second}
	errc := make(chan error, 1)
	go func() { errc <- srv.Serve(ln) }()
	log.WithField("addr", ln.Addr().String()).Info("listening")
	select {
	case err := <-errc:
		if !errors.Is(err, http.ErrServerClosed) {
			return err
		}
	case <-ctx.Done():
	}
	log.Info("shutting down")
	sctx, done := context.WithTimeout(context.Background(), s.ServerShutdownTimeout())
	defer done()
	return srv.Shutdown(sctx)
}

// seedArabic creates the Arabic locale once, with a few translations, so
// the app shows right-to-left pages. Later edits are people's: they're
// never written again.
func seedArabic(tr *i18n.Translator) error {
	switch err := tr.CreateLocale(arabic); {
	case errors.Is(err, i18n.ErrLocaleExists):
		return nil
	case err != nil:
		return err
	}
	for k, v := range map[string]string{
		"branding.title":         "كونفاب",
		"route.dashboard":        "لوحة التحكم",
		"route.events":           "الفعاليات",
		"route.talks":            "المحاضرات",
		"route.speakers":         "المتحدثون",
		"route.settings":         "الإعدادات",
		"route.translations":     "الترجمات",
		"page.nav.programme":     "البرنامج",
		"page.nav.admin":         "الإدارة",
		"page.nav.sign_out":      "تسجيل الخروج",
		"page.dashboard.welcome": "مرحبًا بعودتك",
		"enum.talk_state.accepted": "مقبولة",
	} {
		if err := tr.Set(arabic.Code, k, v); err != nil {
			return err
		}
	}
	return nil
}

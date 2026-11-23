package admin

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/epistemer/cfgx"
	"github.com/epistemer/errors"
	"github.com/epistemer/handlers"
	"github.com/epistemer/i18n"
	"github.com/epistemer/models/column"
	"github.com/epistemer/models/listing"
	"github.com/epistemer/ui"
	"github.com/epistemer/ui/actions"
	"github.com/epistemer/ui/components"
	"github.com/epistemer/ui/forms"
	"github.com/epistemer/ui/forms/rule"
	"github.com/epistemer/ui/tables"
	"github.com/epistemer/ui/theme"

	"github.com/epistemer/auth"

	"github.com/epistemer/example/internal/admin/adminroutes"
	"github.com/epistemer/example/internal/ids"
	"github.com/epistemer/example/internal/models"
	"github.com/epistemer/example/internal/settings"
)

var (
	L10nWelcome       = i18n.Key("page.dashboard.welcome", "Welcome back")
	L10nUnverified    = i18n.Key("page.dashboard.unverified", "You signed in with one proof: pages that change the programme ask for the emailed code too.")
	L10nCFPOpen       = i18n.Key("page.dashboard.cfp_open", "The call for papers is open.")
	L10nCFPShut       = i18n.Key("page.dashboard.cfp_closed", "The call for papers is closed.")
	L10nSearchTitle   = i18n.Key("page.search.title", "Search")
	L10nNoHits        = i18n.Key("page.search.none", "No page matches.")
	L10nSettingsSaved = i18n.Key("page.settings.saved", "The settings were saved; they apply at once.")
	L10nCFPSection    = i18n.Key("page.settings.cfp", "Call for papers")
	L10nThemeSection  = i18n.Key("page.settings.theme", "Look")
	L10nTranslation   = i18n.Key("page.translations.text", "Translation")
	L10nKey           = i18n.Key("page.translations.key", "Key")
	L10nSourceText    = i18n.Key("page.translations.source", "English")
	L10nEdited        = i18n.Key("page.translations.edited", "Edited")
	L10nTransSaved    = i18n.Key("page.translations.saved", "The translation was saved.")
	L10nTransReset    = i18n.Key("page.translations.reset", "Back to the English")
	L10nResetQ        = i18n.Key("page.translations.reset_question", "Drop this translation?")
	L10nLanguage      = i18n.Key("page.translations.language", "Language")
)

//h:key dashboard
//h:method GET
//h:path /
//h:auth unverified
//h:perm staff
//h:search
//h:desc Dashboard
func (h *H) Dashboard(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()
	count := func(f func(context.Context, listing.Source) (int64, error)) int64 {
		n, _ := f(ctx, listing.Query{Limit: 1})
		return n
	}
	ev := count(func(ctx context.Context, s listing.Source) (int64, error) { _, n, err := h.Store.ListEventView(ctx, s); return n, err })
	tk := count(func(ctx context.Context, s listing.Source) (int64, error) { _, n, err := h.Store.ListTalkView(ctx, s); return n, err })
	sp := count(func(ctx context.Context, s listing.Source) (int64, error) { _, n, err := h.Store.ListSpeakerView(ctx, s); return n, err })
	sess, _ := auth.From(ctx)
	return h.page(w, r, http.StatusOK, adminroutes.L10nDashboard, dashboardPage(sess, h.Settings.CFPClosed(), ev, tk, sp))
}

//h:key account
//h:method GET
//h:path /account
//h:auth unverified
//h:perm staff
//h:desc Your account
func (h *H) Account(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()
	sess, _ := auth.From(ctx)
	sessions, err := h.Auth.Sessions(ctx, sess.User)
	if err != nil {
		return err
	}
	return h.page(w, r, http.StatusOK, adminroutes.L10nAccount, accountPage(sess, sessions))
}

//h:key file
//h:method GET
//h:path /files/{key...}
//h:auth session
//h:param key label File
//h:desc A stored file
func (h *H) File(w http.ResponseWriter, r *http.Request, key string) error {
	f, err := h.Files.Open(key)
	if err != nil {
		return err
	}
	defer f.Close()
	st, err := f.Stat()
	if err != nil {
		return err
	}
	w.Header().Set("Cache-Control", "private, max-age=3600")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	http.ServeContent(w, r, "", st.ModTime(), f)
	return nil
}

//h:key search
//h:method GET
//h:path /search
//h:auth session
//h:desc Search pages
func (h *H) Search(w http.ResponseWriter, r *http.Request) error {
	q := strings.TrimSpace(r.URL.Query().Get("q"))
	var hits []handlers.Hit
	if q != "" {
		var err error
		hits, err = h.Routes.Search(r, h.Guard, h.Tr, handlers.Query{
			Text: q, Limit: 20,
			Only: func(rt handlers.Route) bool { return rt.Searchable },
		})
		if err != nil {
			return err
		}
	}
	return h.page(w, r, http.StatusOK, L10nSearchTitle, searchPage(q, hits))
}

// SettingsForm edits the mutable settings: the call for papers and the
// theme. Its rules repeat cfgx's and theme's checks (findings F12).
type SettingsForm struct {
	*forms.Form
	Closed          *forms.Field[bool]
	MaxTalks        *forms.Field[int]
	DefaultLevel    *forms.Field[models.TalkLevel]
	DefaultDuration *forms.Field[time.Duration]
	Brand           *forms.Field[string]
	Danger          *forms.Field[string]
	Warning         *forms.Field[string]
	Success         *forms.Field[string]
	Info            *forms.Field[string]
	FontSans        *forms.Field[string]
	FontMono        *forms.Field[string]
	Radius          *forms.Field[string]
}

func checked(check func(string) error) rule.Rule[string] {
	return rule.New(func(_ context.Context, v string) error { return check(v) })
}

func NewSettingsForm() *SettingsForm {
	f := &SettingsForm{Form: forms.New()}
	// Labels are cfgx's: settings.<path>, generated from the schema.
	f.Closed = f.Checkbox("cfp_closed").Hint(forms.KindToggle).Label(settings.L10nCFPClosed)
	f.MaxTalks = f.Int("cfp_max_talks").Label(settings.L10nCFPMaxTalks).Required().Is(rule.Between(1, 10))
	f.DefaultLevel = f.Enum("cfp_default_level", models.AllTalkLevelValues()).Label(settings.L10nCFPDefaultLevel).Required()
	f.DefaultDuration = f.Int("cfp_default_duration").Label(settings.L10nCFPDefaultDuration).Required().
		Is(rule.Between(5, 240)).As(minutes)
	color := func(name string, l i18n.Ter) *forms.Field[string] {
		return f.Text(name).Label(l).Required().Is(checked(theme.CheckColor))
	}
	f.Brand = color("theme_brand_color", settings.L10nThemeBrandColor)
	f.Danger = color("theme_danger_color", settings.L10nThemeDangerColor)
	f.Warning = color("theme_warning_color", settings.L10nThemeWarningColor)
	f.Success = color("theme_success_color", settings.L10nThemeSuccessColor)
	f.Info = color("theme_info_color", settings.L10nThemeInfoColor)
	f.FontSans = f.Text("theme_font_sans").Label(settings.L10nThemeFontSans).Required().Is(checked(theme.CheckFont))
	f.FontMono = f.Text("theme_font_mono").Label(settings.L10nThemeFontMono).Required().Is(checked(theme.CheckFont))
	f.Radius = f.Text("theme_radius").Label(settings.L10nThemeRadius).Required().Is(checked(theme.CheckRadius))
	return f
}

func (f *SettingsForm) Fill(s *settings.Serve) {
	f.Closed.Set(s.CFPClosed())
	f.MaxTalks.Set(s.CFPMaxTalks())
	f.DefaultLevel.Set(s.CFPDefaultLevel())
	f.DefaultDuration.Set(s.CFPDefaultDuration())
	f.Brand.Set(s.ThemeBrandColor())
	f.Danger.Set(s.ThemeDangerColor())
	f.Warning.Set(s.ThemeWarningColor())
	f.Success.Set(s.ThemeSuccessColor())
	f.Info.Set(s.ThemeInfoColor())
	f.FontSans.Set(s.ThemeFontSans())
	f.FontMono.Set(s.ThemeFontMono())
	f.Radius.Set(s.ThemeRadius())
}

func (f *SettingsForm) apply(u *settings.Update) error {
	u.SetCFPClosed(f.Closed.Value())
	u.SetCFPMaxTalks(f.MaxTalks.Value())
	u.SetCFPDefaultLevel(f.DefaultLevel.Value())
	u.SetCFPDefaultDuration(f.DefaultDuration.Value())
	u.SetThemeBrandColor(f.Brand.Value())
	u.SetThemeDangerColor(f.Danger.Value())
	u.SetThemeWarningColor(f.Warning.Value())
	u.SetThemeSuccessColor(f.Success.Value())
	u.SetThemeInfoColor(f.Info.Value())
	u.SetThemeFontSans(f.FontSans.Value())
	u.SetThemeFontMono(f.FontMono.Value())
	u.SetThemeRadius(f.Radius.Value())
	return nil
}

//h:key settings
//h:method GET
//h:path /settings
//h:perm settings.manage
//h:search
//h:desc Settings
func (h *H) SettingsEdit(w http.ResponseWriter, r *http.Request) error {
	f := NewSettingsForm()
	f.Fill(h.Settings)
	return h.page(w, r, http.StatusOK, adminroutes.L10nSettingsEdit, settingsPage(f))
}

//h:key settings.save
//h:method POST
//h:path /settings
//h:perm settings.manage
//h:desc Save the settings
func (h *H) SettingsSave(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()
	f := NewSettingsForm()
	if !f.Submit(ctx, r) {
		return h.page(w, r, http.StatusUnprocessableEntity, adminroutes.L10nSettingsEdit, settingsPage(f))
	}
	who := "unknown"
	if s, ok := auth.From(ctx); ok {
		who = string(s.User)
	}
	if err := h.Settings.Update(ctx, f.apply, cfgx.By(who)); err != nil {
		// cfgx reports problems by key, as English text (findings F12).
		var probs cfgx.Errors
		if errors.As(err, &probs) {
			f.Fail(ErrSettingsRefused.With(probs.Error()))
		} else {
			f.Fail(err)
		}
		return h.page(w, r, http.StatusUnprocessableEntity, adminroutes.L10nSettingsEdit, settingsPage(f))
	}
	return components.Reply(w, r).Toast(L10nSettingsSaved).Go(adminroutes.SettingsEdit())
}

var ErrSettingsRefused = errors.Key1[string]("errors.confab.settings_refused", "the settings were refused: %s", errors.Invalid)

// Translations: a ui table over i18n's catalogue, not over SQL. Its
// columns are built by hand, as models would generate them.
var (
	trKey = column.Column[i18n.Entry]{Name: "key", Label: dataKey(L10nKey), Kind: column.Text,
		Get: func(e i18n.Entry) column.Value { return column.Value{Text: e.Key} }}
	trSource = column.Column[i18n.Entry]{Name: "source", Label: dataKey(L10nSourceText), Kind: column.Text,
		Get: func(e i18n.Entry) column.Value { return column.Value{Text: e.Source} }}
	trText = column.Column[i18n.Entry]{Name: "text", Label: dataKey(L10nTranslation), Kind: column.Text,
		Get: func(e i18n.Entry) column.Value { return column.Value{Text: e.Text} }}
	trEdited = column.Column[i18n.Entry]{Name: "override", Label: dataKey(L10nEdited), Kind: column.Bool,
		Get: func(e i18n.Entry) column.Value { return column.Value{Bool: e.Override} }}
)

// dataKey is a TKer from a key handle, as column.Column wants.
func dataKey(k i18n.TKer) i18n.TKer { return k }

var TranslationsTable = tables.New("translations", trKey,
	tables.Col(trKey).Searchable(),
	tables.Col(trSource).Searchable(),
	tables.Col(trText).Searchable(),
	tables.Col(trEdited),
).
	Label(adminroutes.L10nTranslations).
	RowLabel(trKey).
	Sizes(25, 50, 100).
	Actions(translationActions)

var TranslationsView = components.TableOf(TranslationsTable)

var (
	EditTranslation  = actions.New(L10nEdit).Dialog()
	ResetTranslation = actions.New(L10nTransReset).Confirm(L10nResetQ)
)

// translationActions needs the locale the page shows: it comes through
// the context, set by the handler.
var pageLocale = ui.NewProvider[ids.Locale]("en")

func translationActions(ctx context.Context, e i18n.Entry) []actions.Item {
	loc := pageLocale.Value(ctx)
	return actions.Resolve(ctx,
		EditTranslation.To(adminroutes.TranslationsEdit(loc, e.Key)),
		ResetTranslation.To(adminroutes.TranslationsReset(loc, e.Key)).If(e.Override),
	)
}

// search runs a table's state against i18n's catalogue: the listing
// query's search and paging, with the total from a second, unpaged read.
func (h *H) searchTranslations(loc ids.Locale, q listing.Query) ([]i18n.Entry, int64, error) {
	all, err := h.Tr.Search(string(loc), i18n.Query{Contains: q.Search})
	if err != nil {
		return nil, 0, err
	}
	total := int64(len(all))
	lo := min(int(q.Offset), len(all))
	hi := len(all)
	if q.Limit > 0 {
		hi = min(lo+int(q.Limit), len(all))
	}
	return all[lo:hi], total, nil
}

//h:key translations
//h:method GET
//h:path /translations/{locale}
//h:perm translations.manage
//h:view
//h:search
//h:param locale label Language
//h:desc Translations
func (h *H) Translations(w http.ResponseWriter, r *http.Request, locale ids.Locale) error {
	l, ok := h.Tr.Locale(string(locale))
	if !ok {
		return i18n.ErrUnknownLocale
	}
	ctx := pageLocale.With(r.Context(), locale)
	r = r.WithContext(ctx)
	st := TranslationsTable.Bind(r)
	v := st.Rejected()
	if st.Valid() {
		rows, total, err := h.searchTranslations(locale, st.ListingQuery())
		if err != nil {
			return err
		}
		v = st.Page(rows, total)
	}
	status := st.Answer(w, r)
	if tables.Targeted(r, TranslationsTable) {
		return render(w, r, status, TranslationsView.Page(v))
	}
	return h.page(w, r, status, adminroutes.L10nTranslations, translationsPage(h.Tr.Locales(), l, TranslationsView.Page(v)))
}

// TranslationForm edits one key in one locale.
type TranslationForm struct {
	*forms.Form
	Body *forms.Field[string] // not Text: the embedded Form has a method Text (findings)
}

func NewTranslationForm(e i18n.Entry) *TranslationForm {
	f := &TranslationForm{Form: forms.New()}
	f.Body = f.Text("text").Raw().Hint(forms.KindMultiline).Label(L10nTranslation).Help(i18n.Static(e.Source)).Required()
	return f
}

func (h *H) entry(locale ids.Locale, key string) (i18n.Entry, error) {
	es, err := h.Tr.Search(string(locale), i18n.Query{Prefix: key})
	if err != nil {
		return i18n.Entry{}, err
	}
	for _, e := range es {
		if e.Key == key {
			return e, nil
		}
	}
	return i18n.Entry{}, i18n.ErrUnknownKey
}

//h:key translations.edit
//h:method GET
//h:path /translations/{locale}/keys/{key...}
//h:perm translations.manage
//h:param locale label Language
//h:param key label Key
//h:desc Edit a translation
func (h *H) TranslationsEdit(w http.ResponseWriter, r *http.Request, locale ids.Locale, key string) error {
	e, err := h.entry(locale, key)
	if err != nil {
		return err
	}
	f := NewTranslationForm(e)
	f.Body.Set(e.Text)
	return h.dialogOrPage(w, r, http.StatusOK, i18n.Static(key), translationBody(locale, e, f))
}

//h:key translations.save
//h:method POST
//h:path /translations/{locale}/keys/{key...}
//h:perm translations.manage
//h:param locale label Language
//h:param key label Key
//h:desc Save a translation
func (h *H) TranslationsSave(w http.ResponseWriter, r *http.Request, locale ids.Locale, key string) error {
	ctx := r.Context()
	e, err := h.entry(locale, key)
	if err != nil {
		return err
	}
	f := NewTranslationForm(e)
	if !f.Submit(ctx, r) {
		return h.dialogOrPage(w, r, http.StatusUnprocessableEntity, i18n.Static(key), translationBody(locale, e, f))
	}
	if err := h.Tr.Set(string(locale), key, f.Body.Value()); err != nil {
		f.Body.Fail(err)
		return h.dialogOrPage(w, r, http.StatusUnprocessableEntity, i18n.Static(key), translationBody(locale, e, f))
	}
	return components.Reply(w, r).Toast(L10nTransSaved).Refresh(TranslationsTable.ID()).Else(adminroutes.Translations(locale))
}

//h:key translations.reset
//h:method DELETE
//h:path /translations/{locale}/keys/{key...}
//h:perm translations.manage
//h:param locale label Language
//h:param key label Key
//h:desc Drop a translation
func (h *H) TranslationsReset(w http.ResponseWriter, r *http.Request, locale ids.Locale, key string) error {
	if _, err := h.Tr.Reset(string(locale), key); err != nil {
		return err
	}
	return components.Reply(w, r).Toast(L10nTransSaved).Refresh(TranslationsTable.ID()).Else(adminroutes.Translations(locale))
}

// LocaleForm picks the language of the pages.
type LocaleForm struct {
	*forms.Form
	Code *forms.Field[string]
}

//h:key locale
//h:method POST
//h:path /locale
//h:auth public
//h:desc Change the language
func (h *H) Locale(w http.ResponseWriter, r *http.Request) error {
	if err := h.Tr.SetCookie(w, r.FormValue("code")); err != nil {
		return err
	}
	back := r.Header.Get("Referer")
	if back == "" || !strings.HasPrefix(back, "/") && !strings.Contains(back, r.Host) {
		back = adminroutes.Dashboard().URL()
	}
	http.Redirect(w, r, back, http.StatusSeeOther)
	return nil
}

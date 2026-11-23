package admin

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/a-h/templ"
	"github.com/epistemer/errors"
	"github.com/epistemer/i18n"
	"github.com/epistemer/models/option"
	"github.com/epistemer/ui"
	"github.com/epistemer/ui/actions"
	"github.com/epistemer/ui/components"
	"github.com/epistemer/ui/forms"
	"github.com/epistemer/ui/forms/files"
	"github.com/epistemer/ui/forms/rule"
	"github.com/epistemer/ui/tables"

	"github.com/epistemer/example/internal/admin/adminroutes"
	"github.com/epistemer/example/internal/confab"
	"github.com/epistemer/example/internal/ids"
	"github.com/epistemer/example/internal/models"
	"github.com/epistemer/example/internal/pg"
)

var (
	L10nNewTalk       = i18n.Key("page.talks.new", "New talk")
	L10nTalkSaved     = i18n.Key("page.talks.saved", "The talk was saved.")
	L10nTalkDeleted   = i18n.Key("page.talks.deleted", "The talk was deleted.")
	L10nDeleteTalkQ   = i18n.Key("page.talks.delete_question", "Delete this talk and its slides?")
	L10nSlides        = i18n.Key("page.talks.slides", "Slides")
	L10nSlidesHelp    = i18n.Key("page.talks.slides_help", "PDF, up to 20 MB each, five at most.")
	L10nTopicsHelp    = i18n.Key("page.talks.topics_help", "Up to three.")
	L10nStartsAtHelp  = i18n.Key("page.talks.starts_at_help", "In UTC. Leave empty until the schedule is made.")
	L10nDurationLabel = i18n.Key("page.talks.duration", "Duration (minutes)")
	L10nAccept        = i18n.Key("page.talks.accept", "Accept")
	L10nReject        = i18n.Key("page.talks.reject", "Reject")
	L10nScheduleTalk  = i18n.Key("page.talks.schedule", "Schedule")
	L10nMove          = i18n.Key("page.talks.move", "Change state")
	L10nRejectQ       = i18n.Key("page.talks.reject_question", "Reject the selected talks?")
	L10nReviewed      = i18n.Key("page.talks.reviewed", "Talks changed: %d.")
	L10nStateChanged  = i18n.Key("page.talks.state_changed", "The talk's state was changed.")
	L10nNewState      = i18n.Key("page.talks.new_state", "New state")
)

// eventOption is an event as a choice of the talk form's select. Its
// label is data, not a translation key (findings F10).
type eventOption struct {
	id   int64
	name string
}

func (o eventOption) Value() string                        { return strconv.FormatInt(o.id, 10) }
func (o eventOption) Label() i18n.TKer                     { return dataLabel(o.name) }
func eventOptions(rows []models.ListEventOptionsRow) []option.SelectOption {
	out := make([]option.SelectOption, len(rows))
	for i, r := range rows {
		out[i] = eventOption{r.ID, r.Name}
	}
	return out
}

// dataLabel is text that comes from data, posing as a key so it fits
// option.SelectOption: K is empty.
type dataLabel string

func (d dataLabel) T(context.Context, ...any) string { return string(d) }
func (d dataLabel) K() string                        { return "" }

// minutes reads an Int field as a time.Duration: As converts the value
// both ways, and rules after it are on durations.
var minutes = forms.Conv[int, time.Duration]{
	Parse:  func(n int) (time.Duration, error) { return time.Duration(n) * time.Minute, nil },
	Format: func(d time.Duration) int { return int(d / time.Minute) },
}

// topics reads the multi-select's values as the enum talk_topic.
var topics = forms.Conv[[]string, []models.TalkTopic]{
	Parse: func(vs []string) ([]models.TalkTopic, error) {
		out := make([]models.TalkTopic, len(vs))
		for i, v := range vs {
			out[i] = models.TalkTopic(v) // Options already refused anything else
		}
		return out, nil
	},
	Format: func(ts []models.TalkTopic) []string {
		out := make([]string, len(ts))
		for i, t := range ts {
			out[i] = string(t)
		}
		return out
	},
}

// TalkForm uses every kind of choice: a select of rows, a remote select,
// enums, a multi-select, a three-state yes/no, and stored uploads.
type TalkForm struct {
	*forms.Form
	Event    *forms.Field[int64]
	Speaker  *forms.Field[int64]
	Title    *forms.Field[string]
	Abstract *forms.Field[string]
	Level    *forms.Field[models.TalkLevel]
	Format   *forms.Field[models.TalkFormat]
	Topics   *forms.Field[[]models.TalkTopic]
	StartsAt *forms.Field[time.Time]
	Duration *forms.Field[time.Duration]
	Seats    *forms.Field[int]
	Recorded *forms.Field[bool]
	Slides   *forms.Field[files.Upload]
}

func NewTalkForm(events []option.SelectOption, speakers speakerLabels, slides []files.Stored) *TalkForm {
	f := &TalkForm{Form: forms.New()}
	f.Event = f.Number[int64]("event").Label(models.L10nTalk_EventId).Required().Options(events...)
	f.Speaker = f.Number[int64]("speaker").Label(models.L10nTalk_SpeakerId).Required().
		Remote(adminroutes.SpeakersSearch(), speakers)
	f.Title = f.Text("title").Label(models.L10nTalk_Title).Required().Is(rule.MinLen(5), rule.MaxLen(150))
	f.Abstract = f.Text("abstract").Hint(forms.KindMultiline).Label(models.L10nTalk_Abstract).Required().
		Is(rule.MinLen(50), rule.MaxLen(3000))
	f.Level = f.Enum("level", models.AllTalkLevelValues()).Label(models.L10nTalk_Level).Required()
	f.Format = f.Enum("format", models.AllTalkFormatValues()).Label(models.L10nTalk_Format).Required()
	f.Topics = f.Texts("topics").Label(models.L10nTalk_Topics).Help(L10nTopicsHelp).
		Options(models.AllTalkTopicOptions()...).Is(rule.MaxItems[string](3)).As(topics)
	f.StartsAt = f.Time("starts_at", "2006-01-02T15:04").Label(models.L10nTalk_StartsAt).Help(L10nStartsAtHelp).
		When(f.Format.In(models.TalkFormatTalk, models.TalkFormatWorkshop, models.TalkFormatLightning, models.TalkFormatPanel))
	f.Duration = f.Int("duration").Label(L10nDurationLabel).Required().Is(rule.Between(5, 240)).As(minutes)
	f.Seats = f.Int("seats").Label(models.L10nTalk_Seats).When(f.Format.Eq(models.TalkFormatWorkshop)).
		Required().Is(rule.Between(1, 200))
	f.Recorded = f.Bool("recorded").Label(models.L10nTalk_Recorded)
	f.Slides = f.Uploads("slides", slides).Label(L10nSlides).Help(L10nSlidesHelp).
		Is(rule.Files(rule.MaxSize(20<<20), rule.ContentType("application/pdf")), rule.MaxFiles(5))
	return f
}

func (f *TalkForm) Input() confab.TalkInput {
	return confab.TalkInput{
		Event: ids.EventID(f.Event.Value()), Speaker: ids.SpeakerID(f.Speaker.Value()),
		Title: f.Title.Value(), Abstract: f.Abstract.Value(), Level: f.Level.Value(), Format: f.Format.Value(),
		Topics: f.Topics.Value(), StartsAt: f.StartsAt.Ptr(), Duration: f.Duration.Value(),
		Seats: f.Seats.Ptr(), Recorded: f.Recorded.Ptr(),
	}
}

func (f *TalkForm) Fill(t models.Talk) {
	f.Event.Set(t.EventID)
	f.Speaker.Set(t.SpeakerID)
	f.Title.Set(t.Title)
	f.Abstract.Set(t.Abstract)
	f.Level.Set(t.Level)
	f.Format.Set(t.Format)
	f.Topics.Set(t.Topics)
	f.StartsAt.SetPtr(pg.TimeOf(t.StartsAt))
	f.Duration.Set(time.Duration(t.DurationMinutes) * time.Minute)
	f.Seats.SetPtr(pg.Int4Of(t.Seats))
	f.Recorded.SetPtr(pg.BoolOf(t.Recorded))
}

func (f *TalkForm) fail(err error) {
	switch {
	case errors.Is(err, confab.ErrOverlap), errors.Is(err, confab.ErrOutsideEvent):
		f.StartsAt.Fail(err)
	case errors.Is(err, confab.ErrTooManyTalks):
		f.Speaker.Fail(err)
	default:
		f.Form.Fail(err)
	}
}

// The talks table: a remote filter on the speaker, enum filters, a bulk
// review and per-row actions that depend on the talk's state.
var tc = models.TalkViewColumns

func (h *H) talksTable() *tables.Table[models.TalkView] {
	h.talksOnce.Do(func() {
		h.talks = tables.New("talks", tc.ID,
			tables.Col(tc.Title).Searchable().Sortable().Link(func(t models.TalkView) ui.Link { return adminroutes.TalksEdit(ids.TalkID(t.ID)) }),
			tables.Col(tc.SpeakerName).Searchable().Sortable().Link(func(t models.TalkView) ui.Link { return adminroutes.SpeakersEdit(ids.SpeakerID(t.SpeakerID)) }),
			tables.Col(tc.SpeakerID).Hidden().Filter().Remote(adminroutes.SpeakersSearch(), speakerLabels{h.Store}),
			tables.Col(tc.EventName).Sortable(),
			tables.Col(tc.State).Filter(),
			tables.Col(tc.Level).Filter(),
			tables.Col(tc.Format).Filter(),
			tables.Col(tc.StartsAt).Sortable().Filter(),
			tables.Col(tc.DurationMinutes).Sortable(),
			tables.Col(tc.Recorded).Filter(),
			tables.Col(tc.CreatedAt).Sortable(),
		).
			Label(adminroutes.L10nTalks).
			RowLabel(tc.Title).
			Sort(tc.CreatedAt, tables.Desc).
			Actions(talkActions).
			Bulk(talkBulk)
		h.talksView = components.TableOf(h.talks).
			Cell(tc.State, statusBadge(func(t models.TalkView) option.SelectOption { return t.State }))
	})
	return h.talks
}

var (
	EditTalk     = actions.New(L10nEdit)
	AcceptTalk   = actions.New(L10nAccept).Primary()
	RejectTalk   = actions.New(L10nReject).Danger().Confirm(L10nRejectQ)
	ScheduleTalk = actions.New(L10nScheduleTalk)
	MoveTalk     = actions.New(L10nMove).Dialog()
	DeleteTalk   = actions.New(L10nDelete).Danger().Confirm(L10nDeleteTalkQ)
)

func talkActions(ctx context.Context, t models.TalkView) []actions.Item {
	id := ids.TalkID(t.ID)
	return actions.Resolve(ctx,
		EditTalk.To(adminroutes.TalksEdit(id)),
		AcceptTalk.To(adminroutes.TalksReview(id, models.TalkStateAccepted)).If(t.State == models.TalkStateSubmitted),
		ScheduleTalk.To(adminroutes.TalksReview(id, models.TalkStateScheduled)).Unless(confab.CanSchedule(t)).If(t.State == models.TalkStateAccepted),
		MoveTalk.To(adminroutes.TalksMove(id)),
		DeleteTalk.To(adminroutes.TalksDelete(id)),
	)
}

func talkBulk(ctx context.Context) []actions.Item {
	return actions.Resolve(ctx,
		AcceptTalk.To(adminroutes.TalksReviewMany(models.TalkStateAccepted)),
		RejectTalk.To(adminroutes.TalksReviewMany(models.TalkStateRejected)),
	)
}

//h:key talks
//h:method GET
//h:path /talks
//h:perm talks.manage
//h:view
//h:search
//h:desc Talks
func (h *H) Talks(w http.ResponseWriter, r *http.Request) error {
	t := h.talksTable()
	st := t.Bind(r)
	v := st.Rejected()
	if st.Valid() {
		rows, total, err := h.Store.ListTalkView(r.Context(), st)
		if err != nil {
			return err
		}
		v = st.Page(rows, total)
	}
	status := st.Answer(w, r)
	if tables.Targeted(r, t) {
		return render(w, r, status, h.talksView.Page(v))
	}
	return h.page(w, r, status, adminroutes.L10nTalks, talksPage(h.talksView.Page(v)))
}

func (h *H) newTalkForm(ctx context.Context, talk ids.TalkID) (*TalkForm, error) {
	evs, err := h.Store.ListEventOptions(ctx)
	if err != nil {
		return nil, err
	}
	var slides []files.Stored
	if talk != 0 {
		fs, err := h.Store.ListTalkFiles(ctx, int64(talk))
		if err != nil {
			return nil, err
		}
		for _, f := range fs {
			slides = append(slides, files.Stored{Key: f.Key, Name: f.Name, Href: adminroutes.File(f.Key)})
		}
	}
	return NewTalkForm(eventOptions(evs), speakerLabels{h.Store}, slides), nil
}

//h:key talks.new
//h:method GET
//h:path /talks/new
//h:perm talks.manage
//h:search
//h:entity talks
//h:desc New talk
func (h *H) TalksNew(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()
	f, err := h.newTalkForm(ctx, 0)
	if err != nil {
		return err
	}
	// The call for papers' settings, changed in the running app, give
	// the new talk its defaults.
	f.Level.Set(h.Settings.CFPDefaultLevel())
	f.Duration.Set(h.Settings.CFPDefaultDuration())
	f.Format.Set(models.TalkFormatTalk)
	return h.page(w, r, http.StatusOK, L10nNewTalk, talkFormPage(f, 0))
}

//h:key talks.create
//h:method POST
//h:path /talks
//h:perm talks.manage
//h:body 110MiB
//h:desc Create a talk
func (h *H) TalksCreate(w http.ResponseWriter, r *http.Request) error {
	return h.saveTalk(w, r, 0)
}

//h:key talks.edit
//h:method GET
//h:path /talks/{id}
//h:perm talks.manage
//h:entity talks
//h:param id label Talk number
//h:desc Edit a talk
func (h *H) TalksEdit(w http.ResponseWriter, r *http.Request, id ids.TalkID) error {
	ctx := r.Context()
	t, err := h.Store.Talk(ctx, id)
	if err != nil {
		return err
	}
	f, err := h.newTalkForm(ctx, id)
	if err != nil {
		return err
	}
	f.Fill(t)
	return h.page(w, r, http.StatusOK, i18n.Static(t.Title), talkFormPage(f, id))
}

//h:key talks.update
//h:method POST
//h:path /talks/{id}
//h:perm talks.manage
//h:body 110MiB
//h:param id label Talk number
//h:desc Save a talk
func (h *H) TalksUpdate(w http.ResponseWriter, r *http.Request, id ids.TalkID) error {
	if _, err := h.Store.Talk(r.Context(), id); err != nil {
		return err
	}
	return h.saveTalk(w, r, id)
}

func (h *H) saveTalk(w http.ResponseWriter, r *http.Request, id ids.TalkID) error {
	ctx := r.Context()
	var title i18n.Ter = L10nNewTalk
	f, err := h.newTalkForm(ctx, id)
	if err != nil {
		return err
	}
	if !f.Submit(ctx, r) {
		var big *http.MaxBytesError
		if errors.As(f.Err(), &big) {
			return f.Err()
		}
		return h.page(w, r, http.StatusUnprocessableEntity, title, talkFormPage(f, id))
	}
	up := f.Slides.Value()
	var added []confab.File
	for _, a := range up.Added {
		stored, err := h.Files.Put(a)
		if err != nil {
			return err
		}
		added = append(added, stored)
	}
	kept := make([]string, len(up.Kept))
	for i, k := range up.Kept {
		kept[i] = k.Key
	}
	saved, dropped, err := h.Store.SaveTalk(ctx, h.Settings, id, f.Input(), kept, added)
	if err != nil {
		for _, a := range added {
			_ = h.Files.Delete(a.Key)
		}
		f.fail(err)
		return h.page(w, r, http.StatusUnprocessableEntity, title, talkFormPage(f, id))
	}
	_ = h.Files.Delete(dropped...)
	return components.Reply(w, r).Toast(L10nTalkSaved).Go(adminroutes.TalksEdit(saved))
}

//h:key talks.review
//h:method POST
//h:path /talks/{id}/review/{state}
//h:perm talks.review
//h:param id label Talk number
//h:param state label New state
//h:desc Accept, reject or schedule a talk
func (h *H) TalksReview(w http.ResponseWriter, r *http.Request, id ids.TalkID, state models.TalkState) error {
	ctx := r.Context()
	t, err := h.Store.TalkView(ctx, id)
	if err != nil {
		return err
	}
	if state == models.TalkStateScheduled {
		if err := confab.CanSchedule(t); err != nil {
			return components.Reply(w, r).Fail(err).Else(adminroutes.Talks())
		}
	}
	if _, err := h.Store.SetState(ctx, []int64{int64(id)}, state); err != nil {
		return err
	}
	t.State = state
	h.talksTable()
	return components.Reply(w, r).Toast(L10nStateChanged).
		Replace(h.talks.RowID(strconv.FormatInt(int64(id), 10)), h.talksView.Row(t)).Else(adminroutes.Talks())
}

//h:key talks.review_many
//h:method POST
//h:path /talks/review/{state}
//h:perm talks.review
//h:param state label New state
//h:desc Accept or reject the chosen talks
func (h *H) TalksReviewMany(w http.ResponseWriter, r *http.Request, state models.TalkState) error {
	t := h.talksTable()
	var chosen []int64
	for _, k := range t.Selected(r) {
		if n, err := strconv.ParseInt(k, 10, 64); err == nil {
			chosen = append(chosen, n)
		}
	}
	n, err := h.Store.SetState(r.Context(), chosen, state)
	if err != nil {
		return err
	}
	return components.Reply(w, r).Toast(i18n.Static(L10nReviewed.T(r.Context(), n))).Refresh(t.ID()).Else(adminroutes.Talks())
}

// MoveForm is the dialog's form: any state, as a select of the enum.
type MoveForm struct {
	*forms.Form
	State *forms.Field[models.TalkState]
}

func NewMoveForm() *MoveForm {
	f := &MoveForm{Form: forms.New()}
	f.State = f.Enum("state", models.AllTalkStateValues()).Label(L10nNewState).Required()
	return f
}

//h:key talks.move
//h:method GET
//h:path /talks/{id}/state
//h:perm talks.review
//h:param id label Talk number
//h:desc Change a talk's state
func (h *H) TalksMove(w http.ResponseWriter, r *http.Request, id ids.TalkID) error {
	t, err := h.Store.TalkView(r.Context(), id)
	if err != nil {
		return err
	}
	f := NewMoveForm()
	f.State.Set(t.State)
	return h.dialogOrPage(w, r, http.StatusOK, L10nMove, moveBody(t, f))
}

//h:key talks.move_put
//h:method PUT
//h:path /talks/{id}/state
//h:perm talks.review
//h:param id label Talk number
//h:desc Save a talk's state
func (h *H) TalksMovePut(w http.ResponseWriter, r *http.Request, id ids.TalkID) error {
	ctx := r.Context()
	t, err := h.Store.TalkView(ctx, id)
	if err != nil {
		return err
	}
	f := NewMoveForm()
	if !f.Submit(ctx, r) {
		return h.dialogOrPage(w, r, http.StatusUnprocessableEntity, L10nMove, moveBody(t, f))
	}
	t.State = f.State.Value()
	if t.State == models.TalkStateScheduled {
		if err := confab.CanSchedule(models.TalkView{State: models.TalkStateAccepted, StartsAt: t.StartsAt}); err != nil {
			f.State.Fail(err)
			return h.dialogOrPage(w, r, http.StatusConflict, L10nMove, moveBody(t, f))
		}
	}
	if _, err := h.Store.SetState(ctx, []int64{int64(id)}, t.State); err != nil {
		return err
	}
	h.talksTable()
	return components.Reply(w, r).Toast(L10nStateChanged).
		Replace(h.talks.RowID(strconv.FormatInt(int64(id), 10)), h.talksView.Row(t)).Else(adminroutes.Talks())
}

//h:key talks.delete
//h:method DELETE
//h:path /talks/{id}
//h:perm talks.manage
//h:param id label Talk number
//h:desc Delete a talk
func (h *H) TalksDelete(w http.ResponseWriter, r *http.Request, id ids.TalkID) error {
	ctx := r.Context()
	fs, err := h.Store.ListTalkFiles(ctx, int64(id))
	if err != nil {
		return err
	}
	if err := h.Store.DeleteTalk(ctx, int64(id)); err != nil {
		return err
	}
	for _, f := range fs {
		_ = h.Files.Delete(f.Key)
	}
	h.talksTable()
	return components.Reply(w, r).Toast(L10nTalkDeleted).
		Remove(h.talks.RowID(strconv.FormatInt(int64(id), 10))).Else(adminroutes.Talks())
}

// dialogOrPage answers a form for ui's action dialog alone, or in a page
// when it was opened without script.
func (h *H) dialogOrPage(w http.ResponseWriter, r *http.Request, status int, title i18n.Ter, body templ.Component) error {
	if components.InActionDialog(r) {
		return render(w, r, status, body)
	}
	return h.page(w, r, status, title, body)
}

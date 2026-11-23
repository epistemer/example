package admin

import (
	"context"
	"encoding/csv"
	"io"
	"mime/multipart"
	"net/http"
	"strconv"
	"strings"

	"github.com/epistemer/errors"
	"github.com/epistemer/i18n"
	"github.com/epistemer/ui"
	"github.com/epistemer/ui/actions"
	"github.com/epistemer/ui/components"
	"github.com/epistemer/ui/forms"
	"github.com/epistemer/ui/forms/files"
	"github.com/epistemer/ui/forms/rule"
	"github.com/epistemer/ui/icons"
	"github.com/epistemer/ui/tables"

	"github.com/epistemer/example/internal/admin/adminroutes"
	"github.com/epistemer/example/internal/confab"
	"github.com/epistemer/example/internal/ids"
	"github.com/epistemer/example/internal/models"
	"github.com/epistemer/example/internal/pg"
)

var (
	L10nNewSpeaker     = i18n.Key("page.speakers.new", "New speaker")
	L10nPhotoHelp      = i18n.Key("page.speakers.photo_help", "A square PNG, JPEG or WebP, up to 2 MB.")
	L10nConsent        = i18n.Key("page.speakers.consent", "The speaker accepted the code of conduct")
	L10nSpeakerSaved   = i18n.Key("page.speakers.saved", "The speaker was saved.")
	L10nSpeakerDeleted = i18n.Key("page.speakers.deleted", "The speaker was deleted.")
	L10nDeleteSpeakerQ = i18n.Key("page.speakers.delete_question", "Delete this speaker?")
	L10nImport         = i18n.Key("page.speakers.import", "Import speakers")
	L10nImportCSV      = i18n.Key("page.speakers.import_csv", "Speakers (CSV)")
	L10nImportCSVHelp  = i18n.Key("page.speakers.import_csv_help", "Columns: name, email, country. A first line of headings is skipped.")
	L10nImportPhotos   = i18n.Key("page.speakers.import_photos", "Photos")
	L10nImportPhotosH  = i18n.Key("page.speakers.import_photos_help", "Named after the speaker's email address, such as ada@example.org.png.")
	L10nImported       = i18n.Key("page.speakers.imported", "Speakers imported: %d, photos: %d.")
)

// SpeakerForm shows a stored upload (the photo) beside the new-file input.
type SpeakerForm struct {
	*forms.Form
	Name    *forms.Field[string]
	Email   *forms.Field[string]
	Phone   *forms.Field[string]
	Country *forms.Field[string]
	Website *forms.Field[string]
	Bio     *forms.Field[string]
	Photo   *forms.Field[files.Upload]
	Consent *forms.Field[bool]
}

// NewSpeakerForm takes the photo the speaker has, if any: forms needs the
// stored files when the form is declared, to accept their keys back.
func NewSpeakerForm(photo []files.Stored) *SpeakerForm {
	f := &SpeakerForm{Form: forms.New()}
	f.Name = f.Text("name").Label(models.L10nSpeaker_Name).Required().Is(rule.MaxLen(120))
	f.Email = f.Text("email").Label(models.L10nSpeaker_Email).Required().Is(rule.Email())
	f.Phone = f.Text("phone").Label(models.L10nSpeaker_Phone).Is(rule.Phone()).Placeholder(i18n.Static("+31201234567"))
	f.Country = f.Text("country").Label(models.L10nSpeaker_Country).Required().Is(rule.CountryCode())
	f.Website = f.Text("website").Label(models.L10nSpeaker_Website).Is(rule.URL())
	f.Bio = f.Text("bio").Hint(forms.KindMultiline).Label(models.L10nSpeaker_Bio).Is(rule.MaxLen(1500))
	f.Photo = f.Upload("photo", photo).Label(models.L10nSpeaker_PhotoKey).Help(L10nPhotoHelp).
		Is(rule.Files(rule.MaxSize(2<<20), rule.ContentType("image/png", "image/jpeg", "image/webp")))
	f.Consent = f.Checkbox("consent").Label(L10nConsent).Required()
	return f
}

func storedPhoto(s models.Speaker) []files.Stored {
	if !s.PhotoKey.Valid {
		return nil
	}
	return []files.Stored{{
		Key: s.PhotoKey.String, Name: s.PhotoName.String,
		Href: adminroutes.File(s.PhotoKey.String), Image: true,
	}}
}

func (f *SpeakerForm) Fill(s models.Speaker) {
	f.Name.Set(s.Name)
	f.Email.Set(s.Email)
	f.Phone.SetPtr(pg.TextPtr(s.Phone))
	f.Country.Set(s.Country)
	f.Website.SetPtr(pg.TextPtr(s.Website))
	f.Bio.Set(s.Bio)
	f.Consent.Set(true)
}

// save stores the new photo, if any, and returns the input and the keys
// of files to delete once the row is saved.
func (f *SpeakerForm) save(fs *confab.Files, old models.Speaker) (confab.SpeakerInput, []string, error) {
	in := confab.SpeakerInput{
		Name: f.Name.Value(), Email: f.Email.Value(), Country: f.Country.Value(), Bio: f.Bio.Value(),
		Phone: f.Phone.Ptr(), Website: f.Website.Ptr(),
	}
	up := f.Photo.Value()
	var drop []string
	switch {
	case len(up.Added) == 1:
		stored, err := fs.Put(up.Added[0])
		if err != nil {
			return in, nil, err
		}
		in.Photo = &confab.StoredFile{Key: stored.Key, Name: stored.Name}
	case len(up.Kept) == 1:
		in.Photo = &confab.StoredFile{Key: up.Kept[0], Name: old.PhotoName.String}
	}
	if old.PhotoKey.Valid && (in.Photo == nil || in.Photo.Key != old.PhotoKey.String) {
		drop = append(drop, old.PhotoKey.String)
	}
	return in, drop, nil
}

func (f *SpeakerForm) fail(err error) {
	if errors.Is(err, confab.ErrEmailTaken) {
		f.Email.Fail(err)
		return
	}
	f.Form.Fail(err)
}

// The speakers table and its actions.
var sc = models.SpeakerViewColumns

var SpeakersTable = tables.New("speakers", sc.ID,
	tables.Col(sc.Name).Searchable().Sortable().Link(func(s models.SpeakerView) ui.Link { return adminroutes.SpeakersEdit(ids.SpeakerID(s.ID)) }),
	tables.Col(sc.Email).Searchable(),
	tables.Col(sc.Country).Filter().Sortable(),
	tables.Col(sc.Website).Link(func(s models.SpeakerView) ui.Link { return ui.URL(s.Website.String) }),
	tables.Col(sc.Talks).Sortable().Filter(),
	tables.Col(sc.CreatedAt).Sortable(),
).
	Label(adminroutes.L10nSpeakers).
	RowLabel(sc.Name).
	Sort(sc.Name, tables.Asc).
	Actions(speakerActions)

var SpeakersView = components.TableOf(SpeakersTable)

var (
	EditSpeaker   = actions.New(L10nEdit).Icon(icons.Pencil())
	DeleteSpeaker = actions.New(L10nDelete).Danger().Confirm(L10nDeleteSpeakerQ)
)

func speakerActions(ctx context.Context, s models.SpeakerView) []actions.Item {
	id := ids.SpeakerID(s.ID)
	var busy error
	if s.Talks > 0 {
		busy = confab.ErrHasTalks.With(int(s.Talks))
	}
	return actions.Resolve(ctx,
		EditSpeaker.To(adminroutes.SpeakersEdit(id)),
		DeleteSpeaker.To(adminroutes.SpeakersDelete(id)).Unless(busy),
	)
}

// Speakers is the listing.Labeler of a speaker id: the remote select on
// the talk form and the talks table's speaker filter name the speaker.
type speakerLabels struct{ s *confab.Store }

func (l speakerLabels) LabelOf(ctx context.Context, id string) (string, bool, error) {
	n, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		return "", false, nil
	}
	sp, err := l.s.Speaker(ctx, ids.SpeakerID(n))
	if errors.Is(err, confab.ErrNotFound) {
		return "", false, nil
	}
	if err != nil {
		return "", false, err
	}
	return sp.Name, true, nil
}

//h:key speakers
//h:method GET
//h:path /speakers
//h:perm speakers.manage
//h:view
//h:search
//h:desc Speakers
func (h *H) Speakers(w http.ResponseWriter, r *http.Request) error {
	st := SpeakersTable.Bind(r)
	v := st.Rejected()
	if st.Valid() {
		rows, total, err := h.Store.ListSpeakerView(r.Context(), st)
		if err != nil {
			return err
		}
		v = st.Page(rows, total)
	}
	status := st.Answer(w, r)
	if tables.Targeted(r, SpeakersTable) {
		return render(w, r, status, SpeakersView.Page(v))
	}
	return h.page(w, r, status, adminroutes.L10nSpeakers, speakersPage(SpeakersView.Page(v)))
}

//h:key speakers.search
//h:method GET
//h:path /speakers/search
//h:perm talks.manage
//h:desc Find a speaker
func (h *H) SpeakersSearch(w http.ResponseWriter, r *http.Request) error {
	rows, err := h.Store.SearchSpeakers(r.Context(), models.SearchSpeakersParams{
		Q: r.URL.Query().Get("q"), Lim: components.SuggestionLimit + 1,
	})
	if err != nil {
		return err
	}
	out := make([]components.Suggestion, len(rows))
	for i, s := range rows {
		out[i] = components.Suggestion{Value: strconv.FormatInt(s.ID, 10), Label: i18n.Static(s.Name + " <" + s.Email + ">")}
	}
	return render(w, r, http.StatusOK, components.Suggestions(out...))
}

//h:key speakers.new
//h:method GET
//h:path /speakers/new
//h:perm speakers.manage
//h:search
//h:entity speakers
//h:desc New speaker
func (h *H) SpeakersNew(w http.ResponseWriter, r *http.Request) error {
	return h.page(w, r, http.StatusOK, L10nNewSpeaker, speakerFormPage(NewSpeakerForm(nil), 0))
}

//h:key speakers.create
//h:method POST
//h:path /speakers
//h:perm speakers.manage
//h:body 3MiB
//h:desc Create a speaker
func (h *H) SpeakersCreate(w http.ResponseWriter, r *http.Request) error {
	return h.saveSpeaker(w, r, 0, models.Speaker{})
}

//h:key speakers.edit
//h:method GET
//h:path /speakers/{id}
//h:perm speakers.manage
//h:entity speakers
//h:param id label Speaker number
//h:desc Edit a speaker
func (h *H) SpeakersEdit(w http.ResponseWriter, r *http.Request, id ids.SpeakerID) error {
	s, err := h.Store.Speaker(r.Context(), id)
	if err != nil {
		return err
	}
	f := NewSpeakerForm(storedPhoto(s))
	f.Fill(s)
	return h.page(w, r, http.StatusOK, i18n.Static(s.Name), speakerFormPage(f, id))
}

//h:key speakers.update
//h:method POST
//h:path /speakers/{id}
//h:perm speakers.manage
//h:body 3MiB
//h:param id label Speaker number
//h:desc Save a speaker
func (h *H) SpeakersUpdate(w http.ResponseWriter, r *http.Request, id ids.SpeakerID) error {
	s, err := h.Store.Speaker(r.Context(), id)
	if err != nil {
		return err
	}
	return h.saveSpeaker(w, r, id, s)
}

func (h *H) saveSpeaker(w http.ResponseWriter, r *http.Request, id ids.SpeakerID, old models.Speaker) error {
	ctx := r.Context()
	title := L10nNewSpeaker
	if id != 0 {
		title = i18n.Static(old.Name)
	}
	f := NewSpeakerForm(storedPhoto(old))
	if !f.Submit(ctx, r) {
		var big *http.MaxBytesError
		if errors.As(f.Err(), &big) {
			return f.Err() // handlers answers 413; the script keeps the files
		}
		return h.page(w, r, http.StatusUnprocessableEntity, title, speakerFormPage(f, id))
	}
	in, drop, err := f.save(h.Files, old)
	if err == nil {
		if id == 0 {
			id, err = h.Store.CreateSpeaker(ctx, in)
		} else {
			err = h.Store.UpdateSpeaker(ctx, id, in)
		}
	}
	if err != nil {
		if in.Photo != nil && in.Photo.Key != old.PhotoKey.String {
			_ = h.Files.Delete(in.Photo.Key)
		}
		f.fail(err)
		return h.page(w, r, handlersStatus(err), title, speakerFormPage(f, id))
	}
	_ = h.Files.Delete(drop...)
	return components.Reply(w, r).Toast(L10nSpeakerSaved).Go(adminroutes.SpeakersEdit(id))
}

//h:key speakers.delete
//h:method DELETE
//h:path /speakers/{id}
//h:perm speakers.manage
//h:param id label Speaker number
//h:desc Delete a speaker
func (h *H) SpeakersDelete(w http.ResponseWriter, r *http.Request, id ids.SpeakerID) error {
	ctx := r.Context()
	s, err := h.Store.Speaker(ctx, id)
	if err != nil {
		return err
	}
	if err := h.Store.DeleteSpeaker(ctx, id); err != nil {
		return components.Reply(w, r).Fail(err).Else(adminroutes.Speakers())
	}
	if s.PhotoKey.Valid {
		_ = h.Files.Delete(s.PhotoKey.String)
	}
	return components.Reply(w, r).Toast(L10nSpeakerDeleted).
		Remove(SpeakersTable.RowID(strconv.FormatInt(int64(id), 10))).Else(adminroutes.Speakers())
}

// ImportForm takes one CSV (a File) and any number of photos (Files):
// raw uploads the handler reads once and doesn't keep as they are.
type ImportForm struct {
	*forms.Form
	CSV    *forms.Field[*multipart.FileHeader]
	Photos *forms.Field[[]*multipart.FileHeader]
}

func NewImportForm() *ImportForm {
	f := &ImportForm{Form: forms.New()}
	f.CSV = f.File("speakers").Label(L10nImportCSV).Help(L10nImportCSVHelp).Required().
		Is(rule.MaxSize(1<<20), rule.ContentType("text/plain"))
	f.Photos = f.Files("photos").Label(L10nImportPhotos).Help(L10nImportPhotosH).
		Is(rule.MaxItems[*multipart.FileHeader](50), rule.Each(rule.MaxSize(2<<20), rule.ContentType("image/*")))
	return f
}

var ErrBadCSV = errors.Key1[int]("errors.confab.bad_csv", "line %d needs a name, an email address and a country", errors.Invalid)

//h:key speakers.import
//h:method GET
//h:path /speakers/import
//h:perm speakers.manage
//h:search
//h:entity speakers
//h:desc Import speakers
func (h *H) SpeakersImport(w http.ResponseWriter, r *http.Request) error {
	return h.page(w, r, http.StatusOK, L10nImport, importPage(NewImportForm(), 0, 0))
}

//h:key speakers.import_post
//h:method POST
//h:path /speakers/import
//h:perm speakers.manage
//h:body 64MiB
//h:desc Read an import
func (h *H) SpeakersImportPost(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()
	f := NewImportForm()
	if !f.Submit(ctx, r) {
		var big *http.MaxBytesError
		if errors.As(f.Err(), &big) {
			return f.Err()
		}
		return h.page(w, r, http.StatusUnprocessableEntity, L10nImport, importPage(f, 0, 0))
	}
	photos := map[string]*multipart.FileHeader{}
	for _, p := range f.Photos.Value() {
		name := strings.ToLower(p.Filename)
		photos[strings.TrimSuffix(name, name[strings.LastIndex(name, "."):])] = p
	}
	src, err := f.CSV.Value().Open()
	if err != nil {
		return err
	}
	defer src.Close()
	rd := csv.NewReader(src)
	rd.FieldsPerRecord = -1
	people, withPhoto := 0, 0
	for line := 1; ; line++ {
		rec, err := rd.Read()
		if err == io.EOF {
			break
		}
		if err != nil || len(rec) < 3 {
			f.CSV.Fail(ErrBadCSV.With(line))
			return h.page(w, r, http.StatusUnprocessableEntity, L10nImport, importPage(f, people, withPhoto))
		}
		if line == 1 && strings.EqualFold(rec[1], "email") {
			continue
		}
		in := confab.SpeakerInput{Name: rec[0], Email: rec[1], Country: strings.ToUpper(rec[2])}
		if p, ok := photos[strings.ToLower(in.Email)]; ok {
			stored, err := h.Files.Put(p)
			if err != nil {
				return err
			}
			in.Photo = &confab.StoredFile{Key: stored.Key, Name: stored.Name}
			withPhoto++
		}
		if _, err := h.Store.CreateSpeaker(ctx, in); err != nil {
			f.CSV.Fail(err)
			return h.page(w, r, handlersStatus(err), L10nImport, importPage(f, people, withPhoto))
		}
		people++
	}
	return h.page(w, r, http.StatusOK, L10nImport, importPage(NewImportForm(), people, withPhoto))
}

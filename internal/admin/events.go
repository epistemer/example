package admin

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/epistemer/errors"
	"github.com/epistemer/i18n"
	"github.com/epistemer/models/option"
	"github.com/epistemer/ui"
	"github.com/epistemer/ui/actions"
	"github.com/epistemer/ui/components"
	"github.com/epistemer/ui/forms"
	"github.com/epistemer/ui/forms/rule"
	"github.com/epistemer/ui/icons"
	"github.com/epistemer/ui/tables"
	"github.com/shopspring/decimal"

	"github.com/epistemer/example/internal/admin/adminroutes"
	"github.com/epistemer/example/internal/confab"
	"github.com/epistemer/example/internal/ids"
	"github.com/epistemer/example/internal/models"
	"github.com/epistemer/example/internal/pg"
)

var (
	L10nVenue       = i18n.Key("page.events.venue", "Venue")
	L10nVenueHelp   = i18n.Key("page.events.venue_help", "Where people go. Not asked for an online event.")
	L10nTiers       = i18n.Key("page.events.tiers", "Ticket tiers")
	L10nSlugHelp    = i18n.Key("page.events.slug_help", "The event's address: lower-case words joined by hyphens.")
	L10nCreated     = i18n.Key("page.events.created", "The event was created.")
	L10nSaved       = i18n.Key("page.events.saved", "The event was saved.")
	L10nDeleted     = i18n.Key("page.events.deleted", "The event was deleted.")
	L10nPublished   = i18n.Key("page.events.published", "The event was published.")
	L10nArchived    = i18n.Key("page.events.archived", "The event was archived.")
	L10nEdit        = i18n.Key("page.actions.edit", "Edit")
	L10nDelete      = i18n.Key("page.actions.delete", "Delete")
	L10nPublish     = i18n.Key("page.events.publish", "Publish")
	L10nArchive     = i18n.Key("page.events.archive", "Archive")
	L10nDeleteEvent = i18n.Key("page.events.delete_question", "Delete this event, its tiers and its talks?")
	L10nArchiveQ    = i18n.Key("page.events.archive_question", "Archive this event? Its page stays, without tickets.")
	L10nSave        = i18n.Key("page.actions.save", "Save")
	L10nCreate      = i18n.Key("page.actions.create", "Create")
	L10nNewEvent    = i18n.Key("page.events.new", "New event")
	L10nPublicPage  = i18n.Key("page.events.public_page", "Public page")
)

// VenueForm is the group "venue": active only for an event that isn't
// online.
type VenueForm struct {
	*forms.Form
	Name      *forms.Field[string]
	Street    *forms.Field[string]
	City      *forms.Field[string]
	Country   *forms.Field[string]
	PostCode  *forms.Field[string]
	Latitude  *forms.Field[float64]
	Longitude *forms.Field[float64]
}

func newVenue(g *forms.Form) *VenueForm {
	v := &VenueForm{Form: g}
	v.Name = g.Text("name").Label(models.L10nEvent_VenueName).Required().Is(rule.MaxLen(120))
	v.Street = g.Text("street").Label(models.L10nEvent_VenueStreet).Required()
	v.City = g.Text("city").Label(models.L10nEvent_VenueCity).Required()
	v.Country = g.Text("country").Label(models.L10nEvent_VenueCountry).Required().Is(rule.CountryCode()).
		Placeholder(i18n.Static("NL"))
	v.PostCode = g.Text("post_code").Label(models.L10nEvent_VenuePostCode)
	v.Latitude = g.Float("latitude").Label(models.L10nEvent_Latitude).Is(rule.Between(-90.0, 90.0))
	v.Longitude = g.Float("longitude").Label(models.L10nEvent_Longitude).Is(rule.Between(-180.0, 180.0))
	// The post code's pattern depends on the country the person chose:
	// a check of the group, since a rule takes a fixed country.
	g.Check(func(ctx context.Context) error {
		if err := postCode(ctx, v.Country.Value(), v.PostCode.Value()); err != nil {
			v.PostCode.Fail(err)
		}
		return nil
	})
	return v
}

// postCode checks a post code against a country chosen at run time.
// rule.PostCode panics for a country outside its table (findings F9), so
// the panic is turned into "no pattern known".
func postCode(ctx context.Context, country, code string) (err error) {
	if code == "" {
		return nil
	}
	defer func() {
		if recover() != nil {
			err = nil
		}
	}()
	return rule.Check(ctx, code, rule.PostCode(country))
}

// TierForm is one item of the list "tiers".
type TierForm struct {
	*forms.Form
	Name       *forms.Field[string]
	Price      *forms.Field[decimal.Decimal]
	Quantity   *forms.Field[int]
	SaleEndsOn *forms.Field[time.Time]
}

func newTier(g *forms.Form) *TierForm {
	t := &TierForm{Form: g}
	t.Name = g.Text("name").Label(models.L10nTicketTier_Name).Required().Is(rule.MaxLen(60))
	t.Price = g.Parsed[decimal.Decimal]("price").Label(models.L10nTicketTier_Price).Required().
		Is(rule.AtLeast(decimal.Zero), rule.AtMost(decimal.NewFromInt(10000)))
	t.Quantity = g.Int("quantity").Label(models.L10nTicketTier_Quantity).Required().Is(rule.Min(1))
	t.SaleEndsOn = g.Date("sale_ends_on").Label(models.L10nTicketTier_SaleEndsOn)
	return t
}

// EventForm declares the event's form once: the new page, the edit page
// and their handlers share it.
type EventForm struct {
	*forms.Form
	Name     *forms.Field[string]
	Slug     *forms.Field[ids.Slug]
	Summary  *forms.Field[string]
	Status   *forms.Field[models.EventStatus]
	StartsOn *forms.Field[time.Time]
	EndsOn   *forms.Field[time.Time]
	Online   *forms.Field[bool]
	Capacity *forms.Field[int32]
	Venue    *VenueForm
	Currency *forms.Field[string]
	Tiers    *forms.List[*TierForm]
	Version  *forms.Field[int]
}

func NewEventForm() *EventForm {
	f := &EventForm{Form: forms.New()}
	f.Name = f.Text("name").Label(models.L10nEvent_Name).Required().Is(rule.MinLen(3), rule.MaxLen(120))
	f.Slug = f.Parsed[ids.Slug]("slug").Label(models.L10nEvent_Slug).Help(L10nSlugHelp).Required()
	f.Summary = f.Text("summary").Hint(forms.KindMultiline).Label(models.L10nEvent_Summary).Is(rule.MaxLen(2000))
	f.Status = f.Enum("status", models.AllEventStatusValues()).Label(models.L10nEvent_Status).Required()
	f.StartsOn = f.Date("starts_on").Label(models.L10nEvent_StartsOn).Required()
	f.EndsOn = f.Date("ends_on").Label(models.L10nEvent_EndsOn).Required()
	f.Online = f.Checkbox("online").Hint(forms.KindToggle).Label(models.L10nEvent_Online)
	f.Capacity = f.Number[int32]("capacity").Label(models.L10nEvent_Capacity).Required().
		Is(rule.Between[int32](1, 100_000))
	f.Venue = f.Group("venue", newVenue)
	f.Venue.Label(L10nVenue).When(f.Online.Eq(false))
	f.Currency = f.Text("currency").Label(models.L10nEvent_Currency).Required().Is(rule.CurrencyCode())
	f.Tiers = f.List("tiers", newTier).Label(L10nTiers).MinItems(1).MaxItems(8)
	f.Version = f.Int("version").Hint(forms.KindHidden)
	f.Check(func(context.Context) error {
		if f.EndsOn.Value().Before(f.StartsOn.Value()) {
			f.EndsOn.Fail(confab.ErrEndsFirst)
		}
		return nil
	})
	return f
}

// Input is what the form holds, for the store.
func (f *EventForm) Input() confab.EventInput {
	in := confab.EventInput{
		Name: f.Name.Value(), Slug: f.Slug.Value(), Summary: f.Summary.Value(), Status: f.Status.Value(),
		StartsOn: f.StartsOn.Value(), EndsOn: f.EndsOn.Value(), Online: f.Online.Value(),
		Capacity: f.Capacity.Value(), Currency: f.Currency.Value(),
	}
	if f.Venue.Active() {
		v := f.Venue
		in.Venue = &confab.Venue{
			Name: v.Name.Value(), Street: v.Street.Value(), City: v.City.Value(),
			Country: v.Country.Value(), PostCode: v.PostCode.Value(),
			Latitude: v.Latitude.Ptr(), Longitude: v.Longitude.Ptr(),
		}
	}
	for _, t := range f.Tiers.Items() {
		in.Tiers = append(in.Tiers, confab.Tier{
			Name: t.Name.Value(), Price: t.Price.Value(), Quantity: t.Quantity.Value(), SaleEndsOn: t.SaleEndsOn.Ptr(),
		})
	}
	return in
}

// Fill sets the form from a stored event, for the edit page.
func (f *EventForm) Fill(e models.Event, tiers []confab.Tier) {
	f.Name.Set(e.Name)
	f.Slug.Set(ids.Slug(e.Slug))
	f.Summary.Set(e.Summary)
	f.Status.Set(e.Status)
	f.StartsOn.Set(e.StartsOn.Time)
	f.EndsOn.Set(e.EndsOn.Time)
	f.Online.Set(e.Online)
	f.Capacity.Set(e.Capacity)
	f.Currency.Set(e.Currency)
	f.Version.Set(int(e.Version))
	if !e.Online {
		v := f.Venue
		v.Name.SetPtr(pg.TextPtr(e.VenueName))
		v.Street.SetPtr(pg.TextPtr(e.VenueStreet))
		v.City.SetPtr(pg.TextPtr(e.VenueCity))
		v.Country.SetPtr(pg.TextPtr(e.VenueCountry))
		v.PostCode.SetPtr(pg.TextPtr(e.VenuePostCode))
		v.Latitude.SetPtr(pg.FloatOf(e.Latitude))
		v.Longitude.SetPtr(pg.FloatOf(e.Longitude))
	}
	for _, t := range tiers {
		item := f.Tiers.Add()
		item.Name.Set(t.Name)
		item.Price.Set(t.Price)
		item.Quantity.Set(t.Quantity)
		item.SaleEndsOn.SetPtr(t.SaleEndsOn)
	}
}

// fail puts a store's error where it belongs: on a field when one field
// caused it, else on the form.
func (f *EventForm) fail(err error) {
	switch {
	case errors.Is(err, confab.ErrSlugTaken):
		f.Slug.Fail(err)
	case errors.Is(err, confab.ErrEndsFirst):
		f.EndsOn.Fail(err)
	default:
		f.Form.Fail(err)
	}
}

// The events table: its columns come from models' column set for
// event_view; searching, sorting and filtering are declared here.
var ec = models.EventViewColumns

var EventsTable = tables.New("events", ec.ID,
	tables.Col(ec.Name).Searchable().Sortable().Link(func(e models.EventView) ui.Link { return adminroutes.EventsEdit(ids.EventID(e.ID)) }),
	tables.Col(ec.Status).Filter(),
	tables.Col(ec.StartsOn).Sortable().Filter(),
	tables.Col(ec.EndsOn),
	tables.Col(ec.Online).Filter(),
	tables.Col(ec.VenueCity).Searchable(),
	tables.Col(ec.Capacity).Sortable().Filter(),
	tables.Col(ec.Talks).Sortable(),
	tables.Col(ec.FromPrice).Sortable().Amount(func(e models.EventView) i18n.Ter {
		if !e.FromPrice.Valid {
			return i18n.Static("")
		}
		return money(pg.Decimal(e.FromPrice), e.Currency)
	}),
).
	Label(adminroutes.L10nEvents).
	RowLabel(ec.Name).
	Sort(ec.StartsOn, tables.Desc).
	Sizes(10, 25, 50).
	Actions(eventActions)

var EventsView = components.TableOf(EventsTable).
	Cell(ec.Status, statusBadge(func(e models.EventView) option.SelectOption { return e.Status }))

var (
	EditEvent    = actions.New(L10nEdit)
	PublishEvent = actions.New(L10nPublish).Primary()
	ArchiveEvent = actions.New(L10nArchive).Confirm(L10nArchiveQ)
	DeleteEvent  = actions.New(L10nDelete).Danger().Confirm(L10nDeleteEvent)
	ViewEvent    = actions.New(L10nPublicPage).Icon(icons.Eye())
)

func eventActions(ctx context.Context, e models.EventView) []actions.Item {
	id := ids.EventID(e.ID)
	return actions.Resolve(ctx,
		EditEvent.To(adminroutes.EventsEdit(id)),
		ViewEvent.To(adminroutes.EventPage(ids.Slug(e.Slug))).If(e.Status == models.EventStatusPublished),
		PublishEvent.To(adminroutes.EventsStatus(id, models.EventStatusPublished)).If(e.Status == models.EventStatusDraft),
		ArchiveEvent.To(adminroutes.EventsStatus(id, models.EventStatusArchived)).If(e.Status == models.EventStatusPublished),
		DeleteEvent.To(adminroutes.EventsDelete(id)).Unless(canDeleteEvent(e)),
	)
}

var ErrEventHasTalks = errors.Key("errors.confab.event_has_talks", "an event with talks can't be deleted: archive it", errors.Conflict)

func canDeleteEvent(e models.EventView) error {
	if e.Talks > 0 {
		return ErrEventHasTalks
	}
	return nil
}

func money(d decimal.Decimal, currency string) i18n.Ter {
	return i18n.Static(currency + " " + d.StringFixed(2))
}

//h:key events
//h:method GET
//h:path /events
//h:perm events.manage
//h:view
//h:search
//h:desc Events
func (h *H) Events(w http.ResponseWriter, r *http.Request) error {
	st := EventsTable.Bind(r)
	v := st.Rejected()
	if st.Valid() {
		rows, total, err := h.Store.ListEventView(r.Context(), st)
		if err != nil {
			return err
		}
		v = st.Page(rows, total)
	}
	status := st.Answer(w, r)
	if tables.Targeted(r, EventsTable) {
		return render(w, r, status, EventsView.Page(v))
	}
	return h.page(w, r, status, adminroutes.L10nEvents, eventsPage(EventsView.Page(v)))
}

//h:key events.new
//h:method GET
//h:path /events/new
//h:perm events.manage
//h:search
//h:entity events
//h:desc New event
func (h *H) EventsNew(w http.ResponseWriter, r *http.Request) error {
	f := NewEventForm()
	f.Status.Set(models.EventStatusDraft)
	f.Currency.Set("EUR")
	f.Tiers.Add().Name.Set("Regular")
	return h.page(w, r, http.StatusOK, L10nNewEvent, eventFormPage(f, 0))
}

//h:key events.create
//h:method POST
//h:path /events
//h:perm events.manage
//h:desc Create an event
func (h *H) EventsCreate(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()
	f := NewEventForm()
	if !f.Submit(ctx, r) {
		return h.page(w, r, http.StatusUnprocessableEntity, L10nNewEvent, eventFormPage(f, 0))
	}
	id, err := h.Store.CreateEvent(ctx, f.Input())
	if err != nil {
		f.fail(err)
		return h.page(w, r, http.StatusUnprocessableEntity, L10nNewEvent, eventFormPage(f, 0))
	}
	return components.Reply(w, r).Toast(L10nCreated).Go(adminroutes.EventsEdit(id))
}

//h:key events.edit
//h:method GET
//h:path /events/{id}/edit
//h:perm events.manage
//h:entity events
//h:param id label Event number
//h:desc Edit an event
func (h *H) EventsEdit(w http.ResponseWriter, r *http.Request, id ids.EventID) error {
	ctx := r.Context()
	e, err := h.Store.Event(ctx, id)
	if err != nil {
		return err
	}
	tiers, err := h.Store.Tiers(ctx, id)
	if err != nil {
		return err
	}
	f := NewEventForm()
	f.Fill(e, tiers)
	return h.page(w, r, http.StatusOK, i18n.Static(e.Name), eventFormPage(f, id))
}

//h:key events.update
//h:method POST
//h:path /events/{id}/edit
//h:perm events.manage
//h:param id label Event number
//h:desc Save an event
func (h *H) EventsUpdate(w http.ResponseWriter, r *http.Request, id ids.EventID) error {
	ctx := r.Context()
	f := NewEventForm()
	if !f.Submit(ctx, r) {
		return h.page(w, r, http.StatusUnprocessableEntity, i18n.Static(f.Name.Value()), eventFormPage(f, id))
	}
	if err := h.Store.UpdateEvent(ctx, id, int32(f.Version.Value()), f.Input()); err != nil {
		f.fail(err)
		return h.page(w, r, http.StatusUnprocessableEntity, i18n.Static(f.Name.Value()), eventFormPage(f, id))
	}
	return components.Reply(w, r).Toast(L10nSaved).Go(adminroutes.EventsEdit(id))
}

//h:key events.status
//h:method PATCH
//h:path /events/{id}/status/{status}
//h:perm events.manage
//h:param id label Event number
//h:param status label New status
//h:desc Publish or archive an event
func (h *H) EventsStatus(w http.ResponseWriter, r *http.Request, id ids.EventID, status models.EventStatus) error {
	ctx := r.Context()
	if err := h.Store.SetEventStatus(ctx, models.SetEventStatusParams{ID: int64(id), Status: status}); err != nil {
		return err
	}
	rows, _, err := h.Store.ListEventView(ctx, eventByID(id))
	if err != nil || len(rows) == 0 {
		return components.Reply(w, r).Toast(L10nSaved).Refresh(EventsTable.ID()).Else(adminroutes.Events())
	}
	done := L10nPublished
	if status == models.EventStatusArchived {
		done = L10nArchived
	}
	return components.Reply(w, r).Toast(done).
		Replace(EventsTable.RowID(strconv.FormatInt(int64(id), 10)), EventsView.Row(rows[0])).
		Else(adminroutes.Events())
}

//h:key events.delete
//h:method DELETE
//h:path /events/{id}
//h:perm events.manage
//h:param id label Event number
//h:desc Delete an event
func (h *H) EventsDelete(w http.ResponseWriter, r *http.Request, id ids.EventID) error {
	ctx := r.Context()
	rows, _, err := h.Store.ListEventView(ctx, eventByID(id))
	if err != nil {
		return err
	}
	if len(rows) == 0 {
		return confab.ErrNotFound
	}
	if err := canDeleteEvent(rows[0]); err != nil {
		return components.Reply(w, r).Fail(err).Else(adminroutes.Events())
	}
	if err := h.Store.DeleteEvent(ctx, int64(id)); err != nil {
		return err
	}
	return components.Reply(w, r).Toast(L10nDeleted).
		Remove(EventsTable.RowID(strconv.FormatInt(int64(id), 10))).Else(adminroutes.Events())
}

//h:key event_page
//h:method GET
//h:path /e/{slug}
//h:auth public
//h:param slug label Event address
//h:desc Event programme
func (h *H) EventPage(w http.ResponseWriter, r *http.Request, slug ids.Slug) error {
	ctx := r.Context()
	e, err := h.Store.EventBySlug(ctx, slug)
	if err != nil {
		return err
	}
	if e.Status != models.EventStatusPublished {
		return confab.ErrPublishedOnly
	}
	talks, err := h.Store.PublicSchedule(ctx, e.ID)
	if err != nil {
		return err
	}
	tiers, err := h.Store.Tiers(ctx, ids.EventID(e.ID))
	if err != nil {
		return err
	}
	return render(w, r, http.StatusOK, publicEventPage(r, e, tiers, talks))
}

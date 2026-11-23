package confab

import (
	"github.com/epistemer/errors"
	"github.com/epistemer/i18n"
)

// The domain's errors people see. Each is an i18n key under
// errors.confab, with a kind and so an HTTP status.
var (
	ErrNotFound      = errors.Key("errors.confab.not_found", "it doesn't exist, or it was deleted", errors.NotFound)
	ErrSlugTaken     = errors.Key1[string]("errors.confab.slug_taken", "the address %s is taken by another event", errors.Conflict)
	ErrEmailTaken    = errors.Key("errors.confab.email_taken", "another speaker has this email address", errors.Conflict)
	ErrStale         = errors.Key("errors.confab.stale", "someone changed this event while you edited it: reload it and try again", errors.Conflict)
	ErrEndsFirst     = errors.Key("errors.confab.ends_first", "the last day can't be before the first", errors.Invalid)
	ErrCFPClosed     = errors.Key("errors.confab.cfp_closed", "the call for papers is closed", errors.Conflict)
	ErrTooManyTalks  = errors.Key1[int]("errors.confab.too_many_talks", "a speaker may give at most %d talks at one event", errors.Conflict)
	ErrOverlap       = errors.Key1[string]("errors.confab.overlap", "the speaker gives %s at that time", errors.Conflict)
	ErrOutsideEvent  = errors.Key2[i18n.Ter, i18n.Ter]("errors.confab.outside_event", "the talk must start between %s and %s", errors.Invalid)
	ErrNotScheduled  = errors.Key("errors.confab.not_scheduled", "only an accepted talk with a start time can be scheduled", errors.Conflict)
	ErrHasTalks      = errors.Key1[int]("errors.confab.has_talks", "the speaker has %d talks: remove them first", errors.Conflict)
	ErrNotSaved      = errors.Key("errors.confab.not_saved", "it couldn't be saved")
	ErrNeedsVenue    = errors.Key("errors.confab.needs_venue", "an event that isn't online needs a venue", errors.Invalid)
	ErrPublishedOnly = errors.Key("errors.confab.published_only", "only a published event has a public schedule", errors.NotFound)
)

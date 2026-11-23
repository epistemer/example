// Code generated from labels by github.com/epistemer/i18n. DO NOT EDIT.

package apiroutes

import "github.com/epistemer/i18n"

var (
	L10nGetSchedule     = i18n.Key("route.api.get_schedule", "An event's public schedule")
	L10nGetScheduleSlug = i18n.Key("route_param.api.get_schedule.slug", "Slug")
	L10nGetMe           = i18n.Key("route.api.get_me", "Who is signed in")
	L10nRefresh         = i18n.Key("route.api.refresh", "Refresh tokens")
	L10nSignIn          = i18n.Key("route.api.sign_in", "Sign in")
	L10nListTalks       = i18n.Key("route.api.list_talks", "List talks")
	L10nUploadSlides    = i18n.Key("route.api.upload_slides", "Add slides to a talk")
	L10nUploadSlidesId  = i18n.Key("route_param.api.upload_slides.id", "Id")
	L10nSetTalkState    = i18n.Key("route.api.talks.review", "Change a talk's state")
	L10nSetTalkStateId  = i18n.Key("route_param.api.talks.review.id", "Id")
)

// l10n holds every label that has a path.
var l10n = map[string]i18n.TKer{
	"route.api.get_schedule":            L10nGetSchedule,
	"route_param.api.get_schedule.slug": L10nGetScheduleSlug,
	"route.api.get_me":                  L10nGetMe,
	"route.api.refresh":                 L10nRefresh,
	"route.api.sign_in":                 L10nSignIn,
	"route.api.list_talks":              L10nListTalks,
	"route.api.upload_slides":           L10nUploadSlides,
	"route_param.api.upload_slides.id":  L10nUploadSlidesId,
	"route.api.talks.review":            L10nSetTalkState,
	"route_param.api.talks.review.id":   L10nSetTalkStateId,
}

// L10n returns the label at path, or nil when there is none.
func L10n(path string) i18n.TKer { return l10n[path] }

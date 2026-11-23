// Package ids holds the types route parameters are read into. They live
// outside the handler package so the generated routes package can import
// them (handlers refuses parameter types declared in the handler package).
package ids

import (
	"regexp"

	"github.com/epistemer/errors"
)

// Record identities, one type each, so a talk's id can't be passed where
// an event's is expected.
type (
	EventID   int64
	SpeakerID int64
	TalkID    int64
)

// Slug is an event's address name: lower-case words joined by hyphens.
// It reads itself from a path through encoding.TextUnmarshaler.
type Slug string

var slugPattern = regexp.MustCompile(`^[a-z0-9]+(-[a-z0-9]+)*$`)

// ErrBadSlug is a slug that isn't lower-case words joined by hyphens.
var ErrBadSlug = errors.Key("errors.confab.bad_slug", "use lower-case letters, digits and single hyphens", errors.Invalid)

func (s Slug) MarshalText() ([]byte, error) { return []byte(s), nil }

func (s *Slug) UnmarshalText(b []byte) error {
	if !slugPattern.Match(b) {
		return ErrBadSlug
	}
	*s = Slug(b)
	return nil
}

func (s Slug) String() string { return string(s) }

// SlugPattern is the slug's pattern, for a form rule.
func SlugPattern() *regexp.Regexp { return slugPattern }

// Locale is a language's BCP 47 code in a path, such as ar or pt-BR.
type Locale string

var localePattern = regexp.MustCompile(`^[a-z]{2,3}(-[A-Za-z0-9]{2,8})*$`)

var ErrBadLocale = errors.Key("errors.confab.bad_locale", "not a language code such as ar or pt-BR", errors.Invalid)

func (l Locale) MarshalText() ([]byte, error) { return []byte(l), nil }

func (l *Locale) UnmarshalText(b []byte) error {
	if !localePattern.Match(b) {
		return ErrBadLocale
	}
	*l = Locale(b)
	return nil
}

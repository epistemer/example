// Code generated from sqlc's request by github.com/epistemer/models. DO NOT EDIT.

package models

import (
	"fmt"

	"github.com/epistemer/i18n"
	"github.com/epistemer/models/option"
)

// Label returns the label of t, or nil when t isn't a value of EventStatus.
func (t EventStatus) Label() i18n.TKer {
	switch t {
	case EventStatusDraft:
		return L10nEventStatusDraft
	case EventStatusPublished:
		return L10nEventStatusPublished
	case EventStatusArchived:
		return L10nEventStatusArchived
	}
	return nil
}

// Value returns t as a string, for option.SelectOption.
func (t EventStatus) Value() string { return string(t) }

// MarshalText returns t's value, or an error when t isn't a value of
// EventStatus.
func (t EventStatus) MarshalText() ([]byte, error) {
	if t.Label() == nil {
		return nil, fmt.Errorf("models: %q is not a value of EventStatus", string(t))
	}
	return []byte(t), nil
}

// UnmarshalText sets t to text, or returns an error when text isn't a
// value of EventStatus.
func (t *EventStatus) UnmarshalText(text []byte) error {
	if EventStatus(text).Label() == nil {
		return fmt.Errorf("models: %q is not a value of EventStatus", string(text))
	}
	*t = EventStatus(text)
	return nil
}

// AllEventStatusOptions returns every value of EventStatus, in the
// schema's order, as select options.
func AllEventStatusOptions() []option.SelectOption {
	return []option.SelectOption{
		EventStatusDraft,
		EventStatusPublished,
		EventStatusArchived,
	}
}

// Label returns the label of t, or nil when t isn't a value of TalkLevel.
func (t TalkLevel) Label() i18n.TKer {
	switch t {
	case TalkLevelBeginner:
		return L10nTalkLevelBeginner
	case TalkLevelIntermediate:
		return L10nTalkLevelIntermediate
	case TalkLevelAdvanced:
		return L10nTalkLevelAdvanced
	}
	return nil
}

// Value returns t as a string, for option.SelectOption.
func (t TalkLevel) Value() string { return string(t) }

// MarshalText returns t's value, or an error when t isn't a value of
// TalkLevel.
func (t TalkLevel) MarshalText() ([]byte, error) {
	if t.Label() == nil {
		return nil, fmt.Errorf("models: %q is not a value of TalkLevel", string(t))
	}
	return []byte(t), nil
}

// UnmarshalText sets t to text, or returns an error when text isn't a
// value of TalkLevel.
func (t *TalkLevel) UnmarshalText(text []byte) error {
	if TalkLevel(text).Label() == nil {
		return fmt.Errorf("models: %q is not a value of TalkLevel", string(text))
	}
	*t = TalkLevel(text)
	return nil
}

// AllTalkLevelOptions returns every value of TalkLevel, in the
// schema's order, as select options.
func AllTalkLevelOptions() []option.SelectOption {
	return []option.SelectOption{
		TalkLevelBeginner,
		TalkLevelIntermediate,
		TalkLevelAdvanced,
	}
}

// Label returns the label of t, or nil when t isn't a value of TalkFormat.
func (t TalkFormat) Label() i18n.TKer {
	switch t {
	case TalkFormatTalk:
		return L10nTalkFormatTalk
	case TalkFormatWorkshop:
		return L10nTalkFormatWorkshop
	case TalkFormatLightning:
		return L10nTalkFormatLightning
	case TalkFormatPanel:
		return L10nTalkFormatPanel
	}
	return nil
}

// Value returns t as a string, for option.SelectOption.
func (t TalkFormat) Value() string { return string(t) }

// MarshalText returns t's value, or an error when t isn't a value of
// TalkFormat.
func (t TalkFormat) MarshalText() ([]byte, error) {
	if t.Label() == nil {
		return nil, fmt.Errorf("models: %q is not a value of TalkFormat", string(t))
	}
	return []byte(t), nil
}

// UnmarshalText sets t to text, or returns an error when text isn't a
// value of TalkFormat.
func (t *TalkFormat) UnmarshalText(text []byte) error {
	if TalkFormat(text).Label() == nil {
		return fmt.Errorf("models: %q is not a value of TalkFormat", string(text))
	}
	*t = TalkFormat(text)
	return nil
}

// AllTalkFormatOptions returns every value of TalkFormat, in the
// schema's order, as select options.
func AllTalkFormatOptions() []option.SelectOption {
	return []option.SelectOption{
		TalkFormatTalk,
		TalkFormatWorkshop,
		TalkFormatLightning,
		TalkFormatPanel,
	}
}

// Label returns the label of t, or nil when t isn't a value of TalkState.
func (t TalkState) Label() i18n.TKer {
	switch t {
	case TalkStateSubmitted:
		return L10nTalkStateSubmitted
	case TalkStateAccepted:
		return L10nTalkStateAccepted
	case TalkStateRejected:
		return L10nTalkStateRejected
	case TalkStateScheduled:
		return L10nTalkStateScheduled
	}
	return nil
}

// Value returns t as a string, for option.SelectOption.
func (t TalkState) Value() string { return string(t) }

// MarshalText returns t's value, or an error when t isn't a value of
// TalkState.
func (t TalkState) MarshalText() ([]byte, error) {
	if t.Label() == nil {
		return nil, fmt.Errorf("models: %q is not a value of TalkState", string(t))
	}
	return []byte(t), nil
}

// UnmarshalText sets t to text, or returns an error when text isn't a
// value of TalkState.
func (t *TalkState) UnmarshalText(text []byte) error {
	if TalkState(text).Label() == nil {
		return fmt.Errorf("models: %q is not a value of TalkState", string(text))
	}
	*t = TalkState(text)
	return nil
}

// AllTalkStateOptions returns every value of TalkState, in the
// schema's order, as select options.
func AllTalkStateOptions() []option.SelectOption {
	return []option.SelectOption{
		TalkStateSubmitted,
		TalkStateAccepted,
		TalkStateRejected,
		TalkStateScheduled,
	}
}

// Label returns the label of t, or nil when t isn't a value of TalkTopic.
func (t TalkTopic) Label() i18n.TKer {
	switch t {
	case TalkTopicGo:
		return L10nTalkTopicGo
	case TalkTopicWeb:
		return L10nTalkTopicWeb
	case TalkTopicData:
		return L10nTalkTopicData
	case TalkTopicSecurity:
		return L10nTalkTopicSecurity
	case TalkTopicOperations:
		return L10nTalkTopicOperations
	case TalkTopicDesign:
		return L10nTalkTopicDesign
	}
	return nil
}

// Value returns t as a string, for option.SelectOption.
func (t TalkTopic) Value() string { return string(t) }

// MarshalText returns t's value, or an error when t isn't a value of
// TalkTopic.
func (t TalkTopic) MarshalText() ([]byte, error) {
	if t.Label() == nil {
		return nil, fmt.Errorf("models: %q is not a value of TalkTopic", string(t))
	}
	return []byte(t), nil
}

// UnmarshalText sets t to text, or returns an error when text isn't a
// value of TalkTopic.
func (t *TalkTopic) UnmarshalText(text []byte) error {
	if TalkTopic(text).Label() == nil {
		return fmt.Errorf("models: %q is not a value of TalkTopic", string(text))
	}
	*t = TalkTopic(text)
	return nil
}

// AllTalkTopicOptions returns every value of TalkTopic, in the
// schema's order, as select options.
func AllTalkTopicOptions() []option.SelectOption {
	return []option.SelectOption{
		TalkTopicGo,
		TalkTopicWeb,
		TalkTopicData,
		TalkTopicSecurity,
		TalkTopicOperations,
		TalkTopicDesign,
	}
}

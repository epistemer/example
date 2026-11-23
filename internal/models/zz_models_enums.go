// Code generated from sqlc's request by github.com/epistemer/models. DO NOT EDIT.

package models

import (
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

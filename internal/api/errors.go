package api

import "github.com/epistemer/errors"

var (
	ErrSignInBody = errors.Key("errors.api.sign_in_body", "send a login and a password, or a flow and a code", errors.Invalid)
	ErrNoFile     = errors.Key("errors.api.no_file", "the request has no part named file", errors.Invalid)
)

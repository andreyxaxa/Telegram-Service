package errs

import "errors"

var (
	// session
	ErrSessionAlreadyExists = errors.New("session already exists")
	ErrSessionNotFound      = errors.New("session not found")
	ErrSessionNotAuthorized = errors.New("session not authorized")

	// token
	ErrUnexpectedTokenType   = errors.New("unexpected token type")
	ErrQRTokenWaitingTimeout = errors.New("timeout waiting for QR token")

	ErrUnexpectedUpdatesType = errors.New("unexpected updates type")
)

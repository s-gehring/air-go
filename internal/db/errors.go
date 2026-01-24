package db

import "errors"

// Standard database errors
var (
	// Connection Errors
	ErrNotConnected         = errors.New("db: not connected to database")
	ErrConnectionTimeout    = errors.New("db: connection timeout")
	ErrAuthenticationFailed = errors.New("db: authentication failed")
	ErrInvalidConfiguration = errors.New("db: invalid configuration")

	// Operation Errors
	ErrOperationTimeout = errors.New("db: operation timeout")
	ErrNotFound         = errors.New("db: document not found")
	ErrDuplicateKey     = errors.New("db: duplicate key error")
	ErrInvalidDocument  = errors.New("db: invalid document")

	// State Errors
	ErrAlreadyConnected    = errors.New("db: already connected")
	ErrDatabaseUnavailable = errors.New("db: database unavailable")
)

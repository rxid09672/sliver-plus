package c2profiles

import "errors"

// Profile validation and loading errors
var (
	ErrInvalidProfileName       = errors.New("profile name is required")
	ErrInvalidProfileVersion    = errors.New("profile version is required")
	ErrInvalidTLSFingerprint    = errors.New("invalid TLS fingerprint (must be: chrome, firefox, ios, android, edge, safari)")
	ErrInvalidInterval          = errors.New("interval must be positive")
	ErrInvalidJitter            = errors.New("jitter must be between 0-99")
	ErrInvalidEncoding          = errors.New("invalid encoding (must be: none, base64, hex, netbios)")
	ErrProfileNotFound          = errors.New("profile not found")
	ErrInvalidProfileFormat     = errors.New("invalid profile format")
)


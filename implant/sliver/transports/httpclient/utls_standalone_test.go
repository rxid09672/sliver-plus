//go:build standalone
// +build standalone

package httpclient_test

import (
	"testing"

	tls "github.com/refraction-networking/utls"
)

// TestUTLSStandalone is a standalone test that verifies utls can be imported
// without depending on other httpclient package code.
// Run with: go test -tags=standalone -v -run TestUTLSStandalone
func TestUTLSStandalone(t *testing.T) {
	// Verify we can reference utls types
	var _ *tls.UConn
	var _ tls.ClientHelloID

	t.Log("✅ utls import successful")
}

// TestUTLSClientHelloIDsStandalone verifies ClientHello IDs are available.
func TestUTLSClientHelloIDsStandalone(t *testing.T) {
	testCases := []struct {
		name string
		id   tls.ClientHelloID
	}{
		{"Chrome", tls.HelloChrome_Auto},
		{"Firefox", tls.HelloFirefox_Auto},
		{"iOS", tls.HelloIOS_Auto},
		{"Android", tls.HelloAndroid_11_OkHttp},
		{"Edge", tls.HelloEdge_Auto},
		{"Safari", tls.HelloSafari_Auto},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.id.Client == "" {
				t.Errorf("%s: ClientHelloID has empty Client field", tc.name)
			}
			t.Logf("✅ %s ClientHelloID available: %s", tc.name, tc.id.Client)
		})
	}
}

// TestUTLSVersionsStandalone verifies TLS version constants.
func TestUTLSVersionsStandalone(t *testing.T) {
	versions := map[string]uint16{
		"TLS 1.0": tls.VersionTLS10,
		"TLS 1.1": tls.VersionTLS11,
		"TLS 1.2": tls.VersionTLS12,
		"TLS 1.3": tls.VersionTLS13,
	}

	for name, version := range versions {
		if version == 0 {
			t.Errorf("%s version constant is zero", name)
		} else {
			t.Logf("✅ %s version: 0x%04x", name, version)
		}
	}
}

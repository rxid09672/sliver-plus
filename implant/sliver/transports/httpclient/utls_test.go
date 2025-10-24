package httpclient

import (
	"testing"

	tls "github.com/refraction-networking/utls"
)

// TestUTLSImport verifies that utls can be imported without errors.
// This is a minimal smoke test to ensure the vendored dependency is accessible.
func TestUTLSImport(t *testing.T) {
	// Verify we can reference utls types
	var _ *tls.UConn
	var _ tls.ClientHelloID
	
	t.Log("utls import successful")
}

// TestUTLSClientHelloIDs verifies that utls provides the expected ClientHello IDs.
// These IDs will be used for TLS fingerprinting in the actual implementation.
func TestUTLSClientHelloIDs(t *testing.T) {
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
			t.Logf("%s ClientHelloID: %+v", tc.name, tc.id)
		})
	}
}

// TestUTLSVersion verifies we can get version information from utls.
func TestUTLSVersion(t *testing.T) {
	// utls doesn't have a Version() function, but we can verify
	// that the package constants are accessible
	versions := []uint16{
		tls.VersionTLS10,
		tls.VersionTLS11,
		tls.VersionTLS12,
		tls.VersionTLS13,
	}

	for _, version := range versions {
		if version == 0 {
			t.Error("TLS version constant is zero")
		}
	}
	
	t.Logf("utls TLS version constants verified: TLS 1.0-1.3")
}


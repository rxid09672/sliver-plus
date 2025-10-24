package c2profiles

/*
	Sliver Implant Framework
	Copyright (C) 2024  Bishop Fox

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

// Profile represents a Malleable C2 profile for customizing network traffic patterns.
// Profiles are defined in YAML format and loaded at implant generation time.
type Profile struct {
	Metadata  ProfileMetadata  `yaml:"profile"`
	HTTP      *HTTPConfig      `yaml:"http,omitempty"`
	TLS       *TLSConfig       `yaml:"tls,omitempty"`
	Timing    *TimingConfig    `yaml:"timing,omitempty"`
	Metadata2 *MetadataConfig  `yaml:"metadata,omitempty"`
	Evasion   *EvasionConfig   `yaml:"evasion,omitempty"`
}

// ProfileMetadata contains basic profile information
type ProfileMetadata struct {
	Name        string `yaml:"name"`
	Author      string `yaml:"author,omitempty"`
	Description string `yaml:"description,omitempty"`
	Version     string `yaml:"version"`
}

// HTTPConfig defines HTTP traffic customization
type HTTPConfig struct {
	UserAgents    []string               `yaml:"user_agents,omitempty"`
	URIs          *URIPatterns           `yaml:"uris,omitempty"`
	Headers       *HeaderConfig          `yaml:"headers,omitempty"`
	ServerHeaders map[string]string      `yaml:"server_headers,omitempty"`
}

// URIPatterns defines URI patterns for different request types
type URIPatterns struct {
	Get  []string `yaml:"get,omitempty"`
	Post []string `yaml:"post,omitempty"`
}

// HeaderConfig defines headers for different request types
type HeaderConfig struct {
	Common map[string]string `yaml:"common,omitempty"`
	Get    map[string]string `yaml:"get,omitempty"`
	Post   map[string]string `yaml:"post,omitempty"`
}

// TLSConfig defines TLS fingerprinting settings
type TLSConfig struct {
	Fingerprint string `yaml:"fingerprint,omitempty"`
	MinVersion  string `yaml:"min_version,omitempty"`
	MaxVersion  string `yaml:"max_version,omitempty"`
}

// TimingConfig defines callback timing and jitter
type TimingConfig struct {
	Interval         int `yaml:"interval,omitempty"`
	IntervalVariance int `yaml:"interval_variance,omitempty"`
	Jitter           int `yaml:"jitter,omitempty"`
	PollTimeout      int `yaml:"poll_timeout,omitempty"`
}

// MetadataConfig defines how C2 metadata is encoded and transmitted
type MetadataConfig struct {
	Location   string `yaml:"location,omitempty"`
	HeaderName string `yaml:"header_name,omitempty"`
	CookieName string `yaml:"cookie_name,omitempty"`
	ParamName  string `yaml:"param_name,omitempty"`
	Encoding   string `yaml:"encoding,omitempty"`
}

// EvasionConfig defines additional evasion techniques
type EvasionConfig struct {
	TrafficMasking *TrafficMaskingConfig `yaml:"traffic_masking,omitempty"`
}

// TrafficMaskingConfig defines settings for prepending legitimate content
type TrafficMaskingConfig struct {
	Enabled     bool   `yaml:"enabled"`
	ContentType string `yaml:"content_type,omitempty"`
	MinSize     int    `yaml:"min_size,omitempty"`
	MaxSize     int    `yaml:"max_size,omitempty"`
}

// Validate performs basic validation on the profile
func (p *Profile) Validate() error {
	if p.Metadata.Name == "" {
		return ErrInvalidProfileName
	}
	if p.Metadata.Version == "" {
		return ErrInvalidProfileVersion
	}
	
	// Validate TLS fingerprint if specified
	if p.TLS != nil && p.TLS.Fingerprint != "" {
		validFingerprints := []string{"chrome", "firefox", "ios", "android", "edge", "safari", "safari-ios", "safari-macos"}
		valid := false
		for _, fp := range validFingerprints {
			if p.TLS.Fingerprint == fp {
				valid = true
				break
			}
		}
		if !valid {
			return ErrInvalidTLSFingerprint
		}
	}
	
	// Validate timing if specified
	if p.Timing != nil {
		if p.Timing.Interval < 0 {
			return ErrInvalidInterval
		}
		if p.Timing.Jitter < 0 || p.Timing.Jitter > 99 {
			return ErrInvalidJitter
		}
	}
	
	// Validate metadata encoding if specified
	if p.Metadata2 != nil && p.Metadata2.Encoding != "" {
		validEncodings := []string{"none", "base64", "hex", "netbios"}
		valid := false
		for _, enc := range validEncodings {
			if p.Metadata2.Encoding == enc {
				valid = true
				break
			}
		}
		if !valid {
			return ErrInvalidEncoding
		}
	}
	
	return nil
}


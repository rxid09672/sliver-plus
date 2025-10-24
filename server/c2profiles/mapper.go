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

import (
	"fmt"
	"time"
	
	"github.com/bishopfox/sliver/protobuf/clientpb"
)

// ApplyProfile applies a Malleable C2 profile to an ImplantConfig
// This maps profile settings to Sliver's configuration structure
func ApplyProfile(profile *Profile, config *clientpb.ImplantConfig) error {
	if profile == nil {
		return fmt.Errorf("profile cannot be nil")
	}
	if config == nil {
		return fmt.Errorf("config cannot be nil")
	}
	
	// Apply TLS fingerprinting settings
	if profile.TLS != nil && profile.TLS.Fingerprint != "" {
		config.EnableTLSFingerprinting = true
		config.TLSFingerprint = profile.TLS.Fingerprint
	}
	
	// Apply timing settings
	if profile.Timing != nil {
		if profile.Timing.Interval > 0 {
			config.ReconnectInterval = int64(profile.Timing.Interval) * int64(time.Second)
		}
		if profile.Timing.PollTimeout > 0 {
			config.PollTimeout = int64(profile.Timing.PollTimeout) * int64(time.Second)
		}
		// Note: Jitter is typically calculated at runtime, but we could add a config field
	}
	
	// Store the full profile name for reference
	// This allows the implant to know which profile it was built with
	config.MalleableC2Profile = profile.Metadata.Name
	
	// TODO: HTTP settings (User-Agents, URIs, Headers)
	// These will require extending the ImplantConfig protobuf or
	// embedding the profile data directly into the implant source
	
	return nil
}

// GetProfileSummary returns a human-readable summary of the profile
func GetProfileSummary(profile *Profile) string {
	if profile == nil {
		return "No profile"
	}
	
	summary := fmt.Sprintf("Profile: %s v%s", profile.Metadata.Name, profile.Metadata.Version)
	
	if profile.TLS != nil && profile.TLS.Fingerprint != "" {
		summary += fmt.Sprintf("\n  TLS: %s", profile.TLS.Fingerprint)
	}
	
	if profile.Timing != nil {
		if profile.Timing.Interval > 0 {
			summary += fmt.Sprintf("\n  Interval: %ds", profile.Timing.Interval)
		}
		if profile.Timing.Jitter > 0 {
			summary += fmt.Sprintf(" (±%d%% jitter)", profile.Timing.Jitter)
		}
	}
	
	if profile.HTTP != nil {
		if len(profile.HTTP.UserAgents) > 0 {
			summary += fmt.Sprintf("\n  User-Agents: %d defined", len(profile.HTTP.UserAgents))
		}
		if profile.HTTP.URIs != nil {
			getCount := len(profile.HTTP.URIs.Get)
			postCount := len(profile.HTTP.URIs.Post)
			if getCount > 0 || postCount > 0 {
				summary += fmt.Sprintf("\n  URIs: %d GET, %d POST", getCount, postCount)
			}
		}
	}
	
	return summary
}

// ValidateProfileForC2(profile *Profile, c2Type string) error validates
// that a profile is suitable for the specified C2 type
func ValidateProfileForC2(profile *Profile, c2Type string) error {
	if profile == nil {
		return fmt.Errorf("profile cannot be nil")
	}
	
	switch c2Type {
	case "http", "https":
		// HTTP(S) profiles should have HTTP config
		if profile.HTTP == nil {
			return fmt.Errorf("profile '%s' has no HTTP configuration for HTTP(S) C2", profile.Metadata.Name)
		}
		// TLS fingerprinting only makes sense for HTTPS
		if c2Type == "https" && profile.TLS != nil && profile.TLS.Fingerprint != "" {
			// Valid combination
		}
	case "mtls", "wg", "dns", "named-pipe":
		// These C2 types don't use HTTP profiles currently
		// But timing settings still apply
	default:
		return fmt.Errorf("unknown C2 type: %s", c2Type)
	}
	
	return nil
}


package generate

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
	"os"
	"path/filepath"
	"time"

	"github.com/bishopfox/sliver/protobuf/clientpb"
	"gopkg.in/yaml.v3"
)

// MalleableC2Profile represents a complete Malleable C2 profile
// This is a minimal client-side representation for loading YAML profiles
type MalleableC2Profile struct {
	Profile struct {
		Name        string `yaml:"name"`
		Version     string `yaml:"version"`
		Description string `yaml:"description"`
		Author      string `yaml:"author"`
		TLS         *struct {
			Fingerprint string `yaml:"fingerprint"`
		} `yaml:"tls,omitempty"`
		HTTP *struct {
			UserAgents []string `yaml:"user_agents,omitempty"`
			URIs       *struct {
				Get  []string `yaml:"get,omitempty"`
				Post []string `yaml:"post,omitempty"`
			} `yaml:"uris,omitempty"`
			Headers map[string]string `yaml:"headers,omitempty"`
		} `yaml:"http,omitempty"`
		Timing *struct {
			Interval    int `yaml:"interval"`
			Jitter      int `yaml:"jitter"`
			PollTimeout int `yaml:"poll_timeout"`
		} `yaml:"timing,omitempty"`
	} `yaml:"profile"`
}

// LoadMalleableProfile loads a Malleable C2 profile from a YAML file
func LoadMalleableProfile(path string) (*MalleableC2Profile, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("profile not found: %s", path)
		}
		return nil, fmt.Errorf("failed to read profile: %w", err)
	}

	var profile MalleableC2Profile
	if err := yaml.Unmarshal(data, &profile); err != nil {
		return nil, fmt.Errorf("invalid profile format: %w", err)
	}

	// Basic validation
	if profile.Profile.Name == "" {
		return nil, fmt.Errorf("profile missing required field: profile.name")
	}

	return &profile, nil
}

// LoadMalleableProfileByName searches for a profile by name in standard locations
func LoadMalleableProfileByName(name string) (*MalleableC2Profile, error) {
	// Search paths (in order of priority):
	// 1. Current directory
	// 2. ./profiles/
	// 3. ~/.sliver/profiles/
	// 4. Recursive search in profiles subdirectories

	searchPaths := []string{
		fmt.Sprintf("%s.yml", name),
		filepath.Join("profiles", fmt.Sprintf("%s.yml", name)),
	}

	// Add user profile directory
	if homeDir, err := os.UserHomeDir(); err == nil {
		searchPaths = append(searchPaths,
			filepath.Join(homeDir, ".sliver", "profiles", fmt.Sprintf("%s.yml", name)),
		)
	}

	// Add recursive search in common profile directories
	profileDirs := []string{
		"profiles",
		filepath.Join("profiles", "examples"),
		filepath.Join("profiles", "services"),
		filepath.Join("profiles", "stealth"),
		filepath.Join("profiles", "apt"),
	}

	for _, dir := range profileDirs {
		searchPaths = append(searchPaths, filepath.Join(dir, fmt.Sprintf("%s.yml", name)))
	}

	// Try each search path
	for _, path := range searchPaths {
		if _, err := os.Stat(path); err == nil {
			return LoadMalleableProfile(path)
		}
	}

	return nil, fmt.Errorf("profile not found: %s (searched %d locations)", name, len(searchPaths))
}

// ApplyMalleableProfile applies a Malleable C2 profile to an ImplantConfig
// This modifies the config with profile settings
func ApplyMalleableProfile(profile *MalleableC2Profile, config *clientpb.ImplantConfig) error {
	if profile == nil {
		return fmt.Errorf("profile cannot be nil")
	}
	if config == nil {
		return fmt.Errorf("config cannot be nil")
	}

	// Apply TLS fingerprinting settings
	// Only override if not already set by explicit flags
	if profile.Profile.TLS != nil && profile.Profile.TLS.Fingerprint != "" {
		if !config.EnableTLSFingerprinting {
			config.EnableTLSFingerprinting = true
			config.TLSFingerprint = profile.Profile.TLS.Fingerprint
		}
	}

	// Apply timing settings
	// Only override defaults, not explicit user choices
	if profile.Profile.Timing != nil {
		if profile.Profile.Timing.Interval > 0 && config.ReconnectInterval == DefaultReconnect*int64(time.Second) {
			config.ReconnectInterval = int64(profile.Profile.Timing.Interval) * int64(time.Second)
		}
		if profile.Profile.Timing.PollTimeout > 0 && config.PollTimeout == DefaultPollTimeout*int64(time.Second) {
			config.PollTimeout = int64(profile.Profile.Timing.PollTimeout) * int64(time.Second)
		}
	}

	// Store the profile name for reference
	config.MalleableC2Profile = profile.Profile.Name

	// Note: HTTP settings (User-Agents, URIs, Headers) will require
	// extending the HTTP C2 profile system or embedding in implant source.
	// For now, we just record the profile name.

	return nil
}

// GetMalleableProfileSummary returns a human-readable summary of the profile
func GetMalleableProfileSummary(profile *MalleableC2Profile) string {
	if profile == nil {
		return "No profile"
	}

	summary := fmt.Sprintf("Profile: %s v%s", profile.Profile.Name, profile.Profile.Version)

	if profile.Profile.TLS != nil && profile.Profile.TLS.Fingerprint != "" {
		summary += fmt.Sprintf("\n  TLS: %s", profile.Profile.TLS.Fingerprint)
	}

	if profile.Profile.Timing != nil {
		if profile.Profile.Timing.Interval > 0 {
			summary += fmt.Sprintf("\n  Interval: %ds", profile.Profile.Timing.Interval)
		}
		if profile.Profile.Timing.Jitter > 0 {
			summary += fmt.Sprintf(" (±%d%% jitter)", profile.Profile.Timing.Jitter)
		}
	}

	if profile.Profile.HTTP != nil {
		if len(profile.Profile.HTTP.UserAgents) > 0 {
			summary += fmt.Sprintf("\n  User-Agents: %d defined", len(profile.Profile.HTTP.UserAgents))
		}
		if profile.Profile.HTTP.URIs != nil {
			getCount := len(profile.Profile.HTTP.URIs.Get)
			postCount := len(profile.Profile.HTTP.URIs.Post)
			if getCount > 0 || postCount > 0 {
				summary += fmt.Sprintf("\n  URIs: %d GET, %d POST", getCount, postCount)
			}
		}
	}

	return summary
}

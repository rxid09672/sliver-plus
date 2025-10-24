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
	"os"
	"path/filepath"
	
	"gopkg.in/yaml.v3"
)

// LoadProfile loads a profile from a YAML file
func LoadProfile(path string) (*Profile, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, ErrProfileNotFound
		}
		return nil, fmt.Errorf("failed to read profile: %w", err)
	}
	
	var profile Profile
	if err := yaml.Unmarshal(data, &profile); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidProfileFormat, err)
	}
	
	// Validate the profile
	if err := profile.Validate(); err != nil {
		return nil, err
	}
	
	return &profile, nil
}

// LoadProfileByName searches for a profile by name in standard locations
func LoadProfileByName(name string) (*Profile, error) {
	// Search paths (in order of priority):
	// 1. Current directory
	// 2. ./profiles/
	// 3. ~/.sliver/profiles/
	// 4. /usr/local/share/sliver/profiles/ (Unix)
	
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
	
	// Try each search path
	for _, path := range searchPaths {
		if _, err := os.Stat(path); err == nil {
			return LoadProfile(path)
		}
	}
	
	return nil, ErrProfileNotFound
}

// ListProfiles returns a list of available profiles in the given directory
func ListProfiles(dir string) ([]string, error) {
	var profiles []string
	
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	
	for _, entry := range entries {
		if entry.IsDir() {
			// Recursively search subdirectories
			subProfiles, err := ListProfiles(filepath.Join(dir, entry.Name()))
			if err == nil {
				profiles = append(profiles, subProfiles...)
			}
		} else if filepath.Ext(entry.Name()) == ".yml" || filepath.Ext(entry.Name()) == ".yaml" {
			// Load profile to get its name
			fullPath := filepath.Join(dir, entry.Name())
			profile, err := LoadProfile(fullPath)
			if err == nil {
				profiles = append(profiles, profile.Metadata.Name)
			}
		}
	}
	
	return profiles, nil
}


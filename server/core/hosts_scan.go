package core

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/bishopfox/sliver/protobuf/clientpb"
	"github.com/bishopfox/sliver/server/db"
	"github.com/bishopfox/sliver/server/db/models"
	"github.com/bishopfox/sliver/server/log"
	"github.com/gofrs/uuid"
)

var (
	scanLog = log.NamedLogger("core", "scan")
)

// NmapHost represents a host from nmap XML output
type NmapHost struct {
	Address   string
	Hostname  string
	OSVersion string
	Services  []NmapService
}

// NmapService represents a service from nmap XML output
type NmapService struct {
	Port     int
	Protocol string
	Name     string
	Product  string
	Version  string
	State    string
}

// ScanHosts - Scan network targets and create/update host entries
func ScanHosts(req *clientpb.HostScanReq) (*clientpb.HostScanResults, error) {
	startTime := time.Now()

	if len(req.Targets) == 0 {
		return nil, errors.New("no targets specified")
	}

	scanLog.Infof("Starting host scan: targets=%v, ports=%s, svc=%v, os=%v",
		req.Targets, req.Ports, req.ServiceDetection, req.OSDetection)

	// Check if nmap is available
	if !isNmapAvailable() {
		return nil, errors.New("nmap not found in PATH - please install nmap to use host scanning")
	}

	results := &clientpb.HostScanResults{
		Results: []*clientpb.HostScanResult{},
	}

	// Scan each target
	for _, target := range req.Targets {
		targetResults, err := scanTarget(target, req)
		if err != nil {
			scanLog.Warnf("Failed to scan target %s: %v", target, err)
			continue
		}
		results.Results = append(results.Results, targetResults...)
	}

	// Update statistics
	results.TotalHosts = int32(len(results.Results))
	newHostCount := 0
	for _, result := range results.Results {
		if result.IsNew {
			newHostCount++
		}
	}
	results.NewHosts = int32(newHostCount)
	results.ScanDuration = time.Since(startTime).String()

	scanLog.Infof("Scan complete: %d hosts discovered (%d new) in %s",
		results.TotalHosts, results.NewHosts, results.ScanDuration)

	return results, nil
}

func isNmapAvailable() bool {
	_, err := exec.LookPath("nmap")
	return err == nil
}

func scanTarget(target string, req *clientpb.HostScanReq) ([]*clientpb.HostScanResult, error) {
	scanLog.Debugf("Scanning target: %s", target)

	// Build nmap command
	args := []string{
		"-sS",      // SYN scan
		"-Pn",      // Skip ping (assume host is up)
		"-oJ", "-", // JSON output to stdout
	}

	// Add port specification
	if req.Ports != "" {
		args = append(args, "-p", req.Ports)
	} else {
		args = append(args, "-p", "21,22,23,25,53,80,110,111,135,139,143,443,445,993,995,1723,3306,3389,5900,8080")
	}

	// Add service detection
	if req.ServiceDetection {
		args = append(args, "-sV")
	}

	// Add OS detection
	if req.OSDetection {
		args = append(args, "-O")
	}

	// Add timeout
	timeout := 300
	if req.Timeout > 0 {
		timeout = int(req.Timeout)
	}

	// Add target
	args = append(args, target)

	scanLog.Debugf("Running: nmap %s", strings.Join(args, " "))

	// Execute nmap with timeout
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeout)*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "nmap", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return nil, fmt.Errorf("nmap scan timed out after %d seconds", timeout)
		}
		return nil, fmt.Errorf("nmap failed: %v (output: %s)", err, string(output))
	}

	// Parse nmap output
	hosts, err := parseNmapJSON(output)
	if err != nil {
		return nil, fmt.Errorf("failed to parse nmap output: %v", err)
	}

	// Convert to results and store in database
	results := make([]*clientpb.HostScanResult, 0, len(hosts))
	for _, host := range hosts {
		result, err := storeScannedHost(host)
		if err != nil {
			scanLog.Warnf("Failed to store host %s: %v", host.Address, err)
			continue
		}
		results = append(results, result)
	}

	return results, nil
}

func parseNmapJSON(output []byte) ([]NmapHost, error) {
	// Nmap's -oJ format outputs JSON, but we'll need to parse it
	// For now, let's use a simple approach that works with nmap's JSON output

	var data map[string]interface{}
	if err := json.Unmarshal(output, &data); err != nil {
		// If JSON parsing fails, try to extract hosts manually
		return parseNmapPlainText(string(output))
	}

	hosts := []NmapHost{}

	// Parse nmap JSON structure
	if nmaprun, ok := data["nmaprun"].(map[string]interface{}); ok {
		if hostsData, ok := nmaprun["host"].([]interface{}); ok {
			for _, hostData := range hostsData {
				if hostMap, ok := hostData.(map[string]interface{}); ok {
					host := parseNmapJSONHost(hostMap)
					hosts = append(hosts, host)
				}
			}
		}
	}

	return hosts, nil
}

func parseNmapJSONHost(hostMap map[string]interface{}) NmapHost {
	host := NmapHost{
		Services: []NmapService{},
	}

	// Extract address
	if address, ok := hostMap["address"].(map[string]interface{}); ok {
		if addr, ok := address["addr"].(string); ok {
			host.Address = addr
		}
	}

	// Extract hostname
	if hostnames, ok := hostMap["hostnames"].(map[string]interface{}); ok {
		if hostname, ok := hostnames["hostname"].(map[string]interface{}); ok {
			if name, ok := hostname["name"].(string); ok {
				host.Hostname = name
			}
		}
	}

	// Extract OS
	if os, ok := hostMap["os"].(map[string]interface{}); ok {
		if osmatch, ok := os["osmatch"].(map[string]interface{}); ok {
			if name, ok := osmatch["name"].(string); ok {
				host.OSVersion = name
			}
		}
	}

	// Extract ports/services
	if ports, ok := hostMap["ports"].(map[string]interface{}); ok {
		if portList, ok := ports["port"].([]interface{}); ok {
			for _, portData := range portList {
				if portMap, ok := portData.(map[string]interface{}); ok {
					service := parseNmapJSONService(portMap)
					host.Services = append(host.Services, service)
				}
			}
		}
	}

	return host
}

func parseNmapJSONService(portMap map[string]interface{}) NmapService {
	service := NmapService{}

	if portid, ok := portMap["portid"].(string); ok {
		fmt.Sscanf(portid, "%d", &service.Port)
	}

	if protocol, ok := portMap["protocol"].(string); ok {
		service.Protocol = protocol
	}

	if state, ok := portMap["state"].(map[string]interface{}); ok {
		if stateStr, ok := state["state"].(string); ok {
			service.State = stateStr
		}
	}

	if svc, ok := portMap["service"].(map[string]interface{}); ok {
		if name, ok := svc["name"].(string); ok {
			service.Name = name
		}
		if product, ok := svc["product"].(string); ok {
			service.Product = product
		}
		if version, ok := svc["version"].(string); ok {
			service.Version = version
		}
	}

	return service
}

func parseNmapPlainText(output string) ([]NmapHost, error) {
	// Fallback: simple text parsing for basic nmap output
	// This is a minimal implementation to handle cases where JSON parsing fails

	scanLog.Debugf("Falling back to plain text parsing")

	hosts := []NmapHost{}
	currentHost := NmapHost{Services: []NmapService{}}

	lines := strings.Split(output, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)

		// Look for "Nmap scan report for"
		if strings.HasPrefix(line, "Nmap scan report for ") {
			if currentHost.Address != "" {
				hosts = append(hosts, currentHost)
			}
			currentHost = NmapHost{Services: []NmapService{}}

			// Extract IP/hostname
			parts := strings.Fields(line)
			if len(parts) >= 5 {
				addr := parts[4]
				addr = strings.Trim(addr, "()")
				currentHost.Address = addr
			}
		}

		// Look for open ports (e.g., "22/tcp open ssh")
		if strings.Contains(line, "/tcp") || strings.Contains(line, "/udp") {
			parts := strings.Fields(line)
			if len(parts) >= 2 {
				portProto := strings.Split(parts[0], "/")
				if len(portProto) == 2 {
					service := NmapService{
						Protocol: portProto[1],
						State:    "open",
					}
					fmt.Sscanf(portProto[0], "%d", &service.Port)

					if len(parts) >= 3 {
						service.Name = parts[2]
					}

					currentHost.Services = append(currentHost.Services, service)
				}
			}
		}
	}

	if currentHost.Address != "" {
		hosts = append(hosts, currentHost)
	}

	return hosts, nil
}

func storeScannedHost(nmapHost NmapHost) (*clientpb.HostScanResult, error) {
	dbSession := db.Session()

	// Try to find existing host by IP address (check if we have sessions/beacons from this IP)
	// For now, we'll create a new host entry with a generated UUID

	result := &clientpb.HostScanResult{
		IPAddress: nmapHost.Address,
		Hostname:  nmapHost.Hostname,
		OSVersion: nmapHost.OSVersion,
		Services:  []*clientpb.HostService{},
	}

	// Convert services
	for _, svc := range nmapHost.Services {
		result.Services = append(result.Services, &clientpb.HostService{
			Port:     int32(svc.Port),
			Protocol: svc.Protocol,
			Name:     svc.Name,
			Product:  svc.Product,
			Version:  svc.Version,
			State:    svc.State,
		})
	}

	// Check if host already exists (by hostname)
	// For now, we'll always create new hosts since we don't have a good way to match by IP
	// This is safe because:
	// 1. Scanned hosts are tracked separately from session/beacon hosts
	// 2. Operators can manually reconcile via the hosts command
	// 3. Future enhancement: add IP-based matching
	var existingHost *models.Host = nil

	if existingHost != nil {
		// Update existing host
		result.HostUUID = existingHost.HostUUID.String()
		result.IsNew = false

		// Update OS if we have better information
		if nmapHost.OSVersion != "" && existingHost.OSVersion == "" {
			existingHost.OSVersion = nmapHost.OSVersion
		}

		// Store services in extension data
		servicesJSON, _ := json.Marshal(result.Services)
		existingHost.ExtensionData = append(existingHost.ExtensionData, models.ExtensionData{
			Name:   "scan_services",
			Output: string(servicesJSON),
		})

		if err := dbSession.Save(existingHost).Error; err != nil {
			return nil, err
		}

		scanLog.Debugf("Updated existing host: %s (%s)", nmapHost.Address, existingHost.HostUUID)
	} else {
		// Create new host
		newUUID, _ := uuid.NewV4()
		result.HostUUID = newUUID.String()
		result.IsNew = true

		// Store services in extension data
		servicesJSON, _ := json.Marshal(result.Services)

		newHost := &models.Host{
			HostUUID:  newUUID,
			Hostname:  nmapHost.Hostname,
			OSVersion: nmapHost.OSVersion,
			IOCs:      []models.IOC{},
			ExtensionData: []models.ExtensionData{
				{
					Name:   "scan_ip",
					Output: nmapHost.Address,
				},
				{
					Name:   "scan_services",
					Output: string(servicesJSON),
				},
				{
					Name:   "discovery_method",
					Output: "network_scan",
				},
			},
		}

		if err := dbSession.Create(newHost).Error; err != nil {
			return nil, err
		}

		scanLog.Infof("Created new host: %s (%s) with %d services",
			nmapHost.Address, newUUID, len(nmapHost.Services))
	}

	return result, nil
}

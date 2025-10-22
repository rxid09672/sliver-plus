package core

import (
	"bufio"
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
	reconLog = log.NamedLogger("core", "recon")
)

// SubfinderResult represents a single subdomain from subfinder
type SubfinderResult struct {
	Host   string `json:"host"`
	IP     string `json:"ip,omitempty"`
	Source string `json:"source,omitempty"`
}

// HarvesterResult represents theHarvester output
type HarvesterResult struct {
	Hosts  []string `json:"hosts,omitempty"`
	IPs    []string `json:"ips,omitempty"`
	Emails []string `json:"emails,omitempty"`
}

// ReconData represents the stored reconnaissance data
type ReconData struct {
	Target     string            `json:"target"`
	Timestamp  string            `json:"timestamp"`
	Tools      []string          `json:"tools"`
	Subdomains []SubfinderResult `json:"subdomains,omitempty"`
	Emails     []string          `json:"emails,omitempty"`
	IPs        []string          `json:"ips,omitempty"`
	Summary    map[string]int    `json:"summary"`
}

// ReconTarget - Run OSINT reconnaissance and store as loot
func ReconTarget(req *clientpb.ReconReq) (*clientpb.ReconResult, error) {
	startTime := time.Now()

	if req.Target == "" {
		return nil, errors.New("no target specified")
	}

	// Validate domain format (basic check)
	if strings.Contains(req.Target, " ") || strings.Contains(req.Target, "/") {
		return nil, errors.New("invalid domain format")
	}

	reconLog.Infof("Starting OSINT reconnaissance: type=%s, target=%s", req.Type, req.Target)

	result := &clientpb.ReconResult{
		Type:   req.Type,
		Target: req.Target,
		Tools:  []string{},
	}

	reconData := &ReconData{
		Target:     req.Target,
		Timestamp:  time.Now().Format(time.RFC3339),
		Tools:      []string{},
		Subdomains: []SubfinderResult{},
		Emails:     []string{},
		IPs:        []string{},
		Summary:    make(map[string]int),
	}

	timeout := 300
	if req.Timeout > 0 {
		timeout = int(req.Timeout)
	}

	// Run reconnaissance based on type
	switch req.Type {
	case "domains":
		if err := runSubdomainEnum(req.Target, timeout, reconData); err != nil {
			reconLog.Warnf("Subdomain enumeration failed: %v", err)
		}

	case "emails":
		if err := runEmailEnum(req.Target, timeout, reconData); err != nil {
			reconLog.Warnf("Email enumeration failed: %v", err)
		}

	case "all":
		// Run all tools
		if err := runSubdomainEnum(req.Target, timeout, reconData); err != nil {
			reconLog.Warnf("Subdomain enumeration failed: %v", err)
		}
		if err := runEmailEnum(req.Target, timeout, reconData); err != nil {
			reconLog.Warnf("Email enumeration failed: %v", err)
		}

	default:
		return nil, fmt.Errorf("unknown recon type: %s", req.Type)
	}

	// Calculate summary
	reconData.Summary["subdomains"] = len(reconData.Subdomains)
	reconData.Summary["emails"] = len(reconData.Emails)
	reconData.Summary["ips"] = len(reconData.IPs)

	// Store as loot
	lootID, itemsFound, err := storeLoot(req.Type, req.Target, reconData)
	if err != nil {
		return nil, fmt.Errorf("failed to store loot: %v", err)
	}

	result.LootID = lootID
	result.ItemsFound = int32(itemsFound)
	result.Duration = time.Since(startTime).String()
	result.Tools = reconData.Tools

	reconLog.Infof("Reconnaissance complete: %d items found in %s", itemsFound, result.Duration)

	return result, nil
}

func runSubdomainEnum(domain string, timeout int, data *ReconData) error {
	// Check if subfinder is available
	if !isToolAvailable("subfinder") {
		return errors.New("subfinder not found in PATH - install: go install -v github.com/projectdiscovery/subfinder/v2/cmd/subfinder@latest")
	}

	reconLog.Infof("Running subfinder for %s", domain)
	data.Tools = append(data.Tools, "subfinder")

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeout)*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "subfinder", "-d", domain, "-oJ", "-silent")
	output, err := cmd.CombinedOutput()
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return fmt.Errorf("subfinder timed out after %d seconds", timeout)
		}
		return fmt.Errorf("subfinder failed: %v (output: %s)", err, string(output))
	}

	// Parse JSON lines output
	scanner := bufio.NewScanner(strings.NewReader(string(output)))
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}

		var result SubfinderResult
		if err := json.Unmarshal([]byte(line), &result); err != nil {
			reconLog.Warnf("Failed to parse subfinder line: %v", err)
			continue
		}

		data.Subdomains = append(data.Subdomains, result)

		// Track unique IPs
		if result.IP != "" && !contains(data.IPs, result.IP) {
			data.IPs = append(data.IPs, result.IP)
		}
	}

	reconLog.Infof("Subfinder found %d subdomains", len(data.Subdomains))
	return nil
}

func runEmailEnum(domain string, timeout int, data *ReconData) error {
	// Check if theHarvester is available
	if !isToolAvailable("theHarvester") {
		return errors.New("theHarvester not found in PATH - install: pip3 install theHarvester")
	}

	reconLog.Infof("Running theHarvester for %s", domain)
	data.Tools = append(data.Tools, "theHarvester")

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeout)*time.Second)
	defer cancel()

	// theHarvester outputs JSON to stdout when using -f with stdout redirect
	cmd := exec.CommandContext(ctx, "theHarvester", "-d", domain, "-b", "all", "-f", "/dev/stdout")
	output, err := cmd.CombinedOutput()
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return fmt.Errorf("theHarvester timed out after %d seconds", timeout)
		}
		// theHarvester often returns non-zero even on success, so just log
		reconLog.Debugf("theHarvester returned error (may be non-fatal): %v", err)
	}

	// Try to parse JSON output
	var harvesterResult HarvesterResult
	if err := json.Unmarshal(output, &harvesterResult); err != nil {
		// If JSON parsing fails, try to extract emails manually from output
		extractEmailsFromText(string(output), data)
		return nil
	}

	// Add emails
	for _, email := range harvesterResult.Emails {
		if !contains(data.Emails, email) {
			data.Emails = append(data.Emails, email)
		}
	}

	// Add any new hosts to subdomains
	for _, host := range harvesterResult.Hosts {
		if !containsHost(data.Subdomains, host) {
			data.Subdomains = append(data.Subdomains, SubfinderResult{
				Host:   host,
				Source: "theharvester",
			})
		}
	}

	// Add IPs
	for _, ip := range harvesterResult.IPs {
		if !contains(data.IPs, ip) {
			data.IPs = append(data.IPs, ip)
		}
	}

	reconLog.Infof("theHarvester found %d emails, %d hosts, %d IPs",
		len(harvesterResult.Emails), len(harvesterResult.Hosts), len(harvesterResult.IPs))

	return nil
}

func extractEmailsFromText(text string, data *ReconData) {
	// Simple email extraction using basic pattern
	// This is a fallback when JSON parsing fails
	lines := strings.Split(text, "\n")
	for _, line := range lines {
		if strings.Contains(line, "@") {
			// Very basic email pattern
			fields := strings.Fields(line)
			for _, field := range fields {
				if strings.Contains(field, "@") && strings.Contains(field, ".") {
					email := strings.Trim(field, "[](){}:,;\"'")
					if isValidEmail(email) && !contains(data.Emails, email) {
						data.Emails = append(data.Emails, email)
					}
				}
			}
		}
	}
}

func isValidEmail(email string) bool {
	// Very basic validation
	return strings.Contains(email, "@") && strings.Contains(email, ".") && len(email) > 5
}

func storeLoot(reconType, target string, data *ReconData) (string, int, error) {
	// Generate loot name
	lootName := fmt.Sprintf("%s_%s_%s", target, reconType, time.Now().Format("20060102_150405"))
	lootType := fmt.Sprintf("recon/%s", reconType)

	// Serialize data to JSON
	jsonData, err := json.Marshal(data)
	if err != nil {
		return "", 0, fmt.Errorf("failed to marshal recon data: %v", err)
	}

	// Create loot entry
	lootID, err := uuid.NewV4()
	if err != nil {
		return "", 0, fmt.Errorf("failed to generate UUID: %v", err)
	}

	loot := &models.Loot{
		LootID:   lootID,
		Name:     lootName,
		FileType: lootType,
		Data:     jsonData,
	}

	dbSession := db.Session()
	if err := dbSession.Create(loot).Error; err != nil {
		return "", 0, fmt.Errorf("failed to save loot: %v", err)
	}

	// Calculate total items found
	itemsFound := len(data.Subdomains) + len(data.Emails) + len(data.IPs)

	reconLog.Infof("Stored recon data as loot: %s (%d items)", lootID.String(), itemsFound)

	return lootID.String(), itemsFound, nil
}

func isToolAvailable(tool string) bool {
	_, err := exec.LookPath(tool)
	return err == nil
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func containsHost(subdomains []SubfinderResult, host string) bool {
	for _, sub := range subdomains {
		if sub.Host == host {
			return true
		}
	}
	return false
}

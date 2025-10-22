package loot

import (
	"context"
	"strings"

	"github.com/bishopfox/sliver/client/console"
	"github.com/bishopfox/sliver/protobuf/clientpb"
	"github.com/spf13/cobra"
)

// LootReconCmd - Run OSINT reconnaissance and store as loot
func LootReconCmd(cmd *cobra.Command, con *console.SliverClient, args []string) {
	// Validate arguments
	if len(args) < 2 {
		con.PrintErrorf("Missing arguments. Usage: loot recon <type> <target>\n")
		con.PrintInfof("Types: domains, emails, all\n")
		con.PrintInfof("Example: loot recon domains example.com\n")
		return
	}

	reconType := args[0]
	target := args[1]

	// Validate recon type
	validTypes := map[string]bool{
		"domains": true,
		"emails":  true,
		"all":     true,
	}
	if !validTypes[reconType] {
		con.PrintErrorf("Invalid recon type: %s\n", reconType)
		con.PrintInfof("Valid types: domains, emails, all\n")
		return
	}

	// Validate target (basic domain format check)
	if strings.Contains(target, " ") || strings.Contains(target, "/") {
		con.PrintErrorf("Invalid domain format: %s\n", target)
		con.PrintInfof("Expected format: example.com\n")
		return
	}

	// Get timeout flag
	timeout, _ := cmd.Flags().GetInt32("timeout")

	// Display start message
	con.PrintInfof("Starting OSINT reconnaissance on %s...\n", target)
	con.PrintInfof("Type: %s\n", reconType)
	if timeout != 300 {
		con.PrintInfof("Timeout: %d seconds\n", timeout)
	}
	con.PrintInfof("This may take a while...\n\n")

	// Build request
	req := &clientpb.ReconReq{
		Type:    reconType,
		Target:  target,
		Timeout: timeout,
	}

	// Execute reconnaissance
	result, err := con.Rpc.Recon(context.Background(), req)
	if err != nil {
		con.PrintErrorf("Reconnaissance failed: %s\n", err)
		return
	}

	// Display results
	displayReconResults(result, con)
}

func displayReconResults(result *clientpb.ReconResult, con *console.SliverClient) {
	con.PrintInfof("Reconnaissance complete!\n\n")

	// Tool execution summary
	if len(result.Tools) > 0 {
		con.PrintInfof("Tools executed:\n")
		for _, tool := range result.Tools {
			con.PrintInfof("  └─ %s\n", tool)
		}
		con.PrintInfof("\n")
	}

	// Items found summary
	con.PrintInfof("Results:\n")
	switch result.Type {
	case "domains":
		con.PrintInfof("  └─ Found %d subdomain(s)\n", result.ItemsFound)
	case "emails":
		con.PrintInfof("  └─ Found %d email(s)\n", result.ItemsFound)
	case "all":
		con.PrintInfof("  └─ Found %d total item(s)\n", result.ItemsFound)
	}
	con.PrintInfof("  └─ Duration: %s\n", result.Duration)
	con.PrintInfof("\n")

	// Loot information
	con.PrintInfof("Results stored as loot:\n")
	con.PrintInfof("  └─ ID: %s\n", result.LootID)
	con.PrintInfof("  └─ Type: recon/%s\n", result.Type)
	con.PrintInfof("\n")

	// Usage instructions
	con.PrintInfof("View results:\n")
	con.PrintInfof("  sliver> loot fetch %s\n", result.LootID)
	con.PrintInfof("  sliver> loot --type recon/%s\n", result.Type)
	con.PrintInfof("\n")

	// Suggestions based on results
	if result.ItemsFound > 0 {
		switch result.Type {
		case "domains":
			con.PrintInfof("💡 Tip: Use 'hosts scan' to actively scan discovered subdomains\n")
		case "emails":
			con.PrintInfof("💡 Tip: Use emails for targeted phishing assessments\n")
		case "all":
			con.PrintInfof("💡 Tip: Review the full dataset with 'loot fetch %s'\n", result.LootID)
		}
	} else {
		con.PrintWarnf("No results found. This could mean:\n")
		con.PrintWarnf("  • Target has no public presence\n")
		con.PrintWarnf("  • Tools couldn't reach data sources\n")
		con.PrintWarnf("  • Network connectivity issues\n")
		con.PrintWarnf("  • API keys not configured (for better results)\n")
	}
}

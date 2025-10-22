package hosts

import (
	"context"
	"fmt"
	"strings"

	"github.com/bishopfox/sliver/client/command/settings"
	"github.com/bishopfox/sliver/client/console"
	"github.com/bishopfox/sliver/protobuf/clientpb"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/spf13/cobra"
)

// HostsScanCmd - Scan network targets and discover hosts
func HostsScanCmd(cmd *cobra.Command, con *console.SliverClient, args []string) {
	// Get flags
	ports, _ := cmd.Flags().GetString("ports")
	serviceDetection, _ := cmd.Flags().GetBool("service-detection")
	osDetection, _ := cmd.Flags().GetBool("os-detection")
	timeout, _ := cmd.Flags().GetInt32("scan-timeout")

	// Validate arguments
	if len(args) == 0 {
		con.PrintErrorf("Missing target(s). Usage: hosts scan <target> [target...]\n")
		con.PrintInfof("Examples:\n")
		con.PrintInfof("  hosts scan 192.168.1.1\n")
		con.PrintInfof("  hosts scan 10.0.0.0/24\n")
		con.PrintInfof("  hosts scan 192.168.1.1 192.168.1.2 10.0.0.0/24\n")
		return
	}

	con.PrintInfof("Starting network scan of %d target(s)...\n", len(args))
	if serviceDetection {
		con.PrintInfof("Service detection: enabled\n")
	}
	if osDetection {
		con.PrintInfof("OS detection: enabled\n")
	}
	if ports != "" {
		con.PrintInfof("Port specification: %s\n", ports)
	}
	con.PrintInfof("This may take a while...\n\n")

	// Build request
	req := &clientpb.HostScanReq{
		Targets:          args,
		Ports:            ports,
		ServiceDetection: serviceDetection,
		OSDetection:      osDetection,
		Timeout:          timeout,
	}

	// Execute scan
	results, err := con.Rpc.HostScan(context.Background(), req)
	if err != nil {
		con.PrintErrorf("Scan failed: %s\n", err)
		return
	}

	// Display results
	if len(results.Results) == 0 {
		con.PrintInfof("No hosts discovered\n")
		return
	}

	con.PrintInfof("Scan completed in %s\n\n", results.ScanDuration)
	con.Printf("%s\n", renderScanResults(results, con))

	// Summary
	con.PrintInfof("\n")
	con.PrintInfof("Total hosts discovered: %d\n", results.TotalHosts)
	con.PrintInfof("New hosts added: %d\n", results.NewHosts)
	con.PrintInfof("\n")
	con.PrintInfof("Use 'hosts' to view all hosts in the database\n")
}

func renderScanResults(results *clientpb.HostScanResults, con *console.SliverClient) string {
	tw := table.NewWriter()
	tw.SetStyle(settings.GetTableStyle(con))
	tw.AppendHeader(table.Row{
		"Status",
		"IP Address",
		"Hostname",
		"OS",
		"Open Ports",
		"UUID",
	})

	for _, result := range results.Results {
		status := "Updated"
		if result.IsNew {
			status = "New"
		}

		// Format open ports
		openPorts := []string{}
		for _, svc := range result.Services {
			if svc.State == "open" {
				portStr := fmt.Sprintf("%d/%s", svc.Port, svc.Protocol)
				if svc.Name != "" {
					portStr += fmt.Sprintf(" (%s)", svc.Name)
				}
				openPorts = append(openPorts, portStr)
			}
		}
		portsStr := strings.Join(openPorts, ", ")
		if len(portsStr) > 50 {
			portsStr = portsStr[:47] + "..."
		}

		hostname := result.Hostname
		if hostname == "" {
			hostname = "-"
		}

		osVersion := result.OSVersion
		if osVersion == "" {
			osVersion = "-"
		}

		// Short UUID
		shortUUID := result.HostUUID
		if len(shortUUID) > 8 {
			shortUUID = shortUUID[:8]
		}

		tw.AppendRow(table.Row{
			status,
			result.IPAddress,
			hostname,
			osVersion,
			portsStr,
			shortUUID,
		})
	}

	return tw.Render()
}

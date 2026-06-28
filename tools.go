package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/oracle/oci-go-sdk/v65/common"
	"github.com/oracle/oci-go-sdk/v65/common/auth"
	"github.com/oracle/oci-go-sdk/v65/databasetools"
)

const ociTimeout = 30 * time.Second

type sqlCollectionClient interface {
	ListDatabaseToolsSqlReports(ctx context.Context, req databasetools.ListDatabaseToolsSqlReportsRequest) (databasetools.ListDatabaseToolsSqlReportsResponse, error)
	GetDatabaseToolsSqlReport(ctx context.Context, req databasetools.GetDatabaseToolsSqlReportRequest) (databasetools.GetDatabaseToolsSqlReportResponse, error)
}

// newOCIClient builds a DatabaseTools client using the requested auth type.
// instance_principal uses the compute instance identity and ignores profile.
// security_token reads a token-based session profile written by `oci session authenticate`.
// api_key (default) uses a standard API key profile from ~/.oci/config.
func newOCIClient(authType, profile string) (sqlCollectionClient, error) {
	var (
		provider common.ConfigurationProvider
		err      error
	)
	switch authType {
	case "instance_principal":
		provider, err = auth.InstancePrincipalConfigurationProvider()
		if err != nil {
			return nil, fmt.Errorf("instance_principal auth: %w", err)
		}
	case "security_token":
		provider = common.CustomProfileSessionTokenConfigProvider("", profile)
	default: // api_key
		provider = common.CustomProfileConfigProvider("", profile)
	}
	return databasetools.NewDatabaseToolsClientWithConfigurationProvider(provider)
}

// listReports calls OCI to list SQL reports in a compartment, then filters exclusions.
func listReports(cfg config, client sqlCollectionClient, args map[string]any) (any, *jsonRPCError) {
	compartment := cfg.compartment
	if v, ok := args["compartment_id"].(string); ok && v != "" {
		compartment = v
	}
	if compartment == "" {
		return nil, &jsonRPCError{Code: -32602, Message: "compartment_id is required"}
	}
	ctx, cancel := context.WithTimeout(context.Background(), ociTimeout)
	defer cancel()
	resp, err := client.ListDatabaseToolsSqlReports(ctx, databasetools.ListDatabaseToolsSqlReportsRequest{
		CompartmentId: &compartment,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "list_reports OCI error: %v\n", err)
		return nil, &jsonRPCError{Code: -32603, Message: "internal error: OCI API call failed"}
	}
	excluded := makeExcludeSet([]string(cfg.excludeReports))
	var items []databasetools.DatabaseToolsSqlReportSummary
	for _, item := range resp.Items {
		if id := item.GetId(); id != nil && !excluded[*id] {
			items = append(items, item)
		}
	}
	text, err := json.MarshalIndent(items, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "list_reports marshal error: %v\n", err)
		return nil, &jsonRPCError{Code: -32603, Message: "internal error: failed to encode response"}
	}
	return mcpTextResult(string(text)), nil
}

// getReport fetches a single SQL report by OCID, respecting the exclusion list.
func getReport(cfg config, client sqlCollectionClient, args map[string]any) (any, *jsonRPCError) {
	reportID, _ := args["report_id"].(string)
	if reportID == "" {
		return nil, &jsonRPCError{Code: -32602, Message: "report_id is required"}
	}
	if makeExcludeSet([]string(cfg.excludeReports))[reportID] {
		return mcpTextResult("No report found (excluded by filter)"), nil
	}
	ctx, cancel := context.WithTimeout(context.Background(), ociTimeout)
	defer cancel()
	resp, err := client.GetDatabaseToolsSqlReport(ctx, databasetools.GetDatabaseToolsSqlReportRequest{
		DatabaseToolsSqlReportId: &reportID,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "get_report OCI error: %v\n", err)
		return nil, &jsonRPCError{Code: -32603, Message: "internal error: OCI API call failed"}
	}
	text, err := json.MarshalIndent(resp.DatabaseToolsSqlReport, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "get_report marshal error: %v\n", err)
		return nil, &jsonRPCError{Code: -32603, Message: "internal error: failed to encode response"}
	}
	return mcpTextResult(string(text)), nil
}

// makeExcludeSet converts a slice of OCIDs to a set for O(1) lookup.
func makeExcludeSet(excludes []string) map[string]bool {
	set := make(map[string]bool, len(excludes))
	for _, id := range excludes {
		set[id] = true
	}
	return set
}

func mcpTextResult(text string) any {
	return map[string]any{
		"content": []map[string]any{
			{"type": "text", "text": text},
		},
	}
}

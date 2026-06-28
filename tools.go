package main

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/oracle/oci-go-sdk/v65/common"
	"github.com/oracle/oci-go-sdk/v65/common/auth"
	"github.com/oracle/oci-go-sdk/v65/databasetools"
)

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
	resp, err := client.ListDatabaseToolsSqlReports(context.Background(), databasetools.ListDatabaseToolsSqlReportsRequest{
		CompartmentId: &compartment,
	})
	if err != nil {
		return nil, &jsonRPCError{Code: -32603, Message: fmt.Sprintf("OCI error: %v", err)}
	}
	excluded := makeExcludeSet([]string(cfg.excludeReports))
	var items []databasetools.DatabaseToolsSqlReportSummary
	for _, item := range resp.Items {
		if id := item.GetId(); id != nil && !excluded[*id] {
			items = append(items, item)
		}
	}
	text, _ := json.MarshalIndent(items, "", "  ")
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
	resp, err := client.GetDatabaseToolsSqlReport(context.Background(), databasetools.GetDatabaseToolsSqlReportRequest{
		DatabaseToolsSqlReportId: &reportID,
	})
	if err != nil {
		return nil, &jsonRPCError{Code: -32603, Message: fmt.Sprintf("OCI error: %v", err)}
	}
	text, _ := json.MarshalIndent(resp.DatabaseToolsSqlReport, "", "  ")
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

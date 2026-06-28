package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/oracle/oci-go-sdk/v65/databasetools"
)

// mockClient stubs OCI calls for testing without live credentials.
type mockClient struct {
	listFunc func(ctx context.Context, req databasetools.ListDatabaseToolsSqlReportsRequest) (databasetools.ListDatabaseToolsSqlReportsResponse, error)
	getFunc  func(ctx context.Context, req databasetools.GetDatabaseToolsSqlReportRequest) (databasetools.GetDatabaseToolsSqlReportResponse, error)
}

func (m *mockClient) ListDatabaseToolsSqlReports(ctx context.Context, req databasetools.ListDatabaseToolsSqlReportsRequest) (databasetools.ListDatabaseToolsSqlReportsResponse, error) {
	return m.listFunc(ctx, req)
}

func (m *mockClient) GetDatabaseToolsSqlReport(ctx context.Context, req databasetools.GetDatabaseToolsSqlReportRequest) (databasetools.GetDatabaseToolsSqlReportResponse, error) {
	return m.getFunc(ctx, req)
}

func TestFlagParsing(t *testing.T) {
	cfg, err := parseFlags([]string{
		"-compartment", "ocid1.compartment.oc1..aaa",
		"-profile", "MYPROFILE",
		"-exclude-report", "ocid1.report.oc1..aaa",
		"-exclude-report", "ocid1.report.oc1..bbb",
	})
	if err != nil {
		t.Fatalf("parseFlags error: %v", err)
	}
	if cfg.compartment != "ocid1.compartment.oc1..aaa" {
		t.Errorf("compartment = %q, want %q", cfg.compartment, "ocid1.compartment.oc1..aaa")
	}
	if cfg.profile != "MYPROFILE" {
		t.Errorf("profile = %q, want %q", cfg.profile, "MYPROFILE")
	}
	if len(cfg.excludeReports) != 2 {
		t.Fatalf("excludeReports len = %d, want 2", len(cfg.excludeReports))
	}
	if cfg.excludeReports[0] != "ocid1.report.oc1..aaa" || cfg.excludeReports[1] != "ocid1.report.oc1..bbb" {
		t.Errorf("excludeReports = %v, unexpected values", cfg.excludeReports)
	}
}

func TestFlagDefaultProfile(t *testing.T) {
	cfg, err := parseFlags([]string{})
	if err != nil {
		t.Fatalf("parseFlags error: %v", err)
	}
	if cfg.profile != "DEFAULT" {
		t.Errorf("default profile = %q, want %q", cfg.profile, "DEFAULT")
	}
}

func TestFlagAuthDefault(t *testing.T) {
	cfg, err := parseFlags([]string{})
	if err != nil {
		t.Fatalf("parseFlags error: %v", err)
	}
	if cfg.authType != "api_key" {
		t.Errorf("default authType = %q, want %q", cfg.authType, "api_key")
	}
}

func TestFlagAuthValues(t *testing.T) {
	for _, auth := range []string{"api_key", "instance_principal", "security_token"} {
		cfg, err := parseFlags([]string{"-auth", auth})
		if err != nil {
			t.Fatalf("parseFlags(%q) error: %v", auth, err)
		}
		if cfg.authType != auth {
			t.Errorf("authType = %q, want %q", cfg.authType, auth)
		}
	}
}

func TestFlagAuthInvalid(t *testing.T) {
	_, err := parseFlags([]string{"-auth", "bad_value"})
	if err == nil {
		t.Error("expected error for invalid -auth value, got nil")
	}
}

func TestFlagInstancePrincipalIgnoresProfile(t *testing.T) {
	// -profile is accepted alongside -auth instance_principal but the provider ignores it.
	_, err := parseFlags([]string{"-auth", "instance_principal", "-profile", "SOME_PROFILE"})
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}
}

func TestMCPInitialize(t *testing.T) {
	input := `{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2025-11-25","capabilities":{},"clientInfo":{"name":"test","version":"0"}}}` + "\n"
	var out bytes.Buffer
	serve(strings.NewReader(input), &out, config{}, nil)

	var resp jsonRPCResponse
	if err := json.Unmarshal(out.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal response: %v (raw: %s)", err, out.String())
	}
	if resp.Error != nil {
		t.Fatalf("unexpected error: %v", resp.Error)
	}
	result, ok := resp.Result.(map[string]any)
	if !ok {
		t.Fatalf("result is not a map: %T", resp.Result)
	}
	if result["protocolVersion"] != "2025-11-25" {
		t.Errorf("protocolVersion = %v, want 2025-11-25", result["protocolVersion"])
	}
	if _, ok := result["capabilities"]; !ok {
		t.Error("missing capabilities in initialize response")
	}
	if _, ok := result["serverInfo"]; !ok {
		t.Error("missing serverInfo in initialize response")
	}
}

func TestMCPToolsList(t *testing.T) {
	input := `{"jsonrpc":"2.0","id":2,"method":"tools/list","params":{}}` + "\n"
	var out bytes.Buffer
	serve(strings.NewReader(input), &out, config{}, nil)

	var resp jsonRPCResponse
	if err := json.Unmarshal(out.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	result := resp.Result.(map[string]any)
	tools := result["tools"].([]any)
	if len(tools) != 2 {
		t.Fatalf("tools count = %d, want 2", len(tools))
	}
	names := map[string]bool{}
	for _, t := range tools {
		tool := t.(map[string]any)
		names[tool["name"].(string)] = true
	}
	if !names["list_reports"] || !names["get_report"] {
		t.Errorf("missing expected tools, got: %v", names)
	}
}

func TestMCPNotificationNoResponse(t *testing.T) {
	// notifications/initialized has no id and must produce no response
	input := `{"jsonrpc":"2.0","method":"notifications/initialized"}` + "\n"
	var out bytes.Buffer
	serve(strings.NewReader(input), &out, config{}, nil)
	if out.Len() != 0 {
		t.Errorf("expected no output for notification, got: %s", out.String())
	}
}

func TestExclusionFilter(t *testing.T) {
	excluded := makeExcludeSet([]string{"ocid1.x", "ocid1.y"})
	if !excluded["ocid1.x"] {
		t.Error("ocid1.x should be excluded")
	}
	if !excluded["ocid1.y"] {
		t.Error("ocid1.y should be excluded")
	}
	if excluded["ocid1.z"] {
		t.Error("ocid1.z should not be excluded")
	}
}

func TestGetReportExcluded(t *testing.T) {
	// get_report with an excluded OCID must not call the OCI API.
	called := false
	client := &mockClient{
		getFunc: func(ctx context.Context, req databasetools.GetDatabaseToolsSqlReportRequest) (databasetools.GetDatabaseToolsSqlReportResponse, error) {
			called = true
			return databasetools.GetDatabaseToolsSqlReportResponse{}, fmt.Errorf("should not be called")
		},
	}
	cfg := config{excludeReports: stringSliceFlag{"ocid1.excluded"}}
	result, rpcErr := getReport(cfg, client, map[string]any{"report_id": "ocid1.excluded"})
	if rpcErr != nil {
		t.Fatalf("unexpected rpc error: %v", rpcErr)
	}
	if called {
		t.Error("OCI client should not have been called for excluded report")
	}
	content := extractTextContent(t, result)
	if !strings.Contains(content, "excluded by filter") {
		t.Errorf("expected filter message, got: %s", content)
	}
}

func TestListReportsExclusion(t *testing.T) {
	id1 := "ocid1.report.oc1..keep"
	id2 := "ocid1.report.oc1..exclude"
	client := &mockClient{
		listFunc: func(ctx context.Context, req databasetools.ListDatabaseToolsSqlReportsRequest) (databasetools.ListDatabaseToolsSqlReportsResponse, error) {
			return databasetools.ListDatabaseToolsSqlReportsResponse{
				DatabaseToolsSqlReportCollection: databasetools.DatabaseToolsSqlReportCollection{
					Items: []databasetools.DatabaseToolsSqlReportSummary{
						databasetools.DatabaseToolsSqlReportSummaryOracleDatabase{Id: &id1},
						databasetools.DatabaseToolsSqlReportSummaryOracleDatabase{Id: &id2},
					},
				},
			}, nil
		},
	}
	cfg := config{
		compartment:    "ocid1.compartment.oc1..aaa",
		excludeReports: stringSliceFlag{id2},
	}
	result, rpcErr := listReports(cfg, client, map[string]any{})
	if rpcErr != nil {
		t.Fatalf("unexpected rpc error: %v", rpcErr)
	}
	text := extractTextContent(t, result)
	if strings.Contains(text, id2) {
		t.Errorf("excluded report %q should not appear in output", id2)
	}
	if !strings.Contains(text, id1) {
		t.Errorf("kept report %q should appear in output", id1)
	}
}

// extractTextContent pulls the text from an MCP tool result.
func extractTextContent(t *testing.T, result any) string {
	t.Helper()
	m, ok := result.(map[string]any)
	if !ok {
		t.Fatalf("result not a map: %T", result)
	}
	content, ok := m["content"].([]map[string]any)
	if !ok {
		t.Fatalf("content not []map[string]any: %T", m["content"])
	}
	if len(content) == 0 {
		t.Fatal("content is empty")
	}
	return content[0]["text"].(string)
}

package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
)

type jsonRPCRequest struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      json.RawMessage `json:"id,omitempty"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

type jsonRPCResponse struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      json.RawMessage `json:"id,omitempty"`
	Result  any             `json:"result,omitempty"`
	Error   any             `json:"error,omitempty"`
}

type jsonRPCError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
}

// serve is the MCP STDIO loop: one JSON-RPC 2.0 message per line in/out.
func serve(r io.Reader, w io.Writer, cfg config, client sqlCollectionClient) {
	enc := json.NewEncoder(w)
	scanner := bufio.NewScanner(r)
	scanner.Buffer(make([]byte, 1<<20), 1<<20) // 1 MB max line
	for scanner.Scan() {
		var req jsonRPCRequest
		if err := json.Unmarshal(scanner.Bytes(), &req); err != nil {
			continue
		}
		// Notifications have no id field; no response is sent.
		if len(req.ID) == 0 {
			continue
		}
		result, rpcErr := dispatch(req, cfg, client)
		var resp jsonRPCResponse
		if rpcErr != nil {
			resp = jsonRPCResponse{JSONRPC: "2.0", ID: req.ID, Error: rpcErr}
		} else {
			resp = jsonRPCResponse{JSONRPC: "2.0", ID: req.ID, Result: result}
		}
		if err := enc.Encode(resp); err != nil {
			fmt.Fprintf(os.Stderr, "write error: %v\n", err)
			return
		}
	}
	if err := scanner.Err(); err != nil {
		fmt.Fprintf(os.Stderr, "stdin read error: %v\n", err)
	}
}

func dispatch(req jsonRPCRequest, cfg config, client sqlCollectionClient) (any, *jsonRPCError) {
	switch req.Method {
	case "initialize":
		// Respond with our supported version regardless of what the client sent.
		// Per the MCP spec the client decides to disconnect if it cannot accept ours.
		return map[string]any{
			"protocolVersion": "2025-11-25",
			"capabilities":    map[string]any{"tools": map[string]any{}},
			"serverInfo":      map[string]any{"name": "mcp-sql-reports", "version": "1.0.0"},
		}, nil
	case "ping":
		return map[string]any{}, nil
	case "tools/list":
		return toolsList(), nil
	case "tools/call":
		return toolsCall(req.Params, cfg, client)
	default:
		return nil, &jsonRPCError{Code: -32601, Message: "method not found: " + truncate(req.Method, 128)}
	}
}

func toolsList() any {
	return map[string]any{
		"tools": []map[string]any{
			{
				"name":        "list_reports",
				"description": "List Database Tools SQL Reports in a compartment",
				"inputSchema": map[string]any{
					"type": "object",
					"properties": map[string]any{
						"compartment_id": map[string]any{
							"type":        "string",
							"description": "OCID of the compartment. Falls back to -compartment CLI flag if omitted.",
						},
					},
				},
			},
			{
				"name":        "get_report",
				"description": "Get details of a Database Tools SQL Report",
				"inputSchema": map[string]any{
					"type": "object",
					"properties": map[string]any{
						"report_id": map[string]any{
							"type":        "string",
							"description": "OCID of the Database Tools SQL Report to fetch.",
						},
					},
					"required": []string{"report_id"},
				},
			},
		},
	}
}

func toolsCall(params json.RawMessage, cfg config, client sqlCollectionClient) (any, *jsonRPCError) {
	var p struct {
		Name      string         `json:"name"`
		Arguments map[string]any `json:"arguments"`
	}
	if err := json.Unmarshal(params, &p); err != nil {
		return nil, &jsonRPCError{Code: -32602, Message: "invalid params"}
	}
	switch p.Name {
	case "list_reports":
		return listReports(cfg, client, p.Arguments)
	case "get_report":
		return getReport(cfg, client, p.Arguments)
	default:
		return nil, &jsonRPCError{Code: -32601, Message: "unknown tool: " + truncate(p.Name, 128)}
	}
}

func truncate(s string, n int) string {
	if len(s) > n {
		return s[:n]
	}
	return s
}

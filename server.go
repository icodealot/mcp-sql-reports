package main

import (
	"bufio"
	"encoding/json"
	"io"
)

type jsonRPCRequest struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      json.RawMessage `json:"id,omitempty"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

type jsonRPCResponse struct {
	JSONRPC string `json:"jsonrpc"`
	ID      any    `json:"id,omitempty"`
	Result  any    `json:"result,omitempty"`
	Error   any    `json:"error,omitempty"`
}

type jsonRPCError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// serve is the MCP STDIO loop: one JSON-RPC 2.0 message per line in/out.
func serve(r io.Reader, w io.Writer, cfg config, client sqlCollectionClient) {
	enc := json.NewEncoder(w)
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		var req jsonRPCRequest
		if err := json.Unmarshal(scanner.Bytes(), &req); err != nil {
			continue
		}
		// Notifications have no id field; no response is sent.
		if len(req.ID) == 0 {
			continue
		}
		var id any
		json.Unmarshal(req.ID, &id)

		result, rpcErr := dispatch(req, cfg, client)
		if rpcErr != nil {
			enc.Encode(jsonRPCResponse{JSONRPC: "2.0", ID: id, Error: rpcErr})
		} else {
			enc.Encode(jsonRPCResponse{JSONRPC: "2.0", ID: id, Result: result})
		}
	}
}

func dispatch(req jsonRPCRequest, cfg config, client sqlCollectionClient) (any, *jsonRPCError) {
	switch req.Method {
	case "initialize":
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
		return nil, &jsonRPCError{Code: -32601, Message: "method not found: " + req.Method}
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
		return nil, &jsonRPCError{Code: -32601, Message: "unknown tool: " + p.Name}
	}
}

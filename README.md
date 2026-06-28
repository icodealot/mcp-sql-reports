# mcp-sql-reports

A local STDIO-mode [Model Context Protocol (MCP)](https://modelcontextprotocol.io) server that exposes [Oracle Cloud Infrastructure (OCI) Database Tools SQL Reports](https://docs.oracle.com/en-us/iaas/database-tools/doc/sql-reports.html) as MCP tools. This lets AI agents (Claude, Codex, Cline, etc.) discover and read your pre-defined SQL reports, giving them well-structured, context-rich queries to work with instead of generating SQL from scratch.

## Prerequisites

- An OCI account with Database Tools SQL Reports created in one or more compartments
- A working OCI SDK configuration at `~/.oci/config` (see [SDK Configuration](https://docs.oracle.com/en-us/iaas/Content/API/Concepts/sdkconfig.htm))
- Go 1.25+ to build from source, or a pre-built binary

## Building

```bash
go build -o mcp-sql-reports .
```

## CLI flags

| Flag | Required | Default | Description |
|------|----------|---------|-------------|
| `-compartment <ocid>` | No | — | Default compartment OCID used by `list_reports` when no `compartment_id` argument is passed by the agent |
| `-profile <name>` | No | `DEFAULT` | OCI config profile to use from `~/.oci/config` |
| `-exclude-report <ocid>` | No | — | Exclude a report OCID from all results; repeatable |

## MCP tools

### `list_reports`

Lists all Database Tools SQL Reports in a compartment.

**Arguments**

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `compartment_id` | string | No | OCID of the compartment to list reports from. Falls back to the `-compartment` CLI flag if omitted. |

**Returns** a JSON array of SQL report summary objects.

---

### `get_report`

Fetches the full details of a single Database Tools SQL Report by OCID.

**Arguments**

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `report_id` | string | Yes | OCID of the SQL report to fetch |

**Returns** the full SQL report object as JSON, or a message indicating the report was excluded by the `-exclude-report` filter.

## MCP server configuration

### Claude Desktop / Claude Code (`~/.mcp.json` or `.mcp.json`)

```json
{
  "mcpServers": {
    "sql-reports": {
      "type": "stdio",
      "command": "/path/to/mcp-sql-reports",
      "args": [
        "-compartment", "ocid1.compartment.oc1..aaaaaaaaexamplecompartmentocid",
        "-profile", "DEFAULT"
      ]
    }
  }
}
```

With a named profile and excluded reports:

```json
{
  "mcpServers": {
    "sql-reports": {
      "type": "stdio",
      "command": "/path/to/mcp-sql-reports",
      "args": [
        "-compartment", "ocid1.compartment.oc1..aaaaaaaaexamplecompartmentocid",
        "-profile", "PROD",
        "-exclude-report", "ocid1.databasetoolssqlreport.oc1..aaaaaaaareport1ocid",
        "-exclude-report", "ocid1.databasetoolssqlreport.oc1..aaaaaaaareport2ocid"
      ]
    }
  }
}
```

### VS Code (`.vscode/mcp.json`)

```json
{
  "servers": {
    "sql-reports": {
      "type": "stdio",
      "command": "/path/to/mcp-sql-reports",
      "args": [
        "-compartment", "ocid1.compartment.oc1..aaaaaaaaexamplecompartmentocid",
        "-profile", "DEFAULT"
      ]
    }
  }
}
```

## Authentication

The server uses the standard OCI SDK authentication chain. It reads `~/.oci/config` automatically — no paths are hard-coded. The `-profile` flag selects which profile block to use. A minimal config looks like:

```ini
[DEFAULT]
user=ocid1.user.oc1..aaaaaaaaexampleuserocid
fingerprint=aa:bb:cc:dd:ee:ff:00:11:22:33:44:55:66:77:88:99
tenancy=ocid1.tenancy.oc1..aaaaaaaaexampletenancyocid
region=us-ashburn-1
key_file=~/.oci/oci_api_key.pem
```

## SDK references

- [OCI Go SDK](https://docs.oracle.com/en-us/iaas/tools/go/latest/index.html)
- [OCI Go SDK — Database Tools](https://docs.oracle.com/en-us/iaas/tools/go/latest/databasetools/index.html)
- [MCP Specification 2025-11-25](https://modelcontextprotocol.io/specification/2025-11-25)

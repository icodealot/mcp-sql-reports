# mcp-sql-reports

A local STDIO-mode [Model Context Protocol (MCP)](https://modelcontextprotocol.io) server that exposes [Oracle Cloud Infrastructure (OCI) Database Tools SQL Reports](https://docs.oracle.com/en-us/iaas/database-tools/doc/sql-reports.html) as MCP tools. This lets AI agents (Claude, Codex, Cline, etc.) discover and read your pre-defined SQL reports, giving them well-structured, context-rich queries to work with instead of generating SQL from scratch.

## Prerequisites

- An OCI account with Database Tools SQL Reports created in one or more compartments
- Go 1.25+ to build from source, or a pre-built binary
- For `api_key` and `security_token` auth: a working OCI SDK configuration at `~/.oci/config` (see [SDK Configuration](https://docs.oracle.com/en-us/iaas/Content/API/Concepts/sdkconfig.htm))
- For `instance_principal` auth: the binary must be running on an OCI Compute instance with an attached instance principal policy

## Building

```bash
go build -o mcp-sql-reports .
```

## CLI flags

| Flag | Required | Default | Description |
|------|----------|---------|-------------|
| `-compartment <ocid>` | No | — | Default compartment OCID used by `list_reports` when no `compartment_id` argument is passed by the agent |
| `-auth <type>` | No | `api_key` | Authentication type. One of `api_key`, `instance_principal`, or `security_token` |
| `-profile <name>` | No | `DEFAULT` | OCI config profile to use from `~/.oci/config`. Ignored when `-auth instance_principal` is set |
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

**API key (default):**

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

**API key with a named profile and excluded reports:**

```json
{
  "mcpServers": {
    "sql-reports": {
      "type": "stdio",
      "command": "/path/to/mcp-sql-reports",
      "args": [
        "-compartment", "ocid1.compartment.oc1..aaaaaaaaexamplecompartmentocid",
        "-auth", "api_key",
        "-profile", "PROD",
        "-exclude-report", "ocid1.databasetoolssqlreport.oc1..aaaaaaaareport1ocid",
        "-exclude-report", "ocid1.databasetoolssqlreport.oc1..aaaaaaaareport2ocid"
      ]
    }
  }
}
```

**Instance principal (OCI Compute only):**

```json
{
  "mcpServers": {
    "sql-reports": {
      "type": "stdio",
      "command": "/path/to/mcp-sql-reports",
      "args": [
        "-compartment", "ocid1.compartment.oc1..aaaaaaaaexamplecompartmentocid",
        "-auth", "instance_principal"
      ]
    }
  }
}
```

**Security token (browser-based session login):**

```json
{
  "mcpServers": {
    "sql-reports": {
      "type": "stdio",
      "command": "/path/to/mcp-sql-reports",
      "args": [
        "-compartment", "ocid1.compartment.oc1..aaaaaaaaexamplecompartmentocid",
        "-auth", "security_token",
        "-profile", "MY_SESSION"
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

Three authentication modes are supported via the `-auth` flag.

### `api_key` (default)

Uses a standard OCI API key profile from `~/.oci/config`. The `-profile` flag selects the profile block. This is the right choice for local development.

```ini
[DEFAULT]
user=ocid1.user.oc1..aaaaaaaaexampleuserocid
fingerprint=aa:bb:cc:dd:ee:ff:00:11:22:33:44:55:66:77:88:99
tenancy=ocid1.tenancy.oc1..aaaaaaaaexampletenancyocid
region=us-ashburn-1
key_file=~/.oci/oci_api_key.pem
```

### `instance_principal`

Uses the identity of the OCI Compute instance the binary is running on. No `~/.oci/config` or `-profile` is needed — credentials are fetched automatically from the instance metadata service. The IAM policy for the instance principal must grant access to the Database Tools service.

```json
"-auth", "instance_principal"
```

### `security_token`

Uses a short-lived session token generated by the OCI CLI's interactive browser login. The token and associated key are written into `~/.oci/config` by the CLI and expire after a configurable period (default 1 hour). Use this when API key distribution is restricted or for temporary elevated access.

Generate a session with:

```bash
oci session authenticate --profile-name MY_SESSION
```

Then pass the resulting profile name:

```json
"-auth", "security_token", "-profile", "MY_SESSION"
```

The `~/.oci/config` block written by `oci session authenticate` looks like:

```ini
[MY_SESSION]
tenancy=ocid1.tenancy.oc1..aaaaaaaaexampletenancyocid
region=us-ashburn-1
key_file=~/.oci/sessions/MY_SESSION/oci_api_key.pem
fingerprint=aa:bb:cc:dd:ee:ff:00:11:22:33:44:55:66:77:88:99
security_token_file=~/.oci/sessions/MY_SESSION/token
```

## SDK references

- [OCI Go SDK](https://docs.oracle.com/en-us/iaas/tools/go/latest/index.html)
- [OCI Go SDK — Database Tools](https://docs.oracle.com/en-us/iaas/tools/go/latest/databasetools/index.html)
- [MCP Specification 2025-11-25](https://modelcontextprotocol.io/specification/2025-11-25)

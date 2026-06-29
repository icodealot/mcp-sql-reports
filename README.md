# mcp-sql-reports

A local STDIO-mode [Model Context Protocol (MCP)](https://modelcontextprotocol.io) server that exposes
Oracle Cloud Infrastructure (OCI) [Database Tools SQL Reports](https://docs.oracle.com/en-us/iaas/database-tools/doc/database-tools-sql-reports.html) 
as MCP tools. This lets AI agents (Claude, Codex, Cline, etc.) discover and read your pre-defined SQL 
reports, giving them well-structured, context-rich queries to work with instead of generating SQL 
from scratch.

There is a demo of creating a Database Tools SQL report using Terraform included in this repo. You can
use this to configure a SQL report that you can then use for testing with this MCP server to get up
and running.

- See the related [README](demo-resources/README.md)


## Prerequisites

- An OCI account with Database Tools SQL Reports created in one or more compartments
- Go 1.25+ to build or install from source
- For `api_key` and `security_token` auth: a working OCI SDK configuration at `~/.oci/config` 
(see [SDK Configuration](https://docs.oracle.com/en-us/iaas/Content/API/Concepts/sdkconfig.htm))
- For `instance_principal` auth: the binary must be running on an OCI Compute instance with an attached
instance principal policy


## Installing


**From source via GitHub:**

To download, build, and install from source: (assumes Golang is installed on your machine)

```bash
go install github.com/icodealot/mcp-sql-reports@latest
```

Note: by default the binary will be in `$(go env GOPATH)/bin` (defaults to `~/go/bin`). Make sure that
directory is on your `PATH` for ease of use.


**Locally from a cloned repo:**

To build and install from cloned source: (assumes Golang is installed on your machine)

```bash
cd <path-to-clone>/mcp-sql-reports
go test ./...
go install
```

Note: by default the binary will be in `$(go env GOPATH)/bin` (defaults to `~/go/bin`). Make sure that
directory is on your `PATH` for ease of use.


## MCP server usage

To use the `mcp-sql-reports` MCP server, you can supply the following options in your MCP configuration.

| Flag | Required | Default | Description |
|------|----------|---------|-------------|
| `-compartment <ocid>` | No | — | Default compartment OCID used by `list_reports` when no `compartment_id` argument is passed by the agent |
| `-auth <type>` | No | `api_key` | Authentication type. One of `api_key`, `instance_principal`, or `security_token` |
| `-profile <name>` | No | `DEFAULT` | OCI config profile to use from `~/.oci/config`. Ignored when `-auth instance_principal` is set |
| `-exclude-report <ocid>` | No | — | Exclude a report OCID from all results; repeatable |

You will find some example MCP configurations using various options below.


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

**Returns** the full SQL report object as JSON, or a message indicating the report was excluded
by the `-exclude-report` filter.


## MCP server configuration examples

### Claude Desktop / Claude Code (`~/.mcp.json` or `.mcp.json`)

**API key (default):**

```json
{
  "mcpServers": {
    "mcp-sql-reports": {
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
    "mcp-sql-reports": {
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
    "mcp-sql-reports": {
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
    "mcp-sql-reports": {
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
    "mcp-sql-reports": {
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

Uses a standard OCI API key profile from `~/.oci/config`. The `-profile` flag selects the profile block.
This is the right choice for local development.

```ini
[DEFAULT]
user=ocid1.user.oc1..aaaaaaaaexampleuserocid
fingerprint=aa:bb:cc:dd:ee:ff:00:11:22:33:44:55:66:77:88:99
tenancy=ocid1.tenancy.oc1..aaaaaaaaexampletenancyocid
region=us-ashburn-1
key_file=~/.oci/oci_api_key.pem
```

### `instance_principal`

Uses the identity of the OCI Compute instance the binary is running on. No `~/.oci/config` or
`-profile` is needed — credentials are fetched automatically from the instance metadata service.
The IAM policy for the instance principal must grant access to `database-tools-sql-reports` or
a resource family that includes it such as `database-tools-family`.

```json
"-auth", "instance_principal"
```

### `security_token`

Uses a short-lived session token generated by the OCI CLI's interactive browser login. The token
and associated key are written into `~/.oci/config` by the CLI and expire after a configurable
period (default 1 hour). 

Use this when API key distribution is restricted or for temporary elevated access.

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

## Debugging access to SQL Reports in OCI

This MCP server assumes that you are working inside a single OCI compartment. (It is just a demo after all!)

After configuring your MCP server if you are getting 404 responses from OCI you should confirm that your OCI
principal (user, instance, etc.) has the necessary IAM policies for the compartment specified. A good way to
validate this would be to use the OCI CLI to list SQL reports in a given compartment.

If your OCI compartment is `ocid1.compartment.oc1..aaaaaaaaexamplecompartmentocid` then try:

```bash
oci --profile <YOUR_OCI_CONFIG_PROFILE> dbtools sql-report list --compartment-id ocid1.compartment.oc1..aaaaaaaaexamplecompartmentocid
```

If you get `404` responses from OCI then you likely have a policy or OCI profile issue.

### Symptom: the MCP server has never worked

This MCP server only does READ operations so an IAM policy such as this should do the trick:

```
allow group '<domain_name>/<group_name>' to read database-tools-sql-reports in compartment <your-compartment>
```

### Symptom: I am using `-auth security_token` and the MCP server stopped working suddenly

Keep in mind that OCI profiles created as session tokens using `oci session authenticate` have an expiration
time. This MCP server does **not** automatically attempt to refresh said tokens. Therefore, you may need to
refresh the token if it is expired! 

You can refresh a session with the OCI CLI by using:

```bash
oci session refresh --profile <YOUR_OCI_CONFIG_PROFILE>
```

This is not a bug, it is a feature of ephemeral session tokens!

Note: I have only tested that `api_key` and `security_token` authentication are wired up correctly. If you
try out instance principal authentication and it works correctly please shoot me a note on social media where
you heard about this repository!


## References

- [OCI Go SDK](https://docs.oracle.com/en-us/iaas/tools/go/latest/index.html)
- [OCI Go SDK — Database Tools](https://docs.oracle.com/en-us/iaas/tools/go/latest/databasetools/index.html)
- [MCP Specification 2025-11-25](https://modelcontextprotocol.io/specification/2025-11-25)

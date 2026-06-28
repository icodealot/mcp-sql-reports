# Demo Resources

This Terraform configuration creates a sample **Database Tools SQL Report** in Oracle Cloud
Infrastructure (OCI) to demonstrate how the `mcp-sql-reports` MCP server consumes them.

## What is a Database Tools SQL Report?

A [Database Tools SQL Report](https://docs.oracle.com/en-us/iaas/database-tools/doc/database-tools-sql-reports.html)
is a first-class OCI resource that pairs a SQL statement with structured metadata — including a
human-readable purpose, per-column descriptions, and parameterized variables — so that AI agents
have enough context to decide *when* and *how* to run the query without any additional prompting.

The MCP server in this project lists and retrieves these reports from a compartment and exposes them
as MCP tools, giving agents like Claude a well-defined NL-to-SQL mapping they can invoke directly.

## What this configuration creates

| Resource | Description |
|---|---|
| `oci_database_tools_database_tools_sql_report.demo` | A sample SQL report that queries `ALL_OBJECTS` to list schema objects filtered by status |

The report includes:
- A descriptive `purpose` and `instructions` field so agents understand when to use it
- `columns` metadata describing every column returned by the query (`OWNER`, `OBJECT_NAME`, `OBJECT_TYPE`, `STATUS`, `LAST_DDL_TIME`, `CREATED`)
- A `variables` block for the `STATUS` bind variable used in the `WHERE` clause (`VALID` or `INVALID`)
- No DBA privileges required — the query runs against `ALL_OBJECTS`

## Prerequisites

- Terraform >= 1.5
- OCI provider >= 8.20
- A valid `~/.oci/config` profile (or equivalent environment auth)
- An OCI compartment OCID where you have permission to manage Database Tools resources

## Usage

```bash
cd demo-resources

terraform init

# Validate the configuration and preview what will be created
terraform plan \
  -var="oci_config_profile=DEFAULT" \
  -var="compartment_id=ocid1.compartment.oc1..example"

# Apply after reviewing the plan output
terraform apply \
  -var="oci_config_profile=DEFAULT" \
  -var="compartment_id=ocid1.compartment.oc1..example"
```

Run `terraform plan` first to validate the configuration and review the resources that will be
created before committing. Terraform will list every resource it intends to add, change, or destroy
and exit without making any changes — a safe way to catch misconfiguration or unexpected diffs.

After apply the SQL report OCID is written to `sql_report_id` in the Terraform output and can be
used to verify the MCP server can discover and read the report.

## Variables

| Name | Description | Default |
|---|---|---|
| `oci_config_profile` | Profile name in `~/.oci/config` to authenticate with | — (required) |
| `compartment_id` | OCID of the OCI compartment to create the report in | — (required) |

## Outputs

| Name | Description |
|---|---|
| `sql_report_id` | OCID of the created Database Tools SQL Report |
| `sql_report_state` | Lifecycle state of the report |

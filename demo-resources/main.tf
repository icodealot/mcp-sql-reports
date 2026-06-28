terraform {
  required_version = ">= 1.3"

  required_providers {
    oci = {
      source  = "oracle/oci"
      version = ">= 8.20"
    }
  }
}

provider "oci" {
  config_file_profile = var.oci_config_profile
}

# Demo SQL report: list schema objects filtered by status.
# Uses ALL_OBJECTS so the query runs without DBA privileges.
# The :STATUS bind variable is exposed through the `variables` block.
resource "oci_database_tools_database_tools_sql_report" "demo" {
  compartment_id = var.compartment_id
  display_name   = "demo-schema-objects"
  type           = "ORACLE_DATABASE"

  purpose = "List database objects visible to the current user, filtered by object status."

  instructions = <<-EOT
    Use this report to inspect objects in the current schema or any schema accessible
    to the connected user. Supply the STATUS variable with one of: VALID, INVALID.
    Run this report when a user asks about compiled objects, invalid packages,
    or the overall state of schema objects.
  EOT

  source = <<-EOT
    SELECT
      o.owner,
      o.object_name,
      o.object_type,
      o.status,
      o.last_ddl_time,
      o.created
    FROM
      all_objects o
    WHERE
      o.status = :STATUS
    ORDER BY
      o.owner,
      o.object_type,
      o.object_name
  EOT

  description = "Returns objects from ALL_OBJECTS matching the supplied STATUS bind variable. No DBA privileges required."

  # Describe each column returned by the query so agents can interpret results.
  columns {
    name        = "OWNER"
    type        = "VARCHAR2"
    description = "Schema that owns the object."
  }

  columns {
    name        = "OBJECT_NAME"
    type        = "VARCHAR2"
    description = "Name of the database object."
  }

  columns {
    name        = "OBJECT_TYPE"
    type        = "VARCHAR2"
    description = "Type of the object: TABLE, VIEW, PACKAGE, PROCEDURE, FUNCTION, INDEX, TRIGGER, etc."
  }

  columns {
    name        = "STATUS"
    type        = "VARCHAR2"
    description = "Compilation status of the object: VALID or INVALID."
  }

  columns {
    name        = "LAST_DDL_TIME"
    type        = "DATE"
    description = "Timestamp of the last DDL change to the object."
  }

  columns {
    name        = "CREATED"
    type        = "DATE"
    description = "Timestamp when the object was created."
  }

  # Expose the bind variable so agents know what parameter to supply.
  variables {
    name        = "STATUS"
    type        = "VARCHAR2"
    description = "Object status filter. Accepted values: VALID, INVALID."
  }

  freeform_tags = {
    "project" = "mcp-sql-reports-demo"
  }
}

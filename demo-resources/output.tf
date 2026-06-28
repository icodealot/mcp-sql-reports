output "sql_report_id" {
  description = "OCID of the created Database Tools SQL Report."
  value       = oci_database_tools_database_tools_sql_report.demo.id
}

output "sql_report_state" {
  description = "Lifecycle state of the created SQL Report."
  value       = oci_database_tools_database_tools_sql_report.demo.state
}

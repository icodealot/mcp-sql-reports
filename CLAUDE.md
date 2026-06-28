# CLAUDE.md

This is an example of a local STDIO mcp server that can be used with
Database Tools SQL Reports to give your AI agents (Claude, Codex, Cline, etc.)
better well-defined SQL statements for your use case. SQL Reports are first-class
resources in Oracle Cloud Infrastructure (OCI) and they can be created using any
of the standard OCI tools such as the OCI CLI, the SDK, or the Terraform provider.

## About This Project

This project contains a sample Terraform configuration as well as an example of
using the OCI Go SDK to create a local STDIO-mode MCP server that can be used
with the Database Tools SQL Report resources to create your own custom natural
language (NL) to structured query language (SQL) mapping. You define your SQL
in the SQL reports and give enough context for the agent with the metadata and
then use this MCP server to list and read SQL reports in a specified compartment.

All of this relies on the standard OCI authentication mechanisms for the SDK so 
you will need a ~/.oci/config file setup with appropriate settings. The STDIO 
MCP server then will send `list_reports` and `get_report` commands to OCI and
return the results to your AI agent in an MCP-compliant way.

## Agent guidelines

- Never perform Git write operations
- You are working collaboratively with a developer so never perform `git reset...`.
- Always develop tests for a feature as a specification, confirm they fail, then implement the minimal amount of code necessary for the tests to pass
- Use as little abstraction as possible
- When a function is larger than 5 lines use minimal Go comments to explain the implementation in idomatic Golang style
- This is a local STDIO-mode MCP server. Ensure that the features included are compliant with the MCP specification at: 
  - https://modelcontextprotocol.io/specification/2025-11-25
- Use the OCI Go SDK documentation as the reference for all SDK calls:
  - https://docs.oracle.com/en-us/iaas/tools/go/latest/index.html
  - https://docs.oracle.com/en-us/iaas/tools/go/latest/databasetools/index.html

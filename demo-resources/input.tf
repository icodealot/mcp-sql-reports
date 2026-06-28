variable "oci_config_profile" {
  type        = string
  description = "Profile name in ~/.oci/config used to authenticate with OCI."

  validation {
    condition     = length(trimspace(var.oci_config_profile)) > 0
    error_message = "oci_config_profile must not be empty."
  }
}

variable "compartment_id" {
  type        = string
  description = "OCID of the OCI compartment in which to create the demo SQL report."

  validation {
    condition     = can(regex("^ocid1\\.compartment\\.", var.compartment_id))
    error_message = "compartment_id must be a valid compartment OCID (starts with 'ocid1.compartment.')."
  }
}

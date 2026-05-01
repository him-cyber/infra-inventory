variable "location" {
  type    = string
  default = "eastus"
}

variable "name_prefix" {
  type    = string
  default = "iinv"
}

variable "api_image" {
  type = string
}

variable "indexer_image" {
  type = string
}

variable "web_image" {
  type = string
}

variable "opensearch_image" {
  type    = string
  default = "opensearchproject/opensearch:3.5.0"
}

variable "azure_tenant_id" {
  type        = string
  description = "Azure Entra ID tenant for OIDC sign-in."
  default     = ""
}

variable "azure_client_id" {
  type        = string
  description = "Azure app registration client ID."
  default     = ""
}

variable "azure_client_secret" {
  type        = string
  description = "Azure app registration client secret."
  sensitive   = true
  default     = ""
}

variable "auth_session_key" {
  type        = string
  description = "Base64-encoded 32-byte key used to encrypt browser session cookies."
  sensitive   = true
  default     = ""
}

variable "auth_redirect_url" {
  type        = string
  description = "OIDC callback URL registered in Azure Entra ID."
  default     = ""
}

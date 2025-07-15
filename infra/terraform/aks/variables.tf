variable "subscription_id" {
  type        = string
  description = "Azure Subscription ID"
}

variable "environment" {
  type        = string
  description = "Environment name, e.g. dev, prod"
}

variable "deploy_aks" {
  type        = bool
  description = "Flag to deploy Azure Kubernetes Service (AKS)"
  default     = true
}

variable "aks_config" {
  type = object({
    kubernetes_version : string
    default_node_count : number
    default_vm_size : string
    default_os_disk_size_gb : number
  })
  description = "Configuration for AKS cluster"
}

variable "aks_eso_infisical_client_id" {
  type        = string
  description = "Infisical identity client ID for accessing secrets in AKS"
  sensitive   = true
}

variable "aks_eso_infisical_client_secret" {
  type        = string
  description = "Infisical identity client secret for accessing secrets in AKS"
  sensitive   = true
}

variable "infisical_project_slug" {
  type        = string
  description = "Infisical project slug for secret management"
}

variable "subscription_id" {
  type        = string
  description = "Azure Subscription ID"
}

variable "status_collector_image_tag" {
  type        = string
  description = "Status collector version tag: e.g. 1.0"
}

variable "environment" {
  type        = string
  description = "Environment name, e.g. dev, prod"
}

variable "environment_variables" {
  type        = map(string)
  description = "Environment variables for the container"
  default     = {}
}

variable "deploy_aci" {
  type        = bool
  description = "Flag to deploy Azure Container Instance (ACI)"
}

variable "subscription_id" {
  type        = string
  description = "Azure Subscription ID"
}

variable "queue_monitor_image_tag" {
  type        = string
  description = "Queue monitor version tag: e.g. 1.0"
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

variable "notification_telegram_broadcast_channel_name" {
  type        = string
  description = "Notification broadcast telegram channel name"
}

variable "notification_telegram_bot_token" {
  type        = string
  description = "Telegram bot token"
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

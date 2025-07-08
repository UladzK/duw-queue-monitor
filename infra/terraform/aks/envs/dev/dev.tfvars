subscription_id = "77a70a5e-2230-43b7-8983-61e7497498a8"
queue_monitor_image_tag= "0.4.1"
environment = "dev"
// todo: get sensitive values from a secret vault
environment_variables = {
    "LOG_LEVEL" = "debug"
    "USE_TELEGRAM_NOTIFICATIONS"="true"
    "STATE_REDIS_CONNECTION_STRING"="redis://localhost:6379/0"
    "STATUS_CHECK_INTERVAL_SECONDS"="300"
}
deploy_aci = false

aks_config = {
    kubernetes_version = "1.33.0"
    default_node_count         = 1
    default_vm_size            = "Standard_B2s"
    default_os_disk_size_gb    = 30
}

infisical_project_slug = "duw-monitor-secrets-default"
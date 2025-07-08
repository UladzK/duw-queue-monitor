locals {
  service_image_name = "queue-monitor"
  service_name       = replace(local.service_image_name, "-", "")
  location           = "Poland Central"
  location_short     = "plc"
  infisical_region   = "eu"

  acr_identity_id  = data.terraform_remote_state.shared.outputs.acr_app_pull_identity_id
  acr_login_server = data.terraform_remote_state.shared.outputs.acr_login_server

  telegram_env_vars = {
    "NOTIFICATION_TELEGRAM_BOT_TOKEN"              = var.notification_telegram_bot_token
    "NOTIFICATION_TELEGRAM_BROADCAST_CHANNEL_NAME" = var.notification_telegram_broadcast_channel_name
  }
}

resource "azurerm_resource_group" "rg_aks" {
  name     = "rg-aks-${var.environment}"
  location = local.location
}

resource "azurerm_kubernetes_cluster" "aks" {
  name                = "aks-duw-${var.environment}-${local.location_short}"
  location            = azurerm_resource_group.rg_aks.location
  resource_group_name = azurerm_resource_group.rg_aks.name
  kubernetes_version  = var.aks_config.kubernetes_version
  sku_tier            = "Free"

  private_cluster_enabled = false
  # TODO: for some reason the aks cluster is changed with empty api_server_access_profile
  # see issue: https://github.com/hashicorp/terraform-provider-azurerm/issues/20085
  api_server_access_profile {
    authorized_ip_ranges = [] # allow access from all IPs
  }
  role_based_access_control_enabled = true
  local_account_disabled            = true
  azure_active_directory_role_based_access_control {
    azure_rbac_enabled     = true
    tenant_id              = data.azurerm_subscription.current.tenant_id
    admin_group_object_ids = [data.terraform_remote_state.shared.outputs.ug_aks_admins_object_id]
  }

  dns_prefix = "aksduw-${var.environment}"

  default_node_pool {
    name            = "default"
    node_count      = var.aks_config.default_node_count
    vm_size         = var.aks_config.default_vm_size
    os_disk_size_gb = var.aks_config.default_os_disk_size_gb
    os_disk_type    = "Ephemeral"
    os_sku          = "Ubuntu"

    upgrade_settings {
      max_surge                     = "100%"
      drain_timeout_in_minutes      = 5
      node_soak_duration_in_minutes = 1
    }
  }

  identity {
    type = "SystemAssigned"
  }

  tags = {
    environment = var.environment
  }
}

resource "azurerm_role_assignment" "aks_acr_pull" {
  scope                = data.terraform_remote_state.shared.outputs.acr_id
  principal_id         = azurerm_kubernetes_cluster.aks.kubelet_identity[0].object_id
  role_definition_name = "AcrPull"
}

resource "kubernetes_secret" "infisical_universal_identity" {
  metadata {
    name = "infisical-universal-auth-credentials"
  }
  type = "Opaque"

  data = {
    clientId     = var.aks_eso_infisical_client_id
    clientSecret = var.aks_eso_infisical_client_secret
  }
}

resource "kubernetes_manifest" "eso_infisical_secret_store" {
  manifest = yamldecode(templatefile("${path.module}/k8s/eso-infisical-secret-store.yml", {
    infisical_universal_auth_credentials_secret_name = kubernetes_secret.infisical_universal_identity.metadata[0].name
    infisical_project_slug                           = var.infisical_project_slug
    infisical_environment_slug                       = var.environment
    infisical_region                                 = local.infisical_region
  }))

  depends_on = [kubernetes_secret.infisical_universal_identity]
}

# resource "azurerm_resource_group" "rg_aci" {
#   name     = "rg-${local.service_name}-${var.environment}"
#   location = local.location
# }

# resource "azurerm_container_group" "aci" {
#   count = var.deploy_aci ? 1 : 0

#   name                = local.service_name
#   location            = azurerm_resource_group.rg_aci.location
#   resource_group_name = azurerm_resource_group.rg_aci.name
#   os_type             = "Linux"
#   ip_address_type     = "Public"
#   restart_policy      = "OnFailure"

#   identity {
#     type         = "UserAssigned"
#     identity_ids = [local.acr_identity_id]
#   }

#   image_registry_credential {
#     server                    = local.acr_login_server
#     user_assigned_identity_id = local.acr_identity_id
#   }

#   container {
#     name   = "aci-${local.service_name}-${var.environment}"
#     image  = "${local.acr_login_server}/${local.service_image_name}:${var.queue_monitor_image_tag}"
#     cpu    = "0.5"
#     memory = "0.5"

#     ports {
#       port     = 80
#       protocol = "TCP"
#     }

#     secure_environment_variables = merge(local.telegram_env_vars, var.environment_variables)
#   }

#   tags = {
#     environment  = var.environment
#     service_name = local.service_name
#   }
# }

locals {
  service_image_name = "queue-monitor"
  service_name       = replace(local.service_image_name, "-", "")
  location           = "Poland Central"
  location_short     = "plc"
  infisical_region   = "eu"

  acr_identity_id  = data.terraform_remote_state.shared.outputs.acr_app_pull_identity_id
  acr_login_server = data.terraform_remote_state.shared.outputs.acr_login_server
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

  private_cluster_enabled = false # it's acceptable to allow access from all IPs. vNet integration is too complex and expensive. access is controlled by RBAC
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
    admin_group_object_ids = [data.terraform_remote_state.shared.outputs.aks_admins_group_object_id]
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

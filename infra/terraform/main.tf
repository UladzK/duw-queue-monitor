locals {
  service_name = "statuscollector"
  location     = "Poland Central"
  telegram_env_vars = {
    "NOTIFICATION_TELEGRAM_BOT_TOKEN"              = var.notification_telegram_bot_token
    "NOTIFICATION_TELEGRAM_BROADCAST_CHANNEL_NAME" = var.notification_telegram_broadcast_channel_name
  }
}

resource "azurerm_resource_group" "rg" {
  name     = "rg-${local.service_name}-${var.environment}"
  location = local.location
}

// todo: switch to a shared ACR
resource "azurerm_container_registry" "acr" {
  name                = "acrduwshared"
  resource_group_name = azurerm_resource_group.rg.name
  location            = azurerm_resource_group.rg.location
  sku                 = "Basic"
}

resource "azurerm_user_assigned_identity" "identity" {
  name                = "uami-${local.service_name}-identity-${var.environment}"
  resource_group_name = azurerm_resource_group.rg.name
  location            = azurerm_resource_group.rg.location
}

resource "azurerm_role_assignment" "acr_pull" {
  scope                = azurerm_container_registry.acr.id
  role_definition_name = "AcrPull"
  principal_id         = azurerm_user_assigned_identity.identity.principal_id
}

resource "azurerm_container_group" "aci" {
  count = var.deploy_aci ? 1 : 0

  name                = local.service_name
  location            = azurerm_resource_group.rg.location
  resource_group_name = azurerm_resource_group.rg.name
  os_type             = "Linux"
  ip_address_type     = "Public"
  restart_policy      = "OnFailure"

  identity {
    type         = "UserAssigned"
    identity_ids = [azurerm_user_assigned_identity.identity.id]
  }

  image_registry_credential {
    server                    = azurerm_container_registry.acr.login_server
    user_assigned_identity_id = azurerm_user_assigned_identity.identity.id
  }

  container {
    name   = "aci-${local.service_name}-${var.environment}"
    image  = "${azurerm_container_registry.acr.login_server}/queue-monitor:${var.status_collector_image_tag}"
    cpu    = "0.5"
    memory = "0.5"

    ports {
      port     = 80
      protocol = "TCP"
    }

    secure_environment_variables = merge(local.telegram_env_vars, var.environment_variables)
  }

  tags = {
    environment  = var.environment
    service_name = local.service_name
  }
}

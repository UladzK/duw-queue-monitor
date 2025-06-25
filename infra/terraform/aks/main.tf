locals {
  service_image_name = "queue-monitor"
  service_name       = replace(local.service_image_name, "-", "")
  location           = "Poland Central"

  acr_identity_id  = data.terraform_remote_state.shared.outputs.acr_app_pull_identity_id
  acr_login_server = data.terraform_remote_state.shared.outputs.acr_login_server

  telegram_env_vars = {
    "NOTIFICATION_TELEGRAM_BOT_TOKEN"              = var.notification_telegram_bot_token
    "NOTIFICATION_TELEGRAM_BROADCAST_CHANNEL_NAME" = var.notification_telegram_broadcast_channel_name
  }
}

resource "azurerm_resource_group" "rg_aci" {
  name     = "rg-${local.service_name}-${var.environment}"
  location = local.location
}

resource "azurerm_container_group" "aci" {
  count = var.deploy_aci ? 1 : 0

  name                = local.service_name
  location            = azurerm_resource_group.rg_aci.location
  resource_group_name = azurerm_resource_group.rg_aci.name
  os_type             = "Linux"
  ip_address_type     = "Public"
  restart_policy      = "OnFailure"

  identity {
    type         = "UserAssigned"
    identity_ids = [local.acr_identity_id]
  }

  image_registry_credential {
    server                    = local.acr_login_server
    user_assigned_identity_id = local.acr_identity_id
  }

  container {
    name   = "aci-${local.service_name}-${var.environment}"
    image  = "${local.acr_login_server}/${local.service_image_name}:${var.queue_monitor_image_tag}"
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

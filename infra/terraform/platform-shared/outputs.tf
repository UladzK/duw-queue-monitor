output "acr_login_server" {
  value = azurerm_container_registry.acr.login_server

  description = "FQDN of the Azure Container Registry (ACR) for pulling images"
}

output "acr_app_pull_identity_id" {
  value = azurerm_user_assigned_identity.app.id

  description = "User Assigned Identity ID for ACR pull access to run app containers"
}

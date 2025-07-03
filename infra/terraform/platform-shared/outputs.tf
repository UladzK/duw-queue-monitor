output "acr_id" {
  value = azurerm_container_registry.acr.id

  description = "ID of the Azure Container Registry (ACR) for pulling images"
}

output "acr_login_server" {
  value = azurerm_container_registry.acr.login_server

  description = "FQDN of the Azure Container Registry (ACR) for pulling images"
}

output "acr_app_pull_identity_id" {
  value = azurerm_user_assigned_identity.app.id

  description = "User Assigned Identity ID for ACR pull access to run app containers"
}

output "ug_aks_admins_object_id" {
  value = azuread_group.aks_admins_group.object_id

  description = "Azure AD group ID for AKS admins"
}

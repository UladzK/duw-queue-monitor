locals {
  location = "Poland Central"
}

resource "azurerm_resource_group" "rg_shared" {
  name     = "rg-shared"
  location = local.location
}
resource "azurerm_container_registry" "acr" {
  name                = "acrduwshared"
  resource_group_name = azurerm_resource_group.rg_shared.name
  location            = azurerm_resource_group.rg_shared.location
  sku                 = "Basic"
}

resource "azurerm_user_assigned_identity" "app" {
  name                = "uami-acr-app-pull-shared"
  resource_group_name = azurerm_resource_group.rg_shared.name
  location            = azurerm_resource_group.rg_shared.location
}
resource "azurerm_role_assignment" "app_pull" {
  scope                = azurerm_container_registry.acr.id
  role_definition_name = "AcrPull"
  principal_id         = azurerm_user_assigned_identity.app.principal_id
}

resource "azurerm_resource_group" "rg_tfstate" {
  name     = "rg-tfstate-shared"
  location = local.location
}

// todo: configure encryption at rest for the storage account
resource "azurerm_storage_account" "sa_tfstate" {
  name                     = "saduwtfstateshared"
  resource_group_name      = azurerm_resource_group.rg_tfstate.name
  location                 = azurerm_resource_group.rg_tfstate.location
  account_tier             = "Standard"
  account_replication_type = "LRS"
  access_tier              = "Cool"

  blob_properties {
    versioning_enabled = true
    delete_retention_policy {
      days = 30
    }
  }

}

resource "azurerm_storage_container" "sc_tfstate" {
  name                  = "scduwtfstate"
  storage_account_id    = azurerm_storage_account.sa_tfstate.id
  container_access_type = "private"
}

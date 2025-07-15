locals {
  infisical_region = "eu"
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

  depends_on = [azurerm_kubernetes_cluster.aks]
}

resource "kubernetes_manifest" "eso_infisical_secret_store" {
  manifest = yamldecode(templatefile("${path.module}/k8s/eso-infisical-secret-store.yml", {
    infisical_universal_auth_credentials_secret_name = kubernetes_secret.infisical_universal_identity.metadata[0].name
    infisical_project_slug                           = var.infisical_project_slug
    infisical_environment_slug                       = var.environment
    infisical_region                                 = local.infisical_region
  }))

  depends_on = [azurerm_kubernetes_cluster.aks, kubernetes_secret.infisical_universal_identity]
}

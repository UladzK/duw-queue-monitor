provider "kubernetes" {
  config_path = "~/.kube/config" # TODO: temporary solution, should be replaced with proper AKS auth to enable CI/CD
}

resource "ovh_cloud_project_alerting" "monthly_budget" {
  service_name      = var.project_id
  email             = var.ovh_billing_alert_email
  monthly_threshold = var.monthly_threshold_eur
  delay             = 86400
}

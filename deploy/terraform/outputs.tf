output "api_url" {
  value = "https://${azurerm_container_app.api.ingress[0].fqdn}"
}

output "web_url" {
  value = "https://${azurerm_container_app.web.ingress[0].fqdn}"
}

output "eventhub_namespace" {
  value = azurerm_eventhub_namespace.main.name
}

output "storage_account" {
  value = azurerm_storage_account.config.name
}

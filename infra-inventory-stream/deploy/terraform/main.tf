resource "random_string" "suffix" {
  length  = 6
  lower   = true
  numeric = true
  special = false
  upper   = false
}

locals {
  name = "${var.name_prefix}-${random_string.suffix.result}"
}

resource "azurerm_resource_group" "main" {
  name     = "rg-${local.name}"
  location = var.location
}

resource "azurerm_log_analytics_workspace" "main" {
  name                = "log-${local.name}"
  location            = azurerm_resource_group.main.location
  resource_group_name = azurerm_resource_group.main.name
  sku                 = "PerGB2018"
  retention_in_days   = 30
}

resource "azurerm_container_app_environment" "main" {
  name                       = "cae-${local.name}"
  location                   = azurerm_resource_group.main.location
  resource_group_name        = azurerm_resource_group.main.name
  log_analytics_workspace_id = azurerm_log_analytics_workspace.main.id
}

resource "azurerm_container_registry" "main" {
  name                = replace("acr${local.name}", "-", "")
  resource_group_name = azurerm_resource_group.main.name
  location            = azurerm_resource_group.main.location
  sku                 = "Basic"
  admin_enabled       = false
}

resource "azurerm_storage_account" "config" {
  name                     = replace("st${local.name}", "-", "")
  resource_group_name      = azurerm_resource_group.main.name
  location                 = azurerm_resource_group.main.location
  account_tier             = "Standard"
  account_replication_type = "LRS"
}

resource "azurerm_storage_container" "config" {
  name                  = "config"
  storage_account_id    = azurerm_storage_account.config.id
  container_access_type = "private"
}

resource "azurerm_storage_blob" "config" {
  name                   = "inventory-config.json"
  storage_account_name   = azurerm_storage_account.config.name
  storage_container_name = azurerm_storage_container.config.name
  type                   = "Block"
  source                 = "${path.module}/../../config/inventory-config.json"
}

resource "azurerm_eventhub_namespace" "main" {
  name                = "ehns-${local.name}"
  location            = azurerm_resource_group.main.location
  resource_group_name = azurerm_resource_group.main.name
  sku                 = "Standard"
  capacity            = 1
}

resource "azurerm_eventhub" "events" {
  name              = "inventory.events"
  namespace_id      = azurerm_eventhub_namespace.main.id
  partition_count   = 2
  message_retention = 1
}

resource "azurerm_eventhub_namespace_authorization_rule" "app" {
  name                = "app-producer-consumer"
  namespace_name      = azurerm_eventhub_namespace.main.name
  resource_group_name = azurerm_resource_group.main.name
  listen              = true
  send                = true
  manage              = false
}

resource "azurerm_container_app" "opensearch" {
  name                         = "opensearch-${local.name}"
  container_app_environment_id = azurerm_container_app_environment.main.id
  resource_group_name          = azurerm_resource_group.main.name
  revision_mode                = "Single"

  ingress {
    external_enabled = false
    target_port      = 9200
    transport        = "http"
    traffic_weight {
      percentage      = 100
      latest_revision = true
    }
  }

  template {
    min_replicas = 1
    max_replicas = 1
    container {
      name   = "opensearch"
      image  = var.opensearch_image
      cpu    = 1
      memory = "2Gi"
      env {
        name  = "discovery.type"
        value = "single-node"
      }
      env {
        name  = "DISABLE_SECURITY_PLUGIN"
        value = "true"
      }
      env {
        name  = "OPENSEARCH_JAVA_OPTS"
        value = "-Xms512m -Xmx512m"
      }
    }
  }
}

resource "azurerm_container_app" "api" {
  name                         = "api-${local.name}"
  container_app_environment_id = azurerm_container_app_environment.main.id
  resource_group_name          = azurerm_resource_group.main.name
  revision_mode                = "Single"

  identity {
    type = "SystemAssigned"
  }

  secret {
    name  = "eventhub-connection-string"
    value = azurerm_eventhub_namespace_authorization_rule.app.primary_connection_string
  }

  secret {
    name  = "azure-client-secret"
    value = var.azure_client_secret
  }

  secret {
    name  = "auth-session-key"
    value = var.auth_session_key
  }

  ingress {
    external_enabled = true
    target_port      = 8080
    transport        = "http"
    traffic_weight {
      percentage      = 100
      latest_revision = true
    }
  }

  template {
    min_replicas = 1
    max_replicas = 1
    container {
      name   = "api"
      image  = var.api_image
      cpu    = 0.5
      memory = "1Gi"
      env {
        name  = "CONFIG_SOURCE"
        value = "azure"
      }
      env {
        name  = "AZURE_STORAGE_ACCOUNT_URL"
        value = azurerm_storage_account.config.primary_blob_endpoint
      }
      env {
        name  = "KAFKA_BROKERS"
        value = "${azurerm_eventhub_namespace.main.name}.servicebus.windows.net:9093"
      }
      env {
        name  = "KAFKA_TOPIC"
        value = azurerm_eventhub.events.name
      }
      env {
        name  = "OPENSEARCH_URL"
        value = "http://${azurerm_container_app.opensearch.ingress[0].fqdn}"
      }
      env {
        name        = "EVENTHUB_CONNECTION_STRING"
        secret_name = "eventhub-connection-string"
      }
      env {
        name  = "AUTH_MODE"
        value = var.azure_tenant_id == "" ? "dev" : "azure"
      }
      env {
        name  = "AUTH_REQUIRED"
        value = var.azure_tenant_id == "" ? "false" : "true"
      }
      env {
        name  = "COOKIE_SECURE"
        value = "true"
      }
      env {
        name  = "AUTH_REDIRECT_URL"
        value = var.auth_redirect_url
      }
      env {
        name  = "AZURE_TENANT_ID"
        value = var.azure_tenant_id
      }
      env {
        name  = "AZURE_CLIENT_ID"
        value = var.azure_client_id
      }
      env {
        name        = "AZURE_CLIENT_SECRET"
        secret_name = "azure-client-secret"
      }
      env {
        name        = "AUTH_SESSION_KEY"
        secret_name = "auth-session-key"
      }
    }
  }
}

resource "azurerm_container_app" "indexer" {
  name                         = "indexer-${local.name}"
  container_app_environment_id = azurerm_container_app_environment.main.id
  resource_group_name          = azurerm_resource_group.main.name
  revision_mode                = "Single"

  secret {
    name  = "eventhub-connection-string"
    value = azurerm_eventhub_namespace_authorization_rule.app.primary_connection_string
  }

  template {
    min_replicas = 1
    max_replicas = 1
    container {
      name   = "indexer"
      image  = var.indexer_image
      cpu    = 0.5
      memory = "1Gi"
      env {
        name  = "KAFKA_BROKERS"
        value = "${azurerm_eventhub_namespace.main.name}.servicebus.windows.net:9093"
      }
      env {
        name  = "KAFKA_TOPIC"
        value = azurerm_eventhub.events.name
      }
      env {
        name  = "KAFKA_GROUP"
        value = "inventory-indexer"
      }
      env {
        name  = "OPENSEARCH_URL"
        value = "http://${azurerm_container_app.opensearch.ingress[0].fqdn}"
      }
      env {
        name        = "EVENTHUB_CONNECTION_STRING"
        secret_name = "eventhub-connection-string"
      }
    }
  }
}

resource "azurerm_container_app" "web" {
  name                         = "web-${local.name}"
  container_app_environment_id = azurerm_container_app_environment.main.id
  resource_group_name          = azurerm_resource_group.main.name
  revision_mode                = "Single"

  ingress {
    external_enabled = true
    target_port      = 80
    transport        = "http"
    traffic_weight {
      percentage      = 100
      latest_revision = true
    }
  }

  template {
    min_replicas = 1
    max_replicas = 1
    container {
      name   = "web"
      image  = var.web_image
      cpu    = 0.25
      memory = "0.5Gi"
    }
  }
}

resource "azurerm_role_assignment" "api_blob_reader" {
  scope                = azurerm_storage_account.config.id
  role_definition_name = "Storage Blob Data Reader"
  principal_id         = azurerm_container_app.api.identity[0].principal_id
}

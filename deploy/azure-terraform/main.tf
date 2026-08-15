# Blacklight on Azure Container Apps — a fully worked Terraform example.
#
# Deploys the published container image to Azure Container Apps (Azure's
# serverless-container service; the Azure equivalent of "Cloud Run"), with:
#
#   * the database and evidence on an Azure Files share in a Storage Account,
#   * the session secret and encryption key generated here and kept in Key Vault.
#
# The two application secrets are handled with Terraform's ephemeral values and
# write-only arguments: they are generated during apply, sent straight to Key
# Vault, and never written to state or to the plan file. See "secrets" below.
#
# The only prerequisite is an `az login` session. See README.md for the full
# walkthrough, the first-administrator step, and the operational caveats.

terraform {
  # 1.11 is the floor for write-only arguments (`value_wo`); ephemeral resources
  # arrived in 1.10.
  required_version = ">= 1.11.0"

  required_providers {
    azurerm = {
      # 4.23.0 is the first release with `azurerm_key_vault_secret.value_wo`.
      source  = "hashicorp/azurerm"
      version = ">= 4.23.0, < 5.0.2"
    }
    random = {
      # 3.7.0 is the first release with the `random_password` ephemeral resource.
      source  = "hashicorp/random"
      version = ">= 3.7.0, < 4.0.0"
    }
  }
}

provider "azurerm" {
  # No credentials configured on purpose: with an `az login` session active, the
  # provider authenticates through the Azure CLI automatically.
  features {}
}

data "azurerm_client_config" "current" {}

# ─── names ────────────────────────────────────────────────────────────────────
#
# Storage accounts and Key Vaults need globally unique names, so they get a
# random suffix. Everything else derives from `var.name`.

resource "random_string" "suffix" {
  length  = 8
  lower   = true
  upper   = false
  numeric = true
  special = false
}

locals {
  # Whether this deployment asks the app to create its first administrator. The
  # address is the switch, here and in the application's own configuration.
  create_admin = var.admin_email != null

  resource_group_name  = "${var.name}-rg"
  log_analytics_name   = "${var.name}-logs"
  environment_name     = "${var.name}-env"
  identity_name        = "${var.name}-identity"
  container_app_name   = var.name
  storage_account_name = "st${random_string.suffix.result}"
  key_vault_name       = "kv${random_string.suffix.result}"
  admin_secret_name    = "blacklight-bootstrap-admin-password"
  file_share_name      = "blacklight-data"
  storage_mount_name   = "blacklight-data"
}

# ─── base: resource group, logs, environment ─────────────────────────────────

resource "azurerm_resource_group" "main" {
  name     = local.resource_group_name
  location = var.location
}

resource "azurerm_log_analytics_workspace" "main" {
  name                = local.log_analytics_name
  location            = azurerm_resource_group.main.location
  resource_group_name = azurerm_resource_group.main.name
  sku                 = "PerGB2018"
  retention_in_days   = 30
}

resource "azurerm_container_app_environment" "main" {
  name                       = local.environment_name
  location                   = azurerm_resource_group.main.location
  resource_group_name        = azurerm_resource_group.main.name
  log_analytics_workspace_id = azurerm_log_analytics_workspace.main.id
}

# ─── storage: the database + evidence live on an Azure Files share ───────────

resource "azurerm_storage_account" "main" {
  name                     = local.storage_account_name
  location                 = azurerm_resource_group.main.location
  resource_group_name      = azurerm_resource_group.main.name
  account_tier             = "Standard"
  account_replication_type = "LRS"
}

resource "azurerm_storage_share" "data" {
  name                 = local.file_share_name
  storage_account_name = azurerm_storage_account.main.name
  quota                = var.storage_share_quota_gb
}

# Links the environment to the share. The container app's volume references this
# link by name (`storage_mount_name`), not the share name.
resource "azurerm_container_app_environment_storage" "data" {
  name                         = local.storage_mount_name
  container_app_environment_id = azurerm_container_app_environment.main.id
  account_name                 = azurerm_storage_account.main.name
  share_name                   = azurerm_storage_share.data.name
  access_key                   = azurerm_storage_account.main.primary_access_key
  access_mode                  = "ReadWrite"
}

# ─── secrets: generated ephemerally, stored only in Key Vault ────────────────
#
# Nothing in this section reaches the state file. `ephemeral` resources exist
# for the duration of one operation and are never persisted, and `value_wo` is
# a write-only argument: the provider sends it to Azure and Terraform drops it.
# The Key Vault is the only place the two values live.

resource "azurerm_key_vault" "main" {
  name                = local.key_vault_name
  location            = azurerm_resource_group.main.location
  resource_group_name = azurerm_resource_group.main.name
  tenant_id           = data.azurerm_client_config.current.tenant_id
  sku_name            = "standard"
}

# Two independent values, so the encryption key can never equal the session
# secret (the server refuses to start if the two match).
#
# These regenerate on every plan and apply, which is harmless: the generated
# value is only written to Key Vault when the matching `..._version` variable
# changes (see `value_wo_version` below). A run that does not bump the version
# produces no diff and leaves the stored secret alone.
ephemeral "random_password" "session_secret" {
  length  = 64
  special = false
  upper   = true
  lower   = true
  numeric = true
}

ephemeral "random_password" "encryption_key" {
  length  = 64
  special = false
  upper   = true
  lower   = true
  numeric = true
}

# The initial password of the first administrator. Shorter than the two keys and
# without symbols, because unlike them it is a password a person reads out of
# Key Vault and types once — 24 characters of mixed case and digits is far past
# the application's twelve-character policy either way.
ephemeral "random_password" "admin_password" {
  length  = 24
  special = false
  upper   = true
  lower   = true
  numeric = true
}

# The principal running `terraform apply` (the az-login user) writes these, so
# it needs data-plane access to the vault before the secrets are created.
resource "azurerm_key_vault_access_policy" "terraform" {
  key_vault_id = azurerm_key_vault.main.id
  tenant_id    = data.azurerm_client_config.current.tenant_id
  object_id    = data.azurerm_client_config.current.object_id

  secret_permissions = ["Get", "List", "Set", "Delete", "Recover", "Purge"]
}

# `value_wo` is write-only — Terraform sends it to Key Vault and keeps neither
# the value nor a hash of it. Because the value is invisible to the plan,
# `value_wo_version` is what drives updates: change the version variable and the
# next apply writes a new secret version; leave it alone and the secret is left
# untouched, whatever the ephemeral generator produced this run.
resource "azurerm_key_vault_secret" "session_secret" {
  name             = "blacklight-session-secret"
  value_wo         = coalesce(var.session_secret, ephemeral.random_password.session_secret.result)
  value_wo_version = var.session_secret_version
  key_vault_id     = azurerm_key_vault.main.id

  depends_on = [azurerm_key_vault_access_policy.terraform]
}

resource "azurerm_key_vault_secret" "encryption_key" {
  name             = "blacklight-encryption-key"
  value_wo         = coalesce(var.encryption_key, ephemeral.random_password.encryption_key.result)
  value_wo_version = var.encryption_key_version
  key_vault_id     = azurerm_key_vault.main.id

  depends_on = [azurerm_key_vault_access_policy.terraform]
}

# The first administrator's initial password, handled exactly as the two secrets
# above are: generated during apply, written to Key Vault write-only, and absent
# from state and from the plan. Key Vault is where the operator reads it from —
# `terraform output admin_password_command` prints the command.
#
# Unlike the two above, this one is *expected* to stop mattering. The server
# uses it once, on a database with no accounts, and ignores it forever after;
# the account's password is whatever its owner changes it to at first sign-in.
resource "azurerm_key_vault_secret" "admin_password" {
  count = local.create_admin ? 1 : 0

  name             = local.admin_secret_name
  value_wo         = coalesce(var.admin_password, ephemeral.random_password.admin_password.result)
  value_wo_version = var.admin_password_version
  key_vault_id     = azurerm_key_vault.main.id

  depends_on = [azurerm_key_vault_access_policy.terraform]
}

# ─── identity: the container app reads the Key Vault secrets with this ───────

resource "azurerm_user_assigned_identity" "app" {
  name                = local.identity_name
  location            = azurerm_resource_group.main.location
  resource_group_name = azurerm_resource_group.main.name
}

resource "azurerm_key_vault_access_policy" "app" {
  key_vault_id = azurerm_key_vault.main.id
  tenant_id    = data.azurerm_client_config.current.tenant_id
  object_id    = azurerm_user_assigned_identity.app.principal_id

  secret_permissions = ["Get"]
}

# ─── the app ─────────────────────────────────────────────────────────────────

resource "azurerm_container_app" "main" {
  name                         = local.container_app_name
  container_app_environment_id = azurerm_container_app_environment.main.id
  resource_group_name          = azurerm_resource_group.main.name
  revision_mode                = "Single"

  # The identity is assigned to the app AND used by the `secret` blocks, so the
  # Key Vault references resolve before the first revision starts.
  identity {
    type         = "UserAssigned"
    identity_ids = [azurerm_user_assigned_identity.app.id]
  }

  secret {
    name                = "session-secret"
    identity            = azurerm_user_assigned_identity.app.id
    key_vault_secret_id = azurerm_key_vault_secret.session_secret.versionless_id
  }
  secret {
    name                = "encryption-key"
    identity            = azurerm_user_assigned_identity.app.id
    key_vault_secret_id = azurerm_key_vault_secret.encryption_key.versionless_id
  }
  # Present only when there is an administrator to create. A Container Apps
  # secret is not readable back through the API or the portal; it is resolved
  # from Key Vault by the app's identity, like the other two.
  dynamic "secret" {
    for_each = azurerm_key_vault_secret.admin_password
    content {
      name                = "bootstrap-admin-password"
      identity            = azurerm_user_assigned_identity.app.id
      key_vault_secret_id = secret.value.versionless_id
    }
  }

  ingress {
    external_enabled = true
    target_port      = 8080
    transport        = "http"

    traffic_weight {
      latest_revision = true
      percentage      = 100
    }
  }

  template {
    # DuckDB gives the database file to one process at a time, so this app must
    # never scale beyond a single replica.
    min_replicas = 1
    max_replicas = 1

    volume {
      name         = "data"
      storage_type = "AzureFile"
      storage_name = azurerm_container_app_environment_storage.data.name
      # The image runs as uid/gid 10001; force the share's files to that owner
      # so the non-root process can write them.
      mount_options = "uid=10001,gid=10001,dir_mode=0750,file_mode=0750"
    }

    # Chromium (PDF rendering) needs more than Container Apps' 64 MiB /dev/shm.
    # An EmptyDir mount replaces it with the replica's ephemeral disk.
    volume {
      name         = "shm"
      storage_type = "EmptyDir"
    }

    container {
      name   = "blacklight"
      image  = "${var.image}:${var.image_tag}"
      cpu    = var.container_cpu
      memory = var.container_memory

      env {
        name  = "BLACKLIGHT_BASE_URL"
        value = "https://${local.container_app_name}.${azurerm_container_app_environment.main.default_domain}"
      }
      env {
        name        = "BLACKLIGHT_SESSION_SECRET"
        secret_name = "session-secret"
      }
      env {
        name        = "BLACKLIGHT_ENCRYPTION_KEY"
        secret_name = "encryption-key"
      }

      # The first administrator. The server acts on these once — on a database
      # with no accounts — and ignores them on every start after that, which is
      # what makes leaving them configured harmless: they cannot reset a
      # password, re-promote a demoted account or revive a disabled one.
      #
      # The password arrives as an environment variable rather than as a file,
      # which is the weaker of the two ways the application accepts it: the
      # value ends up in the process environment of the replica. Container Apps
      # itself can mount secrets as files, but the azurerm volume block pinned
      # here takes only AzureFile and EmptyDir, so that is not reachable from
      # this configuration. Change the password at first sign-in and, if you
      # want the variable off the revision entirely, clear admin_email and apply
      # again.
      dynamic "env" {
        for_each = local.create_admin ? {
          BLACKLIGHT_BOOTSTRAP_ADMIN_EMAIL = var.admin_email
          BLACKLIGHT_BOOTSTRAP_ADMIN_NAME  = var.admin_name
        } : {}
        content {
          name  = env.key
          value = env.value
        }
      }
      dynamic "env" {
        for_each = azurerm_key_vault_secret.admin_password
        content {
          name        = "BLACKLIGHT_BOOTSTRAP_ADMIN_PASSWORD"
          secret_name = "bootstrap-admin-password"
        }
      }

      volume_mounts {
        name = "data"
        path = "/var/lib/blacklight"
      }
      volume_mounts {
        name = "shm"
        path = "/dev/shm"
      }

      readiness_probe {
        transport = "HTTP"
        port      = 8080
        path      = "/api/v1/healthz"
      }
      liveness_probe {
        transport               = "HTTP"
        port                    = 8080
        path                    = "/api/v1/healthz"
        initial_delay           = 30
        interval_seconds        = 30
        timeout                 = 5
        failure_count_threshold = 3
      }
      startup_probe {
        transport               = "HTTP"
        port                    = 8080
        path                    = "/api/v1/healthz"
        initial_delay           = 10
        interval_seconds        = 10
        timeout                 = 5
        failure_count_threshold = 30
      }
    }
  }

  # The app's managed identity must have Get on the vault before the first
  # revision starts resolving the secret references.
  depends_on = [azurerm_key_vault_access_policy.app]
}

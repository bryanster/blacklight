# Deploy to Azure (one click)

[![Deploy to Azure](https://aka.ms/deploytoazurebutton)](https://portal.azure.com/#create/Microsoft.Template/uri/https%3A%2F%2Fraw.githubusercontent.com%2Fbryanster%2Fblacklight%2Fmain%2Fdeploy%2Fazure-arm%2Fazuredeploy.json)

The button opens the Azure portal on `azuredeploy.json` in this directory — an ARM template that
provisions the same deployment as [`deploy/azure-terraform/`](../azure-terraform/README.md), without
a Terraform install or a state file to keep. Pick a subscription, a resource group and a region, and
optionally the address of the first administrator; everything else has a working default.

What it provisions:

| Resource | Purpose |
|---|---|
| Container App (single replica) | Runs the published `ghcr.io/bryanster/blacklight` image |
| Storage Account + Azure Files share | Persistent `/var/lib/blacklight` — the DuckDB database, its WAL and uploaded evidence |
| Key Vault | Holds `BLACKLIGHT_SESSION_SECRET`, `BLACKLIGHT_ENCRYPTION_KEY` and — if you ask for a first administrator — that account's initial password |
| User-assigned managed identity | Lets the app read the Key Vault secrets. It is the only principal the template grants data-plane access to |
| Log Analytics workspace + Container Apps environment | Required base for Container Apps |

The two application secrets are generated during the deployment, stored in Key Vault, and referenced
by the container app through the managed identity — the values never appear in the container app's
own secret list. Each is two fresh GUIDs with the hyphens stripped: 64 hexadecimal characters, 244
bits of randomness, taken from the GUIDs' own bytes rather than through `uniqueString`, whose 64-bit
hash would throw most of that away. They are `securestring` values, so the deployment history records
their names and types and no value; Key Vault is where they live.

The generation happens in a nested, inner-scoped deployment (`<name>-secrets`) rather than in this
template's parameter defaults, and that placement is load-bearing — see
[Troubleshooting](#troubleshooting) for what goes wrong otherwise. Leave `sessionSecret`,
`encryptionKey` and `adminPassword` empty and each is generated; fill one in and that value is used
verbatim.

## Prerequisites

- An Azure subscription where you can create resource groups, Key Vaults and Container Apps
  (Owner, or Contributor — the template grants Key Vault access with an access policy, not a role
  assignment, so no User Access Administrator rights are needed).
- The `Microsoft.App`, `Microsoft.OperationalInsights`, `Microsoft.Storage`, `Microsoft.KeyVault` and
  `Microsoft.ManagedIdentity` resource providers registered. The portal registers them for you if
  your account may; to do it up front:

  ```sh
  for ns in Microsoft.App Microsoft.OperationalInsights Microsoft.Storage Microsoft.KeyVault Microsoft.ManagedIdentity; do
    az provider register --namespace "$ns"
  done
  ```

Deployment takes about five minutes, most of it the Container Apps environment.

## Parameters

Everything is optional; these are the defaults.

| Parameter | Default | What it does |
|---|---|---|
| `name` | `blacklight` | Names the Container App, environment, workspace and identity, and the first label of the app's hostname |
| `locationOverride` | *(empty)* | Region for every resource. Empty means the resource group's own region, which is what you want when the button creates the group. Set it only to place the resources somewhere else — see [Troubleshooting](#troubleshooting) |
| `image` / `imageTag` | `ghcr.io/bryanster/blacklight` / `v1.0.4` | The image to run. Pin a release |
| `containerCpu` / `containerMemory` | `1` / `2Gi` | Container size. Must be a valid Consumption-plan pair |
| `storageShareQuotaGb` | `100` | Size of the Azure Files share |
| `adminEmail` | *(empty)* | Set it and the app creates that first administrator; leave it empty and no account is created |
| `adminName` | `Administrator` | Display name for that account |
| `createSecrets` | `true` | Whether to write the three Key Vault secrets. **Set it to `false` when redeploying** — see below |
| `sessionSecret` / `encryptionKey` / `adminPassword` | *(empty)* | **Leave these empty** — each is generated during the deployment. Fill one in only to reuse a value you already have |

The Storage Account and Key Vault take a generated name (`st…`/`kv…` plus a hash of the resource
group and `name`), because both must be globally unique.

## After it deploys

The **Outputs** tab of the deployment gives you `appUrl`, the Key Vault and storage account names,
and `adminPasswordCommand`.

A fresh deployment has no users and no sign-up (see [`docs/deploy.md`](../../docs/deploy.md)), so
something has to create the first account.

### The first administrator, from the template

Set `adminEmail` in the portal form. The template generates a password, stores it in Key Vault beside
the other two secrets, and passes the address, the name and the password to the app. The app creates
that account — as a platform administrator, active, with a local password — **the first time it
starts against a database with no accounts in it**, and ignores the setting on every start
afterwards. Read the password out of Key Vault and sign in:

```sh
az keyvault secret show --vault-name <keyVaultName from the outputs> \
  --name blacklight-bootstrap-admin-password --query value -o tsv
```

Change the password at first sign-in. Two things worth being clear about, both of which apply
equally to the Terraform version:

- **It is a bootstrap, not an account manager.** The app acts on those variables only when there are
  no accounts at all. It cannot reset a password, re-promote an account somebody demoted, or revive
  one somebody disabled — which is what makes it safe to leave `adminEmail` set for the life of the
  deployment. A genuinely forgotten password is a `blctl` job.
- **The password reaches the app as an environment variable.** The revision holds a *reference* to
  the Key Vault secret rather than the value, but the value ends up in the replica's process
  environment. The application also accepts `BLACKLIGHT_BOOTSTRAP_ADMIN_PASSWORD_FILE`, which is the
  better of the two, but Container Apps volumes take only Azure Files and EmptyDir here, so a secret
  cannot be mounted as a file. Change the password at first sign-in, and clear `adminEmail` and
  redeploy if you want the variable off the revision entirely.

### The first administrator, with `blctl`

The other way, and the only one on a deployment you would rather not hand a password to. `blctl user
create` needs exclusive access to the database — DuckDB gives the file to one process at a time — so
it means stopping the app, running `blctl` once against the same share, and starting it again. The
[Terraform README](../azure-terraform/README.md#with-blctl) has the commands; they are identical
here, with the resource group being the one you chose in the portal.

## Redeploying

**Set `createSecrets` to `false` on any deployment after the first.** The generated defaults are
regenerated every time the template runs, so a redeploy that leaves it `true` writes new values into
Key Vault: a new session secret signs every user out, and a new encryption key makes every enrolled
MFA authenticator, recovery code and service token unreadable. With `createSecrets: false` the
existing Key Vault secrets are left untouched and everything else — image tag, container size, share
quota — still updates. The container app references the secrets by their versionless Key Vault URL,
so it keeps resolving them.

Upgrading the image, from the CLI:

```sh
az deployment group create \
  --resource-group <your-rg> \
  --template-file deploy/azure-arm/azuredeploy.json \
  --parameters @deploy/azure-arm/azuredeploy.parameters.json \
  --parameters createSecrets=false imageTag=v1.0.4
```

`az containerapp update --image ghcr.io/bryanster/blacklight:v1.0.4` does the same thing for that one
field, and cannot touch the secrets at all.

To rotate a secret deliberately, write the new value to Key Vault and restart the revision — Container
Apps caches Key Vault references:

```sh
az keyvault secret set --vault-name <kv> --name blacklight-session-secret --value "$(openssl rand -hex 32)"
az containerapp revision restart --name <app> --resource-group <your-rg> --revision "$(az containerapp show --name <app> --resource-group <your-rg> --query properties.latestRevisionName -o tsv)"
```

Rotating the encryption key is destructive in the way described above. Only do it as part of a
deliberate re-enrollment.

## Backups

`blctl backup` and `blctl migrate` have the same single-process constraint as `blctl user create`:
stop the app, run the command against the same share, start it again. Back up the Key Vault secrets
too — the session secret and encryption key live there and nowhere else, and a restored database
without its encryption key is one where nobody with MFA can sign in.

## Troubleshooting

**The Key Vault secrets contain `[replace(concat(newGuid(),newGuid()),'-','')]`** — the literal
expression rather than a value — and the app never starts: its revision sits in `Activating` and the
health endpoint never answers.

Versions of this template before the nested `<name>-secrets` deployment generated the secrets in
their parameter *defaults*, and that does not survive the portal. The portal cannot evaluate ARM
expressions in the browser, so its deployment form pre-fills every field with the default's raw text
and submits it as an explicit value — and ARM evaluates a `defaultValue` only when no value is
supplied. The expression is stored verbatim as the secret. Neither `az deployment group validate`
nor `what-if` reproduces it, because neither goes through the form; deploying from the CLI does not
either, which is what makes it a trap.

The app refusing to start is the one piece of luck in it: both keys get the same literal string, and
the server rejects a configuration where the session secret equals the encryption key, so nothing is
ever served. To check a deployment you are unsure of:

```sh
az keyvault secret show --vault-name <kv> --name blacklight-session-secret --query value -o tsv
```

If the value starts with `[`, treat all three secrets as public — they are a string from a file in
this repository — and the bootstrap administrator's password as known. Nothing was served, so there
is nothing to salvage: delete the resource group and deploy again with a current template.

```sh
az group delete --name <your-rg>
```

**`RequestDisallowedByAzure` — "The selected region is currently not accepting new customers"**, on
every resource at once:

```json
{"code": "InvalidTemplateDeployment", "details": [
  {"code": "RequestDisallowedByAzure", "target": "blacklight-logs",
   "message": "Resource 'blacklight-logs' was disallowed by Azure: The selected region is currently not accepting new customers"}]}
```

Nothing was created, and nothing is wrong with the template: Azure caps capacity per region and per
subscription, and closes a region to subscriptions that have never deployed there. `westeurope` is a
common one to hit.

The resources are created in the resource group's own region, so this is the region of the group you
picked. They do not have to live there, though, which is what `locationOverride` is for: set it to a
region your subscription can use and everything is created there instead, in the same group.

```sh
az deployment group create --resource-group <your-rg> \
  --template-file deploy/azure-arm/azuredeploy.json --parameters locationOverride=northeurope
```

The override exists because of a portal behaviour worth knowing about: **when you deploy into a group
that already exists, the blade's Region control is locked to that group's region.** There is nothing
else on the form that moves the resources, so on an existing group in a closed region the override is
the only way through — and choosing a region elsewhere on the form will look like it worked and fail
identically. When you let the button create the group, the Region you pick there is the group's
region, the override stays empty, and everything follows it.

Try a neighbouring region — `northeurope` for `westeurope`, `swedencentral` or `uksouth` after that.
There is no API that reports which regions are open to you, so trying is how you find out; a closed
region fails in seconds, at preflight, before anything is provisioned (the attempt is not even
recorded in the group's deployment history).

To check a region supports every resource type in the first place:

```sh
az provider show --namespace Microsoft.App \
  --query "resourceTypes[?resourceType=='managedEnvironments'].locations" -o tsv
```

## Caveats

Everything in the Terraform README's [caveats](../azure-terraform/README.md#caveats) applies —
single replica is load-bearing, DuckDB on SMB, `/dev/shm` as an EmptyDir mount, no scale-to-zero —
plus two that are specific to this template:

- **Secrets are template inputs, not write-only values.** Terraform generates them with ephemeral
  resources and writes them with `value_wo`, so they never reach state. ARM has no equivalent: the
  values are `securestring` parameters, which keeps them out of the deployment history, but they do
  pass through the deployment. `createSecrets` is what stops a redeploy from overwriting them.
- **Deleting and redeploying into the same resource group collides with the soft-deleted Key
  Vault.** The vault name is derived from the resource group and `name` rather than being random, so
  it is the same name on the second deployment while the first is still in Key Vault's 7-day soft-delete
  window. Either purge it (`az keyvault purge --name <kv> --location <region>`) or deploy with a
  different `name`.

## Teardown

```sh
az group delete --name <your-rg>
```

The Key Vault is soft-deleted rather than purged; see the caveat above before redeploying with the
same resource group and `name`.

## Which one should I use?

This template for a first deployment, a demo, or an environment nobody is going to change often —
one click, nothing to install, no state to store. [`deploy/azure-terraform/`](../azure-terraform/)
for anything with a lifecycle: it keeps the secrets out of the deployment entirely (ephemeral values
and write-only arguments), makes rotation an explicit counter, and can plan a change before making
it. Both produce the same resources, so a deployment made with the button can be adopted by
Terraform later with `terraform import` — pass the existing secrets back in with
`TF_VAR_session_secret` and `TF_VAR_encryption_key` so the adoption does not rewrite them.

# terraform-provider-gopass

OpenTofu/Terraform provider for reading secrets from [gopass](https://github.com/gopasspw/gopass)
as **ephemeral values** - credentials that are never stored in state or plan files.

## Features

- 🔐 **Ephemeral-only reading**: Secrets exist only during plan/apply, never persisted
- ✍️ **Write-only storage**: Store generated credentials to gopass without state leakage
- 🔗 **Native gopass integration**: Links directly against gopass Go library - no subprocess spawning
- 🔑 **Hardware token support**: Works with YubiKey, Nitrokey, etc. via GPG
- 📁 **Multiple access patterns**:
  - `ephemeral gopass_secret`: Read single secret by path
  - `ephemeral gopass_env`: Read credential set as key-value map (like `gopassenv`)
  - `resource gopass_secret`: Write secrets with write-only attributes
- 🔄 **No state leakage**: Provider credentials don't end up in terraform.tfstate

## Requirements

- **OpenTofu 1.11+** (ephemeral resources support) or Terraform 1.10+
- **gopass** installed and configured
- GPG key available (hardware token or software key)

## Installation

### From Source

```bash
git clone https://git.ingo-struck.com/opentofu/terraform-provider-gopass.git
cd terraform-provider-gopass
make install
```

### Manual Installation

```bash
# Build
go build -o terraform-provider-gopass

# Install for OpenTofu
mkdir -p ~/.local/share/opentofu/plugins/registry.opentofu.org/istr/gopass/0.1.0/linux_amd64
cp terraform-provider-gopass ~/.local/share/opentofu/plugins/registry.opentofu.org/istr/gopass/0.1.0/linux_amd64/
```

## Usage

### Provider Configuration

```hcl
terraform {
  required_version = ">= 1.11.0"

  required_providers {
    gopass = {
      source  = "registry.opentofu.org/istr/gopass"
      version = "~> 0.1"
    }
  }
}

# Default: uses gopass's native configuration
provider "gopass" {}

# Or specify a custom store path
provider "gopass" {
  store_path = "/home/user/.password-store"
}
```

#### Provider Arguments

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `store_path` | string | no | Path to the gopass password store. If not set, uses gopass default configuration from `~/.config/gopass/config` or the `PASSWORD_STORE_DIR` environment variable. |

### Reading a Credential Set (gopassenv style)

The `gopass_env` ephemeral resource reads all secrets under a path and makes them accessible via dot-notation. It supports both flat and nested/hierarchical path structures.

#### Flat Structure (Immediate Children)

Given this gopass structure:

```
env/terraform/scaleway/istr/
├── SCW_ACCESS_KEY
├── SCW_SECRET_KEY
└── SCW_DEFAULT_PROJECT_ID
```

Read all as a flat object:

```hcl
ephemeral "gopass_env" "scaleway" {
  path = "env/terraform/scaleway/istr"
}

provider "scaleway" {
  access_key = ephemeral.gopass_env.scaleway.credentials.SCW_ACCESS_KEY
  secret_key = ephemeral.gopass_env.scaleway.credentials.SCW_SECRET_KEY
  project_id = ephemeral.gopass_env.scaleway.credentials.SCW_DEFAULT_PROJECT_ID
}
```

#### Nested Structure (Deep Hierarchies)

The provider automatically converts slash-separated paths into nested object structures:

Given this gopass structure:

```
env/terraform/cloud/aws/
├── REGION
├── API/
│   ├── v2/
│   │   ├── ACCESS_KEY
│   │   └── SECRET_KEY
│   └── v1/
│       └── LEGACY_TOKEN
└── database/
    └── prod/
        ├── HOST
        └── PASSWORD
```

Access nested paths using dot-notation:

```hcl
ephemeral "gopass_env" "aws" {
  path = "env/terraform/cloud/aws"
}

provider "aws" {
  region     = ephemeral.gopass_env.aws.credentials.REGION
  # Nested paths: API/v2/ACCESS_KEY becomes credentials.API.v2.ACCESS_KEY
  access_key = ephemeral.gopass_env.aws.credentials.API.v2.ACCESS_KEY
  secret_key = ephemeral.gopass_env.aws.credentials.API.v2.SECRET_KEY
}

# Access database credentials (2 levels deep)
resource "postgresql_database" "main" {
  name = "mydb"
  # database/prod/HOST becomes credentials.database.prod.HOST
  connection_string = "postgres://${ephemeral.gopass_env.aws.credentials.database.prod.HOST}:5432"
}
```

**Key features:**
- 🌳 **Recursive**: Reads all secrets at any depth under the specified path
- 📂 **Automatic nesting**: Slash-separated paths become nested objects (`API/v2/KEY` → `credentials.API.v2.KEY`)
- 🔀 **Mixed depths**: Supports both flat and nested secrets in the same tree
- ⚡ **Efficient**: Single gopass query for entire tree

### Reading Individual Secrets

```hcl
ephemeral "gopass_secret" "db_password" {
  path = "infrastructure/database/admin_password"
}

resource "postgresql_role" "admin" {
  name     = "admin"
  password = ephemeral.gopass_secret.db_password.value  # write-only attribute
}
```

## Ephemeral Resources

### gopass_secret

Reads a single secret from the gopass store.

#### Arguments

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `path` | string | yes | Path to the secret in gopass |

#### Attributes

| Name | Type | Description |
|------|------|-------------|
| `value` | string | The secret value (first line only) |

### gopass_env

Reads all secrets under a path recursively as a nested object structure.

#### Arguments

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `path` | string | yes | Path prefix in gopass store |

#### Attributes

| Name | Type | Description |
|------|------|-------------|
| `credentials` | dynamic object | Nested object with secrets accessible via dot-notation. Slash-separated paths become nested: `API/v2/KEY` → `credentials.API.v2.KEY` |

#### Behavior

- **Recursive**: Includes all secrets at any depth under the path
- **Automatic nesting**: Converts slash-separated paths to nested objects
- **Mixed structures**: Supports both flat and nested secrets in the same tree
- **Dot-notation access**: All secrets accessible via standard Terraform dot-notation

## Managed Resources

### gopass_secret (resource)

Writes a secret to the gopass store using **write-only attributes**. The secret value is never stored in Terraform state.

This is ideal for storing generated credentials like API keys or database passwords.

#### Example: Store API Key

```hcl
# When Scaleway creates an API key, store it in gopass
resource "scaleway_iam_api_key" "infra" {
  application_id = scaleway_iam_application.infra.id
  description    = "Terraform infrastructure manager"
}

resource "gopass_secret" "scw_secret_key" {
  path             = "env/terraform/scaleway/infra-manager/SCW_SECRET_KEY"
  value_wo         = scaleway_iam_api_key.infra.secret_key
  value_wo_version = 1
}
```

#### Example: Store Generated Password

```hcl
# Generate a random password and store it in gopass
ephemeral "random_password" "db_admin" {
  length  = 32
  special = true
}

resource "gopass_secret" "db_password" {
  path             = "infrastructure/database/admin_password"
  value_wo         = ephemeral.random_password.db_admin.result
  value_wo_version = 1
}

# Use the password in the database
resource "postgresql_role" "admin" {
  name     = "admin"
  password = ephemeral.random_password.db_admin.result
}
```

#### Arguments

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `path` | string | yes | Path in the gopass store where the secret will be written |
| `value_wo` | string | no | The secret value to write. **Write-only** - never stored in state. Accepts ephemeral values. |
| `value_wo_version` | int | no | Version number. Increment to trigger a secret update when `value_wo` changes. |
| `delete_on_remove` | bool | no | Whether to delete the secret from gopass on destroy. Default: `true` |

#### Attributes

| Name | Type | Description |
|------|------|-------------|
| `id` | string | The path of the secret |
| `revision_count` | int | Number of gopass revisions (for drift detection) |

#### Drift Detection

The provider tracks the number of revisions in gopass to detect external changes:

- If someone modifies the secret outside of Terraform, the revision count increases
- On the next `tofu plan`, you'll see a warning about the drift
- To reconcile, increment `value_wo_version` to overwrite with your intended value

**Note:** Not all gopass backends support versioning. For backends without version history
(e.g., some mount types), `revision_count` will always be `1` if the secret exists.

#### Write-Only Behavior

The `value_wo` attribute follows the [Terraform write-only attributes pattern](https://developer.hashicorp.com/terraform/language/resources/ephemeral#best-practices-for-working-with-ephemeral-resources):

- The value is sent to gopass but **never stored** in state or plan files
- Terraform cannot detect drift in the actual secret value
- To update the secret, increment `value_wo_version`
- This pattern matches AWS, Azure, and Google providers for sensitive values

#### Import

Existing secrets can be imported:

```bash
tofu import gopass_secret.api_key "env/terraform/scaleway/api_key"
```

After import, set `value_wo` and `value_wo_version` in your configuration.

## How It Works

```
┌─────────────────────────────────────────────────────────────┐
│  tofu plan / tofu apply                                     │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│  1. OpenTofu calls provider's Open() method                 │
│                                                             │
│  2. Provider uses gopass Go library directly:               │
│     ├── api.New(ctx) → initializes store once              │
│     ├── store.List(ctx) → lists secrets                    │
│     └── store.Get(ctx, path, "latest") → retrieves secret  │
│         └── GPG decryption (may require PIN/touch)         │
│                                                             │
│  3. Secret returned to OpenTofu in memory only              │
│     └── Used for provider config, write-only attributes     │
│                                                             │
│  4. Operation completes, secret discarded                   │
│     └── Nothing written to state or plan                    │
│                                                             │
└─────────────────────────────────────────────────────────────┘
```

### Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                     OpenTofu/Terraform                      │
├─────────────────────────────────────────────────────────────┤
│                    terraform-provider-gopass                │
│  ┌─────────────────┐  ┌──────────────────────────────────┐  │
│  │ GopassClient    │  │ Ephemeral Resources              │  │
│  │                 │  │  - gopass_secret (single value)  │  │
│  │ • ensureStore() │──│  - gopass_env (key-value map)    │  │
│  │ • GetSecret()   │  │                                  │  │
│  │ • GetEnvSecrets │  │ Values exist only in memory      │  │
│  └────────┬────────┘  └──────────────────────────────────┘  │
├───────────┼─────────────────────────────────────────────────┤
│           │           gopass Library (linked)               │
│           v                                                 │
│  ┌─────────────────────────────────────────────────────────┐│
│  │ github.com/gopasspw/gopass/pkg/gopass/api               ││
│  │  • api.New(ctx) → Store                                 ││
│  │  • store.List(ctx) → []string                           ││
│  │  • store.Get(ctx, name, revision) → Secret              ││
│  │  • secret.Password() → string                           ││
│  └────────┬────────────────────────────────────────────────┘│
├───────────┼─────────────────────────────────────────────────┤
│           v                                                 │
│  ┌─────────────────────────────────────────────────────────┐│
│  │                    GPG Agent                            ││
│  │         (handles hardware token interaction)            ││
│  └─────────────────────────────────────────────────────────┘│
└─────────────────────────────────────────────────────────────┘
```

## Security Considerations

### What's Protected

- ✅ Secrets never written to `terraform.tfstate`
- ✅ Secrets never written to plan files
- ✅ No subprocess spawning (no secrets in process arguments)
- ✅ Hardware token provides physical authentication factor
- ✅ Each operation requires fresh authentication

### What's NOT Protected

- ⚠️ Secrets exist in memory during execution
- ⚠️ Debug logs might expose paths (not values)
- ⚠️ Process memory could theoretically be dumped
- ⚠️ Resources created with secrets may store them externally

### Recommendations

1. **Use hardware tokens**: YubiKey/Nitrokey with touch requirement
2. **Enable state encryption**: Use OpenTofu's state encryption as defense-in-depth
3. **Audit gopass access**: Monitor GPG agent activity
4. **Prefer write-only attributes**: When passing secrets to resources

## Development

```bash
# Setup
git clone https://git.ingo-struck.com/opentofu/terraform-provider-gopass.git
cd terraform-provider-gopass
go mod download

# Build & Install
make build
make install

# Test
make test

# Format & Lint
make fmt
make lint
```

## Comparison with Alternatives

| Approach | Secrets in State | Subprocess | Hardware Token |
|----------|-----------------|------------|----------------|
| **gopass ephemeral (native)** | ❌ No | ❌ No | ✅ Yes |
| gopass ephemeral (exec) | ❌ No | ✅ Yes | ✅ Yes |
| Environment variables | ❌ No | N/A | ✅ Yes |
| Vault data source | ✅ Yes | ❌ No | Via Vault |
| External data source | ✅ Yes | ✅ Yes | ✅ Yes |
| SOPS provider | ✅ Yes | ✅ Yes | Via GPG |
| State encryption | ✅ Yes (encrypted) | N/A | Via KMS |

## Troubleshooting

### "gopass store not found"

If you see an error like:
```
gopass store not found: ...
```

Possible solutions:

1. **Initialize gopass** (if not done yet):
   ```bash
   gopass init
   ```

2. **Specify store path explicitly** in provider configuration:
   ```hcl
   provider "gopass" {
     store_path = "/home/user/.password-store"
   }
   ```

3. **Set environment variable**:
   ```bash
   export PASSWORD_STORE_DIR=/path/to/store
   ```

4. **Check your gopass configuration**:
   ```bash
   cat ~/.config/gopass/config
   ```

### GPG/Hardware Token Issues

If GPG fails during secret access:
- Ensure `gpg-agent` is running
- If using a hardware token, verify it's connected
- Check that your GPG key is available: `gpg --list-secret-keys`

## API Stability Note

The gopass library includes this warning:

> Feel free to report feedback on API design and missing features but please note that
> bug reports will be silently ignored and the API WILL CHANGE WITHOUT NOTICE until this note is gone.

This provider may need updates when gopass releases new versions. Pin your gopass
dependency version in `go.mod` for stability.

## License

MPL-2.0

## Contributing

Contributions welcome! Please open an issue first to discuss proposed changes.

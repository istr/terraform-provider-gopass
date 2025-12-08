// Copyright (c) Ingo Struck
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/ephemeral"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure GopassProvider satisfies various provider interfaces.
var (
	_ provider.Provider                       = &GopassProvider{}
	_ provider.ProviderWithEphemeralResources = &GopassProvider{}
)

// GopassProvider defines the provider implementation.
type GopassProvider struct {
	version string
}

// GopassProviderModel describes the provider data model.
type GopassProviderModel struct {
	StorePath types.String `tfsdk:"store_path"`
}

// New creates a new provider instance.
func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &GopassProvider{
			version: version,
		}
	}
}

func (p *GopassProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "gopass"
	resp.Version = p.version
}

func (p *GopassProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "The gopass provider enables reading secrets from a gopass password store as ephemeral values. " +
			"Secrets are never stored in Terraform state or plan files.",
		MarkdownDescription: `
The gopass provider enables reading secrets from a [gopass](https://github.com/gopasspw/gopass)
password store as **ephemeral values**.

This provider links directly against the gopass library - no subprocess spawning required.

Ephemeral values are:
- Only available during plan/apply execution
- Never written to state or plan files
- Ideal for credentials that should not persist

## Authentication

The provider uses gopass's native configuration and GPG integration. If you use a hardware
token (YubiKey, Nitrokey), you will be prompted for PIN/touch during each Terraform operation.

## Example Usage

` + "```hcl" + `
# Use default gopass configuration
provider "gopass" {}

# Or specify a custom store path
provider "gopass" {
  store_path = "/path/to/your/password-store"
}

# Read credentials as ephemeral values
ephemeral "gopass_env" "db" {
  path = "infrastructure/database"
}

# Use in provider configuration (ephemeral-aware)
provider "postgresql" {
  password = ephemeral.gopass_env.db.values["password"]
}
` + "```" + `
`,
		Attributes: map[string]schema.Attribute{
			"store_path": schema.StringAttribute{
				Description: "Path to the gopass password store. If not set, gopass uses its default " +
					"configuration from ~/.config/gopass/config or the PASSWORD_STORE_DIR environment variable.",
				MarkdownDescription: "Path to the gopass password store. If not set, gopass uses its default " +
					"configuration from `~/.config/gopass/config` or the `PASSWORD_STORE_DIR` environment variable.",
				Optional: true,
			},
		},
	}
}

func (p *GopassProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var config GopassProviderModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Extract store path if configured
	var storePath string
	if !config.StorePath.IsNull() && !config.StorePath.IsUnknown() {
		storePath = config.StorePath.ValueString()
	}

	// Create gopass client - uses native gopass library
	client := NewGopassClient(storePath)

	// Make client available to data sources, resources, and ephemeral resources
	resp.DataSourceData = client
	resp.ResourceData = client
	resp.EphemeralResourceData = client
}

// Resources returns an empty slice - this provider only provides ephemeral resources.
func (p *GopassProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{}
}

// DataSources returns an empty slice - this provider only provides ephemeral resources.
func (p *GopassProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{}
}

// EphemeralResources returns the ephemeral resources this provider offers.
func (p *GopassProvider) EphemeralResources(ctx context.Context) []func() ephemeral.EphemeralResource {
	return []func() ephemeral.EphemeralResource{
		NewSecretEphemeralResource,
		NewEnvEphemeralResource,
	}
}

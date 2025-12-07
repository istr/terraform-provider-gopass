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
	// GopassBinary allows overriding the gopass binary path
	GopassBinary types.String `tfsdk:"gopass_binary"`
	// Store allows specifying a non-default gopass store
	Store types.String `tfsdk:"store"`
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

Ephemeral values are:
- Only available during plan/apply execution
- Never written to state or plan files
- Ideal for credentials that should not persist

## Authentication

The provider relies on your existing gopass and GPG configuration. If you use a hardware 
token (YubiKey, Nitrokey), you will be prompted for PIN/touch during each Terraform operation.

## Example Usage

` + "```hcl" + `
# Read credentials as ephemeral values
ephemeral "gopass_secret" "db_password" {
  path = "infrastructure/database/password"
}

# Use in provider configuration (ephemeral-aware)
provider "postgresql" {
  password = ephemeral.gopass_secret.db_password.value
}
` + "```" + `
`,
		Attributes: map[string]schema.Attribute{
			"gopass_binary": schema.StringAttribute{
				Description: "Path to gopass binary. Defaults to 'gopass' (found via PATH).",
				Optional:    true,
			},
			"store": schema.StringAttribute{
				Description: "Name of the gopass store to use. Defaults to the default store.",
				Optional:    true,
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

	// Build provider config to pass to resources
	providerConfig := &ProviderConfig{
		GopassBinary: "gopass",
		Store:        "",
	}

	if !config.GopassBinary.IsNull() {
		providerConfig.GopassBinary = config.GopassBinary.ValueString()
	}

	if !config.Store.IsNull() {
		providerConfig.Store = config.Store.ValueString()
	}

	// Make config available to resources and ephemeral resources
	resp.DataSourceData = providerConfig
	resp.ResourceData = providerConfig
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

// ProviderConfig holds the parsed provider configuration.
type ProviderConfig struct {
	GopassBinary string
	Store        string
}

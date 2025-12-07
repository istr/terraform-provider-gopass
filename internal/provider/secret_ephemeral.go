// Copyright (c) Ingo Struck
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"
	"os/exec"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/ephemeral"
	"github.com/hashicorp/terraform-plugin-framework/ephemeral/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure implementation satisfies interface.
var _ ephemeral.EphemeralResource = &SecretEphemeralResource{}

// SecretEphemeralResource reads a single secret from gopass.
type SecretEphemeralResource struct {
	config *ProviderConfig
}

// SecretModel describes the data model.
type SecretModel struct {
	Path  types.String `tfsdk:"path"`
	Value types.String `tfsdk:"value"`
}

// NewSecretEphemeralResource creates a new instance.
func NewSecretEphemeralResource() ephemeral.EphemeralResource {
	return &SecretEphemeralResource{}
}

func (r *SecretEphemeralResource) Metadata(ctx context.Context, req ephemeral.MetadataRequest, resp *ephemeral.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_secret"
}

func (r *SecretEphemeralResource) Schema(ctx context.Context, req ephemeral.SchemaRequest, resp *ephemeral.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Reads a single secret value from the gopass store.",
		MarkdownDescription: `
Reads a single secret value from the gopass store.

The secret is retrieved during each Terraform operation and is **never stored** 
in state or plan files.

## Example Usage

` + "```hcl" + `
ephemeral "gopass_secret" "api_key" {
  path = "services/api/token"
}

# Use the secret value
provider "example" {
  api_key = ephemeral.gopass_secret.api_key.value
}
` + "```" + `

## GPG/Hardware Token

If your gopass store is encrypted with a hardware token (YubiKey, Nitrokey, etc.),
you will be prompted for PIN entry and/or touch confirmation during each 
Terraform operation that accesses the secret.
`,
		Attributes: map[string]schema.Attribute{
			"path": schema.StringAttribute{
				Description:         "Path to the secret in the gopass store (e.g., 'infrastructure/db/password').",
				MarkdownDescription: "Path to the secret in the gopass store (e.g., `infrastructure/db/password`).",
				Required:            true,
			},
			"value": schema.StringAttribute{
				Description:         "The secret value. Only the first line is returned (password convention).",
				MarkdownDescription: "The secret value. Only the first line is returned (password convention).",
				Computed:            true,
				Sensitive:           true,
			},
		},
	}
}

func (r *SecretEphemeralResource) Configure(ctx context.Context, req ephemeral.ConfigureRequest, resp *ephemeral.ConfigureResponse) {
	// Get provider config if available
	if req.ProviderData == nil {
		return
	}

	config, ok := req.ProviderData.(*ProviderConfig)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Provider Data",
			fmt.Sprintf("Expected *ProviderConfig, got: %T", req.ProviderData),
		)
		return
	}

	r.config = config
}

func (r *SecretEphemeralResource) Open(ctx context.Context, req ephemeral.OpenRequest, resp *ephemeral.OpenResponse) {
	var data SecretModel

	// Read configuration
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	path := data.Path.ValueString()
	tflog.Debug(ctx, "Reading secret from gopass", map[string]interface{}{
		"path": path,
	})

	// Determine gopass binary
	gopassBin := "gopass"
	if r.config != nil && r.config.GopassBinary != "" {
		gopassBin = r.config.GopassBinary
	}

	// Build command arguments
	args := []string{"show", "-o"}

	// Add store flag if configured
	if r.config != nil && r.config.Store != "" {
		args = append(args, "--store", r.config.Store)
	}

	args = append(args, path)

	// Execute gopass
	// Note: This may trigger GPG agent / hardware token interaction
	cmd := exec.CommandContext(ctx, gopassBin, args...)
	output, err := cmd.Output()
	if err != nil {
		var stderr string
		if exitErr, ok := err.(*exec.ExitError); ok {
			stderr = string(exitErr.Stderr)
		}
		resp.Diagnostics.AddError(
			"Failed to read secret",
			fmt.Sprintf("gopass show failed for path %q: %s\nStderr: %s", path, err.Error(), stderr),
		)
		return
	}

	// gopass returns the full secret, we take only the first line (password convention)
	value := strings.TrimSpace(string(output))
	if idx := strings.Index(value, "\n"); idx != -1 {
		value = value[:idx]
	}

	data.Value = types.StringValue(value)

	// Set result - this is NEVER written to state
	resp.Diagnostics.Append(resp.Result.Set(ctx, &data)...)

	tflog.Debug(ctx, "Successfully read secret from gopass", map[string]interface{}{
		"path": path,
	})
}

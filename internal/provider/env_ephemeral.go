// Copyright (c) Ingo Struck
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"bufio"
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
var _ ephemeral.EphemeralResource = &EnvEphemeralResource{}

// EnvEphemeralResource reads a subtree from gopass as environment variables.
type EnvEphemeralResource struct {
	config *ProviderConfig
}

// EnvModel describes the data model.
type EnvModel struct {
	Path   types.String `tfsdk:"path"`
	Values types.Map    `tfsdk:"values"`
}

// NewEnvEphemeralResource creates a new instance.
func NewEnvEphemeralResource() ephemeral.EphemeralResource {
	return &EnvEphemeralResource{}
}

func (r *EnvEphemeralResource) Metadata(ctx context.Context, req ephemeral.MetadataRequest, resp *ephemeral.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_env"
}

func (r *EnvEphemeralResource) Schema(ctx context.Context, req ephemeral.SchemaRequest, resp *ephemeral.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Reads all secrets under a path as a key-value map (environment variable style).",
		MarkdownDescription: `
Reads all secrets under a path as a key-value map, similar to how ` + "`gopassenv`" + ` works.

Each secret name under the path becomes a key, and the secret's first line becomes the value.
This is ideal for reading credential sets like:

` + "```" + `
env/terraform/scaleway/istr/
├── SCW_ACCESS_KEY
├── SCW_SECRET_KEY
└── SCW_DEFAULT_PROJECT_ID
` + "```" + `

## Example Usage

` + "```hcl" + `
ephemeral "gopass_env" "scaleway" {
  path = "env/terraform/scaleway/istr"
}

provider "scaleway" {
  access_key = ephemeral.gopass_env.scaleway.values["SCW_ACCESS_KEY"]
  secret_key = ephemeral.gopass_env.scaleway.values["SCW_SECRET_KEY"]
  project_id = ephemeral.gopass_env.scaleway.values["SCW_DEFAULT_PROJECT_ID"]
}
` + "```" + `

## Notes

- Only immediate children of the path are included (not recursive)
- Each secret's first line is used as the value (gopass password convention)
- Secret names become map keys as-is (typically UPPER_SNAKE_CASE for env vars)
`,
		Attributes: map[string]schema.Attribute{
			"path": schema.StringAttribute{
				Description:         "Path prefix in the gopass store (e.g., 'env/terraform/scaleway/istr').",
				MarkdownDescription: "Path prefix in the gopass store (e.g., `env/terraform/scaleway/istr`).",
				Required:            true,
			},
			"values": schema.MapAttribute{
				Description:         "Map of secret names to their values.",
				MarkdownDescription: "Map of secret names to their values.",
				Computed:            true,
				Sensitive:           true,
				ElementType:         types.StringType,
			},
		},
	}
}

func (r *EnvEphemeralResource) Configure(ctx context.Context, req ephemeral.ConfigureRequest, resp *ephemeral.ConfigureResponse) {
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

func (r *EnvEphemeralResource) Open(ctx context.Context, req ephemeral.OpenRequest, resp *ephemeral.OpenResponse) {
	var data EnvModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	basePath := data.Path.ValueString()
	// Ensure path ends without trailing slash for consistent handling
	basePath = strings.TrimSuffix(basePath, "/")

	tflog.Debug(ctx, "Reading env secrets from gopass", map[string]interface{}{
		"path": basePath,
	})

	gopassBin := "gopass"
	if r.config != nil && r.config.GopassBinary != "" {
		gopassBin = r.config.GopassBinary
	}

	// First, list all secrets under the path
	listArgs := []string{"list", "--flat"}
	if r.config != nil && r.config.Store != "" {
		listArgs = append(listArgs, "--store", r.config.Store)
	}
	listArgs = append(listArgs, basePath)

	listCmd := exec.CommandContext(ctx, gopassBin, listArgs...)
	listOutput, err := listCmd.Output()
	if err != nil {
		var stderr string
		if exitErr, ok := err.(*exec.ExitError); ok {
			stderr = string(exitErr.Stderr)
		}
		resp.Diagnostics.AddError(
			"Failed to list secrets",
			fmt.Sprintf("gopass list failed for path %q: %s\nStderr: %s", basePath, err.Error(), stderr),
		)
		return
	}

	// Parse list output and filter to immediate children only
	values := make(map[string]string)
	scanner := bufio.NewScanner(strings.NewReader(string(listOutput)))

	for scanner.Scan() {
		secretPath := strings.TrimSpace(scanner.Text())
		if secretPath == "" {
			continue
		}

		// Check if this is an immediate child of basePath
		if !strings.HasPrefix(secretPath, basePath+"/") {
			continue
		}

		// Get the relative name (key)
		relativePath := strings.TrimPrefix(secretPath, basePath+"/")

		// Skip if it's a nested path (contains /)
		if strings.Contains(relativePath, "/") {
			continue
		}

		// Read the secret value
		showArgs := []string{"show", "-o"}
		if r.config != nil && r.config.Store != "" {
			showArgs = append(showArgs, "--store", r.config.Store)
		}
		showArgs = append(showArgs, secretPath)

		showCmd := exec.CommandContext(ctx, gopassBin, showArgs...)
		showOutput, err := showCmd.Output()
		if err != nil {
			tflog.Warn(ctx, "Failed to read secret, skipping", map[string]interface{}{
				"path":  secretPath,
				"error": err.Error(),
			})
			continue
		}

		// Take first line only
		value := strings.TrimSpace(string(showOutput))
		if idx := strings.Index(value, "\n"); idx != -1 {
			value = value[:idx]
		}

		values[relativePath] = value
	}

	if len(values) == 0 {
		resp.Diagnostics.AddWarning(
			"No secrets found",
			fmt.Sprintf("No immediate child secrets found under path %q", basePath),
		)
	}

	// Convert to types.Map
	mapValue, diags := types.MapValueFrom(ctx, types.StringType, values)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	data.Values = mapValue

	// Set result - NEVER written to state
	resp.Diagnostics.Append(resp.Result.Set(ctx, &data)...)

	tflog.Debug(ctx, "Successfully read env secrets from gopass", map[string]interface{}{
		"path":  basePath,
		"count": len(values),
	})
}

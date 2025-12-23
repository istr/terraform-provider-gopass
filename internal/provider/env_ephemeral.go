// Copyright (c) Ingo Struck
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/ephemeral"
	"github.com/hashicorp/terraform-plugin-framework/ephemeral/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure implementation satisfies interface.
var _ ephemeral.EphemeralResource = &EnvEphemeralResource{}

// EnvEphemeralResource reads a subtree from gopass as environment variables.
type EnvEphemeralResource struct {
	client *GopassClient
}

// EnvModel describes the data model.
type EnvModel struct {
	Path        types.String  `tfsdk:"path"`
	Credentials types.Dynamic `tfsdk:"credentials"`
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
		Description: "Reads all secrets under a path as a nested object structure (environment variable style).",
		MarkdownDescription: `
Reads all secrets under a path as a nested object structure, using the native gopass library.

Each secret under the path becomes accessible via dot-notation. The secret's first line becomes the value.
Supports both flat and nested/deep path structures.

This is ideal for reading credential sets with hierarchical organization:

` + "```" + `
env/terraform/scaleway/istr/
├── SCW_ACCESS_KEY
├── SCW_SECRET_KEY
├── SCW_DEFAULT_PROJECT_ID
└── API/
    └── v2/
        ├── ACCESS_KEY
        └── SECRET_KEY
` + "```" + `

## Example Usage

**Flat paths (immediate children):**

` + "```hcl" + `
ephemeral "gopass_env" "scaleway" {
  path = "env/terraform/scaleway/istr"
}

provider "scaleway" {
  access_key = ephemeral.gopass_env.scaleway.credentials.SCW_ACCESS_KEY
  secret_key = ephemeral.gopass_env.scaleway.credentials.SCW_SECRET_KEY
  project_id = ephemeral.gopass_env.scaleway.credentials.SCW_DEFAULT_PROJECT_ID
}
` + "```" + `

**Nested paths (deep hierarchies):**

` + "```hcl" + `
ephemeral "gopass_env" "aws" {
  path = "env/terraform/aws"
}

provider "aws" {
  region     = ephemeral.gopass_env.aws.credentials.REGION
  # Access nested paths: API/v2/ACCESS_KEY becomes credentials.API.v2.ACCESS_KEY
  access_key = ephemeral.gopass_env.aws.credentials.API.v2.ACCESS_KEY
  secret_key = ephemeral.gopass_env.aws.credentials.API.v2.SECRET_KEY
}
` + "```" + `

## Notes

- **Recursive**: All secrets under the path are included, regardless of depth
- Each secret's first line is used as the value (gopass password convention)
- Nested paths use dot-notation: ` + "`API/v2/KEY`" + ` becomes ` + "`credentials.API.v2.KEY`" + `
- Supports mixed flat and nested structures in the same tree
- No subprocess spawning - direct library access for better performance
`,

		Attributes: map[string]schema.Attribute{
			"path": schema.StringAttribute{
				Description:         "Path prefix in the gopass store (e.g., 'env/terraform/scaleway/istr').",
				MarkdownDescription: "Path prefix in the gopass store (e.g., `env/terraform/scaleway/istr`).",
				Required:            true,
			},
			"credentials": schema.DynamicAttribute{
				Description:         "Object with secret names as attributes (accessible via dot-notation).",
				MarkdownDescription: "Object with secret names as attributes (accessible via dot-notation).",
				Computed:            true,
				Sensitive:           true,
			},
		},
	}
}

func (r *EnvEphemeralResource) Configure(ctx context.Context, req ephemeral.ConfigureRequest, resp *ephemeral.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*GopassClient)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Provider Data",
			fmt.Sprintf("Expected *GopassClient, got: %T", req.ProviderData),
		)
		return
	}

	r.client = client
}

func (r *EnvEphemeralResource) Open(ctx context.Context, req ephemeral.OpenRequest, resp *ephemeral.OpenResponse) {
	var data EnvModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	basePath := data.Path.ValueString()

	tflog.Debug(ctx, "Reading env secrets from gopass", map[string]interface{}{
		"path": basePath,
	})

	// Use native gopass library (now returns recursive/nested paths)
	values, err := r.client.GetEnvSecrets(ctx, basePath)
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to read secrets",
			fmt.Sprintf("Could not read secrets under path %q: %s", basePath, err.Error()),
		)
		return
	}

	if len(values) == 0 {
		resp.Diagnostics.AddWarning(
			"No secrets found",
			fmt.Sprintf("No secrets found under path %q", basePath),
		)
	}

	// Build nested object structure from slash-separated paths
	// This allows accessing "API/v2/ACCESS_KEY" as credentials.API.v2.ACCESS_KEY
	objValue := buildNestedObject(values)

	// Convert to dynamic
	dynamicValue := types.DynamicValue(objValue)
	data.Credentials = dynamicValue

	// Set result - NEVER written to state
	resp.Diagnostics.Append(resp.Result.Set(ctx, &data)...)

	tflog.Debug(ctx, "Successfully read env secrets from gopass", map[string]interface{}{
		"path":  basePath,
		"count": len(values),
	})
}

// buildNestedObject converts a flat map with slash-separated keys into a nested object structure.
// For example:
//
//	{
//	  "REGION": "us-east-1",
//	  "API/v2/ACCESS_KEY": "key123",
//	  "API/v2/SECRET_KEY": "secret456",
//	  "database/prod/HOST": "db.example.com"
//	}
//
// becomes:
//
//	{
//	  REGION: "us-east-1"
//	  API: {
//	    v2: {
//	      ACCESS_KEY: "key123"
//	      SECRET_KEY: "secret456"
//	    }
//	  }
//	  database: {
//	    prod: {
//	      HOST: "db.example.com"
//	    }
//	  }
//	}
func buildNestedObject(flatMap map[string]string) types.Object {
	// Build a tree structure first
	type node struct {
		value    *string          // non-nil for leaf nodes
		children map[string]*node // non-nil for branch nodes
	}

	root := &node{children: make(map[string]*node)}

	// Insert all paths into the tree
	for path, value := range flatMap {
		parts := strings.Split(path, "/")
		current := root

		// Navigate/create the tree structure
		for i, part := range parts {
			isLeaf := (i == len(parts)-1)

			if isLeaf {
				// This is the final part - store the value
				valueCopy := value
				current.children[part] = &node{value: &valueCopy}
			} else {
				// This is an intermediate part - ensure child exists
				if current.children[part] == nil {
					current.children[part] = &node{children: make(map[string]*node)}
				}
				current = current.children[part]
			}
		}
	}

	// Convert tree to Terraform types
	var buildObject func(*node) types.Object
	buildObject = func(n *node) types.Object {
		attrTypes := make(map[string]attr.Type)
		attrValues := make(map[string]attr.Value)

		for key, child := range n.children {
			if child.value != nil {
				// Leaf node - string value
				attrTypes[key] = types.StringType
				attrValues[key] = types.StringValue(*child.value)
			} else {
				// Branch node - nested object
				childObj := buildObject(child)
				attrTypes[key] = childObj.Type(context.Background())
				attrValues[key] = childObj
			}
		}

		// Create object value
		objValue, _ := types.ObjectValue(attrTypes, attrValues)
		return objValue
	}

	return buildObject(root)
}

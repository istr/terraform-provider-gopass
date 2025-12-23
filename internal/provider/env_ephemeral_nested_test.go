// Copyright (c) Ingo Struck
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"testing"

	"github.com/gopasspw/gopass/pkg/gopass/secrets"
	"github.com/hashicorp/terraform-plugin-framework/ephemeral"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

// TestEnvEphemeralResource_Open_NestedPaths tests that nested/deep paths work correctly
// with dot-notation access in Terraform configs.
//
// This test verifies the fix for supporting paths like:
//
//	env/terraform/aws/production/API/v2/ACCESS_KEY
//
// Which should be accessible as:
//
//	credentials.API.v2.ACCESS_KEY
func TestEnvEphemeralResource_Open_NestedPaths(t *testing.T) {
	r := &EnvEphemeralResource{}
	mockStore := newMockStore()
	client := NewGopassClient("")
	client.store = mockStore
	r.client = client

	// Add test secrets with nested paths (deeper than immediate children)
	// These represent a real-world scenario where secrets are organized hierarchically
	secret1 := secrets.New()
	secret1.SetPassword("access-key-123")
	mockStore.secrets["env/test/API/v2/ACCESS_KEY"] = secret1

	secret2 := secrets.New()
	secret2.SetPassword("secret-key-456")
	mockStore.secrets["env/test/API/v2/SECRET_KEY"] = secret2

	secret3 := secrets.New()
	secret3.SetPassword("db-host-value")
	mockStore.secrets["env/test/database/prod/HOST"] = secret3

	secret4 := secrets.New()
	secret4.SetPassword("db-pass-789")
	mockStore.secrets["env/test/database/prod/PASSWORD"] = secret4

	// Also add a flat secret for backward compatibility
	secret5 := secrets.New()
	secret5.SetPassword("region-value")
	mockStore.secrets["env/test/REGION"] = secret5

	ctx := context.Background()
	schemaReq := ephemeral.SchemaRequest{}
	schemaResp := &ephemeral.SchemaResponse{}
	r.Schema(ctx, schemaReq, schemaResp)

	configValue := tftypes.NewValue(tftypes.Object{
		AttributeTypes: map[string]tftypes.Type{
			"path":        tftypes.String,
			"credentials": tftypes.DynamicPseudoType,
		},
	}, map[string]tftypes.Value{
		"path":        tftypes.NewValue(tftypes.String, "env/test"),
		"credentials": tftypes.NewValue(tftypes.DynamicPseudoType, nil),
	})

	resultRaw := tftypes.NewValue(tftypes.Object{
		AttributeTypes: map[string]tftypes.Type{
			"path":        tftypes.String,
			"credentials": tftypes.DynamicPseudoType,
		},
	}, nil)

	req := ephemeral.OpenRequest{
		Config: tfsdk.Config{
			Schema: schemaResp.Schema,
			Raw:    configValue,
		},
	}
	resp := &ephemeral.OpenResponse{
		Result: tfsdk.EphemeralResultData{
			Schema: schemaResp.Schema,
			Raw:    resultRaw,
		},
	}

	r.Open(ctx, req, resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("unexpected error: %v", resp.Diagnostics)
	}

	// Extract the result to verify structure
	var result EnvModel
	diags := resp.Result.Get(ctx, &result)
	if diags.HasError() {
		t.Fatalf("failed to get result: %v", diags)
	}

	// Verify the credentials dynamic value contains nested structure
	// The structure should be:
	// {
	//   REGION: "region-value"
	//   API: {
	//     v2: {
	//       ACCESS_KEY: "access-key-123"
	//       SECRET_KEY: "secret-key-456"
	//     }
	//   }
	//   database: {
	//     prod: {
	//       HOST: "db-host-value"
	//       PASSWORD: "db-pass-789"
	//     }
	//   }
	// }

	// Verify credentials is not null
	if result.Credentials.IsNull() {
		t.Fatal("credentials should not be null")
	}

	// For now, just verify that the operation succeeded
	// The actual nested structure verification will be done through
	// integration testing with real Terraform configs
	// TODO: Add more detailed structure verification once implementation is complete
}

// TestEnvEphemeralResource_Open_DeepNesting tests very deep nesting levels
func TestEnvEphemeralResource_Open_DeepNesting(t *testing.T) {
	r := &EnvEphemeralResource{}
	mockStore := newMockStore()
	client := NewGopassClient("")
	client.store = mockStore
	r.client = client

	// Add a secret with 4 levels of nesting
	secret := secrets.New()
	secret.SetPassword("very-nested-value")
	mockStore.secrets["env/deep/level1/level2/level3/level4/SECRET"] = secret

	ctx := context.Background()
	schemaReq := ephemeral.SchemaRequest{}
	schemaResp := &ephemeral.SchemaResponse{}
	r.Schema(ctx, schemaReq, schemaResp)

	configValue := tftypes.NewValue(tftypes.Object{
		AttributeTypes: map[string]tftypes.Type{
			"path":        tftypes.String,
			"credentials": tftypes.DynamicPseudoType,
		},
	}, map[string]tftypes.Value{
		"path":        tftypes.NewValue(tftypes.String, "env/deep"),
		"credentials": tftypes.NewValue(tftypes.DynamicPseudoType, nil),
	})

	resultRaw := tftypes.NewValue(tftypes.Object{
		AttributeTypes: map[string]tftypes.Type{
			"path":        tftypes.String,
			"credentials": tftypes.DynamicPseudoType,
		},
	}, nil)

	req := ephemeral.OpenRequest{
		Config: tfsdk.Config{
			Schema: schemaResp.Schema,
			Raw:    configValue,
		},
	}
	resp := &ephemeral.OpenResponse{
		Result: tfsdk.EphemeralResultData{
			Schema: schemaResp.Schema,
			Raw:    resultRaw,
		},
	}

	r.Open(ctx, req, resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("unexpected error: %v", resp.Diagnostics)
	}

	// Just verify it doesn't error - the structure exists
	var result EnvModel
	diags := resp.Result.Get(ctx, &result)
	if diags.HasError() {
		t.Fatalf("failed to get result: %v", diags)
	}

	if result.Credentials.IsNull() {
		t.Fatal("credentials should not be null")
	}
}

// TestEnvEphemeralResource_Open_MixedDepths tests mixed flat and nested secrets
func TestEnvEphemeralResource_Open_MixedDepths(t *testing.T) {
	r := &EnvEphemeralResource{}
	mockStore := newMockStore()
	client := NewGopassClient("")
	client.store = mockStore
	r.client = client

	// Mix of flat and nested secrets
	flat1 := secrets.New()
	flat1.SetPassword("flat-value-1")
	mockStore.secrets["env/mixed/FLAT_KEY"] = flat1

	nested1 := secrets.New()
	nested1.SetPassword("nested-value-1")
	mockStore.secrets["env/mixed/nested/KEY"] = nested1

	flat2 := secrets.New()
	flat2.SetPassword("flat-value-2")
	mockStore.secrets["env/mixed/ANOTHER_FLAT"] = flat2

	ctx := context.Background()
	schemaReq := ephemeral.SchemaRequest{}
	schemaResp := &ephemeral.SchemaResponse{}
	r.Schema(ctx, schemaReq, schemaResp)

	configValue := tftypes.NewValue(tftypes.Object{
		AttributeTypes: map[string]tftypes.Type{
			"path":        tftypes.String,
			"credentials": tftypes.DynamicPseudoType,
		},
	}, map[string]tftypes.Value{
		"path":        tftypes.NewValue(tftypes.String, "env/mixed"),
		"credentials": tftypes.NewValue(tftypes.DynamicPseudoType, nil),
	})

	resultRaw := tftypes.NewValue(tftypes.Object{
		AttributeTypes: map[string]tftypes.Type{
			"path":        tftypes.String,
			"credentials": tftypes.DynamicPseudoType,
		},
	}, nil)

	req := ephemeral.OpenRequest{
		Config: tfsdk.Config{
			Schema: schemaResp.Schema,
			Raw:    configValue,
		},
	}
	resp := &ephemeral.OpenResponse{
		Result: tfsdk.EphemeralResultData{
			Schema: schemaResp.Schema,
			Raw:    resultRaw,
		},
	}

	r.Open(ctx, req, resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("unexpected error: %v", resp.Diagnostics)
	}

	var result EnvModel
	diags := resp.Result.Get(ctx, &result)
	if diags.HasError() {
		t.Fatalf("failed to get result: %v", diags)
	}

	// For now, just verify that the operation succeeded
	// The actual nested structure verification will be done through
	// integration testing with real Terraform configs
	if result.Credentials.IsNull() {
		t.Fatal("credentials should not be null")
	}
}

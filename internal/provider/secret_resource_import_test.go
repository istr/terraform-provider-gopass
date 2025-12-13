// Copyright (c) Ingo Struck
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
)

/*
func TestSecretResource_ImportState(t *testing.T) {
	r := &SecretResource{}
	mockStore := newMockStore()
	client := NewGopassClient("")
	client.store = mockStore
	r.client = client

	// Add a secret to the mock store
	mockStore.secrets["test/secret"] = newMockSecret("test-password")
	mockStore.revisions["test/secret"] = []string{"1"}

	ctx := context.Background()

	// Get the schema for the state
	schemaReq := resource.SchemaRequest{}
	schemaResp := &resource.SchemaResponse{}
	r.Schema(ctx, schemaReq, schemaResp)

	req := resource.ImportStateRequest{
		ID: "test/secret",
	}
	resp := &resource.ImportStateResponse{
		State: tfsdk.State{
			Schema: schemaResp.Schema,
		},
	}

	r.ImportState(ctx, req, resp)

	if resp.Diagnostics.HasError() {
		t.Errorf("unexpected error: %v", resp.Diagnostics)
	}

	// Verify the imported attributes
	var idVal attr.Value
	resp.Diagnostics.Append(resp.State.GetAttribute(ctx, path.Root("id"), &idVal)...)
	if resp.Diagnostics.HasError() {
		t.Errorf("failed to get id: %v", resp.Diagnostics)
	}

	if idString := idVal.(types.String); idString.ValueString() != "test/secret" {
		t.Errorf("expected ID 'test/secret', got %q", idString.ValueString())
	}
}
*/

func TestSecretResource_ImportState_NotFound(t *testing.T) {
	r := &SecretResource{}
	mockStore := newMockStore()
	client := NewGopassClient("")
	client.store = mockStore
	r.client = client

	ctx := context.Background()
	req := resource.ImportStateRequest{
		ID: "nonexistent",
	}
	resp := &resource.ImportStateResponse{
		State: tfsdk.State{},
	}

	r.ImportState(ctx, req, resp)

	if !resp.Diagnostics.HasError() {
		t.Error("expected error for non-existent secret")
	}
}

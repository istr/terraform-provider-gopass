// Copyright (c) Ingo Struck
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"
	"testing"

	"github.com/gopasspw/gopass/pkg/gopass"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

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

	// Create an unknown object matching the schema
	objectType := tftypes.Object{
		AttributeTypes: map[string]tftypes.Type{
			"id":               tftypes.String,
			"path":             tftypes.String,
			"value_wo":         tftypes.String,
			"value_wo_version": tftypes.Number,
			"delete_on_remove": tftypes.Bool,
			"revision_count":   tftypes.Number,
		},
	}

	resp := &resource.ImportStateResponse{
		State: tfsdk.State{
			Schema: schemaResp.Schema,
			Raw:    tftypes.NewValue(objectType, tftypes.UnknownValue),
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

type failGetStoreImport struct {
	*mockStore
}

func (m *failGetStoreImport) Get(ctx context.Context, name, revision string) (gopass.Secret, error) {
	return nil, fmt.Errorf("forced failure")
}

func TestSecretResource_ImportState_ExistsError(t *testing.T) {
	r := &SecretResource{}
	baseStore := newMockStore()
	mockStore := &failGetStoreImport{mockStore: baseStore}
	client := NewGopassClient("")
	client.store = mockStore
	r.client = client

	ctx := context.Background()
	req := resource.ImportStateRequest{
		ID: "test/secret",
	}
	resp := &resource.ImportStateResponse{
		State: tfsdk.State{},
	}

	r.ImportState(ctx, req, resp)

	if !resp.Diagnostics.HasError() {
		t.Error("expected error for exists failure")
	}
}

type flakyGetStoreImport struct {
	*mockStore
	calls int
}

func (m *flakyGetStoreImport) Get(ctx context.Context, name, revision string) (gopass.Secret, error) {
	m.calls++
	if m.calls == 2 {
		return nil, fmt.Errorf("flaky failure")
	}
	return m.mockStore.Get(ctx, name, revision)
}

func TestSecretResource_ImportState_GetRevisionCountError(t *testing.T) {
	r := &SecretResource{}
	baseStore := newMockStore()
	baseStore.secrets["test/rev-fail"] = newMockSecret("test")
	mockStore := &flakyGetStoreImport{mockStore: baseStore}
	client := NewGopassClient("")
	client.store = mockStore
	r.client = client

	ctx := context.Background()
	schemaReq := resource.SchemaRequest{}
	schemaResp := &resource.SchemaResponse{}
	r.Schema(ctx, schemaReq, schemaResp)

	req := resource.ImportStateRequest{
		ID: "test/rev-fail",
	}

	// Create an unknown object matching the schema
	objectType := tftypes.Object{
		AttributeTypes: map[string]tftypes.Type{
			"id":               tftypes.String,
			"path":             tftypes.String,
			"value_wo":         tftypes.String,
			"value_wo_version": tftypes.Number,
			"delete_on_remove": tftypes.Bool,
			"revision_count":   tftypes.Number,
		},
	}

	resp := &resource.ImportStateResponse{
		State: tfsdk.State{
			Schema: schemaResp.Schema,
			Raw:    tftypes.NewValue(objectType, tftypes.UnknownValue),
		},
	}

	r.ImportState(ctx, req, resp)

	if resp.Diagnostics.HasError() {
		t.Errorf("unexpected error: %v", resp.Diagnostics)
	}

	// Verify revision count fallback to 1
	var revCountVal attr.Value
	resp.Diagnostics.Append(resp.State.GetAttribute(ctx, path.Root("revision_count"), &revCountVal)...)
	if resp.Diagnostics.HasError() {
		t.Errorf("failed to get revision_count: %v", resp.Diagnostics)
	}

	if revInt := revCountVal.(types.Int64); revInt.ValueInt64() != 1 {
		t.Errorf("expected revision count 1, got %d", revInt.ValueInt64())
	}
}

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

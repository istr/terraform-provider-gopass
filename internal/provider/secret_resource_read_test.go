// Copyright (c) Ingo Struck
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"
	"testing"

	"github.com/gopasspw/gopass/pkg/gopass"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

func TestSecretResource_Read(t *testing.T) {
	r := &SecretResource{}
	mockStore := newMockStore()
	client := NewGopassClient("")
	client.store = mockStore
	r.client = client

	// Add a secret to the mock store
	mockStore.secrets["test/secret"] = newMockSecret("test-password")
	mockStore.revisions["test/secret"] = []string{"1", "2"}

	ctx := context.Background()
	schemaReq := resource.SchemaRequest{}
	schemaResp := &resource.SchemaResponse{}
	r.Schema(ctx, schemaReq, schemaResp)

	stateValue := tftypes.NewValue(tftypes.Object{
		AttributeTypes: map[string]tftypes.Type{
			"id":               tftypes.String,
			"path":             tftypes.String,
			"value_wo":         tftypes.String,
			"value_wo_version": tftypes.Number,
			"delete_on_remove": tftypes.Bool,
			"revision_count":   tftypes.Number,
		},
	}, map[string]tftypes.Value{
		"id":               tftypes.NewValue(tftypes.String, "test/secret"),
		"path":             tftypes.NewValue(tftypes.String, "test/secret"),
		"value_wo":         tftypes.NewValue(tftypes.String, nil),
		"value_wo_version": tftypes.NewValue(tftypes.Number, nil),
		"delete_on_remove": tftypes.NewValue(tftypes.Bool, true),
		"revision_count":   tftypes.NewValue(tftypes.Number, 1),
	})

	req := resource.ReadRequest{
		State: tfsdk.State{
			Schema: schemaResp.Schema,
			Raw:    stateValue,
		},
	}
	resp := &resource.ReadResponse{
		State: tfsdk.State{
			Schema: schemaResp.Schema,
			Raw:    stateValue,
		},
	}

	r.Read(ctx, req, resp)

	if resp.Diagnostics.HasError() {
		t.Errorf("unexpected error: %v", resp.Diagnostics)
	}

	// Verify state was updated with new revision count
	var state SecretResourceModel
	resp.Diagnostics.Append(resp.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		t.Errorf("failed to get state: %v", resp.Diagnostics)
	}

	if state.RevisionCount.ValueInt64() != 2 {
		t.Errorf("expected revision count 2, got %d", state.RevisionCount.ValueInt64())
	}
}

func TestSecretResource_Read_ExistsError(t *testing.T) {
	r := &SecretResource{}
	mockStore := newMockStore()
	mockStore.shouldFail = true
	mockStore.failMsg = "exists check error"
	client := NewGopassClient("")
	client.store = mockStore
	r.client = client

	ctx := context.Background()
	schemaReq := resource.SchemaRequest{}
	schemaResp := &resource.SchemaResponse{}
	r.Schema(ctx, schemaReq, schemaResp)

	stateValue := tftypes.NewValue(tftypes.Object{
		AttributeTypes: map[string]tftypes.Type{
			"id":               tftypes.String,
			"path":             tftypes.String,
			"value_wo":         tftypes.String,
			"value_wo_version": tftypes.Number,
			"delete_on_remove": tftypes.Bool,
			"revision_count":   tftypes.Number,
		},
	}, map[string]tftypes.Value{
		"id":               tftypes.NewValue(tftypes.String, "test/secret"),
		"path":             tftypes.NewValue(tftypes.String, "test/secret"),
		"value_wo":         tftypes.NewValue(tftypes.String, nil),
		"value_wo_version": tftypes.NewValue(tftypes.Number, nil),
		"delete_on_remove": tftypes.NewValue(tftypes.Bool, true),
		"revision_count":   tftypes.NewValue(tftypes.Number, 1),
	})

	req := resource.ReadRequest{
		State: tfsdk.State{
			Schema: schemaResp.Schema,
			Raw:    stateValue,
		},
	}
	resp := &resource.ReadResponse{
		State: tfsdk.State{
			Schema: schemaResp.Schema,
			Raw:    stateValue,
		},
	}

	r.Read(ctx, req, resp)

	if !resp.Diagnostics.HasError() {
		t.Error("expected error for SecretExists failure")
	}
}

func TestSecretResource_Read_NotFound(t *testing.T) {
	r := &SecretResource{}
	mockStore := newMockStore()
	client := NewGopassClient("")
	client.store = mockStore
	r.client = client

	ctx := context.Background()
	schemaReq := resource.SchemaRequest{}
	schemaResp := &resource.SchemaResponse{}
	r.Schema(ctx, schemaReq, schemaResp)

	stateValue := tftypes.NewValue(tftypes.Object{
		AttributeTypes: map[string]tftypes.Type{
			"id":               tftypes.String,
			"path":             tftypes.String,
			"value_wo":         tftypes.String,
			"value_wo_version": tftypes.Number,
			"delete_on_remove": tftypes.Bool,
			"revision_count":   tftypes.Number,
		},
	}, map[string]tftypes.Value{
		"id":               tftypes.NewValue(tftypes.String, "nonexistent"),
		"path":             tftypes.NewValue(tftypes.String, "nonexistent"),
		"value_wo":         tftypes.NewValue(tftypes.String, nil),
		"value_wo_version": tftypes.NewValue(tftypes.Number, nil),
		"delete_on_remove": tftypes.NewValue(tftypes.Bool, true),
		"revision_count":   tftypes.NewValue(tftypes.Number, 1),
	})

	req := resource.ReadRequest{
		State: tfsdk.State{
			Schema: schemaResp.Schema,
			Raw:    stateValue,
		},
	}
	resp := &resource.ReadResponse{
		State: tfsdk.State{
			Schema: schemaResp.Schema,
			Raw:    stateValue,
		},
	}

	r.Read(ctx, req, resp)

	if resp.Diagnostics.HasError() {
		t.Errorf("unexpected error: %v", resp.Diagnostics)
	}

	// State should be removed when secret doesn't exist
	if !resp.State.Raw.IsNull() {
		t.Error("expected state to be removed for non-existent secret")
	}
}

func TestSecretResource_Read_StateGetError(t *testing.T) {
	r := &SecretResource{}
	client := NewGopassClient("")
	r.client = client
	ctx := context.Background()

	// Incompatible schema to force State.Get error
	incompatibleSchema := schema.Schema{
		Attributes: map[string]schema.Attribute{
			"path":             schema.Int64Attribute{Required: true},
			"id":               schema.StringAttribute{Computed: true},
			"value_wo":         schema.StringAttribute{Optional: true},
			"value_wo_version": schema.Int64Attribute{Optional: true},
			"delete_on_remove": schema.BoolAttribute{Optional: true},
			"revision_count":   schema.Int64Attribute{Computed: true},
		},
	}

	stateValue := tftypes.NewValue(tftypes.Object{
		AttributeTypes: map[string]tftypes.Type{
			"path":             tftypes.Number,
			"id":               tftypes.String,
			"value_wo":         tftypes.String,
			"value_wo_version": tftypes.Number,
			"delete_on_remove": tftypes.Bool,
			"revision_count":   tftypes.Number,
		},
	}, map[string]tftypes.Value{
		"path":             tftypes.NewValue(tftypes.Number, 123),
		"id":               tftypes.NewValue(tftypes.String, "id"),
		"value_wo":         tftypes.NewValue(tftypes.String, nil),
		"value_wo_version": tftypes.NewValue(tftypes.Number, nil),
		"delete_on_remove": tftypes.NewValue(tftypes.Bool, true),
		"revision_count":   tftypes.NewValue(tftypes.Number, 1),
	})

	req := resource.ReadRequest{
		State: tfsdk.State{
			Schema: incompatibleSchema,
			Raw:    stateValue,
		},
	}
	resp := &resource.ReadResponse{
		State: tfsdk.State{
			Schema: incompatibleSchema,
		},
	}

	r.Read(ctx, req, resp)

	if !resp.Diagnostics.HasError() {
		t.Error("expected error from State.Get but got none")
	}
}

// flakyGetStore fails on the second call to Get
type flakyGetStore struct {
	*mockStore
	calls int
}

func (m *flakyGetStore) Get(ctx context.Context, name, revision string) (gopass.Secret, error) {
	m.calls++
	if m.calls == 2 {
		return nil, fmt.Errorf("flaky failure")
	}
	return m.mockStore.Get(ctx, name, revision)
}

func TestSecretResource_Read_GetRevisionCountError(t *testing.T) {
	r := &SecretResource{}
	baseStore := newMockStore()
	// Pre-populate store so SecretExists succeeds
	baseStore.secrets["test/flaky"] = newMockSecret("test")
	mockStore := &flakyGetStore{mockStore: baseStore}

	client := NewGopassClient("")
	client.store = mockStore
	r.client = client

	ctx := context.Background()
	schemaReq := resource.SchemaRequest{}
	schemaResp := &resource.SchemaResponse{}
	r.Schema(ctx, schemaReq, schemaResp)

	stateValue := tftypes.NewValue(tftypes.Object{
		AttributeTypes: map[string]tftypes.Type{
			"id":               tftypes.String,
			"path":             tftypes.String,
			"value_wo":         tftypes.String,
			"value_wo_version": tftypes.Number,
			"delete_on_remove": tftypes.Bool,
			"revision_count":   tftypes.Number,
		},
	}, map[string]tftypes.Value{
		"id":               tftypes.NewValue(tftypes.String, "test/flaky"),
		"path":             tftypes.NewValue(tftypes.String, "test/flaky"),
		"value_wo":         tftypes.NewValue(tftypes.String, nil),
		"value_wo_version": tftypes.NewValue(tftypes.Number, nil),
		"delete_on_remove": tftypes.NewValue(tftypes.Bool, true),
		"revision_count":   tftypes.NewValue(tftypes.Number, 1),
	})

	req := resource.ReadRequest{
		State: tfsdk.State{
			Schema: schemaResp.Schema,
			Raw:    stateValue,
		},
	}
	resp := &resource.ReadResponse{
		State: tfsdk.State{
			Schema: schemaResp.Schema,
			Raw:    stateValue,
		},
	}

	r.Read(ctx, req, resp)

	// Should not have error (warning only)
	if resp.Diagnostics.HasError() {
		t.Errorf("unexpected error: %v", resp.Diagnostics)
	}
}

func TestSecretResource_Read_DriftWarning(t *testing.T) {
	r := &SecretResource{}
	mockStore := newMockStore()
	client := NewGopassClient("")
	client.store = mockStore
	r.client = client

	// Secret exists and has 2 revisions
	mockStore.secrets["test/drift"] = newMockSecret("test")
	mockStore.revisions["test/drift"] = []string{"1", "2"}

	ctx := context.Background()
	schemaReq := resource.SchemaRequest{}
	schemaResp := &resource.SchemaResponse{}
	r.Schema(ctx, schemaReq, schemaResp)

	// State has 1 revision
	stateValue := tftypes.NewValue(tftypes.Object{
		AttributeTypes: map[string]tftypes.Type{
			"id":               tftypes.String,
			"path":             tftypes.String,
			"value_wo":         tftypes.String,
			"value_wo_version": tftypes.Number,
			"delete_on_remove": tftypes.Bool,
			"revision_count":   tftypes.Number,
		},
	}, map[string]tftypes.Value{
		"id":               tftypes.NewValue(tftypes.String, "test/drift"),
		"path":             tftypes.NewValue(tftypes.String, "test/drift"),
		"value_wo":         tftypes.NewValue(tftypes.String, nil),
		"value_wo_version": tftypes.NewValue(tftypes.Number, nil),
		"delete_on_remove": tftypes.NewValue(tftypes.Bool, true),
		"revision_count":   tftypes.NewValue(tftypes.Number, 1),
	})

	req := resource.ReadRequest{
		State: tfsdk.State{
			Schema: schemaResp.Schema,
			Raw:    stateValue,
		},
	}
	resp := &resource.ReadResponse{
		State: tfsdk.State{
			Schema: schemaResp.Schema,
			Raw:    stateValue,
		},
	}

	r.Read(ctx, req, resp)

	if resp.Diagnostics.HasError() {
		t.Errorf("unexpected error: %v", resp.Diagnostics)
	}

	// Should have warning
	hasWarning := false
	for _, diag := range resp.Diagnostics {
		if diag.Summary() == "Secret modified outside of Terraform" {
			hasWarning = true
			break
		}
	}
	if !hasWarning {
		t.Error("expected warning about drift")
	}

	// Check updated revision count
	var state SecretResourceModel
	resp.Diagnostics.Append(resp.State.Get(ctx, &state)...)
	if state.RevisionCount.ValueInt64() != 2 {
		t.Errorf("expected updated revision count 2, got %d", state.RevisionCount.ValueInt64())
	}
}

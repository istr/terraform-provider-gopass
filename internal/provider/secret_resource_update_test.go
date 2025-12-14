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

func TestSecretResource_Update(t *testing.T) {
	r := &SecretResource{}
	mockStore := newMockStore()
	client := NewGopassClient("")
	client.store = mockStore
	r.client = client

	// Pre-populate secret
	mockStore.secrets["test/update"] = newMockSecret("old-password")
	mockStore.revisions["test/update"] = []string{"1"}

	ctx := context.Background()
	schemaReq := resource.SchemaRequest{}
	schemaResp := &resource.SchemaResponse{}
	r.Schema(ctx, schemaReq, schemaResp)

	// State: version 1
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
		"id":               tftypes.NewValue(tftypes.String, "test/update"),
		"path":             tftypes.NewValue(tftypes.String, "test/update"),
		"value_wo":         tftypes.NewValue(tftypes.String, nil),
		"value_wo_version": tftypes.NewValue(tftypes.Number, 1),
		"delete_on_remove": tftypes.NewValue(tftypes.Bool, true),
		"revision_count":   tftypes.NewValue(tftypes.Number, 1),
	})

	// Plan: version 2
	planValue := tftypes.NewValue(tftypes.Object{
		AttributeTypes: map[string]tftypes.Type{
			"id":               tftypes.String,
			"path":             tftypes.String,
			"value_wo":         tftypes.String,
			"value_wo_version": tftypes.Number,
			"delete_on_remove": tftypes.Bool,
			"revision_count":   tftypes.Number,
		},
	}, map[string]tftypes.Value{
		"id":               tftypes.NewValue(tftypes.String, "test/update"),
		"path":             tftypes.NewValue(tftypes.String, "test/update"),
		"value_wo":         tftypes.NewValue(tftypes.String, tftypes.UnknownValue), // Unknown in plan?
		"value_wo_version": tftypes.NewValue(tftypes.Number, 2),
		"delete_on_remove": tftypes.NewValue(tftypes.Bool, true),
		"revision_count":   tftypes.NewValue(tftypes.Number, tftypes.UnknownValue),
	})

	// Config: has value
	configValue := tftypes.NewValue(tftypes.Object{
		AttributeTypes: map[string]tftypes.Type{
			"id":               tftypes.String,
			"path":             tftypes.String,
			"value_wo":         tftypes.String,
			"value_wo_version": tftypes.Number,
			"delete_on_remove": tftypes.Bool,
			"revision_count":   tftypes.Number,
		},
	}, map[string]tftypes.Value{
		"id":               tftypes.NewValue(tftypes.String, "test/update"),
		"path":             tftypes.NewValue(tftypes.String, "test/update"),
		"value_wo":         tftypes.NewValue(tftypes.String, "new-password"),
		"value_wo_version": tftypes.NewValue(tftypes.Number, 2),
		"delete_on_remove": tftypes.NewValue(tftypes.Bool, true),
		"revision_count":   tftypes.NewValue(tftypes.Number, nil),
	})

	req := resource.UpdateRequest{
		State:  tfsdk.State{Schema: schemaResp.Schema, Raw: stateValue},
		Plan:   tfsdk.Plan{Schema: schemaResp.Schema, Raw: planValue},
		Config: tfsdk.Config{Schema: schemaResp.Schema, Raw: configValue},
	}
	resp := &resource.UpdateResponse{
		State: tfsdk.State{Schema: schemaResp.Schema},
	}

	r.Update(ctx, req, resp)

	if resp.Diagnostics.HasError() {
		t.Errorf("unexpected error: %v", resp.Diagnostics)
	}

	// Verify secret updated
	secret := mockStore.secrets["test/update"]
	if secret.Password() != "new-password" {
		t.Errorf("expected password 'new-password', got %q", secret.Password())
	}

	// Verify state updated (revision count should be 2, because SetSecret adds revision)
	// mockStore.Set adds revision "1" if new, or appends if exists.
	// Initial was "1". Set appends "2".
	var state SecretResourceModel
	resp.Diagnostics.Append(resp.State.Get(ctx, &state)...)
	if state.RevisionCount.ValueInt64() != 2 {
		t.Errorf("expected revision count 2, got %d", state.RevisionCount.ValueInt64())
	}
}

func TestSecretResource_Update_NoChange(t *testing.T) {
	r := &SecretResource{}
	mockStore := newMockStore()
	client := NewGopassClient("")
	client.store = mockStore
	r.client = client

	mockStore.secrets["test/no-change"] = newMockSecret("old-password")
	mockStore.revisions["test/no-change"] = []string{"1"}

	ctx := context.Background()
	schemaReq := resource.SchemaRequest{}
	schemaResp := &resource.SchemaResponse{}
	r.Schema(ctx, schemaReq, schemaResp)

	// State: version 1
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
		"id":               tftypes.NewValue(tftypes.String, "test/no-change"),
		"path":             tftypes.NewValue(tftypes.String, "test/no-change"),
		"value_wo":         tftypes.NewValue(tftypes.String, nil),
		"value_wo_version": tftypes.NewValue(tftypes.Number, 1),
		"delete_on_remove": tftypes.NewValue(tftypes.Bool, true),
		"revision_count":   tftypes.NewValue(tftypes.Number, 1),
	})

	// Plan: version 1 (unchanged)
	planValue := stateValue // Plan same as state implies no change?
	// Actually plan usually has unknown computed values
	// But critical part is ValueWOVersion is 1 in both

	// Config: value provided, but version same
	configValue := tftypes.NewValue(tftypes.Object{
		AttributeTypes: map[string]tftypes.Type{
			"id":               tftypes.String,
			"path":             tftypes.String,
			"value_wo":         tftypes.String,
			"value_wo_version": tftypes.Number,
			"delete_on_remove": tftypes.Bool,
			"revision_count":   tftypes.Number,
		},
	}, map[string]tftypes.Value{
		"id":               tftypes.NewValue(tftypes.String, "test/no-change"),
		"path":             tftypes.NewValue(tftypes.String, "test/no-change"),
		"value_wo":         tftypes.NewValue(tftypes.String, "new-password-ignored"),
		"value_wo_version": tftypes.NewValue(tftypes.Number, 1),
		"delete_on_remove": tftypes.NewValue(tftypes.Bool, true),
		"revision_count":   tftypes.NewValue(tftypes.Number, nil),
	})

	req := resource.UpdateRequest{
		State:  tfsdk.State{Schema: schemaResp.Schema, Raw: stateValue},
		Plan:   tfsdk.Plan{Schema: schemaResp.Schema, Raw: planValue},
		Config: tfsdk.Config{Schema: schemaResp.Schema, Raw: configValue},
	}
	resp := &resource.UpdateResponse{
		State: tfsdk.State{Schema: schemaResp.Schema},
	}

	r.Update(ctx, req, resp)

	if resp.Diagnostics.HasError() {
		t.Errorf("unexpected error: %v", resp.Diagnostics)
	}

	// Verify secret NOT updated
	secret := mockStore.secrets["test/no-change"]
	if secret.Password() != "old-password" {
		t.Errorf("expected password 'old-password', got %q", secret.Password())
	}
}

func TestSecretResource_Update_Warning(t *testing.T) {
	r := &SecretResource{}
	mockStore := newMockStore()
	client := NewGopassClient("")
	client.store = mockStore
	r.client = client

	mockStore.secrets["test/warn"] = newMockSecret("old")

	ctx := context.Background()
	schemaReq := resource.SchemaRequest{}
	schemaResp := &resource.SchemaResponse{}
	r.Schema(ctx, schemaReq, schemaResp)

	// State: version 1
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
		"id":               tftypes.NewValue(tftypes.String, "test/warn"),
		"path":             tftypes.NewValue(tftypes.String, "test/warn"),
		"value_wo":         tftypes.NewValue(tftypes.String, nil),
		"value_wo_version": tftypes.NewValue(tftypes.Number, 1),
		"delete_on_remove": tftypes.NewValue(tftypes.Bool, true),
		"revision_count":   tftypes.NewValue(tftypes.Number, 1),
	})

	// Plan: version 2
	planValue := tftypes.NewValue(tftypes.Object{
		AttributeTypes: map[string]tftypes.Type{
			"id":               tftypes.String,
			"path":             tftypes.String,
			"value_wo":         tftypes.String,
			"value_wo_version": tftypes.Number,
			"delete_on_remove": tftypes.Bool,
			"revision_count":   tftypes.Number,
		},
	}, map[string]tftypes.Value{
		"id":               tftypes.NewValue(tftypes.String, "test/warn"),
		"path":             tftypes.NewValue(tftypes.String, "test/warn"),
		"value_wo":         tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
		"value_wo_version": tftypes.NewValue(tftypes.Number, 2),
		"delete_on_remove": tftypes.NewValue(tftypes.Bool, true),
		"revision_count":   tftypes.NewValue(tftypes.Number, tftypes.UnknownValue),
	})

	// Config: NO value
	configValue := tftypes.NewValue(tftypes.Object{
		AttributeTypes: map[string]tftypes.Type{
			"id":               tftypes.String,
			"path":             tftypes.String,
			"value_wo":         tftypes.String,
			"value_wo_version": tftypes.Number,
			"delete_on_remove": tftypes.Bool,
			"revision_count":   tftypes.Number,
		},
	}, map[string]tftypes.Value{
		"id":               tftypes.NewValue(tftypes.String, "test/warn"),
		"path":             tftypes.NewValue(tftypes.String, "test/warn"),
		"value_wo":         tftypes.NewValue(tftypes.String, nil), // Null
		"value_wo_version": tftypes.NewValue(tftypes.Number, 2),
		"delete_on_remove": tftypes.NewValue(tftypes.Bool, true),
		"revision_count":   tftypes.NewValue(tftypes.Number, nil),
	})

	req := resource.UpdateRequest{
		State:  tfsdk.State{Schema: schemaResp.Schema, Raw: stateValue},
		Plan:   tfsdk.Plan{Schema: schemaResp.Schema, Raw: planValue},
		Config: tfsdk.Config{Schema: schemaResp.Schema, Raw: configValue},
	}
	resp := &resource.UpdateResponse{
		State: tfsdk.State{Schema: schemaResp.Schema},
	}

	r.Update(ctx, req, resp)

	hasWarning := false
	for _, diag := range resp.Diagnostics {
		if diag.Summary() == "Version changed but no value provided" {
			hasWarning = true
			break
		}
	}
	if !hasWarning {
		t.Error("expected warning")
	}
}

// Local wrapper for flaky Get (fails on 2nd call)
// Reuse same concept as read_test but specialized for Update flow
// Update flow: SetSecret (calls EnsureStore, Set) -> GetRevisionCount (calls EnsureStore, Get)
// So we need Get to fail. SetSecret does NOT call Get.
// So we can just fail Get unconditionally?
// No, SetSecret calls EnsureStore which might do something? No.
// So failing Get is fine.

type failGetStoreUpdate struct {
	*mockStore
}

func (m *failGetStoreUpdate) Get(ctx context.Context, name, revision string) (gopass.Secret, error) {
	return nil, fmt.Errorf("forced failure")
}

func TestSecretResource_Update_GetRevisionCountError(t *testing.T) {
	r := &SecretResource{}
	baseStore := newMockStore()
	baseStore.secrets["test/rev-fail"] = newMockSecret("old")
	baseStore.revisions["test/rev-fail"] = []string{"1"}

	mockStore := &failGetStoreUpdate{mockStore: baseStore}

	client := NewGopassClient("")
	client.store = mockStore
	r.client = client

	ctx := context.Background()
	schemaReq := resource.SchemaRequest{}
	schemaResp := &resource.SchemaResponse{}
	r.Schema(ctx, schemaReq, schemaResp)

	// State: version 1, rev count 1
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
		"id":               tftypes.NewValue(tftypes.String, "test/rev-fail"),
		"path":             tftypes.NewValue(tftypes.String, "test/rev-fail"),
		"value_wo":         tftypes.NewValue(tftypes.String, nil),
		"value_wo_version": tftypes.NewValue(tftypes.Number, 1),
		"delete_on_remove": tftypes.NewValue(tftypes.Bool, true),
		"revision_count":   tftypes.NewValue(tftypes.Number, 1),
	})

	// Plan: version 2
	planValue := tftypes.NewValue(tftypes.Object{
		AttributeTypes: map[string]tftypes.Type{
			"id":               tftypes.String,
			"path":             tftypes.String,
			"value_wo":         tftypes.String,
			"value_wo_version": tftypes.Number,
			"delete_on_remove": tftypes.Bool,
			"revision_count":   tftypes.Number,
		},
	}, map[string]tftypes.Value{
		"id":               tftypes.NewValue(tftypes.String, "test/rev-fail"),
		"path":             tftypes.NewValue(tftypes.String, "test/rev-fail"),
		"value_wo":         tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
		"value_wo_version": tftypes.NewValue(tftypes.Number, 2),
		"delete_on_remove": tftypes.NewValue(tftypes.Bool, true),
		"revision_count":   tftypes.NewValue(tftypes.Number, tftypes.UnknownValue),
	})

	// Config
	configValue := tftypes.NewValue(tftypes.Object{
		AttributeTypes: map[string]tftypes.Type{
			"id":               tftypes.String,
			"path":             tftypes.String,
			"value_wo":         tftypes.String,
			"value_wo_version": tftypes.Number,
			"delete_on_remove": tftypes.Bool,
			"revision_count":   tftypes.Number,
		},
	}, map[string]tftypes.Value{
		"id":               tftypes.NewValue(tftypes.String, "test/rev-fail"),
		"path":             tftypes.NewValue(tftypes.String, "test/rev-fail"),
		"value_wo":         tftypes.NewValue(tftypes.String, "new"),
		"value_wo_version": tftypes.NewValue(tftypes.Number, 2),
		"delete_on_remove": tftypes.NewValue(tftypes.Bool, true),
		"revision_count":   tftypes.NewValue(tftypes.Number, nil),
	})

	req := resource.UpdateRequest{
		State:  tfsdk.State{Schema: schemaResp.Schema, Raw: stateValue},
		Plan:   tfsdk.Plan{Schema: schemaResp.Schema, Raw: planValue},
		Config: tfsdk.Config{Schema: schemaResp.Schema, Raw: configValue},
	}
	resp := &resource.UpdateResponse{
		State: tfsdk.State{Schema: schemaResp.Schema},
	}

	r.Update(ctx, req, resp)

	if resp.Diagnostics.HasError() {
		t.Errorf("unexpected error: %v", resp.Diagnostics)
	}

	// Secret should have been updated
	if baseStore.secrets["test/rev-fail"].Password() != "new" {
		t.Error("secret should be updated")
	}

	// But revision count check failed, so state should keep old count (1)
	var state SecretResourceModel
	resp.Diagnostics.Append(resp.State.Get(ctx, &state)...)
	if state.RevisionCount.ValueInt64() != 1 {
		t.Errorf("expected fallback revision count 1, got %d", state.RevisionCount.ValueInt64())
	}
}

func TestSecretResource_Update_SetSecretError(t *testing.T) {
	r := &SecretResource{}
	mockStore := newMockStore()
	mockStore.shouldFail = true
	mockStore.failMsg = "write error"
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
		"id":               tftypes.NewValue(tftypes.String, "test/err"),
		"path":             tftypes.NewValue(tftypes.String, "test/err"),
		"value_wo":         tftypes.NewValue(tftypes.String, nil),
		"value_wo_version": tftypes.NewValue(tftypes.Number, 1),
		"delete_on_remove": tftypes.NewValue(tftypes.Bool, true),
		"revision_count":   tftypes.NewValue(tftypes.Number, 1),
	})

	planValue := tftypes.NewValue(tftypes.Object{
		AttributeTypes: map[string]tftypes.Type{
			"id":               tftypes.String,
			"path":             tftypes.String,
			"value_wo":         tftypes.String,
			"value_wo_version": tftypes.Number,
			"delete_on_remove": tftypes.Bool,
			"revision_count":   tftypes.Number,
		},
	}, map[string]tftypes.Value{
		"id":               tftypes.NewValue(tftypes.String, "test/err"),
		"path":             tftypes.NewValue(tftypes.String, "test/err"),
		"value_wo":         tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
		"value_wo_version": tftypes.NewValue(tftypes.Number, 2),
		"delete_on_remove": tftypes.NewValue(tftypes.Bool, true),
		"revision_count":   tftypes.NewValue(tftypes.Number, tftypes.UnknownValue),
	})

	configValue := tftypes.NewValue(tftypes.Object{
		AttributeTypes: map[string]tftypes.Type{
			"id":               tftypes.String,
			"path":             tftypes.String,
			"value_wo":         tftypes.String,
			"value_wo_version": tftypes.Number,
			"delete_on_remove": tftypes.Bool,
			"revision_count":   tftypes.Number,
		},
	}, map[string]tftypes.Value{
		"id":               tftypes.NewValue(tftypes.String, "test/err"),
		"path":             tftypes.NewValue(tftypes.String, "test/err"),
		"value_wo":         tftypes.NewValue(tftypes.String, "new"),
		"value_wo_version": tftypes.NewValue(tftypes.Number, 2),
		"delete_on_remove": tftypes.NewValue(tftypes.Bool, true),
		"revision_count":   tftypes.NewValue(tftypes.Number, nil),
	})

	req := resource.UpdateRequest{
		State:  tfsdk.State{Schema: schemaResp.Schema, Raw: stateValue},
		Plan:   tfsdk.Plan{Schema: schemaResp.Schema, Raw: planValue},
		Config: tfsdk.Config{Schema: schemaResp.Schema, Raw: configValue},
	}
	resp := &resource.UpdateResponse{
		State: tfsdk.State{Schema: schemaResp.Schema},
	}

	r.Update(ctx, req, resp)

	if !resp.Diagnostics.HasError() {
		t.Error("expected error from SetSecret")
	}
}

func TestSecretResource_Update_PlanGetError(t *testing.T) {
	r := &SecretResource{}
	client := NewGopassClient("")
	r.client = client
	ctx := context.Background()

	// Incompatible schema
	incompatibleSchema := schema.Schema{
		Attributes: map[string]schema.Attribute{
			"path": schema.Int64Attribute{Required: true},
			// ...
		},
	}

	stateValue := tftypes.NewValue(tftypes.Object{
		AttributeTypes: map[string]tftypes.Type{
			"path": tftypes.Number,
		},
	}, map[string]tftypes.Value{
		"path": tftypes.NewValue(tftypes.Number, 123),
	})

	req := resource.UpdateRequest{
		State:  tfsdk.State{Schema: incompatibleSchema, Raw: stateValue},
		Plan:   tfsdk.Plan{Schema: incompatibleSchema, Raw: stateValue},
		Config: tfsdk.Config{Schema: incompatibleSchema, Raw: stateValue},
	}
	resp := &resource.UpdateResponse{
		State: tfsdk.State{Schema: incompatibleSchema},
	}

	r.Update(ctx, req, resp)

	if !resp.Diagnostics.HasError() {
		t.Error("expected error from Plan.Get but got none")
	}
}

func TestSecretResource_Update_ConfigGetError(t *testing.T) {
	r := &SecretResource{}
	client := NewGopassClient("")
	r.client = client
	ctx := context.Background()

	validSchema := schema.Schema{
		Attributes: map[string]schema.Attribute{
			"path":             schema.StringAttribute{Required: true},
			"id":               schema.StringAttribute{Computed: true},
			"value_wo":         schema.StringAttribute{Optional: true},
			"value_wo_version": schema.Int64Attribute{Optional: true},
			"delete_on_remove": schema.BoolAttribute{Optional: true},
			"revision_count":   schema.Int64Attribute{Computed: true},
		},
	}

	validValue := tftypes.NewValue(tftypes.Object{
		AttributeTypes: map[string]tftypes.Type{
			"path":             tftypes.String,
			"id":               tftypes.String,
			"value_wo":         tftypes.String,
			"value_wo_version": tftypes.Number,
			"delete_on_remove": tftypes.Bool,
			"revision_count":   tftypes.Number,
		},
	}, map[string]tftypes.Value{
		"path":             tftypes.NewValue(tftypes.String, "path"),
		"id":               tftypes.NewValue(tftypes.String, "id"),
		"value_wo":         tftypes.NewValue(tftypes.String, nil),
		"value_wo_version": tftypes.NewValue(tftypes.Number, 1),
		"delete_on_remove": tftypes.NewValue(tftypes.Bool, true),
		"revision_count":   tftypes.NewValue(tftypes.Number, 1),
	})

	incompatibleSchema := schema.Schema{
		Attributes: map[string]schema.Attribute{
			"path": schema.Int64Attribute{Required: true},
		},
	}
	incompatibleValue := tftypes.NewValue(tftypes.Object{
		AttributeTypes: map[string]tftypes.Type{
			"path": tftypes.Number,
		},
	}, map[string]tftypes.Value{
		"path": tftypes.NewValue(tftypes.Number, 123),
	})

	req := resource.UpdateRequest{
		State:  tfsdk.State{Schema: validSchema, Raw: validValue},
		Plan:   tfsdk.Plan{Schema: validSchema, Raw: validValue},
		Config: tfsdk.Config{Schema: incompatibleSchema, Raw: incompatibleValue},
	}
	resp := &resource.UpdateResponse{
		State: tfsdk.State{Schema: validSchema},
	}

	r.Update(ctx, req, resp)

	if !resp.Diagnostics.HasError() {
		t.Error("expected error from Config.Get but got none")
	}
}

func TestSecretResource_Update_StateGetError(t *testing.T) {
	r := &SecretResource{}
	client := NewGopassClient("")
	r.client = client
	ctx := context.Background()

	validSchema := schema.Schema{
		Attributes: map[string]schema.Attribute{
			"path":             schema.StringAttribute{Required: true},
			"id":               schema.StringAttribute{Computed: true},
			"value_wo":         schema.StringAttribute{Optional: true},
			"value_wo_version": schema.Int64Attribute{Optional: true},
			"delete_on_remove": schema.BoolAttribute{Optional: true},
			"revision_count":   schema.Int64Attribute{Computed: true},
		},
	}

	validValue := tftypes.NewValue(tftypes.Object{
		AttributeTypes: map[string]tftypes.Type{
			"path":             tftypes.String,
			"id":               tftypes.String,
			"value_wo":         tftypes.String,
			"value_wo_version": tftypes.Number,
			"delete_on_remove": tftypes.Bool,
			"revision_count":   tftypes.Number,
		},
	}, map[string]tftypes.Value{
		"path":             tftypes.NewValue(tftypes.String, "path"),
		"id":               tftypes.NewValue(tftypes.String, "id"),
		"value_wo":         tftypes.NewValue(tftypes.String, nil),
		"value_wo_version": tftypes.NewValue(tftypes.Number, 1),
		"delete_on_remove": tftypes.NewValue(tftypes.Bool, true),
		"revision_count":   tftypes.NewValue(tftypes.Number, 1),
	})

	incompatibleSchema := schema.Schema{
		Attributes: map[string]schema.Attribute{
			"path": schema.Int64Attribute{Required: true},
		},
	}
	incompatibleValue := tftypes.NewValue(tftypes.Object{
		AttributeTypes: map[string]tftypes.Type{
			"path": tftypes.Number,
		},
	}, map[string]tftypes.Value{
		"path": tftypes.NewValue(tftypes.Number, 123),
	})

	req := resource.UpdateRequest{
		State:  tfsdk.State{Schema: incompatibleSchema, Raw: incompatibleValue},
		Plan:   tfsdk.Plan{Schema: validSchema, Raw: validValue},
		Config: tfsdk.Config{Schema: validSchema, Raw: validValue},
	}
	resp := &resource.UpdateResponse{
		State: tfsdk.State{Schema: incompatibleSchema},
	}

	r.Update(ctx, req, resp)

	if !resp.Diagnostics.HasError() {
		t.Error("expected error from State.Get but got none")
	}
}

func TestSecretResource_Update_AddVersion(t *testing.T) {
	r := &SecretResource{}
	mockStore := newMockStore()
	client := NewGopassClient("")
	client.store = mockStore
	r.client = client

	mockStore.secrets["test/add-ver"] = newMockSecret("old")
	mockStore.revisions["test/add-ver"] = []string{"1"}

	ctx := context.Background()
	schemaReq := resource.SchemaRequest{}
	schemaResp := &resource.SchemaResponse{}
	r.Schema(ctx, schemaReq, schemaResp)

	// State: version is null (was not tracked)
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
		"id":               tftypes.NewValue(tftypes.String, "test/add-ver"),
		"path":             tftypes.NewValue(tftypes.String, "test/add-ver"),
		"value_wo":         tftypes.NewValue(tftypes.String, nil),
		"value_wo_version": tftypes.NewValue(tftypes.Number, nil),
		"delete_on_remove": tftypes.NewValue(tftypes.Bool, true),
		"revision_count":   tftypes.NewValue(tftypes.Number, 1),
	})

	// Plan: version is set
	planValue := tftypes.NewValue(tftypes.Object{
		AttributeTypes: map[string]tftypes.Type{
			"id":               tftypes.String,
			"path":             tftypes.String,
			"value_wo":         tftypes.String,
			"value_wo_version": tftypes.Number,
			"delete_on_remove": tftypes.Bool,
			"revision_count":   tftypes.Number,
		},
	}, map[string]tftypes.Value{
		"id":               tftypes.NewValue(tftypes.String, "test/add-ver"),
		"path":             tftypes.NewValue(tftypes.String, "test/add-ver"),
		"value_wo":         tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
		"value_wo_version": tftypes.NewValue(tftypes.Number, 1),
		"delete_on_remove": tftypes.NewValue(tftypes.Bool, true),
		"revision_count":   tftypes.NewValue(tftypes.Number, tftypes.UnknownValue),
	})

	// Config: value provided
	configValue := tftypes.NewValue(tftypes.Object{
		AttributeTypes: map[string]tftypes.Type{
			"id":               tftypes.String,
			"path":             tftypes.String,
			"value_wo":         tftypes.String,
			"value_wo_version": tftypes.Number,
			"delete_on_remove": tftypes.Bool,
			"revision_count":   tftypes.Number,
		},
	}, map[string]tftypes.Value{
		"id":               tftypes.NewValue(tftypes.String, "test/add-ver"),
		"path":             tftypes.NewValue(tftypes.String, "test/add-ver"),
		"value_wo":         tftypes.NewValue(tftypes.String, "new"),
		"value_wo_version": tftypes.NewValue(tftypes.Number, 1),
		"delete_on_remove": tftypes.NewValue(tftypes.Bool, true),
		"revision_count":   tftypes.NewValue(tftypes.Number, nil),
	})

	req := resource.UpdateRequest{
		State:  tfsdk.State{Schema: schemaResp.Schema, Raw: stateValue},
		Plan:   tfsdk.Plan{Schema: schemaResp.Schema, Raw: planValue},
		Config: tfsdk.Config{Schema: schemaResp.Schema, Raw: configValue},
	}
	resp := &resource.UpdateResponse{
		State: tfsdk.State{Schema: schemaResp.Schema},
	}

	r.Update(ctx, req, resp)

	if resp.Diagnostics.HasError() {
		t.Errorf("unexpected error: %v", resp.Diagnostics)
	}

	// Secret should have been updated
	if mockStore.secrets["test/add-ver"].Password() != "new" {
		t.Error("secret should be updated")
	}
}

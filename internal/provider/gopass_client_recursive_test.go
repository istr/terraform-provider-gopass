// Copyright (c) Ingo Struck
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"strings"
	"testing"

	"github.com/gopasspw/gopass/pkg/gopass/secrets"
)

func TestGopassClient_ListSecretsRecursive(t *testing.T) {
	client := NewGopassClient("")
	mockStore := newMockStore()
	client.store = mockStore

	// Add test secrets at various depths
	secret := secrets.New()
	secret.SetPassword("pass1")
	mockStore.secrets["env/test/secret1"] = secret
	mockStore.secrets["env/test/secret2"] = secret
	mockStore.secrets["env/test/sub/secret3"] = secret
	mockStore.secrets["env/test/sub/deep/secret4"] = secret
	mockStore.secrets["other/secret5"] = secret

	ctx := context.Background()

	results, err := client.ListSecretsRecursive(ctx, "env/test")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	expectedSecrets := []string{
		"env/test/secret1",
		"env/test/secret2",
		"env/test/sub/secret3",
		"env/test/sub/deep/secret4",
	}

	if len(results) != len(expectedSecrets) {
		t.Errorf("expected %d secrets, got %d", len(expectedSecrets), len(results))
	}

	for _, expected := range expectedSecrets {
		found := false
		for _, result := range results {
			if result == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("expected to find secret %q in results", expected)
		}
	}

	// Verify "other/secret5" was NOT included (wrong prefix)
	for _, result := range results {
		if result == "other/secret5" {
			t.Error("should not include secrets from other prefixes")
		}
	}
}

func TestGopassClient_ListSecretsRecursive_Empty(t *testing.T) {
	client := NewGopassClient("")
	mockStore := newMockStore()
	client.store = mockStore

	ctx := context.Background()

	results, err := client.ListSecretsRecursive(ctx, "env/nonexistent")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if len(results) != 0 {
		t.Errorf("expected 0 secrets for non-existent path, got %d", len(results))
	}
}

func TestGopassClient_ListSecretsRecursive_Error(t *testing.T) {
	client := NewGopassClient("")
	mockStore := newMockStore()
	mockStore.shouldFail = true
	mockStore.failMsg = "list error"
	client.store = mockStore

	ctx := context.Background()

	_, err := client.ListSecretsRecursive(ctx, "test/prefix")
	if err == nil {
		t.Error("expected error but got none")
	}

	if !strings.Contains(err.Error(), "failed to list secrets") {
		t.Errorf("expected wrapped error, got %v", err)
	}
}

func TestGopassClient_ListSecretsRecursive_EnsureStoreError(t *testing.T) {
	client := NewGopassClient("/nonexistent/path/for/test")

	ctx := context.Background()

	_, err := client.ListSecretsRecursive(ctx, "test/prefix")
	if err == nil {
		t.Error("expected error but got none")
	}
}

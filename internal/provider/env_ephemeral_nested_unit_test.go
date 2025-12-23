// Copyright (c) Ingo Struck
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
)

// TestBuildNestedObject_EmptyMap tests buildNestedObject with an empty map
func TestBuildNestedObject_EmptyMap(t *testing.T) {
	result := buildNestedObject(map[string]string{})

	if result.IsNull() {
		t.Error("expected non-null object for empty map")
	}

	attrs := result.Attributes()
	if len(attrs) != 0 {
		t.Errorf("expected 0 attributes for empty map, got %d", len(attrs))
	}
}

// TestBuildNestedObject_FlatKeys tests buildNestedObject with flat keys only
func TestBuildNestedObject_FlatKeys(t *testing.T) {
	input := map[string]string{
		"KEY1": "value1",
		"KEY2": "value2",
		"KEY3": "value3",
	}

	result := buildNestedObject(input)

	if result.IsNull() {
		t.Fatal("expected non-null object")
	}

	attrs := result.Attributes()
	if len(attrs) != 3 {
		t.Errorf("expected 3 attributes, got %d", len(attrs))
	}

	// Verify all keys exist and are strings
	for key := range input {
		if _, exists := attrs[key]; !exists {
			t.Errorf("expected key %q to exist", key)
		}
	}
}

// TestBuildNestedObject_SingleLevel tests one level of nesting
func TestBuildNestedObject_SingleLevel(t *testing.T) {
	input := map[string]string{
		"parent/child": "value",
	}

	result := buildNestedObject(input)

	attrs := result.Attributes()
	if len(attrs) != 1 {
		t.Errorf("expected 1 top-level attribute, got %d", len(attrs))
	}

	parentValue, exists := attrs["parent"]
	if !exists {
		t.Fatal("expected 'parent' attribute to exist")
	}

	parentObj, ok := parentValue.(types.Object)
	if !ok {
		t.Fatalf("expected parent to be Object, got %T", parentValue)
	}

	parentAttrs := parentObj.Attributes()
	if len(parentAttrs) != 1 {
		t.Errorf("expected 1 child attribute, got %d", len(parentAttrs))
	}

	childValue, exists := parentAttrs["child"]
	if !exists {
		t.Fatal("expected 'child' attribute to exist")
	}

	childStr, ok := childValue.(types.String)
	if !ok {
		t.Fatalf("expected child to be String, got %T", childValue)
	}

	if childStr.ValueString() != "value" {
		t.Errorf("expected child value 'value', got %q", childStr.ValueString())
	}
}

// TestBuildNestedObject_MultiLevel tests multiple levels of nesting
func TestBuildNestedObject_MultiLevel(t *testing.T) {
	input := map[string]string{
		"level1/level2/level3/leaf": "deep_value",
	}

	result := buildNestedObject(input)

	attrs := result.Attributes()
	if len(attrs) != 1 {
		t.Errorf("expected 1 top-level attribute, got %d", len(attrs))
	}

	// Navigate through levels
	level1, _ := attrs["level1"].(types.Object)
	level1Attrs := level1.Attributes()

	level2, _ := level1Attrs["level2"].(types.Object)
	level2Attrs := level2.Attributes()

	level3, _ := level2Attrs["level3"].(types.Object)
	level3Attrs := level3.Attributes()

	leaf, _ := level3Attrs["leaf"].(types.String)

	if leaf.ValueString() != "deep_value" {
		t.Errorf("expected leaf value 'deep_value', got %q", leaf.ValueString())
	}
}

// TestBuildNestedObject_MixedStructure tests mixed flat and nested keys
func TestBuildNestedObject_MixedStructure(t *testing.T) {
	input := map[string]string{
		"FLAT_KEY":          "flat_value",
		"nested/KEY":        "nested_value",
		"deeply/nested/KEY": "deep_value",
	}

	result := buildNestedObject(input)

	attrs := result.Attributes()
	if len(attrs) != 3 {
		t.Errorf("expected 3 top-level attributes, got %d", len(attrs))
	}

	// Verify flat key
	flatValue, _ := attrs["FLAT_KEY"].(types.String)
	if flatValue.ValueString() != "flat_value" {
		t.Errorf("expected 'flat_value', got %q", flatValue.ValueString())
	}

	// Verify nested key
	nested, _ := attrs["nested"].(types.Object)
	nestedAttrs := nested.Attributes()
	nestedKey, _ := nestedAttrs["KEY"].(types.String)
	if nestedKey.ValueString() != "nested_value" {
		t.Errorf("expected 'nested_value', got %q", nestedKey.ValueString())
	}

	// Verify deeply nested key
	deeply, _ := attrs["deeply"].(types.Object)
	deeplyAttrs := deeply.Attributes()
	deeplyNested, _ := deeplyAttrs["nested"].(types.Object)
	deeplyNestedAttrs := deeplyNested.Attributes()
	deepKey, _ := deeplyNestedAttrs["KEY"].(types.String)
	if deepKey.ValueString() != "deep_value" {
		t.Errorf("expected 'deep_value', got %q", deepKey.ValueString())
	}
}

// TestBuildNestedObject_MultipleSiblings tests multiple children under the same parent
func TestBuildNestedObject_MultipleSiblings(t *testing.T) {
	input := map[string]string{
		"parent/child1": "value1",
		"parent/child2": "value2",
		"parent/child3": "value3",
	}

	result := buildNestedObject(input)

	attrs := result.Attributes()
	parent, _ := attrs["parent"].(types.Object)
	parentAttrs := parent.Attributes()

	if len(parentAttrs) != 3 {
		t.Errorf("expected 3 siblings, got %d", len(parentAttrs))
	}

	// Verify all children exist
	for i := 1; i <= 3; i++ {
		childKey := "child" + string(rune('0'+i))
		if _, exists := parentAttrs[childKey]; !exists {
			t.Errorf("expected child %q to exist", childKey)
		}
	}
}

// TestBuildNestedObject_ComplexStructure tests a complex real-world structure
func TestBuildNestedObject_ComplexStructure(t *testing.T) {
	input := map[string]string{
		"REGION":                 "us-east-1",
		"API/v2/ACCESS_KEY":      "key123",
		"API/v2/SECRET_KEY":      "secret456",
		"API/v1/LEGACY_TOKEN":    "legacy789",
		"database/prod/HOST":     "db.example.com",
		"database/prod/PORT":     "5432",
		"database/prod/PASSWORD": "pass",
		"database/dev/HOST":      "dev.example.com",
	}

	result := buildNestedObject(input)

	attrs := result.Attributes()

	// Verify top-level has 3 keys: REGION, API, database
	if len(attrs) != 3 {
		t.Errorf("expected 3 top-level attributes, got %d", len(attrs))
	}

	// Verify REGION
	region, _ := attrs["REGION"].(types.String)
	if region.ValueString() != "us-east-1" {
		t.Errorf("expected REGION 'us-east-1', got %q", region.ValueString())
	}

	// Verify API structure
	api, _ := attrs["API"].(types.Object)
	apiAttrs := api.Attributes()
	if len(apiAttrs) != 2 {
		t.Errorf("expected 2 API versions, got %d", len(apiAttrs))
	}

	// Verify API/v2
	v2, _ := apiAttrs["v2"].(types.Object)
	v2Attrs := v2.Attributes()
	if len(v2Attrs) != 2 {
		t.Errorf("expected 2 v2 keys, got %d", len(v2Attrs))
	}

	// Verify database structure
	database, _ := attrs["database"].(types.Object)
	dbAttrs := database.Attributes()
	if len(dbAttrs) != 2 {
		t.Errorf("expected 2 database environments, got %d", len(dbAttrs))
	}

	// Verify database/prod has 3 keys
	prod, _ := dbAttrs["prod"].(types.Object)
	prodAttrs := prod.Attributes()
	if len(prodAttrs) != 3 {
		t.Errorf("expected 3 prod keys, got %d", len(prodAttrs))
	}
}

// TestBuildNestedObject_SpecialCharacters tests keys with various characters
func TestBuildNestedObject_SpecialCharacters(t *testing.T) {
	input := map[string]string{
		"KEY_WITH_UNDERSCORES": "value1",
		"KEY-WITH-DASHES":      "value2",
		"123_NUMERIC_START":    "value3",
		"parent/CHILD_KEY":     "value4",
	}

	result := buildNestedObject(input)

	if result.IsNull() {
		t.Fatal("expected non-null object")
	}

	attrs := result.Attributes()
	if len(attrs) != 4 {
		t.Errorf("expected 4 top-level attributes, got %d", len(attrs))
	}
}

// TestBuildNestedObject_TypeConsistency verifies type information is correct
func TestBuildNestedObject_TypeConsistency(t *testing.T) {
	input := map[string]string{
		"leaf":        "value",
		"branch/leaf": "value",
	}

	result := buildNestedObject(input)

	ctx := context.Background()
	objType := result.Type(ctx)

	if objType.String() == "" {
		t.Error("expected valid type string")
	}
}

// TestBuildNestedObject_SharedParentPath tests multiple paths sharing intermediate nodes
// This ensures the case where current.children already exists is covered
func TestBuildNestedObject_SharedParentPath(t *testing.T) {
	input := map[string]string{
		"parent/branch1/leaf1": "value1",
		"parent/branch1/leaf2": "value2",
		"parent/branch2/leaf3": "value3",
	}

	result := buildNestedObject(input)

	attrs := result.Attributes()
	parent, _ := attrs["parent"].(types.Object)
	parentAttrs := parent.Attributes()

	// Verify parent has 2 branches
	if len(parentAttrs) != 2 {
		t.Errorf("expected 2 branches, got %d", len(parentAttrs))
	}

	// Verify branch1 has 2 leaves
	branch1, _ := parentAttrs["branch1"].(types.Object)
	branch1Attrs := branch1.Attributes()
	if len(branch1Attrs) != 2 {
		t.Errorf("expected 2 leaves in branch1, got %d", len(branch1Attrs))
	}

	// Verify branch2 has 1 leaf
	branch2, _ := parentAttrs["branch2"].(types.Object)
	branch2Attrs := branch2.Attributes()
	if len(branch2Attrs) != 1 {
		t.Errorf("expected 1 leaf in branch2, got %d", len(branch2Attrs))
	}
}

// TestBuildNestedObject_ReusedIntermediateNodes tests the path where
// intermediate nodes already have children set
func TestBuildNestedObject_ReusedIntermediateNodes(t *testing.T) {
	input := map[string]string{
		"a/b/c": "value1",
		"a/b/d": "value2",
		"a/e/f": "value3",
	}

	result := buildNestedObject(input)

	attrs := result.Attributes()
	a, _ := attrs["a"].(types.Object)
	aAttrs := a.Attributes()

	// "a" should have 2 children: "b" and "e"
	if len(aAttrs) != 2 {
		t.Errorf("expected 2 children of 'a', got %d", len(aAttrs))
	}

	// "a/b" should have 2 children: "c" and "d"
	b, _ := aAttrs["b"].(types.Object)
	bAttrs := b.Attributes()
	if len(bAttrs) != 2 {
		t.Errorf("expected 2 children of 'b', got %d", len(bAttrs))
	}
}

// Copyright (c) Ingo Struck
// SPDX-License-Identifier: MPL-2.0

package provider

// Test helpers for secret resource tests

// newMockSecret creates a mock secret for testing.
func newMockSecret(password string) *mockSecret {
	return &mockSecret{
		password: password,
		fields:   make(map[string]string),
	}
}

// mockSecret implements gopass.Secret interface for testing
type mockSecret struct {
	password string
	fields   map[string]string
}

func (s *mockSecret) Password() string      { return s.password }
func (s *mockSecret) SetPassword(pw string) { s.password = pw }
func (s *mockSecret) Keys() []string {
	keys := make([]string, 0, len(s.fields))
	for k := range s.fields {
		keys = append(keys, k)
	}
	return keys
}
func (s *mockSecret) Get(key string) (string, bool) {
	val, ok := s.fields[key]
	return val, ok
}
func (s *mockSecret) Values(key string) ([]string, bool) {
	val, ok := s.fields[key]
	if !ok {
		return nil, false
	}
	return []string{val}, true
}
func (s *mockSecret) Set(key string, value interface{}) error {
	s.fields[key] = value.(string)
	return nil
}
func (s *mockSecret) Add(key string, value interface{}) error {
	return s.Set(key, value)
}
func (s *mockSecret) Del(key string) bool {
	_, exists := s.fields[key]
	delete(s.fields, key)
	return exists
}
func (s *mockSecret) Body() string { return "" }
func (s *mockSecret) Bytes() []byte {
	result := s.password + "\n"
	for k, v := range s.fields {
		result += k + ": " + v + "\n"
	}
	return []byte(result)
}

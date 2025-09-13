package did

import (
	"testing"
)

func TestParseDID(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		expected  *DID
		expectErr bool
	}{
		{
			name:  "valid did:key DID",
			input: "did:key:z6MkhaXgBZDvotDkL5257faiztiGiC2QtKLGpbnnEGta2doK",
			expected: &DID{
				Method:     "key",
				Identifier: "z6MkhaXgBZDvotDkL5257faiztiGiC2QtKLGpbnnEGta2doK",
			},
			expectErr: false,
		},
		{
			name:  "valid did:web DID",
			input: "did:web:example.com",
			expected: &DID{
				Method:     "web",
				Identifier: "example.com",
			},
			expectErr: false,
		},
		{
			name:  "DID with fragment",
			input: "did:key:z6MkhaXgBZDvotDkL5257faiztiGiC2QtKLGpbnnEGta2doK#keys-1",
			expected: &DID{
				Method:     "key",
				Identifier: "z6MkhaXgBZDvotDkL5257faiztiGiC2QtKLGpbnnEGta2doK",
				Fragment:   "keys-1",
			},
			expectErr: false,
		},
		{
			name:  "DID with path",
			input: "did:web:example.com/path/to/resource",
			expected: &DID{
				Method:     "web",
				Identifier: "example.com",
				Path:       "path/to/resource",
			},
			expectErr: false,
		},
		{
			name:  "DID with query",
			input: "did:web:example.com?service=agent",
			expected: &DID{
				Method:     "web",
				Identifier: "example.com",
				Query:      "service=agent",
			},
			expectErr: false,
		},
		{
			name:  "DID with path, query, and fragment",
			input: "did:web:example.com/path?service=agent#key-1",
			expected: &DID{
				Method:     "web",
				Identifier: "example.com",
				Path:       "path",
				Query:      "service=agent",
				Fragment:   "key-1",
			},
			expectErr: false,
		},
		{
			name:      "empty string",
			input:     "",
			expected:  nil,
			expectErr: true,
		},
		{
			name:      "invalid syntax - no method",
			input:     "did:",
			expected:  nil,
			expectErr: true,
		},
		{
			name:      "invalid syntax - no identifier",
			input:     "did:key:",
			expected:  nil,
			expectErr: true,
		},
		{
			name:      "invalid method name - uppercase",
			input:     "did:KEY:z6MkhaXgBZDvotDkL5257faiztiGiC2QtKLGpbnnEGta2doK",
			expected:  nil,
			expectErr: true,
		},
		{
			name:      "not a DID - missing did prefix",
			input:     "key:z6MkhaXgBZDvotDkL5257faiztiGiC2QtKLGpbnnEGta2doK",
			expected:  nil,
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParseDID(tt.input)
			
			if tt.expectErr {
				if err == nil {
					t.Errorf("expected error but got none")
				}
				return
			}
			
			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}
			
			if result == nil {
				t.Errorf("expected result but got nil")
				return
			}
			
			if result.Method != tt.expected.Method {
				t.Errorf("expected method %s, got %s", tt.expected.Method, result.Method)
			}
			
			if result.Identifier != tt.expected.Identifier {
				t.Errorf("expected identifier %s, got %s", tt.expected.Identifier, result.Identifier)
			}
			
			if result.Path != tt.expected.Path {
				t.Errorf("expected path %s, got %s", tt.expected.Path, result.Path)
			}
			
			if result.Query != tt.expected.Query {
				t.Errorf("expected query %s, got %s", tt.expected.Query, result.Query)
			}
			
			if result.Fragment != tt.expected.Fragment {
				t.Errorf("expected fragment %s, got %s", tt.expected.Fragment, result.Fragment)
			}
		})
	}
}

func TestDIDString(t *testing.T) {
	tests := []struct {
		name     string
		did      *DID
		expected string
	}{
		{
			name: "basic DID",
			did: &DID{
				Method:     "key",
				Identifier: "z6MkhaXgBZDvotDkL5257faiztiGiC2QtKLGpbnnEGta2doK",
			},
			expected: "did:key:z6MkhaXgBZDvotDkL5257faiztiGiC2QtKLGpbnnEGta2doK",
		},
		{
			name: "DID with fragment",
			did: &DID{
				Method:     "key",
				Identifier: "z6MkhaXgBZDvotDkL5257faiztiGiC2QtKLGpbnnEGta2doK",
				Fragment:   "keys-1",
			},
			expected: "did:key:z6MkhaXgBZDvotDkL5257faiztiGiC2QtKLGpbnnEGta2doK#keys-1",
		},
		{
			name: "DID with all components",
			did: &DID{
				Method:     "web",
				Identifier: "example.com",
				Path:       "path/to/resource",
				Query:      "service=agent",
				Fragment:   "key-1",
			},
			expected: "did:web:example.com/path/to/resource?service=agent#key-1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.did.String()
			if result != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, result)
			}
		})
	}
}

func TestIsValidDID(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{
			name:     "valid did:key",
			input:    "did:key:z6MkhaXgBZDvotDkL5257faiztiGiC2QtKLGpbnnEGta2doK",
			expected: true,
		},
		{
			name:     "valid did:web",
			input:    "did:web:example.com",
			expected: true,
		},
		{
			name:     "invalid - empty",
			input:    "",
			expected: false,
		},
		{
			name:     "invalid - no method",
			input:    "did:",
			expected: false,
		},
		{
			name:     "invalid - uppercase method",
			input:    "did:KEY:z6MkhaXgBZDvotDkL5257faiztiGiC2QtKLGpbnnEGta2doK",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsValidDID(tt.input)
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestDIDEquals(t *testing.T) {
	did1 := &DID{
		Method:     "key",
		Identifier: "z6MkhaXgBZDvotDkL5257faiztiGiC2QtKLGpbnnEGta2doK",
		Fragment:   "keys-1",
	}

	did2 := &DID{
		Method:     "key",
		Identifier: "z6MkhaXgBZDvotDkL5257faiztiGiC2QtKLGpbnnEGta2doK",
		Fragment:   "keys-1",
	}

	did3 := &DID{
		Method:     "key",
		Identifier: "z6MkhaXgBZDvotDkL5257faiztiGiC2QtKLGpbnnEGta2doK",
		Fragment:   "keys-2",
	}

	if !did1.Equals(did2) {
		t.Error("expected did1 to equal did2")
	}

	if did1.Equals(did3) {
		t.Error("expected did1 not to equal did3")
	}

	if did1.Equals(nil) {
		t.Error("expected did1 not to equal nil")
	}
}

func TestDIDCoreEquals(t *testing.T) {
	did1 := &DID{
		Method:     "key",
		Identifier: "z6MkhaXgBZDvotDkL5257faiztiGiC2QtKLGpbnnEGta2doK",
		Fragment:   "keys-1",
	}

	did2 := &DID{
		Method:     "key",
		Identifier: "z6MkhaXgBZDvotDkL5257faiztiGiC2QtKLGpbnnEGta2doK",
		Fragment:   "keys-2",
	}

	did3 := &DID{
		Method:     "web",
		Identifier: "example.com",
	}

	if !did1.CoreEquals(did2) {
		t.Error("expected did1 core to equal did2 core")
	}

	if did1.CoreEquals(did3) {
		t.Error("expected did1 core not to equal did3 core")
	}
}

func TestDIDQueryParams(t *testing.T) {
	did := &DID{
		Method:     "web",
		Identifier: "example.com",
		Query:      "service=agent&version=1.0",
	}

	// Test getting existing parameter
	service := did.GetQueryParam("service")
	if service != "agent" {
		t.Errorf("expected service=agent, got service=%s", service)
	}

	version := did.GetQueryParam("version")
	if version != "1.0" {
		t.Errorf("expected version=1.0, got version=%s", version)
	}

	// Test getting non-existent parameter
	missing := did.GetQueryParam("missing")
	if missing != "" {
		t.Errorf("expected empty string for missing parameter, got %s", missing)
	}

	// Test setting parameter
	newDID := did.SetQueryParam("new", "value")
	newValue := newDID.GetQueryParam("new")
	if newValue != "value" {
		t.Errorf("expected new=value, got new=%s", newValue)
	}

	// Original DID should be unchanged
	if did.GetQueryParam("new") != "" {
		t.Error("original DID was modified")
	}
}

func TestNormalizeDID(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
		wantErr  bool
	}{
		{
			name:     "DID with fragment",
			input:    "did:key:z6MkhaXgBZDvotDkL5257faiztiGiC2QtKLGpbnnEGta2doK#keys-1",
			expected: "did:key:z6MkhaXgBZDvotDkL5257faiztiGiC2QtKLGpbnnEGta2doK",
			wantErr:  false,
		},
		{
			name:     "DID with path and query",
			input:    "did:web:example.com/path?service=agent",
			expected: "did:web:example.com",
			wantErr:  false,
		},
		{
			name:     "basic DID",
			input:    "did:key:z6MkhaXgBZDvotDkL5257faiztiGiC2QtKLGpbnnEGta2doK",
			expected: "did:key:z6MkhaXgBZDvotDkL5257faiztiGiC2QtKLGpbnnEGta2doK",
			wantErr:  false,
		},
		{
			name:     "invalid DID",
			input:    "not-a-did",
			expected: "",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := NormalizeDID(tt.input)
			
			if tt.wantErr {
				if err == nil {
					t.Error("expected error but got none")
				}
				return
			}
			
			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}
			
			if result != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, result)
			}
		})
	}
}
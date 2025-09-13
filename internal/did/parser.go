package did

import (
	"regexp"
	"strings"
)

// DID syntax according to W3C DID specification:
// did = "did:" method-name ":" method-specific-id
// method-name = 1*method-char
// method-char = %x61-7A / DIGIT ; a-z / 0-9
// method-specific-id = *( *idchar ":" ) 1*idchar
// idchar = ALPHA / DIGIT / "." / "-" / "_" / pct-encoded

var (
	// didRegex matches the DID syntax
	didRegex = regexp.MustCompile(`^did:([a-z0-9]+):([a-zA-Z0-9._%-]+)(?:/([^?#]*))?(?:\?([^#]*))?(?:#(.*))?$`)
	
	// methodNameRegex validates method names
	methodNameRegex = regexp.MustCompile(`^[a-z0-9]+$`)
)

// ParseDID parses a DID string into a DID struct
func ParseDID(didString string) (*DID, error) {
	if didString == "" {
		return nil, NewDIDError(ErrorInvalidDID, "DID string is empty")
	}
	
	matches := didRegex.FindStringSubmatch(didString)
	if matches == nil {
		return nil, NewDIDError(ErrorInvalidDID, "invalid DID syntax: "+didString)
	}
	
	method := matches[1]
	identifier := matches[2]
	path := matches[3]
	query := matches[4]
	fragment := matches[5]
	
	// Validate method name
	if !methodNameRegex.MatchString(method) {
		return nil, NewDIDError(ErrorInvalidDID, "invalid method name: "+method)
	}
	
	// Validate identifier is not empty
	if identifier == "" {
		return nil, NewDIDError(ErrorInvalidDID, "method-specific identifier is empty")
	}
	
	return &DID{
		Method:     method,
		Identifier: identifier,
		Path:       path,
		Query:      query,
		Fragment:   fragment,
	}, nil
}

// IsValidDID checks if a string is a valid DID
func IsValidDID(didString string) bool {
	_, err := ParseDID(didString)
	return err == nil
}

// IsValidMethod checks if a method name is valid
func IsValidMethod(method string) bool {
	return methodNameRegex.MatchString(method)
}

// NormalizeDID normalizes a DID string by removing unnecessary components
func NormalizeDID(didString string) (string, error) {
	parsed, err := ParseDID(didString)
	if err != nil {
		return "", err
	}
	
	// Return just the core DID without path, query, or fragment
	return "did:" + parsed.Method + ":" + parsed.Identifier, nil
}

// ExtractMethod extracts the method from a DID string
func ExtractMethod(didString string) (string, error) {
	parsed, err := ParseDID(didString)
	if err != nil {
		return "", err
	}
	
	return parsed.Method, nil
}

// ExtractIdentifier extracts the method-specific identifier from a DID string
func ExtractIdentifier(didString string) (string, error) {
	parsed, err := ParseDID(didString)
	if err != nil {
		return "", err
	}
	
	return parsed.Identifier, nil
}

// JoinDIDComponents creates a DID string from components
func JoinDIDComponents(method, identifier string) string {
	return "did:" + method + ":" + identifier
}

// SplitDIDComponents splits a DID string into its components
func SplitDIDComponents(didString string) (method, identifier string, err error) {
	parsed, err := ParseDID(didString)
	if err != nil {
		return "", "", err
	}
	
	return parsed.Method, parsed.Identifier, nil
}

// AddFragment adds a fragment to a DID
func (d *DID) AddFragment(fragment string) *DID {
	newDID := *d
	newDID.Fragment = fragment
	return &newDID
}

// AddPath adds a path to a DID
func (d *DID) AddPath(path string) *DID {
	newDID := *d
	newDID.Path = path
	return &newDID
}

// AddQuery adds a query to a DID
func (d *DID) AddQuery(query string) *DID {
	newDID := *d
	newDID.Query = query
	return &newDID
}

// WithoutFragment returns a copy of the DID without the fragment
func (d *DID) WithoutFragment() *DID {
	newDID := *d
	newDID.Fragment = ""
	return &newDID
}

// WithoutPath returns a copy of the DID without the path
func (d *DID) WithoutPath() *DID {
	newDID := *d
	newDID.Path = ""
	return &newDID
}

// WithoutQuery returns a copy of the DID without the query
func (d *DID) WithoutQuery() *DID {
	newDID := *d
	newDID.Query = ""
	return &newDID
}

// Core returns the core DID (without path, query, and fragment)
func (d *DID) Core() *DID {
	return &DID{
		Method:     d.Method,
		Identifier: d.Identifier,
	}
}

// Equals checks if two DIDs are equal
func (d *DID) Equals(other *DID) bool {
	if other == nil {
		return false
	}
	
	return d.Method == other.Method &&
		d.Identifier == other.Identifier &&
		d.Path == other.Path &&
		d.Query == other.Query &&
		d.Fragment == other.Fragment
}

// CoreEquals checks if two DIDs have the same core (method and identifier)
func (d *DID) CoreEquals(other *DID) bool {
	if other == nil {
		return false
	}
	
	return d.Method == other.Method && d.Identifier == other.Identifier
}

// IsFragment checks if this DID represents a fragment reference
func (d *DID) IsFragment() bool {
	return d.Fragment != ""
}

// HasPath checks if this DID has a path component
func (d *DID) HasPath() bool {
	return d.Path != ""
}

// HasQuery checks if this DID has a query component
func (d *DID) HasQuery() bool {
	return d.Query != ""
}

// GetQueryParam gets a specific query parameter from the DID
func (d *DID) GetQueryParam(param string) string {
	if d.Query == "" {
		return ""
	}
	
	// Simple query parameter parsing
	params := strings.Split(d.Query, "&")
	for _, p := range params {
		parts := strings.SplitN(p, "=", 2)
		if len(parts) == 2 && parts[0] == param {
			return parts[1]
		}
	}
	
	return ""
}

// SetQueryParam sets a query parameter on the DID
func (d *DID) SetQueryParam(param, value string) *DID {
	newDID := *d
	
	if newDID.Query == "" {
		newDID.Query = param + "=" + value
	} else {
		// Replace existing parameter or add new one
		params := strings.Split(newDID.Query, "&")
		found := false
		
		for i, p := range params {
			parts := strings.SplitN(p, "=", 2)
			if len(parts) >= 1 && parts[0] == param {
				params[i] = param + "=" + value
				found = true
				break
			}
		}
		
		if !found {
			params = append(params, param+"="+value)
		}
		
		newDID.Query = strings.Join(params, "&")
	}
	
	return &newDID
}
package did

import (
	"context"
	"strings"
)

// DefaultDocumentHelper provides utility methods for working with DID documents
type DefaultDocumentHelper struct{}

// NewDocumentHelper creates a new document helper
func NewDocumentHelper() DocumentHelper {
	return &DefaultDocumentHelper{}
}

// GetVerificationMethod retrieves a verification method by ID
func (h *DefaultDocumentHelper) GetVerificationMethod(document *DIDDocument, methodID string) (*VerificationMethod, error) {
	if document == nil {
		return nil, NewDIDError(ErrorInvalidDocument, "document is nil")
	}
	
	// Check if methodID is relative (fragment only) and resolve it
	resolvedID := methodID
	if strings.HasPrefix(methodID, "#") {
		resolvedID = document.ID + methodID
	}
	
	// Search in verification methods
	for i := range document.VerificationMethod {
		if document.VerificationMethod[i].ID == resolvedID {
			return &document.VerificationMethod[i], nil
		}
	}
	
	return nil, NewDIDError(ErrorNotFound, "verification method not found: "+methodID)
}

// GetVerificationMethodsForPurpose gets all verification methods for a specific purpose
func (h *DefaultDocumentHelper) GetVerificationMethodsForPurpose(document *DIDDocument, purpose VerificationRelationship) ([]*VerificationMethod, error) {
	if document == nil {
		return nil, NewDIDError(ErrorInvalidDocument, "document is nil")
	}
	
	var methods []*VerificationMethod
	var references []interface{}
	
	// Get the appropriate references array based on purpose
	switch purpose {
	case Authentication:
		references = document.Authentication
	case AssertionMethod:
		references = document.AssertionMethod
	case KeyAgreement:
		references = document.KeyAgreement
	case CapabilityInvocation:
		references = document.CapabilityInvocation
	case CapabilityDelegation:
		references = document.CapabilityDelegation
	default:
		return nil, NewDIDError(ErrorInvalidDocument, "invalid verification relationship: "+string(purpose))
	}
	
	// Process each reference
	for _, ref := range references {
		switch r := ref.(type) {
		case string:
			// Reference by ID - find the method
			method, err := h.GetVerificationMethod(document, r)
			if err == nil {
				methods = append(methods, method)
			}
			
		case map[string]interface{}:
			// Embedded verification method - convert to struct
			method, err := h.mapToVerificationMethod(r)
			if err == nil {
				methods = append(methods, method)
			}
			
		case VerificationMethod:
			// Direct verification method
			methods = append(methods, &r)
		}
	}
	
	return methods, nil
}

// GetService retrieves a service by ID
func (h *DefaultDocumentHelper) GetService(document *DIDDocument, serviceID string) (*Service, error) {
	if document == nil {
		return nil, NewDIDError(ErrorInvalidDocument, "document is nil")
	}
	
	// Check if serviceID is relative (fragment only) and resolve it
	resolvedID := serviceID
	if strings.HasPrefix(serviceID, "#") {
		resolvedID = document.ID + serviceID
	}
	
	// Search in services
	for i := range document.Service {
		if document.Service[i].ID == resolvedID {
			return &document.Service[i], nil
		}
	}
	
	return nil, NewDIDError(ErrorNotFound, "service not found: "+serviceID)
}

// GetServicesByType gets all services of a specific type
func (h *DefaultDocumentHelper) GetServicesByType(document *DIDDocument, serviceType string) ([]*Service, error) {
	if document == nil {
		return nil, NewDIDError(ErrorInvalidDocument, "document is nil")
	}
	
	var services []*Service
	
	for i := range document.Service {
		if document.Service[i].Type == serviceType {
			services = append(services, &document.Service[i])
		}
	}
	
	return services, nil
}

// AddVerificationMethod adds a verification method to the document
func (h *DefaultDocumentHelper) AddVerificationMethod(document *DIDDocument, method *VerificationMethod, purposes []VerificationRelationship) error {
	if document == nil {
		return NewDIDError(ErrorInvalidDocument, "document is nil")
	}
	
	if method == nil {
		return NewDIDError(ErrorInvalidKey, "verification method is nil")
	}
	
	// Validate the verification method
	if err := h.ValidateVerificationMethod(method); err != nil {
		return err
	}
	
	// Check if method already exists
	for i := range document.VerificationMethod {
		if document.VerificationMethod[i].ID == method.ID {
			return NewDIDError(ErrorInvalidDocument, "verification method already exists: "+method.ID)
		}
	}
	
	// Add to verification methods
	document.VerificationMethod = append(document.VerificationMethod, *method)
	
	// Add to purposes
	for _, purpose := range purposes {
		if err := h.addMethodToPurpose(document, method.ID, purpose); err != nil {
			return err
		}
	}
	
	return nil
}

// RemoveVerificationMethod removes a verification method from the document
func (h *DefaultDocumentHelper) RemoveVerificationMethod(document *DIDDocument, methodID string) error {
	if document == nil {
		return NewDIDError(ErrorInvalidDocument, "document is nil")
	}
	
	// Remove from verification methods
	found := false
	for i := range document.VerificationMethod {
		if document.VerificationMethod[i].ID == methodID {
			document.VerificationMethod = append(document.VerificationMethod[:i], document.VerificationMethod[i+1:]...)
			found = true
			break
		}
	}
	
	if !found {
		return NewDIDError(ErrorNotFound, "verification method not found: "+methodID)
	}
	
	// Remove from all purposes
	h.removeMethodFromPurpose(&document.Authentication, methodID)
	h.removeMethodFromPurpose(&document.AssertionMethod, methodID)
	h.removeMethodFromPurpose(&document.KeyAgreement, methodID)
	h.removeMethodFromPurpose(&document.CapabilityInvocation, methodID)
	h.removeMethodFromPurpose(&document.CapabilityDelegation, methodID)
	
	return nil
}

// AddService adds a service to the document
func (h *DefaultDocumentHelper) AddService(document *DIDDocument, service *Service) error {
	if document == nil {
		return NewDIDError(ErrorInvalidDocument, "document is nil")
	}
	
	if service == nil {
		return NewDIDError(ErrorInvalidDocument, "service is nil")
	}
	
	// Validate the service
	if err := h.ValidateService(service); err != nil {
		return err
	}
	
	// Check if service already exists
	for i := range document.Service {
		if document.Service[i].ID == service.ID {
			return NewDIDError(ErrorInvalidDocument, "service already exists: "+service.ID)
		}
	}
	
	// Add to services
	document.Service = append(document.Service, *service)
	
	return nil
}

// RemoveService removes a service from the document
func (h *DefaultDocumentHelper) RemoveService(document *DIDDocument, serviceID string) error {
	if document == nil {
		return NewDIDError(ErrorInvalidDocument, "document is nil")
	}
	
	// Remove from services
	for i := range document.Service {
		if document.Service[i].ID == serviceID {
			document.Service = append(document.Service[:i], document.Service[i+1:]...)
			return nil
		}
	}
	
	return NewDIDError(ErrorNotFound, "service not found: "+serviceID)
}

// IsDeactivated checks if a DID document is deactivated
func (h *DefaultDocumentHelper) IsDeactivated(document *DIDDocument) bool {
	if document == nil {
		return false
	}
	
	return document.Deactivated != nil && *document.Deactivated
}

// ValidateVerificationMethod validates a verification method
func (h *DefaultDocumentHelper) ValidateVerificationMethod(method *VerificationMethod) error {
	if method.ID == "" {
		return NewDIDError(ErrorInvalidKey, "verification method ID is required")
	}
	
	if method.Type == "" {
		return NewDIDError(ErrorInvalidKey, "verification method type is required")
	}
	
	if method.Controller == "" {
		return NewDIDError(ErrorInvalidKey, "verification method controller is required")
	}
	
	// Check that at least one key representation is present
	if method.PublicKeyMultibase == nil &&
		method.PublicKeyJwk == nil &&
		method.PublicKeyBase58 == nil &&
		method.PublicKeyBase64 == nil &&
		method.PublicKeyHex == nil &&
		method.BlockchainAccountId == nil {
		return NewDIDError(ErrorInvalidKey, "verification method must have at least one key representation")
	}
	
	return nil
}

// ValidateService validates a service
func (h *DefaultDocumentHelper) ValidateService(service *Service) error {
	if service.ID == "" {
		return NewDIDError(ErrorInvalidDocument, "service ID is required")
	}
	
	if service.Type == "" {
		return NewDIDError(ErrorInvalidDocument, "service type is required")
	}
	
	if service.ServiceEndpoint == nil {
		return NewDIDError(ErrorInvalidDocument, "service endpoint is required")
	}
	
	return nil
}

// ValidateContext validates the @context field
func (h *DefaultDocumentHelper) ValidateContext(context []string) error {
	if len(context) == 0 {
		return NewDIDError(ErrorInvalidDocument, "@context is required")
	}
	
	// First context must be the DID v1 context
	if context[0] != "https://www.w3.org/ns/did/v1" {
		return NewDIDError(ErrorInvalidDocument, "first @context must be https://www.w3.org/ns/did/v1")
	}
	
	return nil
}

// addMethodToPurpose adds a method ID to a verification relationship
func (h *DefaultDocumentHelper) addMethodToPurpose(document *DIDDocument, methodID string, purpose VerificationRelationship) error {
	switch purpose {
	case Authentication:
		document.Authentication = append(document.Authentication, methodID)
	case AssertionMethod:
		document.AssertionMethod = append(document.AssertionMethod, methodID)
	case KeyAgreement:
		document.KeyAgreement = append(document.KeyAgreement, methodID)
	case CapabilityInvocation:
		document.CapabilityInvocation = append(document.CapabilityInvocation, methodID)
	case CapabilityDelegation:
		document.CapabilityDelegation = append(document.CapabilityDelegation, methodID)
	default:
		return NewDIDError(ErrorInvalidDocument, "invalid verification relationship: "+string(purpose))
	}
	
	return nil
}

// removeMethodFromPurpose removes a method ID from a verification relationship array
func (h *DefaultDocumentHelper) removeMethodFromPurpose(purposes *[]interface{}, methodID string) {
	for i := 0; i < len(*purposes); i++ {
		switch ref := (*purposes)[i].(type) {
		case string:
			if ref == methodID {
				*purposes = append((*purposes)[:i], (*purposes)[i+1:]...)
				i-- // Adjust index after removal
			}
		case map[string]interface{}:
			if id, exists := ref["id"]; exists && id == methodID {
				*purposes = append((*purposes)[:i], (*purposes)[i+1:]...)
				i-- // Adjust index after removal
			}
		case VerificationMethod:
			if ref.ID == methodID {
				*purposes = append((*purposes)[:i], (*purposes)[i+1:]...)
				i-- // Adjust index after removal
			}
		}
	}
}

// mapToVerificationMethod converts a map to a VerificationMethod struct
func (h *DefaultDocumentHelper) mapToVerificationMethod(m map[string]interface{}) (*VerificationMethod, error) {
	method := &VerificationMethod{}
	
	if id, ok := m["id"].(string); ok {
		method.ID = id
	}
	
	if typ, ok := m["type"].(string); ok {
		method.Type = typ
	}
	
	if controller, ok := m["controller"].(string); ok {
		method.Controller = controller
	}
	
	if pkm, ok := m["publicKeyMultibase"].(string); ok {
		method.PublicKeyMultibase = &pkm
	}
	
	if pkb58, ok := m["publicKeyBase58"].(string); ok {
		method.PublicKeyBase58 = &pkb58
	}
	
	// Add other fields as needed
	
	return method, nil
}

// DefaultDocumentValidator provides document validation
type DefaultDocumentValidator struct {
	helper DocumentHelper
}

// NewDocumentValidator creates a new document validator
func NewDocumentValidator() DocumentValidator {
	return &DefaultDocumentValidator{
		helper: NewDocumentHelper(),
	}
}

// Validate validates a DID document
func (v *DefaultDocumentValidator) Validate(ctx context.Context, document *DIDDocument) error {
	if document == nil {
		return NewDIDError(ErrorInvalidDocument, "document is nil")
	}
	
	// Validate @context
	if err := v.ValidateContext(document.Context); err != nil {
		return err
	}
	
	// Validate ID
	if document.ID == "" {
		return NewDIDError(ErrorInvalidDocument, "document ID is required")
	}
	
	if !IsValidDID(document.ID) {
		return NewDIDError(ErrorInvalidDID, "document ID is not a valid DID")
	}
	
	// Validate verification methods
	for i := range document.VerificationMethod {
		if err := v.ValidateVerificationMethod(&document.VerificationMethod[i]); err != nil {
			return err
		}
	}
	
	// Validate services
	for i := range document.Service {
		if err := v.ValidateService(&document.Service[i]); err != nil {
			return err
		}
	}
	
	return nil
}

// ValidateVerificationMethod validates a verification method
func (v *DefaultDocumentValidator) ValidateVerificationMethod(method *VerificationMethod) error {
	return v.helper.(*DefaultDocumentHelper).ValidateVerificationMethod(method)
}

// ValidateService validates a service
func (v *DefaultDocumentValidator) ValidateService(service *Service) error {
	return v.helper.(*DefaultDocumentHelper).ValidateService(service)
}

// ValidateContext validates the @context field
func (v *DefaultDocumentValidator) ValidateContext(context []string) error {
	return v.helper.(*DefaultDocumentHelper).ValidateContext(context)
}
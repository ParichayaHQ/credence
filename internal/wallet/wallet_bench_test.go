package wallet

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/ParichayaHQ/credence/internal/did"
	"github.com/ParichayaHQ/credence/internal/vc"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func BenchmarkWallet_GenerateKey(b *testing.B) {
	wallet := setupBenchmarkWallet(b)
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := wallet.GenerateKey(did.KeyTypeEd25519)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkWallet_CreateDID(b *testing.B) {
	wallet := setupBenchmarkWallet(b)
	
	// Pre-generate keys
	keys := make([]*KeyPair, b.N)
	for i := 0; i < b.N; i++ {
		key, err := wallet.GenerateKey(did.KeyTypeEd25519)
		if err != nil {
			b.Fatal(err)
		}
		keys[i] = key
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := wallet.CreateDID(keys[i].ID, "key")
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkWallet_StoreCredential(b *testing.B) {
	wallet := setupBenchmarkWallet(b)
	
	// Pre-generate credentials
	credentials := make([]*vc.VerifiableCredential, b.N)
	for i := 0; i < b.N; i++ {
		credentials[i] = &vc.VerifiableCredential{
			Context:      []string{"https://www.w3.org/2018/credentials/v1"},
			ID:           fmt.Sprintf("cred-%d", i),
			Type:         []string{"VerifiableCredential"},
			Issuer:       fmt.Sprintf("did:key:issuer-%d", i),
			IssuanceDate: time.Now().Format(time.RFC3339),
			CredentialSubject: map[string]interface{}{
				"id":   fmt.Sprintf("did:key:subject-%d", i),
				"name": fmt.Sprintf("Subject %d", i),
			},
		}
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := wallet.StoreCredential(credentials[i])
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkWallet_GetCredential(b *testing.B) {
	wallet := setupBenchmarkWallet(b)
	
	// Pre-store credentials
	credentialIDs := make([]string, 1000)
	for i := 0; i < 1000; i++ {
		cred := &vc.VerifiableCredential{
			Context:      []string{"https://www.w3.org/2018/credentials/v1"},
			ID:           fmt.Sprintf("cred-%d", i),
			Type:         []string{"VerifiableCredential"},
			Issuer:       fmt.Sprintf("did:key:issuer-%d", i),
			IssuanceDate: time.Now().Format(time.RFC3339),
			CredentialSubject: map[string]interface{}{
				"id": "did:key:subject",
			},
		}
		record, err := wallet.StoreCredential(cred)
		if err != nil {
			b.Fatal(err)
		}
		credentialIDs[i] = record.ID
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := wallet.GetCredential(credentialIDs[i%1000])
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkWallet_ListCredentials(b *testing.B) {
	wallet := setupBenchmarkWallet(b)
	
	// Pre-store credentials with different types and issuers
	for i := 0; i < 1000; i++ {
		credType := "TestCredential"
		if i%3 == 0 {
			credType = "AnotherCredential"
		}
		
		cred := &vc.VerifiableCredential{
			Context:      []string{"https://www.w3.org/2018/credentials/v1"},
			ID:           fmt.Sprintf("cred-%d", i),
			Type:         []string{"VerifiableCredential", credType},
			Issuer:       fmt.Sprintf("did:key:issuer-%d", i%10), // 10 different issuers
			IssuanceDate: time.Now().Format(time.RFC3339),
			CredentialSubject: map[string]interface{}{
				"id": fmt.Sprintf("did:key:subject-%d", i),
			},
		}
		_, err := wallet.StoreCredential(cred)
		if err != nil {
			b.Fatal(err)
		}
	}
	
	// Benchmark different list operations
	b.Run("ListAll", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, err := wallet.ListCredentials(nil)
			if err != nil {
				b.Fatal(err)
			}
		}
	})
	
	b.Run("ListByType", func(b *testing.B) {
		filter := &CredentialFilter{
			Type: []string{"TestCredential"},
		}
		for i := 0; i < b.N; i++ {
			_, err := wallet.ListCredentials(filter)
			if err != nil {
				b.Fatal(err)
			}
		}
	})
	
	b.Run("ListByIssuer", func(b *testing.B) {
		filter := &CredentialFilter{
			Issuer: "did:key:issuer-0",
		}
		for i := 0; i < b.N; i++ {
			_, err := wallet.ListCredentials(filter)
			if err != nil {
				b.Fatal(err)
			}
		}
	})
	
	b.Run("ListWithLimit", func(b *testing.B) {
		filter := &CredentialFilter{
			Limit: 10,
		}
		for i := 0; i < b.N; i++ {
			_, err := wallet.ListCredentials(filter)
			if err != nil {
				b.Fatal(err)
			}
		}
	})
}

func BenchmarkWallet_CreatePresentation(b *testing.B) {
	wallet := setupBenchmarkWallet(b)
	
	// Pre-store credentials and keys
	credentialIDs := make([]string, 100)
	for i := 0; i < 100; i++ {
		cred := &vc.VerifiableCredential{
			Context:      []string{"https://www.w3.org/2018/credentials/v1"},
			ID:           fmt.Sprintf("cred-%d", i),
			Type:         []string{"VerifiableCredential"},
			Issuer:       "did:key:issuer",
			IssuanceDate: time.Now().Format(time.RFC3339),
			CredentialSubject: map[string]interface{}{
				"id": "did:key:subject",
			},
		}
		record, err := wallet.StoreCredential(cred)
		if err != nil {
			b.Fatal(err)
		}
		credentialIDs[i] = record.ID
	}
	
	key, err := wallet.GenerateKey(did.KeyTypeEd25519)
	if err != nil {
		b.Fatal(err)
	}
	
	options := &PresentationOptions{
		Holder:    "did:key:holder",
		KeyID:     key.ID,
		Algorithm: "EdDSA",
		Challenge: "benchmark-challenge",
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Use 3 random credentials for each presentation
		selectedCreds := []string{
			credentialIDs[i%100],
			credentialIDs[(i+1)%100],
			credentialIDs[(i+2)%100],
		}
		
		_, err := wallet.CreatePresentation(selectedCreds, options)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkWallet_LockUnlock(b *testing.B) {
	wallet := setupBenchmarkWallet(b)
	password := "benchmark-password"
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err := wallet.Lock(password)
		if err != nil {
			b.Fatal(err)
		}
		
		err = wallet.Unlock(password)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkWallet_Export(b *testing.B) {
	wallet := setupBenchmarkWallet(b)
	
	// Add some data to export
	for i := 0; i < 10; i++ {
		key, err := wallet.GenerateKey(did.KeyTypeEd25519)
		if err != nil {
			b.Fatal(err)
		}
		
		_, err = wallet.CreateDID(key.ID, "key")
		if err != nil {
			b.Fatal(err)
		}
		
		cred := &vc.VerifiableCredential{
			Context:      []string{"https://www.w3.org/2018/credentials/v1"},
			Type:         []string{"VerifiableCredential"},
			Issuer:       fmt.Sprintf("did:key:issuer-%d", i),
			IssuanceDate: time.Now().Format(time.RFC3339),
			CredentialSubject: map[string]interface{}{
				"id": "did:key:subject",
			},
		}
		_, err = wallet.StoreCredential(cred)
		if err != nil {
			b.Fatal(err)
		}
	}
	
	password := "export-password"
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := wallet.Export(password)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkWallet_Import(b *testing.B) {
	// Setup source wallet with data
	sourceWallet := setupBenchmarkWallet(b)
	
	for i := 0; i < 10; i++ {
		key, err := sourceWallet.GenerateKey(did.KeyTypeEd25519)
		if err != nil {
			b.Fatal(err)
		}
		
		_, err = sourceWallet.CreateDID(key.ID, "key")
		if err != nil {
			b.Fatal(err)
		}
		
		cred := &vc.VerifiableCredential{
			Context:      []string{"https://www.w3.org/2018/credentials/v1"},
			Type:         []string{"VerifiableCredential"},
			Issuer:       fmt.Sprintf("did:key:issuer-%d", i),
			IssuanceDate: time.Now().Format(time.RFC3339),
			CredentialSubject: map[string]interface{}{
				"id": "did:key:subject",
			},
		}
		_, err = sourceWallet.StoreCredential(cred)
		if err != nil {
			b.Fatal(err)
		}
	}
	
	password := "import-password"
	exportData, err := sourceWallet.Export(password)
	if err != nil {
		b.Fatal(err)
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		targetWallet := setupBenchmarkWallet(b)
		err := targetWallet.Import(exportData, password)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkPresentationService_CreatePresentation(b *testing.B) {
	wallet := setupBenchmarkWallet(b)
	resolver := &MockDIDResolver{}
	verifier := &MockCredentialVerifier{}
	
	ps := NewPresentationService(wallet, verifier, resolver)
	
	// Pre-store credentials
	credentialIDs := make([]string, 100)
	for i := 0; i < 100; i++ {
		cred := &vc.VerifiableCredential{
			Context:      []string{"https://www.w3.org/2018/credentials/v1"},
			ID:           fmt.Sprintf("cred-%d", i),
			Type:         []string{"VerifiableCredential"},
			Issuer:       "did:key:issuer",
			IssuanceDate: time.Now().Format(time.RFC3339),
			CredentialSubject: map[string]interface{}{
				"id":   fmt.Sprintf("did:key:subject-%d", i),
				"name": fmt.Sprintf("Subject %d", i),
				"age":  25 + i,
			},
		}
		record, err := wallet.StoreCredential(cred)
		if err != nil {
			b.Fatal(err)
		}
		credentialIDs[i] = record.ID
	}
	
	key, err := wallet.GenerateKey(did.KeyTypeEd25519)
	if err != nil {
		b.Fatal(err)
	}
	
	request := &PresentationRequest{
		CredentialIDs: []string{credentialIDs[0]},
		Holder:        "did:key:holder",
		KeyID:         key.ID,
		Algorithm:     "EdDSA",
		Challenge:     "benchmark-challenge",
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		request.CredentialIDs = []string{credentialIDs[i%100]}
		_, err := ps.CreatePresentation(context.Background(), request)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkPresentationService_CreatePresentationWithSelectiveDisclosure(b *testing.B) {
	wallet := setupBenchmarkWallet(b)
	resolver := &MockDIDResolver{}
	verifier := &MockCredentialVerifier{}
	
	ps := NewPresentationService(wallet, verifier, resolver)
	
	// Pre-store credentials with multiple fields
	credentialIDs := make([]string, 100)
	for i := 0; i < 100; i++ {
		cred := &vc.VerifiableCredential{
			Context:      []string{"https://www.w3.org/2018/credentials/v1"},
			ID:           fmt.Sprintf("cred-%d", i),
			Type:         []string{"VerifiableCredential"},
			Issuer:       "did:key:issuer",
			IssuanceDate: time.Now().Format(time.RFC3339),
			CredentialSubject: map[string]interface{}{
				"id":      fmt.Sprintf("did:key:subject-%d", i),
				"name":    fmt.Sprintf("Subject %d", i),
				"email":   fmt.Sprintf("subject%d@example.com", i),
				"age":     25 + i,
				"address": fmt.Sprintf("Address %d", i),
				"phone":   fmt.Sprintf("555-000-%04d", i),
			},
		}
		record, err := wallet.StoreCredential(cred)
		if err != nil {
			b.Fatal(err)
		}
		credentialIDs[i] = record.ID
	}
	
	key, err := wallet.GenerateKey(did.KeyTypeEd25519)
	if err != nil {
		b.Fatal(err)
	}
	
	request := &PresentationRequest{
		CredentialIDs: []string{credentialIDs[0]},
		Holder:        "did:key:holder",
		KeyID:         key.ID,
		Algorithm:     "EdDSA",
		Challenge:     "benchmark-challenge",
		SelectiveDisclosure: map[string][]string{
			"credential_0": {"name", "age"}, // Only include name and age
		},
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		request.CredentialIDs = []string{credentialIDs[i%100]}
		_, err := ps.CreatePresentation(context.Background(), request)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkInMemoryStorage_StoreCredential(b *testing.B) {
	storage := NewInMemoryStorage()
	
	// Pre-generate credentials
	credentials := make([]*CredentialRecord, b.N)
	for i := 0; i < b.N; i++ {
		credentials[i] = &CredentialRecord{
			ID: fmt.Sprintf("cred-%d", i),
			Credential: &vc.VerifiableCredential{
				Context:      []string{"https://www.w3.org/2018/credentials/v1"},
				ID:           fmt.Sprintf("cred-%d", i),
				Type:         []string{"VerifiableCredential"},
				Issuer:       fmt.Sprintf("did:key:issuer-%d", i),
				IssuanceDate: time.Now().Format(time.RFC3339),
				CredentialSubject: map[string]interface{}{
					"id": "did:key:subject",
				},
			},
			Created: time.Now(),
			Status:  CredentialStatusValid,
		}
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err := storage.StoreCredential(credentials[i])
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkInMemoryStorage_ListCredentials(b *testing.B) {
	storage := NewInMemoryStorage()
	
	// Pre-store credentials
	for i := 0; i < 1000; i++ {
		record := &CredentialRecord{
			ID: fmt.Sprintf("cred-%d", i),
			Credential: &vc.VerifiableCredential{
				Context:      []string{"https://www.w3.org/2018/credentials/v1"},
				ID:           fmt.Sprintf("cred-%d", i),
				Type:         []string{"VerifiableCredential", fmt.Sprintf("Type%d", i%5)},
				Issuer:       fmt.Sprintf("did:key:issuer-%d", i%10),
				IssuanceDate: time.Now().Format(time.RFC3339),
				CredentialSubject: map[string]interface{}{
					"id": "did:key:subject",
				},
			},
			Created: time.Now(),
			Status:  CredentialStatusValid,
		}
		err := storage.StoreCredential(record)
		if err != nil {
			b.Fatal(err)
		}
	}
	
	b.Run("ListAll", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, err := storage.ListCredentials(nil)
			if err != nil {
				b.Fatal(err)
			}
		}
	})
	
	b.Run("ListFiltered", func(b *testing.B) {
		filter := &CredentialFilter{
			Type: []string{"Type0"},
		}
		for i := 0; i < b.N; i++ {
			_, err := storage.ListCredentials(filter)
			if err != nil {
				b.Fatal(err)
			}
		}
	})
}

// Helper function to setup wallet for benchmarks
func setupBenchmarkWallet(b *testing.B) Wallet {
	config := DefaultWalletConfig()
	config.StorageType = "memory"
	config.AutoLockTimeout = 0 // Disable auto-lock for benchmarks
	config.EncryptionEnabled = false // Start unlocked for benchmarks
	
	storage := NewInMemoryStorage()
	keyManager := &MockKeyManager{}
	
	// Setup mock key manager
	keyManager.On("GenerateKey", mock.AnythingOfType("did.KeyType")).Return(func(keyType did.KeyType) interface{} {
		return fmt.Sprintf("key-%d", time.Now().UnixNano()) // Return a simple key identifier
	}, nil)
	
	keyManager.On("Sign", mock.AnythingOfType("string"), mock.AnythingOfType("[]uint8")).Return([]byte("signature"), nil)
	
	wallet, err := NewDefaultWallet(config, storage, keyManager)
	require.NoError(b, err)
	
	return wallet
}
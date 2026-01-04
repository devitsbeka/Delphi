package services

import (
	"context"
	"fmt"
	"time"

	"github.com/delphi-platform/delphi/backend/internal/models"
	"github.com/delphi-platform/delphi/backend/internal/providers"
	"github.com/delphi-platform/delphi/backend/internal/repository"
	"github.com/delphi-platform/delphi/backend/pkg/crypto"
	"github.com/delphi-platform/delphi/backend/pkg/logger"
	"github.com/google/uuid"
)

// APIKeyService handles API key operations (full implementation)
type APIKeyServiceImpl struct {
	repos     *repository.Repositories
	encryptor *crypto.Encryptor
	manager   *providers.Manager
	log       *logger.Logger
}

// NewAPIKeyServiceImpl creates a new API key service
func NewAPIKeyServiceImpl(repos *repository.Repositories, encryptor *crypto.Encryptor, manager *providers.Manager, log *logger.Logger) *APIKeyServiceImpl {
	return &APIKeyServiceImpl{
		repos:     repos,
		encryptor: encryptor,
		manager:   manager,
		log:       log,
	}
}

// CreateAPIKeyRequest represents a request to create an API key
type CreateAPIKeyRequest struct {
	Provider models.AIProvider `json:"provider"`
	Name     string            `json:"name"`
	Key      string            `json:"key"`
}

// Create creates a new API key
func (s *APIKeyServiceImpl) Create(ctx context.Context, tenantID uuid.UUID, req *CreateAPIKeyRequest) (*models.APIKey, error) {
	// Validate the key first
	if err := s.validateKey(ctx, req.Provider, req.Key); err != nil {
		return nil, fmt.Errorf("invalid API key: %w", err)
	}

	// Encrypt the key
	var encryptedKey string
	var err error
	if s.encryptor != nil {
		encryptedKey, err = s.encryptor.Encrypt(req.Key)
		if err != nil {
			return nil, fmt.Errorf("failed to encrypt key: %w", err)
		}
	} else {
		// In development, store as-is (NOT FOR PRODUCTION)
		encryptedKey = req.Key
	}

	apiKey := &models.APIKey{
		ID:           uuid.New(),
		TenantID:     tenantID,
		Provider:     req.Provider,
		Name:         req.Name,
		EncryptedKey: encryptedKey,
		IsValid:      true,
		CreatedAt:    time.Now(),
	}

	if err := s.repos.APIKeys.Create(ctx, apiKey); err != nil {
		return nil, fmt.Errorf("failed to create API key: %w", err)
	}

	s.log.Infow("API key created", "tenant_id", tenantID, "provider", req.Provider)

	// Don't return the encrypted key
	apiKey.EncryptedKey = ""
	return apiKey, nil
}

// List returns all API keys for a tenant
func (s *APIKeyServiceImpl) List(ctx context.Context, tenantID uuid.UUID) ([]*models.APIKey, error) {
	keys, err := s.repos.APIKeys.ListByTenant(ctx, tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to list API keys: %w", err)
	}

	// Ensure encrypted keys are not exposed
	for _, key := range keys {
		key.EncryptedKey = ""
	}

	return keys, nil
}

// Delete deletes an API key
func (s *APIKeyServiceImpl) Delete(ctx context.Context, tenantID, keyID uuid.UUID) error {
	// Verify the key belongs to the tenant
	key, err := s.repos.APIKeys.GetByID(ctx, keyID)
	if err != nil {
		return fmt.Errorf("failed to get API key: %w", err)
	}
	if key == nil || key.TenantID != tenantID {
		return fmt.Errorf("API key not found")
	}

	if err := s.repos.APIKeys.Delete(ctx, keyID); err != nil {
		return fmt.Errorf("failed to delete API key: %w", err)
	}

	s.log.Infow("API key deleted", "tenant_id", tenantID, "key_id", keyID)
	return nil
}

// Validate validates an API key
func (s *APIKeyServiceImpl) Validate(ctx context.Context, tenantID, keyID uuid.UUID) (bool, error) {
	key, err := s.repos.APIKeys.GetByID(ctx, keyID)
	if err != nil {
		return false, fmt.Errorf("failed to get API key: %w", err)
	}
	if key == nil || key.TenantID != tenantID {
		return false, fmt.Errorf("API key not found")
	}

	// Decrypt the key
	var plainKey string
	if s.encryptor != nil {
		plainKey, err = s.encryptor.Decrypt(key.EncryptedKey)
		if err != nil {
			return false, fmt.Errorf("failed to decrypt key: %w", err)
		}
	} else {
		plainKey = key.EncryptedKey
	}

	// Validate with provider
	if err := s.validateKey(ctx, key.Provider, plainKey); err != nil {
		return false, nil
	}

	return true, nil
}

// GetDecryptedKey retrieves and decrypts an API key
func (s *APIKeyServiceImpl) GetDecryptedKey(ctx context.Context, tenantID uuid.UUID, provider models.AIProvider) (string, error) {
	key, err := s.repos.APIKeys.GetByTenantAndProvider(ctx, tenantID, provider)
	if err != nil {
		return "", fmt.Errorf("failed to get API key: %w", err)
	}
	if key == nil {
		return "", fmt.Errorf("no API key found for provider: %s", provider)
	}

	// Decrypt the key
	var plainKey string
	if s.encryptor != nil {
		plainKey, err = s.encryptor.Decrypt(key.EncryptedKey)
		if err != nil {
			return "", fmt.Errorf("failed to decrypt key: %w", err)
		}
	} else {
		plainKey = key.EncryptedKey
	}

	// Update last used
	if err := s.repos.APIKeys.UpdateLastUsed(ctx, key.ID); err != nil {
		s.log.Warnw("failed to update last used", "key_id", key.ID, "error", err)
	}

	return plainKey, nil
}

// GetProviderForTenant creates a provider instance for a tenant
func (s *APIKeyServiceImpl) GetProviderForTenant(ctx context.Context, tenantID uuid.UUID, providerName models.AIProvider, baseURL string) (providers.Provider, error) {
	apiKey, err := s.GetDecryptedKey(ctx, tenantID, providerName)
	if err != nil {
		return nil, err
	}

	return s.manager.CreateProviderWithKey(providerName, apiKey, baseURL)
}

// validateKey validates an API key with the provider
func (s *APIKeyServiceImpl) validateKey(ctx context.Context, providerName models.AIProvider, key string) error {
	provider, err := s.manager.CreateProviderWithKey(providerName, key, "")
	if err != nil {
		return err
	}

	return provider.ValidateAPIKey(ctx, key)
}


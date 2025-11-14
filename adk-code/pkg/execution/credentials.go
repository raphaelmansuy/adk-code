package execution

import (
	"context"
	"fmt"
	"strings"
	"time"
)

// Secret represents a credential secret
type Secret struct {
	// Name is the name of the secret
	Name string

	// Value is the secret value (should be kept secure)
	Value string

	// Type is the secret type (api_key, token, password, etc.)
	Type string

	// CreatedAt is when the secret was created
	CreatedAt time.Time

	// ExpiresAt is when the secret expires (0 = never)
	ExpiresAt time.Time

	// Tags are labels for categorizing secrets
	Tags []string

	// Masked indicates whether the secret should be masked in logs
	Masked bool
}

// IsExpired checks if the secret has expired
func (s *Secret) IsExpired() bool {
	if s.ExpiresAt.IsZero() {
		return false
	}
	return time.Now().After(s.ExpiresAt)
}

// MaskedValue returns a masked version of the value
func (s *Secret) MaskedValue() string {
	if !s.Masked || len(s.Value) == 0 {
		return s.Value
	}

	if len(s.Value) <= 4 {
		return "****"
	}

	return s.Value[:2] + strings.Repeat("*", len(s.Value)-4) + s.Value[len(s.Value)-2:]
}

// CredentialStore represents the storage backend for credentials
type CredentialStore interface {
	// Store saves a secret
	Store(ctx context.Context, secret *Secret) error

	// Retrieve gets a secret by name
	Retrieve(ctx context.Context, name string) (*Secret, error)

	// List returns all secret names
	List(ctx context.Context) ([]string, error)

	// Delete removes a secret
	Delete(ctx context.Context, name string) error

	// Clear removes all secrets
	Clear(ctx context.Context) error
}

// InMemoryCredentialStore is an in-memory implementation of CredentialStore
type InMemoryCredentialStore struct {
	secrets map[string]*Secret
}

// NewInMemoryCredentialStore creates a new in-memory credential store
func NewInMemoryCredentialStore() *InMemoryCredentialStore {
	return &InMemoryCredentialStore{
		secrets: make(map[string]*Secret),
	}
}

// Store saves a secret
func (s *InMemoryCredentialStore) Store(ctx context.Context, secret *Secret) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	if secret == nil {
		return fmt.Errorf("secret cannot be nil")
	}

	if secret.Name == "" {
		return fmt.Errorf("secret name cannot be empty")
	}

	if secret.CreatedAt.IsZero() {
		secret.CreatedAt = time.Now()
	}

	s.secrets[secret.Name] = secret
	return nil
}

// Retrieve gets a secret by name
func (s *InMemoryCredentialStore) Retrieve(ctx context.Context, name string) (*Secret, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	if name == "" {
		return nil, fmt.Errorf("secret name cannot be empty")
	}

	secret, exists := s.secrets[name]
	if !exists {
		return nil, fmt.Errorf("secret %q not found", name)
	}

	if secret.IsExpired() {
		return nil, fmt.Errorf("secret %q has expired", name)
	}

	return secret, nil
}

// List returns all secret names
func (s *InMemoryCredentialStore) List(ctx context.Context) ([]string, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	names := make([]string, 0, len(s.secrets))
	for name := range s.secrets {
		names = append(names, name)
	}

	return names, nil
}

// Delete removes a secret
func (s *InMemoryCredentialStore) Delete(ctx context.Context, name string) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	if name == "" {
		return fmt.Errorf("secret name cannot be empty")
	}

	delete(s.secrets, name)
	return nil
}

// Clear removes all secrets
func (s *InMemoryCredentialStore) Clear(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	s.secrets = make(map[string]*Secret)
	return nil
}

// CredentialManager manages secrets for execution
type CredentialManager struct {
	store CredentialStore
}

// NewCredentialManager creates a new credential manager
func NewCredentialManager(store CredentialStore) *CredentialManager {
	if store == nil {
		store = NewInMemoryCredentialStore()
	}

	return &CredentialManager{
		store: store,
	}
}

// AddSecret adds a new secret
func (cm *CredentialManager) AddSecret(ctx context.Context, name, value, secretType string) error {
	secret := &Secret{
		Name:      name,
		Value:     value,
		Type:      secretType,
		CreatedAt: time.Now(),
		Masked:    true,
	}

	return cm.store.Store(ctx, secret)
}

// AddSecretWithExpiry adds a secret with an expiration time
func (cm *CredentialManager) AddSecretWithExpiry(ctx context.Context, name, value, secretType string, expiresAt time.Time) error {
	secret := &Secret{
		Name:      name,
		Value:     value,
		Type:      secretType,
		CreatedAt: time.Now(),
		ExpiresAt: expiresAt,
		Masked:    true,
	}

	return cm.store.Store(ctx, secret)
}

// GetSecret retrieves a secret
func (cm *CredentialManager) GetSecret(ctx context.Context, name string) (*Secret, error) {
	return cm.store.Retrieve(ctx, name)
}

// GetSecretValue retrieves a secret value
func (cm *CredentialManager) GetSecretValue(ctx context.Context, name string) (string, error) {
	secret, err := cm.store.Retrieve(ctx, name)
	if err != nil {
		return "", err
	}

	return secret.Value, nil
}

// ListSecrets returns all secret names
func (cm *CredentialManager) ListSecrets(ctx context.Context) ([]string, error) {
	return cm.store.List(ctx)
}

// RemoveSecret removes a secret
func (cm *CredentialManager) RemoveSecret(ctx context.Context, name string) error {
	return cm.store.Delete(ctx, name)
}

// InjectIntoEnvironment injects secrets into environment variables
func (cm *CredentialManager) InjectIntoEnvironment(ctx context.Context, env map[string]string) (map[string]string, error) {
	if env == nil {
		env = make(map[string]string)
	}

	secrets, err := cm.store.List(ctx)
	if err != nil {
		return env, err
	}

	for _, name := range secrets {
		secret, err := cm.store.Retrieve(ctx, name)
		if err != nil {
			// Skip expired or inaccessible secrets
			continue
		}

		// Use secret name as environment variable key
		env[name] = secret.Value
	}

	return env, nil
}

// ClearAllSecrets removes all secrets
func (cm *CredentialManager) ClearAllSecrets(ctx context.Context) error {
	return cm.store.Clear(ctx)
}

// MaskOutput masks sensitive values in output
func (cm *CredentialManager) MaskOutput(ctx context.Context, output string) (string, error) {
	secrets, err := cm.store.List(ctx)
	if err != nil {
		return output, err
	}

	maskedOutput := output
	for _, name := range secrets {
		secret, err := cm.store.Retrieve(ctx, name)
		if err != nil {
			continue
		}

		if secret.Masked && len(secret.Value) > 0 {
			// Replace secret value with masked version
			masked := secret.MaskedValue()
			maskedOutput = strings.ReplaceAll(maskedOutput, secret.Value, masked)
		}
	}

	return maskedOutput, nil
}

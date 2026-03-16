package keystore

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/google/uuid"

	"smart-qrcode-desktop/internal/crypto"
	"smart-qrcode-desktop/internal/model"
)

type KeyStore struct {
	baseDir string
}

func New() (*KeyStore, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get home directory: %w", err)
	}
	dir := filepath.Join(home, ".smart-qrcode", "keys")
	if err := os.MkdirAll(dir, 0700); err != nil {
		return nil, fmt.Errorf("failed to create key store directory: %w", err)
	}
	return &KeyStore{baseDir: dir}, nil
}

func (ks *KeyStore) GenerateAndStore(name string) (*model.KeyPairInfo, error) {
	pub, priv, err := crypto.GenerateECDSAKeyPair()
	if err != nil {
		return nil, err
	}

	id := uuid.New().String()
	now := time.Now().Local().Format(time.RFC3339)

	stored := model.StoredKeyPair{
		ID:         id,
		Name:       name,
		PublicKey:  pub,  // 65 bytes (uncompressed ECDSA P-256)
		PrivateKey: priv, // 32 bytes (D scalar)
		CreatedAt:  now,
		IsDefault:  false,
	}

	// If this is the first key, make it default
	keys, _ := ks.ListKeys()
	if len(keys) == 0 {
		stored.IsDefault = true
	}

	if err := ks.save(stored); err != nil {
		return nil, err
	}

	return &model.KeyPairInfo{
		ID:            id,
		Name:          name,
		PublicKey:     base64.RawURLEncoding.EncodeToString(pub),
		Fingerprint:   crypto.PublicKeyFingerprint(pub),
		CreatedAt:     now,
		IsDefault:     stored.IsDefault,
		HasPrivateKey: true,
	}, nil
}

func (ks *KeyStore) ListKeys() ([]model.KeyPairInfo, error) {
	entries, err := os.ReadDir(ks.baseDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read key store: %w", err)
	}

	var keys []model.KeyPairInfo
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".json" {
			continue
		}
		stored, err := ks.load(entry.Name()[:len(entry.Name())-5])
		if err != nil {
			continue
		}
		// Skip keys with invalid public key format (e.g. old Ed25519 keys)
		if len(stored.PublicKey) != 65 {
			continue
		}
		keys = append(keys, model.KeyPairInfo{
			ID:            stored.ID,
			Name:          stored.Name,
			PublicKey:     base64.RawURLEncoding.EncodeToString(stored.PublicKey),
			Fingerprint:   crypto.PublicKeyFingerprint(stored.PublicKey),
			CreatedAt:     stored.CreatedAt,
			IsDefault:     stored.IsDefault,
			HasPrivateKey: len(stored.PrivateKey) > 0,
		})
	}
	return keys, nil
}

func (ks *KeyStore) GetKeyPair(id string) (*model.StoredKeyPair, error) {
	return ks.load(id)
}

func (ks *KeyStore) DeleteKeyPair(id string) error {
	path := filepath.Join(ks.baseDir, id+".json")
	if err := os.Remove(path); err != nil {
		return fmt.Errorf("failed to delete key pair: %w", err)
	}
	return nil
}

func (ks *KeyStore) SetDefault(id string) error {
	entries, err := os.ReadDir(ks.baseDir)
	if err != nil {
		return fmt.Errorf("failed to read key store: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".json" {
			continue
		}
		kid := entry.Name()[:len(entry.Name())-5]
		stored, err := ks.load(kid)
		if err != nil {
			continue
		}
		stored.IsDefault = (kid == id)
		if err := ks.save(*stored); err != nil {
			return err
		}
	}
	return nil
}

func (ks *KeyStore) GetDefaultKeyPair() (*model.StoredKeyPair, error) {
	entries, err := os.ReadDir(ks.baseDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read key store: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".json" {
			continue
		}
		stored, err := ks.load(entry.Name()[:len(entry.Name())-5])
		if err != nil {
			continue
		}
		if stored.IsDefault {
			return stored, nil
		}
	}
	return nil, fmt.Errorf("tidak ada kunci default. Silakan buat kunci terlebih dahulu")
}

func (ks *KeyStore) ExportPublicKey(id string, outputPath string) error {
	stored, err := ks.load(id)
	if err != nil {
		return err
	}
	encoded := base64.RawURLEncoding.EncodeToString(stored.PublicKey)
	return os.WriteFile(outputPath, []byte(encoded), 0644)
}

func (ks *KeyStore) ExportPrivateKey(id string, outputPath string) error {
	stored, err := ks.load(id)
	if err != nil {
		return err
	}
	if len(stored.PrivateKey) == 0 {
		return fmt.Errorf("kunci ini tidak memiliki private key")
	}
	// Export format: D (32 bytes) + PublicKey (65 bytes) = 97 bytes
	exportData := make([]byte, 0, 97)
	exportData = append(exportData, stored.PrivateKey...)
	exportData = append(exportData, stored.PublicKey...)
	encoded := base64.RawURLEncoding.EncodeToString(exportData)
	return os.WriteFile(outputPath, []byte(encoded), 0600)
}

func (ks *KeyStore) ImportPrivateKey(filePath string, name string) (*model.KeyPairInfo, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("gagal membaca file private key: %w", err)
	}

	decoded, err := base64.RawURLEncoding.DecodeString(string(data))
	if err != nil {
		return nil, fmt.Errorf("format private key tidak valid: %w", err)
	}

	var dBytes, pubBytes []byte

	switch len(decoded) {
	case 97:
		// Full export format: D (32) + PublicKey (65)
		dBytes = decoded[:32]
		pubBytes = decoded[32:]
	case 32:
		// D scalar only — derive public key
		dBytes = decoded
		pubBytes, err = crypto.DerivePublicKey(dBytes)
		if err != nil {
			return nil, fmt.Errorf("gagal menurunkan kunci publik: %w", err)
		}
	default:
		return nil, fmt.Errorf("ukuran private key tidak valid: %d bytes (harus 32 atau 97)", len(decoded))
	}

	// Validate the key pair
	if _, err := crypto.ParsePublicKey(pubBytes); err != nil {
		return nil, fmt.Errorf("kunci publik tidak valid: %w", err)
	}

	id := uuid.New().String()
	now := time.Now().Local().Format(time.RFC3339)

	stored := model.StoredKeyPair{
		ID:         id,
		Name:       name,
		PublicKey:  pubBytes,
		PrivateKey: dBytes,
		CreatedAt:  now,
		IsDefault:  false,
	}

	if err := ks.save(stored); err != nil {
		return nil, err
	}

	return &model.KeyPairInfo{
		ID:            id,
		Name:          name,
		PublicKey:     base64.RawURLEncoding.EncodeToString(pubBytes),
		Fingerprint:   crypto.PublicKeyFingerprint(pubBytes),
		CreatedAt:     now,
		IsDefault:     false,
		HasPrivateKey: true,
	}, nil
}

func (ks *KeyStore) ImportPublicKey(filePath string, name string) (*model.KeyPairInfo, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read public key file: %w", err)
	}

	pubBytes, err := base64.RawURLEncoding.DecodeString(string(data))
	if err != nil {
		return nil, fmt.Errorf("format kunci publik tidak valid: %w", err)
	}

	if len(pubBytes) != 65 {
		return nil, fmt.Errorf("ukuran kunci publik tidak valid: %d bytes (harus 65 bytes)", len(pubBytes))
	}

	// Validate it's a valid point on the curve
	if _, err := crypto.ParsePublicKey(pubBytes); err != nil {
		return nil, fmt.Errorf("kunci publik tidak valid: %w", err)
	}

	id := uuid.New().String()
	now := time.Now().Local().Format(time.RFC3339)

	stored := model.StoredKeyPair{
		ID:         id,
		Name:       name,
		PublicKey:  pubBytes,
		PrivateKey: nil,
		CreatedAt:  now,
		IsDefault:  false,
	}

	if err := ks.save(stored); err != nil {
		return nil, err
	}

	return &model.KeyPairInfo{
		ID:            id,
		Name:          name,
		PublicKey:     base64.RawURLEncoding.EncodeToString(pubBytes),
		Fingerprint:   crypto.PublicKeyFingerprint(pubBytes),
		CreatedAt:     now,
		IsDefault:     false,
		HasPrivateKey: false,
	}, nil
}

func (ks *KeyStore) save(kp model.StoredKeyPair) error {
	data, err := json.MarshalIndent(kp, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to serialize key pair: %w", err)
	}
	path := filepath.Join(ks.baseDir, kp.ID+".json")
	return os.WriteFile(path, data, 0600)
}

func (ks *KeyStore) load(id string) (*model.StoredKeyPair, error) {
	path := filepath.Join(ks.baseDir, id+".json")
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read key pair: %w", err)
	}
	var kp model.StoredKeyPair
	if err := json.Unmarshal(data, &kp); err != nil {
		return nil, fmt.Errorf("failed to parse key pair: %w", err)
	}
	return &kp, nil
}

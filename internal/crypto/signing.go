package crypto

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"math/big"
	"os"
)

// GenerateECDSAKeyPair generates a new ECDSA P-256 key pair.
// Returns publicKey (65 bytes uncompressed) and privateKey (32 bytes D scalar).
func GenerateECDSAKeyPair() (publicKey []byte, privateKey []byte, err error) {
	privKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to generate ECDSA key pair: %w", err)
	}
	pubBytes := elliptic.Marshal(elliptic.P256(), privKey.PublicKey.X, privKey.PublicKey.Y)
	dBytes := padTo32(privKey.D.Bytes())
	return pubBytes, dBytes, nil
}

// DerivePublicKey derives the ECDSA public key (65 bytes) from a 32-byte D scalar.
func DerivePublicKey(privateKeyD []byte) ([]byte, error) {
	d := new(big.Int).SetBytes(privateKeyD)
	x, y := elliptic.P256().ScalarBaseMult(d.Bytes())
	if x == nil {
		return nil, fmt.Errorf("invalid private key scalar")
	}
	return elliptic.Marshal(elliptic.P256(), x, y), nil
}

// HashDocument computes SHA-256 hash of a file and returns hex string.
func HashDocument(filePath string) (string, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to open file: %w", err)
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", fmt.Errorf("failed to hash file: %w", err)
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

// HashBytes computes SHA-256 hash of raw bytes and returns hex string.
func HashBytes(data []byte) string {
	h := sha256.Sum256(data)
	return hex.EncodeToString(h[:])
}

// SignHash signs a pre-computed 32-byte hash with ECDSA P-256.
// Returns raw r||s signature (64 bytes).
func SignHash(hash []byte, privateKeyD []byte, publicKey []byte) ([]byte, error) {
	privKey, err := reconstructPrivateKey(privateKeyD, publicKey)
	if err != nil {
		return nil, err
	}
	r, s, err := ecdsa.Sign(rand.Reader, privKey, hash)
	if err != nil {
		return nil, fmt.Errorf("ECDSA sign failed: %w", err)
	}
	sig := make([]byte, 64)
	copy(sig[:32], padTo32(r.Bytes()))
	copy(sig[32:], padTo32(s.Bytes()))
	return sig, nil
}

// SignData hashes data with SHA-256 then signs with ECDSA P-256.
// Returns raw r||s signature (64 bytes).
func SignData(data []byte, privateKeyD []byte, publicKey []byte) ([]byte, error) {
	h := sha256.Sum256(data)
	return SignHash(h[:], privateKeyD, publicKey)
}

// VerifyHash verifies an ECDSA signature against a pre-computed 32-byte hash.
func VerifyHash(hash []byte, signature []byte, publicKey []byte) bool {
	pubKey, err := ParsePublicKey(publicKey)
	if err != nil {
		return false
	}
	if len(signature) != 64 {
		return false
	}
	r := new(big.Int).SetBytes(signature[:32])
	s := new(big.Int).SetBytes(signature[32:])
	return ecdsa.Verify(pubKey, hash, r, s)
}

// VerifySignature hashes data with SHA-256 then verifies the ECDSA signature.
func VerifySignature(data []byte, signature []byte, publicKey []byte) bool {
	h := sha256.Sum256(data)
	return VerifyHash(h[:], signature, publicKey)
}

// PublicKeyFingerprint returns the first 16 hex chars of SHA-256 of the public key.
func PublicKeyFingerprint(publicKey []byte) string {
	h := sha256.Sum256(publicKey)
	return hex.EncodeToString(h[:])[:16]
}

// ParsePublicKey reconstructs an ECDSA public key from raw uncompressed bytes (65 bytes).
func ParsePublicKey(pubBytes []byte) (*ecdsa.PublicKey, error) {
	x, y := elliptic.Unmarshal(elliptic.P256(), pubBytes)
	if x == nil {
		return nil, fmt.Errorf("invalid ECDSA public key")
	}
	return &ecdsa.PublicKey{Curve: elliptic.P256(), X: x, Y: y}, nil
}

func reconstructPrivateKey(dBytes []byte, pubBytes []byte) (*ecdsa.PrivateKey, error) {
	pubKey, err := ParsePublicKey(pubBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse public key: %w", err)
	}
	d := new(big.Int).SetBytes(dBytes)
	return &ecdsa.PrivateKey{PublicKey: *pubKey, D: d}, nil
}

func padTo32(b []byte) []byte {
	if len(b) >= 32 {
		return b[:32]
	}
	padded := make([]byte, 32)
	copy(padded[32-len(b):], b)
	return padded
}

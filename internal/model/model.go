package model

// SecurePayload is the data embedded in the QR code.
type SecurePayload struct {
	Version          int    `json:"version"`
	SecureID         string `json:"secure_id"`
	DocumentHash     string `json:"document_hash"`
	EncryptedPayload string `json:"encrypted_payload"`
	OuterSignature   string `json:"outer_signature"`
	Timestamp        int64  `json:"timestamp"`
	ValidFrom        int64  `json:"valid_from"`
	ValidUntil       int64  `json:"valid_until"`
	PublicKey        string `json:"public_key"`
}

// VerificationResult is returned by the verification flow.
type VerificationResult struct {
	Status       string `json:"status"` // authentic, tampered, not_yet_valid, expired, error
	Message      string `json:"message"`
	DocumentHash string `json:"document_hash"`
	FileName     string `json:"file_name"`
	FileSize     int64  `json:"file_size"`
	IssuerID     string `json:"issuer_id"`
	IssuedAt     string `json:"issued_at"`
	ValidFrom    string `json:"valid_from"`
	ValidUntil   string `json:"valid_until"`
	Metadata     string `json:"metadata"`
	PublicKeyHex string `json:"public_key_hex"`
	ScanCount    int    `json:"scan_count"`
}

type KeyPairInfo struct {
	ID            string `json:"id"`
	Name          string `json:"name"`
	PublicKey     string `json:"public_key"`
	Fingerprint   string `json:"fingerprint"`
	CreatedAt     string `json:"created_at"`
	IsDefault     bool   `json:"is_default"`
	HasPrivateKey bool   `json:"has_private_key"`
}

type StoredKeyPair struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	PublicKey  []byte `json:"public_key"`
	PrivateKey []byte `json:"private_key"`
	CreatedAt  string `json:"created_at"`
	IsDefault  bool   `json:"is_default"`
}

type IssuerConfig struct {
	FilePath   string  `json:"file_path"`
	KeyPairID  string  `json:"key_pair_id"`
	ValidFrom  string  `json:"valid_from"`
	ValidUntil string  `json:"valid_until"`
	Metadata   string  `json:"metadata"`
	IssuerID   string  `json:"issuer_id"`
	QRPosition string  `json:"qr_position"`
	QRPage     int     `json:"qr_page"`
	QRSize     float64 `json:"qr_size"`
}

type QRGenerationResult struct {
	Success        bool   `json:"success"`
	QRCodeBase64   string `json:"qr_code_base64"`
	SecureID       string `json:"secure_id"`
	DocumentHash   string `json:"document_hash"`
	SignedFilePath string `json:"signed_file_path,omitempty"`
	IsPDF          bool   `json:"is_pdf"`
	ErrorMessage   string `json:"error_message,omitempty"`
}

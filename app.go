package main

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
	wailsRuntime "github.com/wailsapp/wails/v2/pkg/runtime"

	"smart-qrcode-desktop/internal/crypto"
	"smart-qrcode-desktop/internal/database"
	"smart-qrcode-desktop/internal/keystore"
	"smart-qrcode-desktop/internal/model"
	"smart-qrcode-desktop/internal/pdf"
	"smart-qrcode-desktop/internal/qr"
)

type App struct {
	ctx      context.Context
	keyStore *keystore.KeyStore
}

func NewApp() *App {
	return &App{}
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	ks, err := keystore.New()
	if err != nil {
		fmt.Println("Warning: failed to initialize key store:", err)
		return
	}
	a.keyStore = ks

	if err := database.InitDB(); err != nil {
		fmt.Println("Warning: failed to initialize database:", err)
	}
}

func (a *App) shutdown(ctx context.Context) {
	database.CloseDB()
}

// === File Dialog Methods ===

func (a *App) SelectDocument() (string, error) {
	path, err := wailsRuntime.OpenFileDialog(a.ctx, wailsRuntime.OpenDialogOptions{
		Title: "Pilih Dokumen",
		Filters: []wailsRuntime.FileFilter{
			{DisplayName: "Semua File", Pattern: "*.*"},
			{DisplayName: "Dokumen", Pattern: "*.pdf;*.doc;*.docx;*.txt;*.xlsx;*.pptx"},
			{DisplayName: "Gambar", Pattern: "*.png;*.jpg;*.jpeg;*.gif;*.bmp"},
		},
	})
	if err != nil {
		return "", err
	}
	return path, nil
}

func (a *App) SelectQRImage() (string, error) {
	path, err := wailsRuntime.OpenFileDialog(a.ctx, wailsRuntime.OpenDialogOptions{
		Title: "Pilih Gambar QR Code",
		Filters: []wailsRuntime.FileFilter{
			{DisplayName: "Gambar", Pattern: "*.png;*.jpg;*.jpeg"},
		},
	})
	if err != nil {
		return "", err
	}
	return path, nil
}

func (a *App) SaveQRImage(pngBase64 string) error {
	path, err := wailsRuntime.SaveFileDialog(a.ctx, wailsRuntime.SaveDialogOptions{
		Title:           "Simpan QR Code",
		DefaultFilename: "secure-qr.png",
		Filters: []wailsRuntime.FileFilter{
			{DisplayName: "PNG Image", Pattern: "*.png"},
		},
	})
	if err != nil {
		return err
	}
	if path == "" {
		return nil
	}

	data, err := base64.StdEncoding.DecodeString(pngBase64)
	if err != nil {
		return fmt.Errorf("failed to decode QR image: %w", err)
	}
	return os.WriteFile(path, data, 0644)
}

// === Document Info ===

func (a *App) GetDocumentInfo(filePath string) (map[string]interface{}, error) {
	info, err := os.Stat(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to get file info: %w", err)
	}
	return map[string]interface{}{
		"name": info.Name(),
		"size": info.Size(),
		"path": filePath,
	}, nil
}

func (a *App) HashDocumentFile(filePath string) (string, error) {
	return crypto.HashDocument(filePath)
}

// === Issuer Flow ===

func (a *App) GenerateSecureQR(config model.IssuerConfig) (*model.QRGenerationResult, error) {
	if a.keyStore == nil {
		return nil, fmt.Errorf("key store belum diinisialisasi")
	}

	// 1. Get signing key
	kp, err := a.keyStore.GetKeyPair(config.KeyPairID)
	if err != nil {
		return nil, fmt.Errorf("gagal memuat kunci: %w", err)
	}
	if kp.PrivateKey == nil {
		return nil, fmt.Errorf("kunci ini hanya memiliki public key, tidak bisa digunakan untuk menandatangani")
	}

	// 2. Hash document (SHA-256)
	docHash, err := crypto.HashDocument(config.FilePath)
	if err != nil {
		return nil, fmt.Errorf("gagal menghash dokumen: %w", err)
	}

	fileInfo, err := os.Stat(config.FilePath)
	if err != nil {
		return nil, fmt.Errorf("gagal membaca info file: %w", err)
	}

	// 3. Parse time window
	validFrom, err := crypto.ParseTimeConfig(config.ValidFrom)
	if err != nil {
		return nil, fmt.Errorf("format waktu mulai tidak valid: %w", err)
	}
	validUntil, err := crypto.ParseTimeConfig(config.ValidUntil)
	if err != nil {
		return nil, fmt.Errorf("format waktu berakhir tidak valid: %w", err)
	}
	if !validUntil.After(validFrom) {
		return nil, fmt.Errorf("waktu berakhir harus setelah waktu mulai")
	}

	// 4. Inner signature: ECDSA sign (hash + validFrom + validUntil)
	//    Binding time bounds into the signature prevents DB tampering of expiration.
	hashBytes, err := hex.DecodeString(docHash)
	if err != nil {
		return nil, fmt.Errorf("gagal decode hash: %w", err)
	}
	var innerDataToSign bytes.Buffer
	innerDataToSign.Write(hashBytes)
	binary.Write(&innerDataToSign, binary.LittleEndian, validFrom.Unix())
	binary.Write(&innerDataToSign, binary.LittleEndian, validUntil.Unix())
	innerSig, err := crypto.SignData(innerDataToSign.Bytes(), kp.PrivateKey, kp.PublicKey)
	if err != nil {
		return nil, fmt.Errorf("gagal menandatangani hash dokumen: %w", err)
	}

	// 5. Build inner payload binary: UUID(16) + Hash(32) + ValidFrom(8) + ValidUntil(8) + InnerSig(64) + FileName
	secureID := uuid.New().String()
	parsedUUID, _ := uuid.Parse(secureID)
	var innerPayload bytes.Buffer
	innerPayload.Write(parsedUUID[:])                                      // 16 bytes
	innerPayload.Write(hashBytes)                                          // 32 bytes
	binary.Write(&innerPayload, binary.LittleEndian, validFrom.Unix())     // 8 bytes
	binary.Write(&innerPayload, binary.LittleEndian, validUntil.Unix())    // 8 bytes
	innerPayload.Write(innerSig)                                           // 64 bytes
	innerPayload.WriteString(fileInfo.Name())                              // variable

	// 6. Time-lock encrypt with Drand (unlock at validFrom)
	duration := time.Until(validFrom)
	if duration < 30*time.Second {
		duration = 30 * time.Second
	}

	encryptedPayload, _, err := crypto.TimeLockEncrypt(innerPayload.Bytes(), duration)
	if err != nil {
		return nil, fmt.Errorf("gagal mengenkripsi payload: %w", err)
	}

	// 7. Prepare timestamp (the unlock time)
	timestamp := validFrom.Unix()
	var timestampBytes bytes.Buffer
	binary.Write(&timestampBytes, binary.LittleEndian, timestamp)

	// 8. Outer signature: ECDSA sign (timestamp + encrypted payload)
	var outerDataToSign bytes.Buffer
	outerDataToSign.Write(timestampBytes.Bytes())
	outerDataToSign.Write(encryptedPayload)
	outerSig, err := crypto.SignData(outerDataToSign.Bytes(), kp.PrivateKey, kp.PublicKey)
	if err != nil {
		return nil, fmt.Errorf("gagal menandatangani outer payload: %w", err)
	}

	// 9. Store payload in SQLite (dynamic QR — payload never in QR itself)
	dbPayload := &database.QRPayload{
		SecureID:         secureID,
		Version:          2,
		DocumentHash:     docHash,
		FileName:         fileInfo.Name(),
		FileSize:         fileInfo.Size(),
		EncryptedPayload: encryptedPayload,
		OuterSignature:   outerSig,
		Timestamp:        timestamp,
		ValidFrom:        validFrom.Unix(),
		ValidUntil:       validUntil.Unix(),
		PublicKey:        kp.PublicKey,
		IssuerID:         config.IssuerID,
		Metadata:         config.Metadata,
		CreatedAt:        time.Now().UTC().Format(time.RFC3339),
	}
	if err := database.SaveQRPayload(dbPayload); err != nil {
		return nil, fmt.Errorf("gagal menyimpan payload ke database: %w", err)
	}

	// 10. Generate QR code containing ONLY the secure_id (anti-cloning: no payload in QR)
	pngBytes, err := qr.GenerateQRCode(secureID)
	if err != nil {
		return nil, err
	}

	qrBase64 := base64.StdEncoding.EncodeToString(pngBytes)
	isPDF := strings.HasSuffix(strings.ToLower(config.FilePath), ".pdf")

	// 11. Embed QR into PDF (store secure_id as PDF property for extraction)
	var signedPath string
	if isPDF {
		signedPath, err = pdf.EmbedQR(config.FilePath, pngBytes, secureID, config.QRPage, config.QRPosition, config.QRSize)
	} else {
		signedPath, err = pdf.CreateQRPDF(pngBytes, secureID, config.QRPosition, config.QRSize)
	}
	if err != nil {
		return nil, fmt.Errorf("gagal membuat PDF dengan QR: %w", err)
	}

	return &model.QRGenerationResult{
		Success:        true,
		QRCodeBase64:   qrBase64,
		SecureID:       secureID,
		DocumentHash:   docHash,
		SignedFilePath: signedPath,
		IsPDF:          isPDF,
	}, nil
}

// === Verifier Flow (Dynamic QR: lookup from DB + scan logging) ===

func (a *App) VerifyQRFromImage(imagePath string) (*model.VerificationResult, error) {
	secureID, err := qr.ReadQRCodeFromImage(imagePath)
	if err != nil {
		return &model.VerificationResult{
			Status:  "error",
			Message: fmt.Sprintf("Gagal membaca QR code: %s", err.Error()),
		}, nil
	}
	return a.verifyBySecureID(secureID)
}

func (a *App) VerifyQRFromData(qrData string) (*model.VerificationResult, error) {
	// QR data is now just the secure_id string
	secureID := strings.TrimSpace(qrData)
	if secureID == "" {
		return &model.VerificationResult{
			Status:  "error",
			Message: "QR code kosong",
		}, nil
	}
	return a.verifyBySecureID(secureID)
}

func (a *App) verifyBySecureID(secureID string) (*model.VerificationResult, error) {
	// Step 1: Log the scan (anti-cloning tracking)
	_ = database.LogScan(secureID, "desktop")

	// Step 2: Look up payload from database
	payload, err := database.GetQRPayload(secureID)
	if err != nil {
		return &model.VerificationResult{
			Status:  "error",
			Message: "QR code tidak ditemukan dalam sistem. Kemungkinan QR code palsu atau dari perangkat lain.",
		}, nil
	}

	// Step 3: Get scan count for anti-cloning detection
	scanCount, _ := database.GetScanCount(secureID)

	// Step 4: Verify outer signature FIRST (timestamp + encrypted payload)
	var timestampBytes bytes.Buffer
	binary.Write(&timestampBytes, binary.LittleEndian, payload.Timestamp)

	var outerDataToVerify bytes.Buffer
	outerDataToVerify.Write(timestampBytes.Bytes())
	outerDataToVerify.Write(payload.EncryptedPayload)

	if !crypto.VerifySignature(outerDataToVerify.Bytes(), payload.OuterSignature, payload.PublicKey) {
		return &model.VerificationResult{
			Status:    "tampered",
			Message:   "Tanda tangan luar TIDAK VALID. Dokumen mungkin telah dimodifikasi!",
			ScanCount: scanCount,
		}, nil
	}

	// Step 5: Decrypt time-locked payload (Drand enforces validFrom cryptographically)
	decrypted, err := crypto.TimeLockDecrypt(payload.EncryptedPayload)
	if err != nil {
		// Drand round not yet published — cryptographic enforcement of validFrom
		dbFrom := time.Unix(payload.ValidFrom, 0).Local()
		dbUntil := time.Unix(payload.ValidUntil, 0).Local()
		return &model.VerificationResult{
			Status:     "not_yet_valid",
			Message:    "Dokumen belum dapat diverifikasi — waktu validasi belum tercapai",
			ValidFrom:  dbFrom.Format(time.RFC3339),
			ValidUntil: dbUntil.Format(time.RFC3339),
			ScanCount:  scanCount,
		}, nil
	}

	// Step 6: Parse inner payload: UUID(16) + Hash(32) + ValidFrom(8) + ValidUntil(8) + InnerSig(64) + FileName
	if len(decrypted) < 128 {
		return &model.VerificationResult{
			Status:    "tampered",
			Message:   "Data terenkripsi tidak valid — dokumen mungkin rusak",
			ScanCount: scanCount,
		}, nil
	}

	innerHash := decrypted[16:48]
	innerValidFrom := int64(binary.LittleEndian.Uint64(decrypted[48:56]))
	innerValidUntil := int64(binary.LittleEndian.Uint64(decrypted[56:64]))
	innerSig := decrypted[64:128]
	innerFileName := string(decrypted[128:])

	// Step 7: Verify inner signature (covers hash + validFrom + validUntil)
	var innerDataToVerify bytes.Buffer
	innerDataToVerify.Write(innerHash)
	binary.Write(&innerDataToVerify, binary.LittleEndian, innerValidFrom)
	binary.Write(&innerDataToVerify, binary.LittleEndian, innerValidUntil)

	if !crypto.VerifySignature(innerDataToVerify.Bytes(), innerSig, payload.PublicKey) {
		return &model.VerificationResult{
			Status:    "tampered",
			Message:   "Tanda tangan dalam TIDAK VALID — dokumen telah dimodifikasi!",
			ScanCount: scanCount,
		}, nil
	}

	// Step 8: Verify hash consistency
	if hex.EncodeToString(innerHash) != payload.DocumentHash {
		return &model.VerificationResult{
			Status:    "tampered",
			Message:   "Hash dokumen tidak cocok — data telah diubah!",
			ScanCount: scanCount,
		}, nil
	}

	// Step 9: Verify time bounds from inner payload match DB (detect DB tampering)
	if innerValidFrom != payload.ValidFrom || innerValidUntil != payload.ValidUntil {
		return &model.VerificationResult{
			Status:    "tampered",
			Message:   "Batas waktu validasi tidak konsisten — data mungkin telah dimanipulasi!",
			ScanCount: scanCount,
		}, nil
	}

	// Step 10: Check time validity using cryptographically-verified bounds
	validFrom := time.Unix(innerValidFrom, 0)
	validUntil := time.Unix(innerValidUntil, 0)
	localFrom := validFrom.Local()
	localUntil := validUntil.Local()

	now := time.Now()
	if now.Before(validFrom) {
		return &model.VerificationResult{
			Status:     "not_yet_valid",
			Message:    "Dokumen belum memasuki masa berlaku",
			ValidFrom:  localFrom.Format(time.RFC3339),
			ValidUntil: localUntil.Format(time.RFC3339),
			ScanCount:  scanCount,
		}, nil
	}
	if now.After(validUntil) {
		return &model.VerificationResult{
			Status:     "expired",
			Message:    "Dokumen sudah melewati masa berlaku",
			ValidFrom:  localFrom.Format(time.RFC3339),
			ValidUntil: localUntil.Format(time.RFC3339),
			ScanCount:  scanCount,
		}, nil
	}

	// Build anti-cloning message
	message := "Dokumen TERVERIFIKASI. Tanda tangan digital valid dan dokumen tidak dimodifikasi."
	if scanCount > 5 {
		message += fmt.Sprintf(" PERINGATAN: QR ini telah dipindai %d kali — kemungkinan duplikasi/kloning.", scanCount)
	}

	return &model.VerificationResult{
		Status:       "authentic",
		Message:      message,
		DocumentHash: payload.DocumentHash,
		FileName:     innerFileName,
		FileSize:     payload.FileSize,
		IssuerID:     payload.IssuerID,
		IssuedAt:     localFrom.Format(time.RFC3339),
		ValidFrom:    localFrom.Format(time.RFC3339),
		ValidUntil:   localUntil.Format(time.RFC3339),
		Metadata:     payload.Metadata,
		PublicKeyHex: crypto.PublicKeyFingerprint(payload.PublicKey),
		ScanCount:    scanCount,
	}, nil
}

func (a *App) VerifyDocument() (*model.VerificationResult, error) {
	path, err := wailsRuntime.OpenFileDialog(a.ctx, wailsRuntime.OpenDialogOptions{
		Title: "Pilih Dokumen untuk Verifikasi",
		Filters: []wailsRuntime.FileFilter{
			{DisplayName: "PDF", Pattern: "*.pdf"},
		},
	})
	if err != nil {
		return nil, err
	}
	if path == "" {
		return nil, fmt.Errorf("tidak ada file yang dipilih")
	}

	// ExtractQR returns the secure_id stored as PDF property or scanned from QR image
	secureID, err := pdf.ExtractQR(path)
	if err != nil {
		return &model.VerificationResult{
			Status:  "error",
			Message: fmt.Sprintf("Gagal menemukan QR code dalam dokumen: %s", err.Error()),
		}, nil
	}

	return a.verifyBySecureID(secureID)
}

func (a *App) SaveSignedDocument(sourcePath string) error {
	path, err := wailsRuntime.SaveFileDialog(a.ctx, wailsRuntime.SaveDialogOptions{
		Title:           "Simpan Dokumen Tertandatangani",
		DefaultFilename: filepath.Base(sourcePath),
		Filters: []wailsRuntime.FileFilter{
			{DisplayName: "PDF", Pattern: "*.pdf"},
		},
	})
	if err != nil {
		return err
	}
	if path == "" {
		return nil
	}
	data, err := os.ReadFile(sourcePath)
	if err != nil {
		return fmt.Errorf("gagal membaca file: %w", err)
	}
	return os.WriteFile(path, data, 0644)
}

// === Key Management ===

func (a *App) GenerateKeyPairCmd(name string) (*model.KeyPairInfo, error) {
	if a.keyStore == nil {
		return nil, fmt.Errorf("key store belum diinisialisasi")
	}
	return a.keyStore.GenerateAndStore(name)
}

func (a *App) ListKeyPairs() ([]model.KeyPairInfo, error) {
	if a.keyStore == nil {
		return nil, fmt.Errorf("key store belum diinisialisasi")
	}
	return a.keyStore.ListKeys()
}

func (a *App) DeleteKeyPair(id string) error {
	if a.keyStore == nil {
		return fmt.Errorf("key store belum diinisialisasi")
	}
	return a.keyStore.DeleteKeyPair(id)
}

func (a *App) SetDefaultKeyPair(id string) error {
	if a.keyStore == nil {
		return fmt.Errorf("key store belum diinisialisasi")
	}
	return a.keyStore.SetDefault(id)
}

func (a *App) ExportPublicKey(id string) error {
	if a.keyStore == nil {
		return fmt.Errorf("key store belum diinisialisasi")
	}
	path, err := wailsRuntime.SaveFileDialog(a.ctx, wailsRuntime.SaveDialogOptions{
		Title:           "Ekspor Kunci Publik",
		DefaultFilename: "public-key.txt",
		Filters: []wailsRuntime.FileFilter{
			{DisplayName: "Text File", Pattern: "*.txt"},
		},
	})
	if err != nil {
		return err
	}
	if path == "" {
		return nil
	}
	return a.keyStore.ExportPublicKey(id, path)
}

func (a *App) ExportPrivateKey(id string) error {
	if a.keyStore == nil {
		return fmt.Errorf("key store belum diinisialisasi")
	}
	path, err := wailsRuntime.SaveFileDialog(a.ctx, wailsRuntime.SaveDialogOptions{
		Title:           "Ekspor Private Key",
		DefaultFilename: "private-key.txt",
		Filters: []wailsRuntime.FileFilter{
			{DisplayName: "Text File", Pattern: "*.txt"},
		},
	})
	if err != nil {
		return err
	}
	if path == "" {
		return nil
	}
	return a.keyStore.ExportPrivateKey(id, path)
}

func (a *App) ImportPrivateKey(name string) (*model.KeyPairInfo, error) {
	if a.keyStore == nil {
		return nil, fmt.Errorf("key store belum diinisialisasi")
	}
	path, err := wailsRuntime.OpenFileDialog(a.ctx, wailsRuntime.OpenDialogOptions{
		Title: "Impor Private Key",
		Filters: []wailsRuntime.FileFilter{
			{DisplayName: "Text File", Pattern: "*.txt"},
			{DisplayName: "Semua File", Pattern: "*.*"},
		},
	})
	if err != nil {
		return nil, err
	}
	if path == "" {
		return nil, fmt.Errorf("tidak ada file yang dipilih")
	}
	return a.keyStore.ImportPrivateKey(filepath.Clean(path), name)
}

func (a *App) ImportPublicKey(name string) (*model.KeyPairInfo, error) {
	if a.keyStore == nil {
		return nil, fmt.Errorf("key store belum diinisialisasi")
	}
	path, err := wailsRuntime.OpenFileDialog(a.ctx, wailsRuntime.OpenDialogOptions{
		Title: "Impor Kunci Publik",
		Filters: []wailsRuntime.FileFilter{
			{DisplayName: "Text File", Pattern: "*.txt"},
			{DisplayName: "Semua File", Pattern: "*.*"},
		},
	})
	if err != nil {
		return nil, err
	}
	if path == "" {
		return nil, fmt.Errorf("tidak ada file yang dipilih")
	}
	return a.keyStore.ImportPublicKey(filepath.Clean(path), name)
}

package database

import (
	"fmt"
	"time"
)

type QRPayload struct {
	SecureID         string `json:"secure_id"`
	Version          int    `json:"version"`
	DocumentHash     string `json:"document_hash"`
	FileName         string `json:"file_name"`
	FileSize         int64  `json:"file_size"`
	EncryptedPayload []byte `json:"-"`
	OuterSignature   []byte `json:"-"`
	Timestamp        int64  `json:"timestamp"`
	ValidFrom        int64  `json:"valid_from"`
	ValidUntil       int64  `json:"valid_until"`
	PublicKey        []byte `json:"-"`
	IssuerID         string `json:"issuer_id"`
	Metadata         string `json:"metadata"`
	CreatedAt        string `json:"created_at"`
}

type ScanLog struct {
	ID        int    `json:"id"`
	SecureID  string `json:"secure_id"`
	ScannedAt string `json:"scanned_at"`
	Source    string `json:"source"`
}

func SaveQRPayload(p *QRPayload) error {
	_, err := DB.Exec(
		`INSERT INTO qr_payloads (secure_id, version, document_hash, file_name, file_size, encrypted_payload, outer_signature, timestamp, valid_from, valid_until, public_key, issuer_id, metadata, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		p.SecureID, p.Version, p.DocumentHash, p.FileName, p.FileSize,
		p.EncryptedPayload, p.OuterSignature, p.Timestamp,
		p.ValidFrom, p.ValidUntil, p.PublicKey,
		p.IssuerID, p.Metadata, p.CreatedAt,
	)
	return err
}

func GetQRPayload(secureID string) (*QRPayload, error) {
	p := &QRPayload{}
	err := DB.QueryRow(
		`SELECT secure_id, version, document_hash, file_name, file_size, encrypted_payload, outer_signature, timestamp, valid_from, valid_until, public_key, issuer_id, metadata, created_at
		FROM qr_payloads WHERE secure_id = ?`,
		secureID,
	).Scan(&p.SecureID, &p.Version, &p.DocumentHash, &p.FileName, &p.FileSize,
		&p.EncryptedPayload, &p.OuterSignature, &p.Timestamp,
		&p.ValidFrom, &p.ValidUntil, &p.PublicKey,
		&p.IssuerID, &p.Metadata, &p.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("QR payload not found: %w", err)
	}
	return p, nil
}

func LogScan(secureID, source string) error {
	now := time.Now().UTC().Format(time.RFC3339)
	_, err := DB.Exec(`INSERT INTO scan_logs (secure_id, scanned_at, source) VALUES (?, ?, ?)`, secureID, now, source)
	return err
}

func GetScanCount(secureID string) (int, error) {
	var count int
	err := DB.QueryRow(`SELECT COUNT(*) FROM scan_logs WHERE secure_id = ?`, secureID).Scan(&count)
	return count, err
}

func GetScanLogs(secureID string) ([]ScanLog, error) {
	rows, err := DB.Query(`SELECT id, secure_id, scanned_at, source FROM scan_logs WHERE secure_id = ? ORDER BY scanned_at DESC`, secureID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var logs []ScanLog
	for rows.Next() {
		var l ScanLog
		if err := rows.Scan(&l.ID, &l.SecureID, &l.ScannedAt, &l.Source); err != nil {
			return nil, err
		}
		logs = append(logs, l)
	}
	if logs == nil {
		logs = []ScanLog{}
	}
	return logs, nil
}

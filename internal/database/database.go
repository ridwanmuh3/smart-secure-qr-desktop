package database

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite"
)

var DB *sql.DB

func InitDB() error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %w", err)
	}

	dataDir := filepath.Join(homeDir, ".smart-qrcode")
	if err := os.MkdirAll(dataDir, 0700); err != nil {
		return fmt.Errorf("failed to create data directory: %w", err)
	}

	dbPath := filepath.Join(dataDir, "data.db")
	DB, err = sql.Open("sqlite", dbPath)
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}

	DB.SetMaxOpenConns(1) // SQLite doesn't support concurrent writes

	if err := migrate(); err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}
	return nil
}

func migrate() error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS qr_payloads (
			secure_id TEXT PRIMARY KEY,
			version INTEGER NOT NULL DEFAULT 2,
			document_hash TEXT NOT NULL,
			file_name TEXT NOT NULL,
			file_size INTEGER NOT NULL,
			encrypted_payload BLOB NOT NULL,
			outer_signature BLOB NOT NULL,
			timestamp INTEGER NOT NULL,
			valid_from INTEGER NOT NULL,
			valid_until INTEGER NOT NULL,
			public_key BLOB NOT NULL,
			issuer_id TEXT DEFAULT '',
			metadata TEXT DEFAULT '',
			created_at TEXT NOT NULL DEFAULT (datetime('now'))
		)`,
		`CREATE TABLE IF NOT EXISTS scan_logs (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			secure_id TEXT NOT NULL,
			scanned_at TEXT NOT NULL DEFAULT (datetime('now')),
			source TEXT DEFAULT 'desktop',
			FOREIGN KEY (secure_id) REFERENCES qr_payloads(secure_id)
		)`,
		`CREATE INDEX IF NOT EXISTS idx_scan_logs_secure_id ON scan_logs(secure_id)`,
	}

	for _, q := range queries {
		if _, err := DB.Exec(q); err != nil {
			return err
		}
	}
	return nil
}

func CloseDB() {
	if DB != nil {
		DB.Close()
	}
}

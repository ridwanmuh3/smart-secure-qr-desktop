package crypto

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/drand/tlock"
	"github.com/drand/tlock/networks/http"
)

const (
	DrandURL  = "https://api.drand.sh/"
	ChainHash = "52db9ba70e0cc0f6eaf7803dd07447a1f5477735fd3f661792ba94600c84e971"
)

// TimeLockEncrypt encrypts data with time-lock encryption using the Drand network.
// The data will only be decryptable after the specified duration from now,
// when the corresponding Drand round beacon is published.
func TimeLockEncrypt(data []byte, duration time.Duration) ([]byte, uint64, error) {
	reader := bytes.NewReader(data)
	network, err := http.NewNetwork(DrandURL, ChainHash)
	if err != nil {
		return nil, 0, fmt.Errorf("gagal terhubung ke jaringan enkripsi: %w", err)
	}

	roundNumber := network.RoundNumber(time.Now().Add(duration))

	var cipherData bytes.Buffer
	if err := tlock.New(network).Encrypt(&cipherData, reader, roundNumber); err != nil {
		return nil, 0, fmt.Errorf("gagal mengenkripsi payload: %w", err)
	}

	return cipherData.Bytes(), roundNumber, nil
}

// TimeLockDecrypt decrypts time-locked data.
// Will fail if the Drand round hasn't been published yet (unlock time not reached).
func TimeLockDecrypt(ciphertext []byte) ([]byte, error) {
	reader := bytes.NewReader(ciphertext)
	network, err := http.NewNetwork(DrandURL, ChainHash)
	if err != nil {
		return nil, fmt.Errorf("gagal terhubung ke jaringan dekripsi: %w", err)
	}

	var plainData bytes.Buffer
	if err := tlock.New(network).Decrypt(&plainData, reader); err != nil {
		return nil, fmt.Errorf("gagal mendekripsi payload: %w", err)
	}

	return plainData.Bytes(), nil
}

// ParseTimeConfig parses various time format inputs into time.Time.
func ParseTimeConfig(input string) (time.Time, error) {
	input = strings.TrimSpace(input)

	if strings.EqualFold(input, "now") {
		return time.Now().Local(), nil
	}

	if strings.HasPrefix(input, "+") {
		offset := input[1:]
		if strings.HasSuffix(offset, "d") {
			days, err := strconv.Atoi(strings.TrimSuffix(offset, "d"))
			if err != nil {
				return time.Time{}, fmt.Errorf("format hari tidak valid: %s", input)
			}
			return time.Now().Local().Add(time.Duration(days) * 24 * time.Hour), nil
		}
		d, err := time.ParseDuration(offset)
		if err != nil {
			return time.Time{}, fmt.Errorf("format durasi tidak valid: %s", input)
		}
		return time.Now().Local().Add(d), nil
	}

	// Try multiple time formats in order of specificity
	formats := []string{
		time.RFC3339,                 // 2006-01-02T15:04:05Z07:00
		"2006-01-02T15:04:05-07:00",  // ISO 8601 with offset
		"2006-01-02T15:04:05-0700",   // ISO 8601 compact offset
		"2006-01-02 15:04:05 -07:00", // space-separated with full offset
		"2006-01-02 15:04:05 -0700",  // space-separated with compact offset
		"2006-01-02 15:04:05 -07",    // space-separated with short offset (+07)
		"2006-01-02 15:04:05",        // space-separated no timezone
		"2006-01-02T15:04:05",        // T-separated with seconds
		"2006-01-02T15:04",           // datetime-local (HTML input)
		"2006-01-02 15:04",           // space-separated no seconds
	}

	for _, layout := range formats {
		t, err := time.Parse(layout, input)
		if err == nil {
			// Formats without timezone info: treat as local time
			if !strings.Contains(layout, "07") {
				t, _ = time.ParseInLocation(layout, input, time.Local)
			}
			return t.Local(), nil
		}
	}

	return time.Time{}, fmt.Errorf("format waktu tidak valid: %s", input)
}

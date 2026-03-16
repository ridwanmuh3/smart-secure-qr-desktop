package qr

import (
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"os"

	"github.com/makiuchi-d/gozxing"
	"github.com/makiuchi-d/gozxing/qrcode"
	goqrcode "github.com/skip2/go-qrcode"
	_ "golang.org/x/image/bmp"
	_ "golang.org/x/image/webp"
)

// GenerateQRCode creates a QR code PNG containing only the secure_id.
// Dynamic QR: the full payload is stored in SQLite, not in the QR itself.
func GenerateQRCode(secureID string) ([]byte, error) {
	if secureID == "" {
		return nil, fmt.Errorf("secure_id cannot be empty")
	}

	png, err := goqrcode.Encode(secureID, goqrcode.Low, 1024)
	if err != nil {
		return nil, fmt.Errorf("failed to generate QR code: %w", err)
	}
	return png, nil
}

// ReadQRCodeFromImage reads a QR code from an image file and returns the decoded text (secure_id).
func ReadQRCodeFromImage(imagePath string) (string, error) {
	f, err := os.Open(imagePath)
	if err != nil {
		return "", fmt.Errorf("failed to open image: %w", err)
	}
	defer f.Close()

	img, _, err := image.Decode(f)
	if err != nil {
		return "", fmt.Errorf("failed to decode image: %w", err)
	}

	bmp, err := gozxing.NewBinaryBitmapFromImage(img)
	if err != nil {
		return "", fmt.Errorf("failed to create bitmap: %w", err)
	}

	hints := map[gozxing.DecodeHintType]interface{}{
		gozxing.DecodeHintType_TRY_HARDER: true,
	}

	reader := qrcode.NewQRCodeReader()
	result, err := reader.Decode(bmp, hints)
	if err != nil {
		return "", fmt.Errorf("gagal membaca QR code dari gambar: %w", err)
	}
	return result.GetText(), nil
}

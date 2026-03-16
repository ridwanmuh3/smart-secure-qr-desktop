package pdf

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/pdfcpu/pdfcpu/pkg/api"

	"smart-qrcode-desktop/internal/qr"
)

var positionMap = map[string]string{
	"bottom-right": "br",
	"bottom-left":  "bl",
	"top-right":    "tr",
	"top-left":     "tl",
	"center":       "c",
}

// EmbedQR stamps a QR code image onto a specific page of a PDF and stores the
// raw QR data as a custom PDF property for reliable extraction.
// Returns the path to the output signed PDF.
func EmbedQR(pdfPath string, qrPNG []byte, qrData string, page int, position string, sizeMM float64) (string, error) {
	if sizeMM <= 0 {
		sizeMM = 30
	}

	// Write QR image to temp file (pdfcpu needs a file path)
	tmpQR, err := os.CreateTemp("", "smartqr-*.png")
	if err != nil {
		return "", fmt.Errorf("gagal membuat file sementara: %w", err)
	}
	tmpQRPath := tmpQR.Name()
	defer os.Remove(tmpQRPath)
	if _, err := tmpQR.Write(qrPNG); err != nil {
		tmpQR.Close()
		return "", fmt.Errorf("gagal menulis QR image: %w", err)
	}
	tmpQR.Close()

	// Build output path: original_signed.pdf
	ext := filepath.Ext(pdfPath)
	base := strings.TrimSuffix(pdfPath, ext)
	outPath := base + "_signed" + ext

	// Map position name to pdfcpu anchor
	pos, ok := positionMap[position]
	if !ok {
		pos = "br"
	}

	// Convert mm to points (1mm ~ 2.835pt)
	sizePoints := int(sizeMM * 2.835)

	// Dynamic offset: base 15pt + 10% of QR size, keeps QR inside page margins
	offset := 15 + sizePoints/10

	// Stamp description: place image on top at given position
	// QR image is 1024x1024px, scale to desired points size
	scaleFactor := float64(sizePoints) / 1024.0
	desc := fmt.Sprintf("position:%s, offset:%d %d, scalefactor:%.4f abs, rotation:0", pos, offset, offset, scaleFactor)

	// Determine pages
	var pages []string
	if page <= 0 {
		pages = []string{"l"} // last page
	} else {
		pages = []string{fmt.Sprintf("%d", page)}
	}

	// Step 1: Add QR image stamp to the PDF (onTop=true)
	if err := api.AddImageWatermarksFile(pdfPath, outPath, pages, true, tmpQRPath, desc, nil); err != nil {
		return "", fmt.Errorf("gagal menyisipkan QR ke PDF: %w", err)
	}

	// Step 2: Store raw QR data as a PDF property for programmatic extraction
	props := map[string]string{
		"SmartQRData": qrData,
	}
	tmpPropOut := outPath + ".tmp"
	if err := api.AddPropertiesFile(outPath, tmpPropOut, props, nil); err != nil {
		// Non-fatal: QR is still visually embedded
		fmt.Println("Warning: gagal menyimpan metadata QR:", err)
	} else {
		// Replace original output with the one containing properties
		os.Remove(outPath)
		os.Rename(tmpPropOut, outPath)
	}

	return outPath, nil
}

// CreateQRPDF creates a new A4 PDF with the QR code stamped at the chosen position.
// Used for non-PDF document inputs so users always get a PDF with embedded QR.
func CreateQRPDF(qrPNG []byte, qrData string, position string, sizeMM float64) (string, error) {
	if sizeMM <= 0 {
		sizeMM = 30
	}

	// Create a blank A4 PDF as base
	blankPDF, err := os.CreateTemp("", "smartqr-blank-*.pdf")
	if err != nil {
		return "", fmt.Errorf("gagal membuat file sementara: %w", err)
	}
	blankPath := blankPDF.Name()
	defer os.Remove(blankPath)

	if err := writeBlankA4(blankPath); err != nil {
		return "", fmt.Errorf("gagal membuat PDF kosong: %w", err)
	}

	// Use EmbedQR to stamp QR onto the blank PDF
	outPath := filepath.Join(os.TempDir(), fmt.Sprintf("smartqr-certificate-%d.pdf", os.Getpid()))
	tmpQR, err := os.CreateTemp("", "smartqr-*.png")
	if err != nil {
		return "", fmt.Errorf("gagal membuat file sementara: %w", err)
	}
	tmpQRPath := tmpQR.Name()
	defer os.Remove(tmpQRPath)
	if _, err := tmpQR.Write(qrPNG); err != nil {
		tmpQR.Close()
		return "", fmt.Errorf("gagal menulis QR image: %w", err)
	}
	tmpQR.Close()

	pos, ok := positionMap[position]
	if !ok {
		pos = "br"
	}
	sizePoints := int(sizeMM * 2.835)
	offset := 15 + sizePoints/10
	scaleFactor := float64(sizePoints) / 1024.0
	desc := fmt.Sprintf("position:%s, offset:%d %d, scalefactor:%.4f abs, rotation:0", pos, offset, offset, scaleFactor)

	if err := api.AddImageWatermarksFile(blankPath, outPath, []string{"1"}, true, tmpQRPath, desc, nil); err != nil {
		return "", fmt.Errorf("gagal membuat QR PDF: %w", err)
	}

	// Store QR data as property
	props := map[string]string{"SmartQRData": qrData}
	tmpPropOut := outPath + ".tmp"
	if err := api.AddPropertiesFile(outPath, tmpPropOut, props, nil); err != nil {
		fmt.Println("Warning: gagal menyimpan metadata QR:", err)
	} else {
		os.Remove(outPath)
		os.Rename(tmpPropOut, outPath)
	}

	return outPath, nil
}

// writeBlankA4 creates a minimal valid blank A4 PDF file.
func writeBlankA4(path string) error {
	var buf bytes.Buffer
	buf.WriteString("%PDF-1.4\n")

	off1 := buf.Len()
	buf.WriteString("1 0 obj\n<< /Type /Catalog /Pages 2 0 R >>\nendobj\n")
	off2 := buf.Len()
	buf.WriteString("2 0 obj\n<< /Type /Pages /Kids [3 0 R] /Count 1 >>\nendobj\n")
	off3 := buf.Len()
	buf.WriteString("3 0 obj\n<< /Type /Page /Parent 2 0 R /MediaBox [0 0 595.28 841.89] /Resources << >> >>\nendobj\n")

	xrefOff := buf.Len()
	buf.WriteString("xref\n")
	buf.WriteString("0 4\n")
	buf.WriteString("0000000000 65535 f \n")
	buf.WriteString(fmt.Sprintf("%010d 00000 n \n", off1))
	buf.WriteString(fmt.Sprintf("%010d 00000 n \n", off2))
	buf.WriteString(fmt.Sprintf("%010d 00000 n \n", off3))
	buf.WriteString("trailer\n<< /Size 4 /Root 1 0 R >>\n")
	buf.WriteString("startxref\n")
	buf.WriteString(fmt.Sprintf("%d\n", xrefOff))
	buf.WriteString("%%EOF\n")

	return os.WriteFile(path, buf.Bytes(), 0644)
}

// ExtractQR reads a QR code from a PDF document.
// It first tries the SmartQRData custom property (fast path),
// then falls back to extracting images and scanning them (slow path).
func ExtractQR(pdfPath string) (string, error) {
	// Fast path: read from PDF custom property
	data, err := extractFromProperty(pdfPath)
	if err == nil && data != "" {
		return data, nil
	}

	// Slow path: extract images and scan for QR codes
	return extractFromImages(pdfPath)
}

func extractFromProperty(pdfPath string) (string, error) {
	f, err := os.Open(pdfPath)
	if err != nil {
		return "", err
	}
	defer f.Close()

	props, err := api.Properties(f, nil)
	if err != nil {
		return "", err
	}
	data, ok := props["SmartQRData"]
	if !ok || data == "" {
		return "", fmt.Errorf("properti SmartQRData tidak ditemukan")
	}
	return data, nil
}

func extractFromImages(pdfPath string) (string, error) {
	tmpDir, err := os.MkdirTemp("", "smartqr-extract-*")
	if err != nil {
		return "", fmt.Errorf("gagal membuat direktori sementara: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	if err := api.ExtractImagesFile(pdfPath, tmpDir, nil, nil); err != nil {
		return "", fmt.Errorf("gagal mengekstrak gambar dari PDF: %w", err)
	}

	// Walk all extracted images and scan for QR codes
	var qrData string
	err = filepath.Walk(tmpDir, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil
		}
		data, scanErr := qr.ReadQRCodeFromImage(path)
		if scanErr == nil && data != "" {
			qrData = data
			return filepath.SkipAll
		}
		return nil
	})
	if err != nil {
		return "", err
	}

	if qrData == "" {
		return "", fmt.Errorf("tidak ditemukan QR code dalam dokumen PDF")
	}
	return qrData, nil
}

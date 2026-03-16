package bench_test

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"math"
	"os"
	"runtime"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/drand/tlock"
	tlockhttp "github.com/drand/tlock/networks/http"
	"github.com/google/uuid"

	"smart-qrcode-desktop/internal/crypto"
	"smart-qrcode-desktop/internal/qr"
)

const (
	totalIterations = 101
	warmupSkip      = 1 // exclude iteration 1 from calculation
	effectiveN      = totalIterations - warmupSkip
)

// metric collects timing measurements for a single operation.
type metric struct {
	name  string
	times []time.Duration
}

func newMetric(name string) *metric {
	return &metric{name: name, times: make([]time.Duration, 0, effectiveN)}
}

func (m *metric) add(d time.Duration) {
	m.times = append(m.times, d)
}

func (m *metric) stats() (min, max, avg, median time.Duration, stddev float64) {
	n := len(m.times)
	if n == 0 {
		return
	}

	sorted := make([]time.Duration, n)
	copy(sorted, m.times)
	sort.Slice(sorted, func(i, j int) bool { return sorted[i] < sorted[j] })

	min = sorted[0]
	max = sorted[n-1]

	var sum int64
	for _, t := range sorted {
		sum += t.Nanoseconds()
	}
	avg = time.Duration(sum / int64(n))

	mid := n / 2
	if n%2 == 0 {
		median = (sorted[mid-1] + sorted[mid]) / 2
	} else {
		median = sorted[mid]
	}

	avgF := float64(sum) / float64(n)
	var variance float64
	for _, t := range sorted {
		diff := float64(t.Nanoseconds()) - avgF
		variance += diff * diff
	}
	stddev = math.Sqrt(variance / float64(n))

	return
}

func fmtDur(d time.Duration) string {
	ns := float64(d.Nanoseconds())
	switch {
	case ns < 1_000:
		return fmt.Sprintf("%.2f ns", ns)
	case ns < 1_000_000:
		return fmt.Sprintf("%.2f us", ns/1_000)
	case ns < 1_000_000_000:
		return fmt.Sprintf("%.2f ms", ns/1_000_000)
	default:
		return fmt.Sprintf("%.3f s", ns/1_000_000_000)
	}
}

func fmtSD(ns float64) string {
	switch {
	case ns < 1_000:
		return fmt.Sprintf("%.2f ns", ns)
	case ns < 1_000_000:
		return fmt.Sprintf("%.2f us", ns/1_000)
	case ns < 1_000_000_000:
		return fmt.Sprintf("%.2f ms", ns/1_000_000)
	default:
		return fmt.Sprintf("%.3f s", ns/1_000_000_000)
	}
}

func fmtBytes(b uint64) string {
	switch {
	case b < 1024:
		return fmt.Sprintf("%d B", b)
	case b < 1024*1024:
		return fmt.Sprintf("%.2f KB", float64(b)/1024)
	case b < 1024*1024*1024:
		return fmt.Sprintf("%.2f MB", float64(b)/(1024*1024))
	default:
		return fmt.Sprintf("%.2f GB", float64(b)/(1024*1024*1024))
	}
}

func printTable(t *testing.T, metrics []*metric) {
	sep := "+" + strings.Repeat("-", 40) + "+" + strings.Repeat("-", 6) + "+" +
		strings.Repeat("-", 14) + "+" + strings.Repeat("-", 14) + "+" +
		strings.Repeat("-", 14) + "+" + strings.Repeat("-", 14) + "+" +
		strings.Repeat("-", 14) + "+"

	hdr := fmt.Sprintf("| %-38s | %4s | %12s | %12s | %12s | %12s | %12s |",
		"Metric", "N", "Min", "Max", "Avg", "Median", "Std Dev")

	t.Log("")
	t.Log(sep)
	t.Log(hdr)
	t.Log(sep)

	for _, m := range metrics {
		min, max, avg, median, sd := m.stats()
		row := fmt.Sprintf("| %-38s | %4d | %12s | %12s | %12s | %12s | %12s |",
			m.name, len(m.times),
			fmtDur(min), fmtDur(max),
			fmtDur(avg), fmtDur(median),
			fmtSD(sd))
		t.Log(row)
	}
	t.Log(sep)
}

// TestPerformanceBenchmark runs 101 iterations of each crypto/QR operation.
// Iteration 1 is excluded from statistics (warm-up).
//
// Run with:
//
//	go test -v -run TestPerformanceBenchmark ./internal/bench/ -timeout 30m
func TestPerformanceBenchmark(t *testing.T) {
	t.Logf("Preparing benchmark environment...")

	// ── Setup ────────────────────────────────────────────────────────────

	// 1. Generate ECDSA P-256 key pair
	pubKey, privKey, err := crypto.GenerateECDSAKeyPair()
	if err != nil {
		t.Fatalf("keygen: %v", err)
	}

	// 2. Create 1 MB test document
	testData := make([]byte, 1024*1024)
	for i := range testData {
		testData[i] = byte(i % 256)
	}
	tmpFile, err := os.CreateTemp("", "bench-doc-*.bin")
	if err != nil {
		t.Fatalf("tmpfile: %v", err)
	}
	tmpFile.Write(testData)
	tmpFile.Close()
	defer os.Remove(tmpFile.Name())

	// 3. Pre-compute document hash
	docHash, err := crypto.HashDocument(tmpFile.Name())
	if err != nil {
		t.Fatalf("hash: %v", err)
	}
	hashBytes, _ := hex.DecodeString(docHash)

	// 4. Pre-compute inner signature
	innerSig, err := crypto.SignHash(hashBytes, privKey, pubKey)
	if err != nil {
		t.Fatalf("sign: %v", err)
	}

	// 5. Build representative inner payload: UUID(16) + Hash(32) + Sig(64) + FileName
	sampleUUID := uuid.New()
	var innerPayload bytes.Buffer
	innerPayload.Write(sampleUUID[:])
	innerPayload.Write(hashBytes)
	innerPayload.Write(innerSig)
	innerPayload.WriteString("benchmark-document.pdf")
	innerPayloadBytes := innerPayload.Bytes()
	t.Logf("Inner payload size: %d bytes", len(innerPayloadBytes))

	// 6. Prepare outer signing data (timestamp + some ciphertext)
	timestamp := time.Now().Unix()
	var tsBytes bytes.Buffer
	binary.Write(&tsBytes, binary.LittleEndian, timestamp)

	// 7. Pre-encrypt data to an already-published Drand round for decryption benchmark
	t.Logf("Pre-encrypting data to past Drand round for decryption benchmark...")
	network, err := tlockhttp.NewNetwork(crypto.DrandURL, crypto.ChainHash)
	if err != nil {
		t.Fatalf("drand network: %v", err)
	}
	pastRound := network.RoundNumber(time.Now().Add(-5 * time.Minute))
	var pastEncBuf bytes.Buffer
	if err := tlock.New(network).Encrypt(&pastEncBuf, bytes.NewReader(innerPayloadBytes), pastRound); err != nil {
		t.Fatalf("tlock encrypt past: %v", err)
	}
	decryptableCipher := pastEncBuf.Bytes()
	t.Logf("Decryptable ciphertext size: %d bytes", len(decryptableCipher))

	// 8. Pre-compute outer signature over (timestamp + ciphertext)
	var outerData bytes.Buffer
	outerData.Write(tsBytes.Bytes())
	outerData.Write(decryptableCipher)
	outerDataBytes := outerData.Bytes()

	outerSig, err := crypto.SignData(outerDataBytes, privKey, pubKey)
	if err != nil {
		t.Fatalf("outer sign: %v", err)
	}

	// 9. Secure ID for QR generation
	secureID := uuid.New().String()

	// Capture memory baseline
	runtime.GC()
	var memBefore runtime.MemStats
	runtime.ReadMemStats(&memBefore)

	t.Logf("Setup complete. Running %d iterations (first excluded as warm-up)...\n", totalIterations)

	// ── Define metrics ──────────────────────────────────────────────────

	mHash := newMetric("Document Hashing (SHA-256, 1MB)")
	mKeyGen := newMetric("ECDSA P-256 Key Generation")
	mSignInner := newMetric("ECDSA P-256 Sign (inner)")
	mSignOuter := newMetric("ECDSA P-256 Sign (outer)")
	mVerifyInner := newMetric("ECDSA P-256 Verify (inner)")
	mVerifyOuter := newMetric("ECDSA P-256 Verify (outer)")
	mQRGen := newMetric("QR Code Generation")
	mTLockEnc := newMetric("Time-Lock Encrypt (Drand)")
	mTLockDec := newMetric("Time-Lock Decrypt (Drand)")
	mFullGen := newMetric("Full QR Generation (end-to-end)")
	mFullVerify := newMetric("Full Verification (end-to-end)")

	// ── Benchmark: Document Hashing ─────────────────────────────────────

	t.Log("[1/11] Document Hashing (SHA-256, 1MB)...")
	for i := 0; i < totalIterations; i++ {
		start := time.Now()
		_, err := crypto.HashDocument(tmpFile.Name())
		elapsed := time.Since(start)
		if err != nil {
			t.Fatalf("iter %d: hash: %v", i, err)
		}
		if i >= warmupSkip {
			mHash.add(elapsed)
		}
	}

	// ── Benchmark: Key Generation ───────────────────────────────────────

	t.Log("[2/11] ECDSA P-256 Key Generation...")
	for i := 0; i < totalIterations; i++ {
		start := time.Now()
		_, _, err := crypto.GenerateECDSAKeyPair()
		elapsed := time.Since(start)
		if err != nil {
			t.Fatalf("iter %d: keygen: %v", i, err)
		}
		if i >= warmupSkip {
			mKeyGen.add(elapsed)
		}
	}

	// ── Benchmark: Inner Signing (SignHash) ──────────────────────────────

	t.Log("[3/11] ECDSA P-256 Sign (inner - SignHash)...")
	for i := 0; i < totalIterations; i++ {
		start := time.Now()
		_, err := crypto.SignHash(hashBytes, privKey, pubKey)
		elapsed := time.Since(start)
		if err != nil {
			t.Fatalf("iter %d: signhash: %v", i, err)
		}
		if i >= warmupSkip {
			mSignInner.add(elapsed)
		}
	}

	// ── Benchmark: Outer Signing (SignData = SHA256 + Sign) ─────────────

	t.Log("[4/11] ECDSA P-256 Sign (outer - SignData)...")
	for i := 0; i < totalIterations; i++ {
		start := time.Now()
		_, err := crypto.SignData(outerDataBytes, privKey, pubKey)
		elapsed := time.Since(start)
		if err != nil {
			t.Fatalf("iter %d: signdata: %v", i, err)
		}
		if i >= warmupSkip {
			mSignOuter.add(elapsed)
		}
	}

	// ── Benchmark: Inner Verification (VerifyHash) ──────────────────────

	t.Log("[5/11] ECDSA P-256 Verify (inner - VerifyHash)...")
	for i := 0; i < totalIterations; i++ {
		start := time.Now()
		ok := crypto.VerifyHash(hashBytes, innerSig, pubKey)
		elapsed := time.Since(start)
		if !ok {
			t.Fatalf("iter %d: verify failed", i)
		}
		if i >= warmupSkip {
			mVerifyInner.add(elapsed)
		}
	}

	// ── Benchmark: Outer Verification (VerifySignature = SHA256 + Verify)

	t.Log("[6/11] ECDSA P-256 Verify (outer - VerifySignature)...")
	for i := 0; i < totalIterations; i++ {
		start := time.Now()
		ok := crypto.VerifySignature(outerDataBytes, outerSig, pubKey)
		elapsed := time.Since(start)
		if !ok {
			t.Fatalf("iter %d: verify failed", i)
		}
		if i >= warmupSkip {
			mVerifyOuter.add(elapsed)
		}
	}

	// ── Benchmark: QR Code Generation ───────────────────────────────────

	t.Log("[7/11] QR Code Generation (secure_id only)...")
	for i := 0; i < totalIterations; i++ {
		start := time.Now()
		_, err := qr.GenerateQRCode(secureID)
		elapsed := time.Since(start)
		if err != nil {
			t.Fatalf("iter %d: qrgen: %v", i, err)
		}
		if i >= warmupSkip {
			mQRGen.add(elapsed)
		}
	}

	// ── Benchmark: Time-Lock Encryption ─────────────────────────────────

	t.Log("[8/11] Time-Lock Encrypt (Drand network)...")
	for i := 0; i < totalIterations; i++ {
		start := time.Now()
		_, _, err := crypto.TimeLockEncrypt(innerPayloadBytes, 30*time.Second)
		elapsed := time.Since(start)
		if err != nil {
			t.Fatalf("iter %d: tlock enc: %v", i, err)
		}
		if i >= warmupSkip {
			mTLockEnc.add(elapsed)
		}
	}

	// ── Benchmark: Time-Lock Decryption ─────────────────────────────────

	t.Log("[9/11] Time-Lock Decrypt (Drand network)...")
	for i := 0; i < totalIterations; i++ {
		start := time.Now()
		_, err := crypto.TimeLockDecrypt(decryptableCipher)
		elapsed := time.Since(start)
		if err != nil {
			t.Fatalf("iter %d: tlock dec: %v", i, err)
		}
		if i >= warmupSkip {
			mTLockDec.add(elapsed)
		}
	}

	// ── Benchmark: Full QR Generation Flow ──────────────────────────────

	t.Log("[10/11] Full QR Generation (hash + sign + tlock encrypt + outer sign + QR gen)...")
	for i := 0; i < totalIterations; i++ {
		start := time.Now()

		// 1. Hash document
		dh, _ := crypto.HashDocument(tmpFile.Name())
		hb, _ := hex.DecodeString(dh)

		// 2. Inner signature (sign hash)
		iSig, _ := crypto.SignHash(hb, privKey, pubKey)

		// 3. Build inner payload
		sid := uuid.New()
		var ip bytes.Buffer
		ip.Write(sid[:])
		ip.Write(hb)
		ip.Write(iSig)
		ip.WriteString("benchmark-document.pdf")

		// 4. Time-lock encrypt (Drand)
		enc, _, _ := crypto.TimeLockEncrypt(ip.Bytes(), 30*time.Second)

		// 5. Outer signature
		ts := time.Now().Unix()
		var tb bytes.Buffer
		binary.Write(&tb, binary.LittleEndian, ts)
		var od bytes.Buffer
		od.Write(tb.Bytes())
		od.Write(enc)
		crypto.SignData(od.Bytes(), privKey, pubKey)

		// 6. Generate QR code
		qr.GenerateQRCode(sid.String())

		elapsed := time.Since(start)
		if i >= warmupSkip {
			mFullGen.add(elapsed)
		}
	}

	// ── Benchmark: Full Verification Flow ───────────────────────────────

	t.Log("[11/11] Full Verification (outer verify + tlock decrypt + inner verify + hash compare)...")
	for i := 0; i < totalIterations; i++ {
		start := time.Now()

		// 1. Verify outer signature
		crypto.VerifySignature(outerDataBytes, outerSig, pubKey)

		// 2. Decrypt time-lock
		dec, _ := crypto.TimeLockDecrypt(decryptableCipher)

		// 3. Parse inner payload
		iHash := dec[16:48]
		iSig := dec[48:112]

		// 4. Verify inner signature
		crypto.VerifyHash(iHash, iSig, pubKey)

		// 5. Compare hashes
		_ = hex.EncodeToString(iHash) == docHash

		elapsed := time.Since(start)
		if i >= warmupSkip {
			mFullVerify.add(elapsed)
		}
	}

	// ── Collect memory stats ────────────────────────────────────────────

	runtime.GC()
	var memAfter runtime.MemStats
	runtime.ReadMemStats(&memAfter)

	// ── Print results ───────────────────────────────────────────────────

	allMetrics := []*metric{
		mHash, mKeyGen,
		mSignInner, mSignOuter,
		mVerifyInner, mVerifyOuter,
		mQRGen,
		mTLockEnc, mTLockDec,
		mFullGen, mFullVerify,
	}

	t.Log("")
	t.Log(strings.Repeat("=", 120))
	t.Log("PERFORMANCE BENCHMARK RESULTS")
	t.Logf("Total iterations: %d | Warm-up excluded: %d | Effective: %d", totalIterations, warmupSkip, effectiveN)
	t.Logf("Document size: 1 MB | Inner payload: %d bytes", len(innerPayloadBytes))
	t.Logf("Crypto: ECDSA P-256 (secp256r1) | Hash: SHA-256 | Time-Lock: Drand tlock")
	t.Log(strings.Repeat("=", 120))

	printTable(t, allMetrics)

	t.Log("")
	t.Log("MEMORY USAGE")
	t.Log(strings.Repeat("-", 60))
	t.Logf("  Total Allocated:     %s", fmtBytes(memAfter.TotalAlloc-memBefore.TotalAlloc))
	t.Logf("  Heap In Use:         %s", fmtBytes(memAfter.HeapInuse))
	t.Logf("  Heap Objects:        %d", memAfter.HeapObjects)
	t.Logf("  Stack In Use:        %s", fmtBytes(memAfter.StackInuse))
	t.Logf("  System Memory:       %s", fmtBytes(memAfter.Sys))
	t.Logf("  GC Cycles:           %d", memAfter.NumGC-memBefore.NumGC)
	t.Log(strings.Repeat("-", 60))

	// ── Summary mapping to evaluation table ─────────────────────────────

	t.Log("")
	t.Log("EVALUATION SUMMARY")
	t.Log(strings.Repeat("-", 80))

	_, _, avgFullGen, _, _ := mFullGen.stats()
	_, _, avgSign, _, _ := mSignInner.stats()
	_, _, avgSignOuter, _, _ := mSignOuter.stats()
	_, _, avgFullVerify, _, _ := mFullVerify.stats()
	_, _, avgTLockEnc, _, _ := mTLockEnc.stats()

	t.Logf("  %-30s %s", "QR generation time:", fmtDur(avgFullGen))
	t.Logf("  %-30s %s (inner) / %s (outer)", "Signature time:", fmtDur(avgSign), fmtDur(avgSignOuter))
	t.Logf("  %-30s %s", "Verification time:", fmtDur(avgFullVerify))
	t.Logf("  %-30s %s", "Encryption overhead:", fmtDur(avgTLockEnc))
	t.Logf("  %-30s %s", "Memory usage:", fmtBytes(memAfter.TotalAlloc-memBefore.TotalAlloc))
	t.Log(strings.Repeat("-", 80))
}

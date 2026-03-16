# Smart Secure QR Code - Desktop Application

A research-driven desktop application for generating cryptographically secure, tamper-proof QR codes with multi-layer digital signatures and time-lock encryption. This project is developed collaboratively by a research team investigating advanced document authentication mechanisms that go beyond conventional static QR code systems.

The system employs a dual-signature architecture (inner + outer ECDSA P-256 signatures), SHA-256 document hashing, and Drand-based time-lock encryption to ensure document integrity, temporal validity enforcement, and anti-cloning protection. Unlike traditional QR implementations that embed raw data directly, this system generates dynamic QR codes containing only a secure reference identifier, while the full cryptographic payload is stored server-side — mitigating data exfiltration and QR duplication attacks.

## Research Team (FAST RG)
- Ir. Randi Rizal, Ph.D
- Fauzan Alvin Mubarok
- Ridwan Muhammad Raihan

## Tech Stack

### Backend (Desktop Core)

| Component            | Technology                                                                               |
| -------------------- | ---------------------------------------------------------------------------------------- |
| Language             | Go 1.25                                                                                  |
| Desktop Framework    | [Wails v2](https://wails.io/) (native webview bindings)                                  |
| Digital Signature    | ECDSA P-256 (secp256r1) via Go `crypto/ecdsa`                                            |
| Document Hashing     | SHA-256                                                                                  |
| Time-Lock Encryption | [Drand tlock](https://github.com/drand/tlock) (threshold network randomness beacon)      |
| QR Generation        | [go-qrcode](https://github.com/skip2/go-qrcode)                                          |
| QR Decoding          | [gozxing](https://github.com/makiuchi-d/gozxing)                                         |
| Database             | SQLite via [modernc.org/sqlite](https://pkg.go.dev/modernc.org/sqlite) (pure Go, no CGO) |
| PDF Manipulation     | [pdfcpu](https://github.com/pdfcpu/pdfcpu)                                               |

### Frontend (UI Layer)

| Component   | Technology                                                                                     |
| ----------- | ---------------------------------------------------------------------------------------------- |
| Framework   | [Vue 3](https://vuejs.org/) + [TypeScript](https://www.typescriptlang.org/)                    |
| Build Tool  | [Vite 7](https://vite.dev/)                                                                    |
| Styling     | [Tailwind CSS 4](https://tailwindcss.com/)                                                     |
| QR Scanning | [html5-qrcode](https://github.com/mebjas/html5-qrcode)                                         |
| Obfuscation | [vite-plugin-bundle-obfuscator](https://github.com/AaronYin0514/vite-plugin-bundle-obfuscator) |

## Installation

### Prerequisites

- **Go** >= 1.25
- **Node.js** >= 18 (or [Bun](https://bun.sh/) runtime)
- **Wails CLI** v2

```bash
go install github.com/wailsapp/wails/v2/cmd/wails@latest
```

- Platform-specific dependencies for Wails:
  - **Linux**: `libgtk-3-dev`, `libwebkit2gtk-4.0-dev`
  - **macOS**: Xcode command line tools
  - **Windows**: WebView2 runtime (bundled with Windows 10+)

### Build & Run

```bash
# Clone the repository
git clone https://github.com/<your-org>/smart-secure-qr-code.git
cd smart-secure-qr-code/smart-qrcode-desktop

# Install frontend dependencies
cd frontend && bun install && cd ..

# Run in development mode (hot-reload enabled)
wails dev

# Build production binary
wails build
```

The compiled binary will be located in `build/bin/`.

## Performance Test

Benchmarks are measured at the **integration/functional level** — each cryptographic primitive and composed pipeline is invoked directly in isolation, not through the full application UI or network stack. The test executes **101 iterations** per operation (first excluded as warm-up, effective **N=100**) against a **1 MB document** using ECDSA P-256 + SHA-256 + Drand tlock.

```bash
go test -v -run TestPerformanceBenchmark ./internal/bench/ -timeout 30m
```

### Benchmark Results

| Metric                          | N   | Min       | Max      | Avg           |
| ------------------------------- | --- | --------- | -------- | ------------- |
| Document Hashing (SHA-256, 1MB) | 100 | 3.56 ms   | 16.00 ms | 5.38 ms       |
| ECDSA P-256 Key Generation      | 100 | 0.02 ms   | 0.06 ms  | 0.02 ms       |
| ECDSA P-256 Sign (inner)        | 100 | 0.05 ms   | 0.13 ms  | 0.06 ms       |
| ECDSA P-256 Sign (outer)        | 100 | 0.05 ms   | 0.17 ms  | 0.08 ms       |
| ECDSA P-256 Verify (inner)      | 100 | 0.10 ms   | 0.22 ms  | 0.12 ms       |
| ECDSA P-256 Verify (outer)      | 100 | 0.10 ms   | 0.52 ms  | 0.13 ms       |
| QR Code Generation              | 100 | 15.16 ms  | 65.47 ms | 19.79 ms      |
| Time-Lock Encrypt (Drand)       | 100 | 759.49 ms | 1.944 s  | 827.56 ms     |
| Time-Lock Decrypt (Drand)       | 100 | 945.40 ms | 1.192 s  | 1.012 s       |
| **Full QR Generation**          | 100 | 778.95 ms | 1.063 s  | **850.00 ms** |
| **Full Verification**           | 100 | 952.34 ms | 1.251 s  | **1.010 s**   |

### Evaluation Summary

| Metric              | Value                                             |
| ------------------- | ------------------------------------------------- |
| QR generation time  | 850.00 ms                                         |
| Signature time      | 0.06 ms (inner) / 0.08 ms (outer)                 |
| Verification time   | 1.010 s                                           |
| Encryption overhead | 827.56 ms                                         |
| Memory usage        | 544.97 MB (total allocated across all iterations) |

> **Key observation:** Local crypto operations (sign, verify, hash) are sub-millisecond. The Drand network round-trip dominates both generation (~828ms) and verification (~1.01s), accounting for >97% of end-to-end time.

## Future Works

- **Blockchain-Based Verification** — Anchor document hashes and signature proofs on-chain (e.g., Ethereum or a permissioned ledger) to provide an immutable, decentralized audit trail. This eliminates single-point-of-trust dependency on the issuing server and enables third-party verifiability without backend access.

- **Post-Quantum Digital Signatures** — Migrate from ECDSA P-256 to lattice-based or hash-based signature schemes (e.g., CRYSTALS-Dilithium, SPHINCS+) to ensure long-term cryptographic resilience against quantum adversaries. The dual-signature architecture is designed to accommodate algorithm-agile upgrades without breaking the verification protocol.

- **Decentralized Verification Network** — Replace the centralized SQLite backend with a distributed verification protocol (e.g., IPFS-backed payload storage combined with smart contract-based access control) so that any participant in the network can independently verify a QR code without relying on a single server or database instance.

## License

This project is licensed under the [MIT License](LICENSE).

```
MIT License

Copyright (c) 2025 FAST Research Team

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
```

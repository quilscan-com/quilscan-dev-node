package main

import (
	"encoding/base64"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/Mercer335/test/internal/signing"
)

func main() {
	checksumsPath := flag.String("checksums", "dist/SHA256SUMS", "path to SHA256SUMS")
	outDir := flag.String("out-dir", "dist", "output directory")
	flag.Parse()

	privateKeyRaw := os.Getenv("DEV_NODE_SIGNING_PRIVATE_KEY")
	if privateKeyRaw == "" {
		fatalf("DEV_NODE_SIGNING_PRIVATE_KEY is required")
	}
	privateKey, err := signing.PrivateKeyFromBase64(privateKeyRaw)
	if err != nil {
		fatalf("parse private key: %v", err)
	}
	checksum, err := signing.LoadChecksum(*checksumsPath)
	if err != nil {
		fatalf("load checksum: %v", err)
	}
	binaryPath := filepath.Join(*outDir, checksum.Artifact)
	binary, err := os.ReadFile(binaryPath)
	if err != nil {
		fatalf("read artifact %s: %v", binaryPath, err)
	}
	sig, err := signing.SignBinary(privateKey, binary)
	if err != nil {
		fatalf("sign artifact: %v", err)
	}
	if err := os.MkdirAll(*outDir, 0o755); err != nil {
		fatalf("create output dir: %v", err)
	}
	writeFile(binaryPath+".sig", []byte(base64.StdEncoding.EncodeToString(sig)+"\n"), 0o644)
	writeFile(filepath.Join(*outDir, "public-key.txt"), []byte(signing.PublicKeyBase64(privateKey)+"\n"), 0o644)
	fmt.Printf("signed %s\n", checksum.Artifact)
}

func writeFile(path string, data []byte, mode os.FileMode) {
	if err := os.WriteFile(path, data, mode); err != nil {
		fatalf("write %s: %v", path, err)
	}
}

func fatalf(format string, args ...any) {
	fmt.Fprintf(os.Stderr, format+"\n", args...)
	os.Exit(1)
}

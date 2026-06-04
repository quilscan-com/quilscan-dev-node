package main

import (
	"encoding/base64"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/Mercer335/test/internal/signing"
)

func main() {
	manifestPath := flag.String("manifest", "", "path to manifest.json")
	signaturePath := flag.String("signature", "", "path to manifest.json.sig")
	publicKeyPath := flag.String("public-key", "keys/build-signing-public-key.txt", "path to base64 Ed25519 public key")
	archiveDir := flag.String("archive-dir", "", "directory containing the dependency archive; omit to verify only the manifest signature")
	expectName := flag.String("expect-name", "", "expected dependency name")
	expectPlatform := flag.String("expect-platform", "", "expected target platform")
	printArchive := flag.Bool("print-archive", false, "print archive name after successful verification")
	flag.Parse()

	if *manifestPath == "" {
		fatalf("-manifest is required")
	}
	if *signaturePath == "" {
		fatalf("-signature is required")
	}
	manifest, err := signing.LoadDependencyManifest(*manifestPath)
	if err != nil {
		fatalf("load manifest: %v", err)
	}
	if *expectName != "" && manifest.Name != *expectName {
		fatalf("manifest name mismatch: got %s want %s", manifest.Name, *expectName)
	}
	if *expectPlatform != "" && manifest.Platform != *expectPlatform {
		fatalf("manifest platform mismatch: got %s want %s", manifest.Platform, *expectPlatform)
	}
	publicKeyRaw, err := os.ReadFile(*publicKeyPath)
	if err != nil {
		fatalf("read public key: %v", err)
	}
	publicKey, err := signing.PublicKeyFromBase64(string(publicKeyRaw))
	if err != nil {
		fatalf("parse public key: %v", err)
	}
	signatureRaw, err := os.ReadFile(*signaturePath)
	if err != nil {
		fatalf("read signature: %v", err)
	}
	signature, err := base64.StdEncoding.DecodeString(strings.TrimSpace(string(signatureRaw)))
	if err != nil {
		fatalf("parse signature: %v", err)
	}
	if !signing.VerifyManifestSignature(publicKey, manifest, signature) {
		fatalf("manifest signature verification failed")
	}
	if *archiveDir != "" {
		if err := signing.VerifyManifestArchive(manifest, *archiveDir); err != nil {
			fatalf("archive verification failed: %v", err)
		}
	}
	if *printArchive {
		fmt.Println(manifest.Archive)
	} else {
		fmt.Printf("verified %s\n", manifest.Archive)
	}
}

func fatalf(format string, args ...any) {
	fmt.Fprintf(os.Stderr, format+"\n", args...)
	os.Exit(1)
}

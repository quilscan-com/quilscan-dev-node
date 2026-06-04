package signing

import (
	"crypto/ed25519"
	"crypto/rand"
	"os"
	"path/filepath"
	"testing"
)

func TestSignedDependencyManifestVerifiesArchive(t *testing.T) {
	publicKey, privateKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatalf("GenerateKey returned error: %v", err)
	}
	dir := t.TempDir()
	archivePath := filepath.Join(dir, "flint-3.0-darwin-arm64.tar.gz")
	if err := os.WriteFile(archivePath, []byte("flint archive bytes"), 0o644); err != nil {
		t.Fatalf("WriteFile returned error: %v", err)
	}

	archiveSHA, err := FileSHA256(archivePath)
	if err != nil {
		t.Fatalf("FileSHA256 returned error: %v", err)
	}
	manifest := DependencyManifest{
		Name:         "flint",
		Version:      "flint-3.0",
		Platform:     "darwin-arm64",
		Archive:      "flint-3.0-darwin-arm64.tar.gz",
		SHA256:       archiveSHA,
		SourceRepo:   "https://github.com/flintlib/flint.git",
		SourceRef:    "flint-3.0",
		SourceCommit: "0123456789abcdef0123456789abcdef01234567",
		Metadata: map[string]string{
			"image_tag": "unused-for-flint",
		},
	}

	signature, err := SignManifest(privateKey, manifest)
	if err != nil {
		t.Fatalf("SignManifest returned error: %v", err)
	}
	if !VerifyManifestSignature(publicKey, manifest, signature) {
		t.Fatal("VerifyManifestSignature returned false for a valid manifest")
	}
	if err := VerifyManifestArchive(manifest, dir); err != nil {
		t.Fatalf("VerifyManifestArchive returned error: %v", err)
	}

	tamperedManifest := manifest
	tamperedManifest.SHA256 = "0000000000000000000000000000000000000000000000000000000000000000"
	if VerifyManifestSignature(publicKey, tamperedManifest, signature) {
		t.Fatal("VerifyManifestSignature returned true for a tampered manifest")
	}

	tamperedMetadata := manifest
	tamperedMetadata.Metadata = map[string]string{
		"image_tag": "tampered",
	}
	if VerifyManifestSignature(publicKey, tamperedMetadata, signature) {
		t.Fatal("VerifyManifestSignature returned true for tampered metadata")
	}

	if err := os.WriteFile(archivePath, []byte("tampered archive bytes"), 0o644); err != nil {
		t.Fatalf("WriteFile returned error: %v", err)
	}
	if err := VerifyManifestArchive(manifest, dir); err == nil {
		t.Fatal("VerifyManifestArchive returned nil for a tampered archive")
	}
}

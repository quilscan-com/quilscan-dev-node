package signing

import (
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

type DependencyManifest struct {
	Name         string            `json:"name"`
	Version      string            `json:"version"`
	Platform     string            `json:"platform"`
	Archive      string            `json:"archive"`
	SHA256       string            `json:"sha256"`
	SourceRepo   string            `json:"source_repo"`
	SourceRef    string            `json:"source_ref"`
	SourceCommit string            `json:"source_commit"`
	Metadata     map[string]string `json:"metadata,omitempty"`
}

func FileSHA256(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hasher := sha256.New()
	if _, err := io.Copy(hasher, file); err != nil {
		return "", err
	}
	return hex.EncodeToString(hasher.Sum(nil)), nil
}

func LoadDependencyManifest(path string) (DependencyManifest, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return DependencyManifest{}, err
	}
	var manifest DependencyManifest
	if err := json.Unmarshal(data, &manifest); err != nil {
		return DependencyManifest{}, err
	}
	if err := manifest.Validate(); err != nil {
		return DependencyManifest{}, err
	}
	return manifest, nil
}

func (m DependencyManifest) Validate() error {
	if m.Name == "" {
		return fmt.Errorf("manifest name is required")
	}
	if m.Version == "" {
		return fmt.Errorf("manifest version is required")
	}
	if m.Platform == "" {
		return fmt.Errorf("manifest platform is required")
	}
	if m.Archive == "" {
		return fmt.Errorf("manifest archive is required")
	}
	if filepath.Base(m.Archive) != m.Archive {
		return fmt.Errorf("manifest archive must be a file name, got %q", m.Archive)
	}
	if len(m.SHA256) != 64 {
		return fmt.Errorf("manifest sha256 must be 64 hex characters")
	}
	if _, err := hex.DecodeString(m.SHA256); err != nil {
		return fmt.Errorf("manifest sha256 is not hex: %w", err)
	}
	if m.SourceRepo == "" {
		return fmt.Errorf("manifest source_repo is required")
	}
	if m.SourceRef == "" {
		return fmt.Errorf("manifest source_ref is required")
	}
	if m.SourceCommit == "" {
		return fmt.Errorf("manifest source_commit is required")
	}
	return nil
}

func MarshalManifest(m DependencyManifest) ([]byte, error) {
	if err := m.Validate(); err != nil {
		return nil, err
	}
	return json.MarshalIndent(m, "", "  ")
}

func SignManifest(privateKey ed25519.PrivateKey, manifest DependencyManifest) ([]byte, error) {
	if len(privateKey) != ed25519.PrivateKeySize {
		return nil, fmt.Errorf("invalid Ed25519 private key length: got %d want %d", len(privateKey), ed25519.PrivateKeySize)
	}
	payload, err := MarshalManifest(manifest)
	if err != nil {
		return nil, err
	}
	return ed25519.Sign(privateKey, payload), nil
}

func VerifyManifestSignature(publicKey ed25519.PublicKey, manifest DependencyManifest, signature []byte) bool {
	if len(publicKey) != ed25519.PublicKeySize || len(signature) != ed25519.SignatureSize {
		return false
	}
	payload, err := MarshalManifest(manifest)
	if err != nil {
		return false
	}
	return ed25519.Verify(publicKey, payload, signature)
}

func VerifyManifestArchive(manifest DependencyManifest, archiveDir string) error {
	if err := manifest.Validate(); err != nil {
		return err
	}
	actualSHA, err := FileSHA256(filepath.Join(archiveDir, manifest.Archive))
	if err != nil {
		return err
	}
	if !strings.EqualFold(actualSHA, manifest.SHA256) {
		return fmt.Errorf("archive sha256 mismatch for %s: got %s want %s", manifest.Archive, actualSHA, manifest.SHA256)
	}
	return nil
}

package main

import (
	"encoding/base64"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/Mercer335/test/internal/signing"
)

func main() {
	var metadata metadataFlags
	name := flag.String("name", "", "dependency name")
	version := flag.String("version", "", "dependency version")
	platform := flag.String("platform", "", "target platform")
	archivePath := flag.String("archive", "", "path to dependency archive")
	sourceRepo := flag.String("source-repo", "", "source repository URL")
	sourceRef := flag.String("source-ref", "", "source repository ref")
	sourceCommit := flag.String("source-commit", "", "source repository commit")
	outDir := flag.String("out-dir", "dist", "output directory")
	flag.Var(&metadata, "metadata", "signed metadata entry in key=value form; repeatable")
	flag.Parse()

	if *archivePath == "" {
		fatalf("-archive is required")
	}
	privateKeyRaw := os.Getenv("BUILD_SIGNING_PRIVATE_KEY")
	if privateKeyRaw == "" {
		privateKeyRaw = os.Getenv("DEV_NODE_SIGNING_PRIVATE_KEY")
	}
	if privateKeyRaw == "" {
		fatalf("BUILD_SIGNING_PRIVATE_KEY or DEV_NODE_SIGNING_PRIVATE_KEY is required")
	}
	privateKey, err := signing.PrivateKeyFromBase64(privateKeyRaw)
	if err != nil {
		fatalf("parse private key: %v", err)
	}
	archiveSHA, err := signing.FileSHA256(*archivePath)
	if err != nil {
		fatalf("hash archive: %v", err)
	}
	manifest := signing.DependencyManifest{
		Name:         *name,
		Version:      *version,
		Platform:     *platform,
		Archive:      filepath.Base(*archivePath),
		SHA256:       archiveSHA,
		SourceRepo:   *sourceRepo,
		SourceRef:    *sourceRef,
		SourceCommit: *sourceCommit,
		Metadata:     metadata.Values,
	}
	manifestJSON, err := signing.MarshalManifest(manifest)
	if err != nil {
		fatalf("marshal manifest: %v", err)
	}
	signature, err := signing.SignManifest(privateKey, manifest)
	if err != nil {
		fatalf("sign manifest: %v", err)
	}
	if err := os.MkdirAll(*outDir, 0o755); err != nil {
		fatalf("create output dir: %v", err)
	}
	writeFile(filepath.Join(*outDir, "manifest.json"), append(manifestJSON, '\n'), 0o644)
	writeFile(filepath.Join(*outDir, "manifest.json.sig"), []byte(base64.StdEncoding.EncodeToString(signature)+"\n"), 0o644)
	writeFile(filepath.Join(*outDir, "public-key.txt"), []byte(signing.PublicKeyBase64(privateKey)+"\n"), 0o644)
	fmt.Printf("signed manifest for %s\n", manifest.Archive)
}

type metadataFlags struct {
	Values map[string]string
}

func (m *metadataFlags) String() string {
	if m == nil {
		return ""
	}
	return fmt.Sprint(m.Values)
}

func (m *metadataFlags) Set(raw string) error {
	key, value, ok := strings.Cut(raw, "=")
	if !ok {
		return fmt.Errorf("metadata must be key=value")
	}
	key = strings.TrimSpace(key)
	if key == "" {
		return fmt.Errorf("metadata key is required")
	}
	if m.Values == nil {
		m.Values = map[string]string{}
	}
	m.Values[key] = value
	return nil
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

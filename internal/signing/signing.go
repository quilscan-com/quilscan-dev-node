package signing

import (
	"crypto/ed25519"
	"encoding/base64"
	"fmt"
	"os"
	"strings"
)

type ArtifactChecksum struct {
	SHA256   string
	Artifact string
}

func LoadChecksum(path string) (ArtifactChecksum, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return ArtifactChecksum{}, err
	}
	return ParseChecksum(string(data))
}

func ParseChecksum(raw string) (ArtifactChecksum, error) {
	for _, line := range strings.Split(raw, "\n") {
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}
		sha := strings.TrimSpace(fields[0])
		artifact := strings.TrimPrefix(strings.TrimSpace(fields[1]), "*")
		if len(sha) != 64 {
			return ArtifactChecksum{}, fmt.Errorf("invalid sha256 length for %s", artifact)
		}
		return ArtifactChecksum{SHA256: strings.ToLower(sha), Artifact: artifact}, nil
	}
	return ArtifactChecksum{}, fmt.Errorf("missing checksum line")
}

func PrivateKeyFromBase64(raw string) (ed25519.PrivateKey, error) {
	key, err := base64.StdEncoding.DecodeString(strings.TrimSpace(raw))
	if err != nil {
		return nil, err
	}
	if len(key) != ed25519.PrivateKeySize {
		return nil, fmt.Errorf("invalid Ed25519 private key length: got %d want %d", len(key), ed25519.PrivateKeySize)
	}
	return ed25519.PrivateKey(key), nil
}

func PublicKeyFromBase64(raw string) (ed25519.PublicKey, error) {
	key, err := base64.StdEncoding.DecodeString(strings.TrimSpace(raw))
	if err != nil {
		return nil, err
	}
	if len(key) != ed25519.PublicKeySize {
		return nil, fmt.Errorf("invalid Ed25519 public key length: got %d want %d", len(key), ed25519.PublicKeySize)
	}
	return ed25519.PublicKey(key), nil
}

func PublicKeyBase64(privateKey ed25519.PrivateKey) string {
	publicKey := privateKey.Public().(ed25519.PublicKey)
	return base64.StdEncoding.EncodeToString(publicKey)
}

func SignBinary(privateKey ed25519.PrivateKey, binary []byte) ([]byte, error) {
	if len(privateKey) != ed25519.PrivateKeySize {
		return nil, fmt.Errorf("invalid Ed25519 private key length: got %d want %d", len(privateKey), ed25519.PrivateKeySize)
	}
	return ed25519.Sign(privateKey, binary), nil
}

func VerifyBinary(publicKey ed25519.PublicKey, binary, signature []byte) bool {
	if len(publicKey) != ed25519.PublicKeySize || len(signature) != ed25519.SignatureSize {
		return false
	}
	return ed25519.Verify(publicKey, binary, signature)
}

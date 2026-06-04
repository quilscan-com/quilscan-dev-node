package signing

import (
	"crypto/ed25519"
	"crypto/rand"
	"testing"
)

func TestSignAndVerifyBinary(t *testing.T) {
	publicKey, privateKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatalf("GenerateKey returned error: %v", err)
	}
	binary := []byte("node binary bytes")

	signature, err := SignBinary(privateKey, binary)
	if err != nil {
		t.Fatalf("SignBinary returned error: %v", err)
	}
	if !VerifyBinary(publicKey, binary, signature) {
		t.Fatal("VerifyBinary returned false for a valid signature")
	}
	tampered := append([]byte(nil), binary...)
	tampered[0] ^= 0xff
	if VerifyBinary(publicKey, tampered, signature) {
		t.Fatal("VerifyBinary returned true for a tampered binary")
	}
}

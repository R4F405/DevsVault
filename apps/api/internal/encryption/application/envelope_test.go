package application

import (
	"bytes"
	"testing"

	encinfra "github.com/devsvault/devsvault/apps/api/internal/encryption/infrastructure"
)

func TestEnvelopeEncryptDecrypt(t *testing.T) {
	key := []byte("01234567890123456789012345678901")
	service := NewEnvelopeService(encinfra.NewStaticKEKProvider("test-key", key))

	payload, err := service.Encrypt([]byte("super-secret"), []byte("secret-id:v1"))
	if err != nil {
		t.Fatalf("encrypt failed: %v", err)
	}
	if bytes.Contains(payload.Ciphertext, []byte("super-secret")) {
		t.Fatal("ciphertext contains plaintext")
	}

	plaintext, err := service.Decrypt(payload, []byte("secret-id:v1"))
	if err != nil {
		t.Fatalf("decrypt failed: %v", err)
	}
	if string(plaintext) != "super-secret" {
		t.Fatalf("unexpected plaintext: %q", plaintext)
	}
}

func TestEnvelopeRejectsWrongAAD(t *testing.T) {
	key := []byte("01234567890123456789012345678901")
	service := NewEnvelopeService(encinfra.NewStaticKEKProvider("test-key", key))

	payload, err := service.Encrypt([]byte("super-secret"), []byte("secret-id:v1"))
	if err != nil {
		t.Fatalf("encrypt failed: %v", err)
	}
	if _, err := service.Decrypt(payload, []byte("secret-id:v2")); err == nil {
		t.Fatal("expected decrypt failure")
	}
}

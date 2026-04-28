package application

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"errors"
	"io"

	encdomain "github.com/devsvault/devsvault/apps/api/internal/encryption/domain"
)

const AlgorithmAES256GCM = "AES-256-GCM"

var ErrDecryptFailed = errors.New("decrypt failed")

type KEKProvider interface {
	ActiveKey() (keyID string, key []byte, err error)
	Key(keyID string) ([]byte, error)
}

type EnvelopeService struct {
	keks KEKProvider
}

func NewEnvelopeService(keks KEKProvider) *EnvelopeService {
	return &EnvelopeService{keks: keks}
}

func (s *EnvelopeService) Encrypt(plaintext []byte, aad []byte) (encdomain.EncryptedPayload, error) {
	dek := make([]byte, 32)
	if _, err := io.ReadFull(rand.Reader, dek); err != nil {
		return encdomain.EncryptedPayload{}, err
	}

	ciphertext, nonce, err := seal(dek, plaintext, aad)
	if err != nil {
		return encdomain.EncryptedPayload{}, err
	}

	keyID, kek, err := s.keks.ActiveKey()
	if err != nil {
		return encdomain.EncryptedPayload{}, err
	}
	wrappedDEK, dekNonce, err := seal(kek, dek, []byte(keyID))
	if err != nil {
		return encdomain.EncryptedPayload{}, err
	}

	return encdomain.EncryptedPayload{
		Ciphertext: ciphertext,
		Nonce:      nonce,
		WrappedDEK: wrappedDEK,
		DEKNonce:   dekNonce,
		KeyID:      keyID,
		Algorithm:  AlgorithmAES256GCM,
	}, nil
}

func (s *EnvelopeService) Decrypt(payload encdomain.EncryptedPayload, aad []byte) ([]byte, error) {
	if payload.Algorithm != AlgorithmAES256GCM {
		return nil, ErrDecryptFailed
	}
	kek, err := s.keks.Key(payload.KeyID)
	if err != nil {
		return nil, ErrDecryptFailed
	}
	dek, err := open(kek, payload.WrappedDEK, payload.DEKNonce, []byte(payload.KeyID))
	if err != nil {
		return nil, ErrDecryptFailed
	}
	plaintext, err := open(dek, payload.Ciphertext, payload.Nonce, aad)
	if err != nil {
		return nil, ErrDecryptFailed
	}
	return plaintext, nil
}

func seal(key []byte, plaintext []byte, aad []byte) ([]byte, []byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, nil, err
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, nil, err
	}
	return gcm.Seal(nil, nonce, plaintext, aad), nonce, nil
}

func open(key []byte, ciphertext []byte, nonce []byte, aad []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	return gcm.Open(nil, nonce, ciphertext, aad)
}

package domain

type EncryptedPayload struct {
	Ciphertext []byte `json:"ciphertext"`
	Nonce      []byte `json:"nonce"`
	WrappedDEK []byte `json:"wrapped_dek"`
	DEKNonce   []byte `json:"dek_nonce"`
	KeyID      string `json:"key_id"`
	Algorithm  string `json:"algorithm"`
}

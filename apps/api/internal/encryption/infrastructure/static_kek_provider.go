package infrastructure

import "errors"

type StaticKEKProvider struct {
	keyID string
	key   []byte
}

func NewStaticKEKProvider(keyID string, key []byte) *StaticKEKProvider {
	copyKey := append([]byte(nil), key...)
	return &StaticKEKProvider{keyID: keyID, key: copyKey}
}

func (p *StaticKEKProvider) ActiveKey() (string, []byte, error) {
	if len(p.key) != 32 {
		return "", nil, errors.New("kek must be 32 bytes")
	}
	return p.keyID, append([]byte(nil), p.key...), nil
}

func (p *StaticKEKProvider) Key(keyID string) ([]byte, error) {
	if keyID != p.keyID || len(p.key) != 32 {
		return nil, errors.New("kek not found")
	}
	return append([]byte(nil), p.key...), nil
}

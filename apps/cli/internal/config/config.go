package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

var ErrNotLoggedIn = errors.New("not logged in, run: devsvault login")

var homeDir = os.UserHomeDir

type Session struct {
	APIURL      string    `json:"api_url"`
	AccessToken string    `json:"access_token"`
	ExpiresAt   time.Time `json:"expires_at"`
}

func Save(session Session) error {
	if session.APIURL == "" || session.AccessToken == "" || session.ExpiresAt.IsZero() {
		return errors.New("invalid session")
	}
	path, err := SessionPath()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return err
	}
	data, err := json.MarshalIndent(session, "", "  ")
	if err != nil {
		return err
	}
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o600)
	if err != nil {
		return err
	}
	_, writeErr := file.Write(data)
	closeErr := file.Close()
	if writeErr != nil {
		return writeErr
	}
	if closeErr != nil {
		return closeErr
	}
	if err := os.Chmod(path, 0o600); err != nil {
		return err
	}
	return nil
}

func Load() (Session, error) {
	path, err := SessionPath()
	if err != nil {
		return Session{}, err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return Session{}, ErrNotLoggedIn
		}
		return Session{}, err
	}
	var session Session
	if err := json.Unmarshal(data, &session); err != nil {
		return Session{}, fmt.Errorf("invalid session file: %w", err)
	}
	if session.APIURL == "" || session.AccessToken == "" {
		return Session{}, ErrNotLoggedIn
	}
	if time.Now().UTC().After(session.ExpiresAt) {
		return Session{}, errors.New("session expired, run: devsvault login")
	}
	return session, nil
}

func Clear() error {
	path, err := SessionPath()
	if err != nil {
		return err
	}
	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}

func SessionPath() (string, error) {
	home, err := homeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".devsvault", "session.json"), nil
}

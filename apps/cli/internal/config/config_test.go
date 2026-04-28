package config

import (
	"errors"
	"os"
	"runtime"
	"testing"
	"time"
)

func TestSaveLoadAndClearSession(t *testing.T) {
	withTempHome(t)
	session := Session{APIURL: "http://localhost:8080", AccessToken: "test-token", ExpiresAt: time.Now().UTC().Add(time.Hour)}

	if err := Save(session); err != nil {
		t.Fatalf("save failed: %v", err)
	}
	path, err := SessionPath()
	if err != nil {
		t.Fatalf("session path failed: %v", err)
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("stat failed: %v", err)
	}
	if mode := info.Mode().Perm(); runtime.GOOS != "windows" && mode != 0o600 {
		t.Fatalf("expected 0600, got %o", mode)
	}

	loaded, err := Load()
	if err != nil {
		t.Fatalf("load failed: %v", err)
	}
	if loaded.APIURL != session.APIURL || loaded.AccessToken != session.AccessToken {
		t.Fatalf("unexpected session: %#v", loaded)
	}

	if err := Clear(); err != nil {
		t.Fatalf("clear failed: %v", err)
	}
	if _, err := Load(); !errors.Is(err, ErrNotLoggedIn) {
		t.Fatalf("expected ErrNotLoggedIn, got %v", err)
	}
}

func TestLoadExpiredSession(t *testing.T) {
	withTempHome(t)
	if err := Save(Session{APIURL: "http://localhost:8080", AccessToken: "test-token", ExpiresAt: time.Now().UTC().Add(-time.Minute)}); err != nil {
		t.Fatalf("save failed: %v", err)
	}
	if _, err := Load(); err == nil || err.Error() != "session expired, run: devsvault login" {
		t.Fatalf("expected expired session error, got %v", err)
	}
}

func withTempHome(t *testing.T) {
	t.Helper()
	original := homeDir
	home := t.TempDir()
	homeDir = func() (string, error) { return home, nil }
	t.Cleanup(func() { homeDir = original })
}

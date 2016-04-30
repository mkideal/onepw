package core

import (
	"testing"
)

func TestNewPassword(t *testing.T) {
	pw := NewPassword("category", "account", "password", "site")
	if pw.Category != "category" ||
		pw.PlainAccount != "account" ||
		pw.PlainPassword != "password" ||
		pw.Site != "site" {
		t.Errorf("NewPassword incorrect")
	}
}

func TestPasswordShortID(t *testing.T) {
	pw := NewPassword("category", "account", "password", "site")
	for _, tt := range []struct {
		ID      string
		ShortID string
	}{
		{"123", "123"},
		{"1234567", "1234567"},
		{"12345678", "1234567"},
	} {
		pw.ID = tt.ID
		if got := pw.ShortID(); got != tt.ShortID {
			t.Errorf("short of %s should be %s, got %s", pw.ID, tt.ShortID, got)
		}
	}
}

func TestCheckPassword(t *testing.T) {
	for _, tt := range []struct {
		passwd string
		err    error
	}{
		{"1", errPasswordTooShort},
		{"12", errPasswordTooShort},
		{"123", errPasswordTooShort},
		{"1234", errPasswordTooShort},
		{"12345", errPasswordTooShort},
	} {
		if got := CheckPassword(tt.passwd); got != tt.err {
			t.Errorf("want %v, got %v", tt.err, got)
		}
	}
}

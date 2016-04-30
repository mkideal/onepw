package core

import (
	"crypto/aes"
	"sort"
	"strconv"
	"testing"
)

func bytesEqual(b1, b2 []byte) bool {
	if len(b1) != len(b2) {
		return false
	}

	for i := range b1 {
		if b1[i] != b2[i] {
			return false
		}
	}
	return true
}

func stringsEqual(s1, s2 []string) bool {
	if len(s1) != len(s2) {
		return false
	}
	for i := range s1 {
		if s1[i] != s2[i] {
			return false
		}
	}
	return true
}

func TestSortPasswords(t *testing.T) {
	passwords := []Password{
		*NewEmptyPassword(),
		*NewEmptyPassword(),
		*NewEmptyPassword(),
	}
	for i := range passwords {
		passwords[i].ID = strconv.Itoa(10000000 - i)
	}
	sort.Sort(passwordSlice(passwords))
	for i := 1; i < len(passwords); i++ {
		if passwords[i-1].ID >= passwords[i].ID {
			t.Errorf("sort passwords incorrect")
		}
	}
	table := passwordSlice(passwords)
	if rc := table.RowCount(); rc != 3 {
		t.Errorf("table RowCount want %d, got %d", 3, rc)
	}
	if cc := table.ColCount(); cc != 5 {
		t.Errorf("table ColCount want %d, got %d", 5, cc)
	}

	passwords2 := []*Password{
		NewEmptyPassword(),
		NewEmptyPassword(),
		NewEmptyPassword(),
	}
	for i := range passwords2 {
		passwords2[i].ID = strconv.Itoa(10000000 - i)
	}
	sort.Sort(passwordPtrSlice(passwords2))
	for i := 1; i < len(passwords2); i++ {
		if passwords2[i-1].ID >= passwords2[i].ID {
			t.Errorf("sort passwords2 incorrect")
		}
	}
	table2 := passwordPtrSlice(passwords2)
	if rc := table2.RowCount(); rc != 3 {
		t.Errorf("table2 RowCount want %d, got %d", 3, rc)
	}
	if cc := table2.ColCount(); cc != 5 {
		t.Errorf("table2 ColCount want %d, got %d", 5, cc)
	}
}

func TestEncryptDecrypt(t *testing.T) {
	pw := NewPassword("category", "account", "password", "site")
	pw.ID = "1234567"
	pw.AccountIV = make([]byte, aes.BlockSize)
	for i := range pw.AccountIV {
		pw.AccountIV[i] = byte(i)
	}
	pw.PasswordIV = make([]byte, aes.BlockSize)
	for i := range pw.PasswordIV {
		pw.PasswordIV[i] = byte(2 * i)
	}

	box := NewBox(NewMemRepository([]byte{}))
	box.masterPassword = "123456"

	box.encrypt(pw)

	wantCipherAccount := []byte{228, 58, 249, 147, 129, 167, 175}
	wantCipherPassword := []byte{158, 190, 63, 132, 121, 169, 38, 195}
	if !bytesEqual(pw.CipherAccount, wantCipherAccount) {
		t.Errorf("CipherAccount want %v, got %v", wantCipherAccount, pw.CipherAccount)
	}
	if !bytesEqual(pw.CipherPassword, wantCipherPassword) {
		t.Errorf("CipherPassword want %v, got %v", wantCipherPassword, pw.CipherPassword)
	}

	pw.PlainAccount = ""
	pw.PlainPassword = ""

	box.decrypt(pw)
	if pw.PlainAccount != "account" {
		t.Errorf("PlainAccount want %s, got %s", "account", pw.PlainAccount)
	}
	if pw.PlainPassword != "password" {
		t.Errorf("PlainPassword want %s, got %s", "account", pw.PlainPassword)
	}
}

func TestAdd(t *testing.T) {
	box := NewBox(NewMemRepository([]byte{}))
	box.masterPassword = "123456"
	box.store.Version = 1
	pw1 := NewPassword("category", "account", "password", "site")
	pw1.ID = "1234567"
	pw2 := NewPassword("CATEGORY", "ACCOUNT", "PASSWORD", "SITE")
	pw2.ID = "1234568"
	pw3 := NewPassword("replace", "replace", "replace", "replace")
	pw3.ID = "1234568"
	passwords := []*Password{pw1, pw2, pw3}
	for _, pw := range passwords {
		if _, _, err := box.Add(pw); err != nil {
			t.Errorf("add password %s error: %v", pw.ID, err)
			continue
		}
		var ok bool
		pw, ok = box.passwords[pw.ID]
		if !ok {
			t.Errorf("add password %s fail", pw.ID)
			continue
		}
		wantPlainAccount, wantPlainPassword := pw.PlainAccount, pw.PlainPassword
		pw.PlainAccount = ""
		pw.PlainPassword = ""
		box.decrypt(pw)
		if pw.PlainAccount != wantPlainAccount {
			t.Errorf("PlainAccount want %s, got %s", wantPlainAccount, pw.PlainAccount)
		}
		if pw.PlainPassword != wantPlainPassword {
			t.Errorf("PlainPassword want %s, got %s", wantPlainPassword, pw.PlainPassword)
		}
	}
	if len(box.passwords) != 2 {
		t.Errorf("passwords size want %d, got %d", 2, len(box.passwords))
	}
}

func TestRemove(t *testing.T) {
	box := NewBox(NewMemRepository([]byte{}))
	box.masterPassword = "123456"
	box.store.Version = 1
	genPasswords := func() map[string]*Password {
		pws := map[string]*Password{
			"1234567": NewPassword("category", "account", "password", "site"),
			"1234568": NewPassword("CATEGORY", "ACCOUNT", "PASSWORD", "SITE"),
			"1234569": NewPassword("CATEGORY", "ACCOUNT", "PASSWORD2", ""),
		}
		for id, pw := range pws {
			pw.ID = id
		}
		return pws
	}
	box.passwords = genPasswords()
	if _, err := box.Remove([]string{"12"}, false); err == nil {
		t.Errorf("Remove passwords want error, got nil")
		return
	}

	if ids, err := box.Remove([]string{"1234569"}, false); err != nil {
		t.Errorf("Remove passwords want nil, got %v", err)
		return
	} else if !stringsEqual(ids, []string{"1234569"}) {
		t.Errorf("Remove passwords incorrect")
	}

	if ids, err := box.Remove([]string{"12"}, true); err != nil {
		t.Errorf("Remove passwords want nil, got %v", err)
		return
	} else if !stringsEqual(ids, []string{"1234567", "1234568"}) {
		t.Errorf("Remove passwords incorrect")
	}

	box.passwords = genPasswords()
	if _, err := box.RemoveByAccount("category", "not_found", false); err == nil {
		t.Errorf("RemoveByAccount want error, got nil")
		return
	}
	if ids, err := box.RemoveByAccount("category", "account", false); err != nil {
		t.Errorf("RemoveByAccount want nil, got %v", err)
		return
	} else if !stringsEqual(ids, []string{"1234567"}) {
		t.Errorf("RemoveByAccount passwords incorrect")
		return
	}
	if _, err := box.RemoveByAccount("CATEGORY", "ACCOUNT", false); err == nil {
		t.Errorf("RemoveByAccount want error, got nil")
		return
	}
	if ids, err := box.RemoveByAccount("CATEGORY", "ACCOUNT", true); err != nil {
		t.Errorf("RemoveByAccount want nil, got %v", err)
		return
	} else if !stringsEqual(ids, []string{"1234568", "1234569"}) {
		t.Errorf("RemoveByAccount passwords incorrect")
	}
}

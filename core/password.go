package core

import (
	"crypto/cipher"
	"encoding/json"
	"io"
	"strings"
	"time"
)

const shortIDLength = 7

type passwordInspect struct {
	ID            string
	Category      string
	Account       string
	Password      string
	Site          string
	Tags          []string
	Ext           string
	CreatedAt     string
	LastUpdatedAt string
}

// PasswordBasic is basic of Password
type PasswordBasic struct {
	// Category of password
	Category string `cli:"c,category" usage:"category of password"`

	// Plain account and password
	PlainAccount  string `json:"-" cli:"u,account" usage:"account of password"`
	PlainPassword string `json:"-" cli:"-"`

	// Website address for web password
	Site string `cli:"site" usage:"website of password"`

	// Password tags
	Tags []string `cli:"tag" usage:"tags of password"`

	// Extension information: JSON base64 string
	Ext string `cli:"-"`
}

// Password represents entity of password
type Password struct {
	PasswordBasic

	// Unique id of password
	ID string `cli:"id" usage:"password id for updating"`

	// IVs
	AccountIV  []byte `cli:"-"`
	PasswordIV []byte `cli:"-"`

	// Ciphers
	CipherAccount  []byte `cli:"-"`
	CipherPassword []byte `cli:"-"`

	// Created time stamp
	CreatedAt int64 `cli:"-"`

	// Last updated time stamp
	LastUpdatedAt int64 `cli:"-"`
}

var passwordHeader = []string{"ID", "CATEGORY", "ACCOUNT", "PASSWORD", "UPDATED_AT"}

func (pw Password) get(i int) string {
	switch i {
	case 0:
		return pw.ShortID()
	case 1:
		return pw.Category
	case 2:
		return pw.PlainAccount
	case 3:
		return pw.PlainPassword
	case 4:
		return time.Unix(pw.LastUpdatedAt, 0).Format(time.RFC3339)
	}
	panic("unreachable")
}

func (pw Password) colCount() int {
	return 5
}

func (pw Password) match(word string) bool {
	if strings.Contains(pw.ID, word) {
		return true
	}
	if strings.Contains(pw.Category, word) {
		return true
	}
	if strings.Contains(pw.PlainAccount, word) {
		return true
	}
	if strings.Contains(pw.PlainPassword, word) {
		return true
	}
	if strings.Contains(pw.Site, word) {
		return true
	}
	if pw.Tags != nil {
		for _, tag := range pw.Tags {
			if strings.Contains(tag, word) {
				return true
			}
		}
	}
	return false
}

// NewEmptyPassword creates a empty Password entity
func NewEmptyPassword() *Password {
	return NewPassword("", "", "", "")
}

// NewPassword creates a Password entity
func NewPassword(category, account, passwd, site string) *Password {
	now := time.Now().Unix()
	pw := &Password{
		PasswordBasic: PasswordBasic{
			Category:      category,
			PlainAccount:  account,
			PlainPassword: passwd,
			Site:          site,
			Tags:          []string{},
		},
		AccountIV:      []byte{},
		PasswordIV:     []byte{},
		CipherAccount:  []byte{},
		CipherPassword: []byte{},
		CreatedAt:      now,
		LastUpdatedAt:  now,
	}
	return pw
}

// ShortID returns short length id string
func (pw *Password) ShortID() string {
	if len(pw.ID) > shortIDLength {
		return pw.ID[:shortIDLength]
	}
	return pw.ID
}

func (pw *Password) migrate(from *Password) {
	copyNonEmptyString(&pw.PasswordBasic.Category, from.PasswordBasic.Category)
	copyNonEmptyString(&pw.PasswordBasic.Ext, from.PasswordBasic.Ext)
	copyNonEmptyString(&pw.PasswordBasic.PlainAccount, from.PasswordBasic.PlainAccount)
	copyNonEmptyString(&pw.PasswordBasic.PlainPassword, from.PasswordBasic.PlainPassword)
	copyNonEmptyString(&pw.PasswordBasic.Site, from.PasswordBasic.Site)

	if from.PasswordBasic.Tags != nil && len(from.PasswordBasic.Tags) != 0 {
		pw.PasswordBasic.Tags = make([]string, len(from.PasswordBasic.Tags))
		copy(pw.PasswordBasic.Tags, from.PasswordBasic.Tags)
	}
}

func (pw *Password) inspect(w io.Writer, prefix string) {
	v := new(passwordInspect)
	v.ID = pw.ID
	v.Account = pw.PlainAccount
	v.Category = pw.Category
	v.Password = pw.PlainPassword
	v.Site = pw.Site
	v.Tags = pw.Tags
	v.Ext = pw.Ext
	v.CreatedAt = time.Unix(pw.CreatedAt, 0).Format(time.RFC3339)
	v.LastUpdatedAt = time.Unix(pw.LastUpdatedAt, 0).Format(time.RFC3339)
	if data, err := json.MarshalIndent(v, prefix, "    "); err == nil {
		w.Write(data)
	}
}

// CheckPassword validate password string
func CheckPassword(passwd string) error {
	if len(passwd) < 6 {
		return errPasswordTooShort
	}
	return nil
}

func cfbEncrypt(block cipher.Block, iv, src []byte) []byte {
	cfb := cipher.NewCFBEncrypter(block, iv)
	dst := make([]byte, len(src))
	cfb.XORKeyStream(dst, src)
	return dst
}

func cfbDecrypt(block cipher.Block, iv, src []byte) []byte {
	cfb := cipher.NewCFBDecrypter(block, iv)
	dst := make([]byte, len(src))
	cfb.XORKeyStream(dst, src)
	return dst
}

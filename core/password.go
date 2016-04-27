package core

import (
	"crypto/cipher"
	"fmt"
	"io"
	"time"
)

const shortIdLength = 7

type PasswordBasic struct {
	// Label of password
	Label string `cli:"label" usage:"label of password"`

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

type Password struct {
	PasswordBasic

	// Unique id of password
	Id string `cli:"id" usage:"password id for updating"`

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

func NewEmptyPassword() *Password {
	return NewPassword("", "", "", "")
}

func NewPassword(label, account, passwd, site string) *Password {
	now := time.Now().Unix()
	pw := &Password{
		PasswordBasic: PasswordBasic{
			Label:         label,
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

func (pw *Password) Brief(w io.Writer, format string) {
	fmt.Fprintf(w, format, pw.ShortId(), pw.Label, pw.PlainAccount, pw.PlainPassword, time.Unix(pw.LastUpdatedAt, 0).Format(time.RFC3339))
}

func (pw *Password) ShortId() string {
	if len(pw.Id) > shortIdLength {
		return pw.Id[:shortIdLength]
	}
	return pw.Id
}

func (pw *Password) migrate(from *Password) {
	pw.PasswordBasic = from.PasswordBasic
	pw.PasswordBasic.Tags = make([]string, len(from.PasswordBasic.Tags))
	copy(pw.PasswordBasic.Tags, from.PasswordBasic.Tags)
}

func CheckPassword(passwd string) error {
	if len(passwd) < 6 {
		return ErrPasswordTooShort
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

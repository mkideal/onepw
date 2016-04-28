package core

import (
	"bytes"
	"errors"
	"fmt"
	"sort"
)

var (
	ErrAmbiguous              = errors.New("ambiguous")
	ErrAllocateId             = errors.New("allocate id fail")
	ErrEmptyMasterPassword    = errors.New("master password is empty")
	ErrMasterPasswordTooShort = errors.New("master password too short")
	ErrPasswordTooShort       = errors.New("password too short")
	ErrNotFullBlock           = errors.New("cipher bytes not full block")
	ErrLengthOfIV             = errors.New("IV length not equal to block size")
)

func newErrAmbiguous(passwords []*Password) error {
	buf := bytes.NewBufferString("ambiguous:")
	sort.Stable(passwordPtrSlice(passwords))
	for _, pw := range passwords {
		buf.WriteString("\n")
		pw.Brief(buf, passwordFormat)
	}
	return errors.New(buf.String())
}

func newErrPasswordNotFound(id string) error {
	return fmt.Errorf("password %s not found", id)
}

func newErrPasswordNotFoundWithAccount(category, account string) error {
	return fmt.Errorf("password by (category=%s,account=%s) not found", category, account)
}

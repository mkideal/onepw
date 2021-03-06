package core

import (
	"bytes"
	"errors"
	"fmt"
	"sort"

	"github.com/mkideal/pkg/textutil"
)

var (
	errAmbiguous                   = errors.New("ambiguous")
	errAllocateID                  = errors.New("allocate id fail")
	errEmptyMasterPassword         = errors.New("master password is empty")
	errMasterPasswordTooShort      = errors.New("master password too short")
	errPasswordTooShort            = errors.New("password too short")
	errNotFullBlock                = errors.New("cipher bytes not full block")
	errLengthOfIV                  = errors.New("IV length not equal to block size")
	errMissingMasterPasswordInBook = errors.New("master password not found in password book")
	errMasterPassword              = errors.New("incorrect master password")
)

func newErrAmbiguous(passwords []*Password) error {
	buf := bytes.NewBufferString("ambiguous:\n")
	table := passwordPtrSlice(passwords)
	sort.Stable(table)
	textutil.WriteTable(buf, table, nil)
	return errors.New(buf.String())
}

func newErrPasswordNotFound(id string) error {
	return fmt.Errorf("password %s not found", id)
}

func newErrPasswordNotFoundWithAccount(category, account string) error {
	return fmt.Errorf("password by (category=%s,account=%s) not found", category, account)
}

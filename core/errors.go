package core

import (
	"errors"
)

var (
	ErrAmbiguous              = errors.New("ambiguous")
	ErrPasswordNotFound       = errors.New("password not found")
	ErrAllocateId             = errors.New("allocate id fail")
	ErrEmptyMasterPassword    = errors.New("master password is empty")
	ErrMasterPasswordTooShort = errors.New("master password too short")
	ErrPasswordTooShort       = errors.New("password too short")
	ErrNotFullBlock           = errors.New("cipher bytes not full block")
	ErrLengthOfIV             = errors.New("IV length not equal to block size")
)

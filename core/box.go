package core

import (
	"crypto/aes"
	"crypto/md5"
	crand "crypto/rand"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/mkideal/pkg/debug"
)

const passwordFormat = "%-10s%-15s%-16s%-16s%-20s"

func init() {
	rand.Seed(time.Now().UnixNano())
}

func md5sum(i interface{}) string {
	switch v := i.(type) {
	case string:
		return fmt.Sprintf("%x", md5.Sum([]byte(v)))

	case []byte:
		return fmt.Sprintf("%x", md5.Sum(v))

	default:
		return fmt.Sprintf("%x", md5.Sum([]byte(fmt.Sprintf("%v", v))))
	}
}

type BoxRepository interface {
	Load() ([]byte, error)
	Save([]byte) error
}

type Box struct {
	sync.RWMutex
	masterPassword string
	repo           BoxRepository
	passwords      map[string]*Password
}

func (box *Box) Init(masterPassword string) error {
	//TODO: check masterPassword
	if len(masterPassword) < 6 {
		return ErrMasterPasswordTooShort
	}
	box.Lock()
	defer box.Unlock()
	box.masterPassword = masterPassword
	if err := box.load(); err != nil {
		return err
	}
	for _, pw := range box.passwords {
		if err := box.encrypt(pw); err != nil {
			return err
		}
	}
	return box.save()
}

func NewBox(repo BoxRepository) *Box {
	box := &Box{
		repo:      repo,
		passwords: map[string]*Password{},
	}
	return box
}

func (box *Box) Load() error {
	box.Lock()
	defer box.Unlock()
	return box.load()
}

func (box *Box) load() error {
	data, err := box.repo.Load()
	if err != nil {
		return err
	}
	return box.unmarshal(data)
}

func (box *Box) Save() error {
	box.Lock()
	defer box.Unlock()
	return box.save()
}

func (box *Box) save() error {
	data, err := box.marshal()
	if err != nil {
		return err
	}
	debug.Debugf("marshal result: %v", string(data))
	return box.repo.Save(data)
}

func (box *Box) Add(pw *Password) (id string, new bool, err error) {
	debug.Debugf("Add new password: %v", pw)
	box.Lock()
	defer box.Unlock()
	if box.masterPassword == "" {
		err = ErrEmptyMasterPassword
		return
	}
	if old, ok := box.passwords[pw.Id]; ok {
		old.LastUpdatedAt = time.Now().Unix()
		old.migrate(pw)
		pw = old
		new = false
	} else if pw.Id != "" {
		err = newErrPasswordNotFound(pw.Id)
		return
	} else {
		id, err = box.allocId()
		if err != nil {
			return
		}
		pw.Id = id
		new = true
	}
	if err = box.encrypt(pw); err != nil {
		return
	}
	box.passwords[pw.Id] = pw
	debug.Debugf("add new password: %v", pw)
	err = box.save()
	return
}

func (box *Box) Remove(ids []string, all bool) ([]string, error) {
	box.Lock()
	defer box.Unlock()
	if box.masterPassword == "" {
		return nil, ErrEmptyMasterPassword
	}
	deletedIds := []string{}
	passwords := make([]*Password, 0)

	for _, id := range ids {
		size := len(deletedIds)
		if foundPw, ok := box.passwords[id]; !ok {
			for _, pw := range box.passwords {
				if strings.HasPrefix(pw.Id, id) {
					deletedIds = append(deletedIds, pw.Id)
					passwords = append(passwords, pw)
				}
			}
		} else {
			deletedIds = append(deletedIds, id)
			passwords = append(passwords, foundPw)
		}
		if len(deletedIds) == size {
			return nil, newErrPasswordNotFound(id)
		}
		if len(deletedIds) > 1+size && !all {
			return nil, newErrAmbiguous(passwords[size:])
		}
	}
	deleted := make([]string, 0, len(deletedIds))
	for _, id := range deletedIds {
		if _, ok := box.passwords[id]; ok {
			delete(box.passwords, id)
			deleted = append(deleted, id)
		}
	}
	return deleted, box.save()
}

func (box *Box) RemoveByAccount(category, account string, all bool) ([]string, error) {
	box.Lock()
	defer box.Unlock()
	if box.masterPassword == "" {
		return nil, ErrEmptyMasterPassword
	}
	passwords := box.find(func(pw *Password) bool {
		return pw.Category == category && pw.PlainAccount == account
	})
	if len(passwords) == 0 {
		return nil, newErrPasswordNotFoundWithAccount(category, account)
	}
	if len(passwords) > 1 && !all {
		return nil, newErrAmbiguous(passwords)
	}
	ids := []string{}
	for _, pw := range passwords {
		delete(box.passwords, pw.Id)
		ids = append(ids, pw.Id)
	}
	return ids, box.save()
}

func (box *Box) Clear() ([]string, error) {
	box.Lock()
	defer box.Unlock()
	ids := make([]string, 0, len(box.passwords))
	for _, pw := range box.passwords {
		ids = append(ids, pw.Id)
		delete(box.passwords, pw.Id)
	}
	if len(ids) > 0 {
		return ids, box.save()
	}
	return ids, nil
}

func (box *Box) find(cond func(*Password) bool) []*Password {
	ret := []*Password{}
	for _, pw := range box.passwords {
		if cond(pw) {
			ret = append(ret, pw)
		}
	}
	return ret
}

func (box *Box) List(w io.Writer, noHeader bool) error {
	box.RLock()
	defer box.RUnlock()
	if box.masterPassword == "" {
		return ErrEmptyMasterPassword
	}
	if !noHeader {
		fmt.Fprintf(w, passwordFormat, "ID", "CATEGORY", "ACCOUNT", "PASSWORD", "UPDATED_AT")
		w.Write([]byte{'\n'})
	}
	for _, pw := range box.sortedPasswords() {
		pw.Brief(w, passwordFormat)
		w.Write([]byte{'\n'})
	}
	return nil
}

func (box *Box) sortedPasswords() []Password {
	passwords := make([]Password, 0, len(box.passwords))
	for _, pw := range box.passwords {
		passwords = append(passwords, *pw)
	}
	sort.Stable(passwordSlice(passwords))
	return passwords
}

func (box *Box) allocId() (string, error) {
	count := 0
	for count < 10 {
		id := md5sum(rand.Int63())
		if _, ok := box.passwords[id]; !ok {
			return id, nil
		}
	}
	return "", ErrAllocateId
}

func (box *Box) marshal() ([]byte, error) {
	for _, pw := range box.passwords {
		if err := box.encrypt(pw); err != nil {
			return nil, err
		}
	}
	passwords := box.sortedPasswords()
	return json.MarshalIndent(passwords, "", "    ")
}

func (box *Box) unmarshal(data []byte) error {
	if data == nil || len(data) == 0 {
		return nil
	}
	passwords := make([]Password, 0)
	err := json.Unmarshal(data, &passwords)
	if err != nil {
		return err
	}
	debug.Debugf("unmarshal result: %v", passwords)

	for i, _ := range passwords {
		pw := &(passwords[i])
		if box.masterPassword != "" {
			if err := box.decrypt(pw); err != nil {
				return err
			}
		}
		box.passwords[pw.Id] = pw
	}
	debug.Debugf("load result: %v", box.passwords)
	return nil
}

func (box *Box) encrypt(pw *Password) error {
	block, err := aes.NewCipher([]byte(md5sum(box.masterPassword)))
	if err != nil {
		return err
	}
	if len(pw.AccountIV) != block.BlockSize() {
		pw.AccountIV = make([]byte, block.BlockSize())
		if _, err := crand.Read(pw.AccountIV); err != nil {
			return err
		}
	}
	if len(pw.PasswordIV) != block.BlockSize() {
		pw.PasswordIV = make([]byte, block.BlockSize())
		if _, err := crand.Read(pw.PasswordIV); err != nil {
			return err
		}
	}
	pw.CipherAccount = cfbEncrypt(block, pw.AccountIV, []byte(pw.PlainAccount))
	pw.CipherPassword = cfbEncrypt(block, pw.PasswordIV, []byte(pw.PlainPassword))
	return nil
}

func (box *Box) decrypt(pw *Password) error {
	block, err := aes.NewCipher([]byte(md5sum(box.masterPassword)))
	if err != nil {
		return err
	}
	if len(pw.AccountIV) != block.BlockSize() {
		return ErrLengthOfIV
	}
	if len(pw.PasswordIV) != block.BlockSize() {
		return ErrLengthOfIV
	}
	pw.PlainAccount = string(cfbDecrypt(block, pw.AccountIV, pw.CipherAccount))
	pw.PlainPassword = string(cfbDecrypt(block, pw.PasswordIV, pw.CipherPassword))
	return nil
}

func (box *Box) DecryptAll(masterPassword string) error {
	box.masterPassword = masterPassword
	for _, pw := range box.passwords {
		err := box.decrypt(pw)
		if err != nil {
			return err
		}
	}
	return box.save()
}

// sort passwords by Id
type passwordSlice []Password

func (ps passwordSlice) Len() int           { return len(ps) }
func (ps passwordSlice) Less(i, j int) bool { return ps[i].Id < ps[j].Id }
func (ps passwordSlice) Swap(i, j int)      { ps[i], ps[j] = ps[j], ps[i] }

type passwordPtrSlice []*Password

func (ps passwordPtrSlice) Len() int           { return len(ps) }
func (ps passwordPtrSlice) Less(i, j int) bool { return ps[i].Id < ps[j].Id }
func (ps passwordPtrSlice) Swap(i, j int)      { ps[i], ps[j] = ps[j], ps[i] }

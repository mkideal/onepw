package core

import (
	"crypto/aes"
	crand "crypto/rand"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"sort"
	"strings"
	"sync"
	"time"

	"golang.org/x/crypto/scrypt"

	"github.com/mkideal/pkg/debug"
	"github.com/mkideal/pkg/textutil"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

const (
	masterPasswordID = "0"
	currentVersion   = 3
)

// BoxRepository define repo for storing passwords
type BoxRepository interface {
	Load() ([]byte, error)
	Save([]byte) error
}

type boxStore struct {
	Version   int
	Salt      []byte
	Master    Password
	Passwords []Password
}

func (store *boxStore) clear() {
	store.Passwords = store.Passwords[0:0]
}

// Box represents password box
type Box struct {
	sync.RWMutex
	masterPassword string
	repo           BoxRepository
	passwords      map[string]*Password

	store *boxStore
}

// Init initialize box with master password
func (box *Box) Init(masterPassword string) error {
	if err := CheckPassword(masterPassword); err != nil {
		return err
	}
	box.Lock()
	defer box.Unlock()
	box.masterPassword = masterPassword
	if err := box.load(); err != nil {
		return err
	}

	if err := box.encryptAll(); err != nil {
		return err
	}
	return box.save()
}

// Update updates master password
func (box *Box) Update(newMasterPassword string) error {
	box.Lock()
	defer box.Unlock()
	if err := CheckPassword(newMasterPassword); err != nil {
		return err
	}
	box.masterPassword = newMasterPassword
	if box.store.Version > 0 {
		var err error
		box.store.Master, err = box.generateMasterPasswordEntity()
		if err != nil {
			return err
		}
	}
	if err := box.encryptAll(); err != nil {
		return err
	}
	return box.save()
}

// NewBox creates box with repo
func NewBox(repo BoxRepository) *Box {
	box := &Box{
		repo:      repo,
		passwords: map[string]*Password{},
		store:     &boxStore{Version: currentVersion, Passwords: []Password{}},
	}
	return box
}

func (box *Box) generateMasterPasswordEntity() (Password, error) {
	randomAccount := make([]byte, 64)
	n, err := crand.Read(randomAccount)
	if err != nil {
		return Password{}, err
	}
	if err := box.initSalt(true); err != nil {
		return Password{}, err
	}
	dk, err := derivedKey(box.masterPassword, box.store.Salt)
	if err != nil {
		return Password{}, err
	}
	pw := NewPassword("master", string(randomAccount[:n]), string(dk), "")
	pw.ID = masterPasswordID
	return *pw, nil
}

func (box *Box) load() error {
	data, err := box.repo.Load()
	if err != nil {
		return err
	}
	if err := box.unmarshal(data); err != nil {
		return err
	}

	// decrypt master password
	if box.store.Master.ID != "" {
		if err := box.decrypt(&box.store.Master, nil); err != nil {
			return err
		}
	}

	// check something by Version
	if box.store.Version >= 1 {
		// check master password since version 1
		if box.store.Master.ID == "" {
			box.store.Master, err = box.generateMasterPasswordEntity()
			if err != nil {
				return err
			}
		} else {
			salt := box.store.Salt
			got := ""
			if salt == nil || len(salt) == 0 {
				got = sha1sum([]byte(box.masterPassword))
			} else {
				dk, err := derivedKey(box.masterPassword, box.store.Salt)
				if err != nil {
					return err
				}
				got = string(dk)
			}
			if box.store.Master.PlainPassword != got {
				return errMasterPassword
			}
		}
	}

	return nil
}

func (box *Box) save() error {
	// encrypt master password
	if box.store.Master.ID != "" {
		if err := box.encrypt(&box.store.Master, nil); err != nil {
			return err
		}
	}
	data, err := box.marshal()
	if err != nil {
		return err
	}
	debug.Debugf("marshal result: %v", string(data))
	return box.repo.Save(data)
}

// Upgrade upgrade to current version
func (box *Box) Upgrade() (from, to int, err error) {
	box.Lock()
	defer box.Unlock()

	from, to = box.store.Version, currentVersion
	box.store.Master, err = box.generateMasterPasswordEntity()
	if err != nil {
		return
	}
	box.store.Version = to
	if err = box.initSalt(false); err != nil {
		return
	}
	if err = box.encryptAll(); err != nil {
		return
	}
	err = box.save()
	return
}

// Add adds a new password to box
func (box *Box) Add(pw *Password) (id string, new bool, err error) {
	box.Lock()
	defer box.Unlock()

	debug.Debugf("add password: %v", pw)
	if box.masterPassword == "" {
		err = errEmptyMasterPassword
		return
	}
	var passwords []*Password
	if pw.ID != "" {
		passwords = box.find(func(p *Password) bool {
			return strings.HasPrefix(p.ID, pw.ID)
		})
	} else {
		passwords = []*Password{}
	}
	if len(passwords) > 1 {
		err = newErrAmbiguous(passwords)
		return
	} else if len(passwords) == 1 {
		old := passwords[0]
		old.LastUpdatedAt = time.Now().Unix()
		old.migrate(pw)
		pw = old
		new = false
	} else {
		if len(pw.ID) < shortIDLength {
			id, err = box.allocID()
			if err != nil {
				return
			}
			pw.ID = id
		}
		new = true
	}
	if err = box.encrypt(pw, nil); err != nil {
		return
	}
	box.passwords[pw.ID] = pw
	id = pw.ID
	err = box.save()
	debug.Debugf("add new password: %v", pw)
	return
}

// Remove removes passwords by ids
func (box *Box) Remove(ids []string, all bool) ([]string, error) {
	box.Lock()
	defer box.Unlock()
	if box.masterPassword == "" {
		return nil, errEmptyMasterPassword
	}
	passwords, err := box.findPasswords(ids, all)
	if err != nil {
		return nil, err
	}
	deleted := make([]string, 0, len(passwords))
	for _, pw := range passwords {
		id := pw.ID
		if _, ok := box.passwords[id]; ok {
			delete(box.passwords, id)
			deleted = append(deleted, id)
		}
	}
	return deleted, box.save()
}

func (box *Box) findPasswords(ids []string, all bool) ([]*Password, error) {
	passwords := make([]*Password, 0, len(ids))
	for _, id := range ids {
		size := len(passwords)
		if foundPw, ok := box.passwords[id]; !ok {
			for _, pw := range box.passwords {
				if strings.HasPrefix(pw.ID, id) {
					passwords = append(passwords, pw)
				}
			}
		} else {
			passwords = append(passwords, foundPw)
		}
		if len(passwords) == size {
			return nil, newErrPasswordNotFound(id)
		}
		if len(passwords) > 1+size && !all {
			return nil, newErrAmbiguous(passwords[size:])
		}
	}
	return passwords, nil
}

// RemoveByAccount removes passwords by category and account
func (box *Box) RemoveByAccount(category, account string, all bool) ([]string, error) {
	box.Lock()
	defer box.Unlock()
	if box.masterPassword == "" {
		return nil, errEmptyMasterPassword
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
		delete(box.passwords, pw.ID)
		ids = append(ids, pw.ID)
	}
	return ids, box.save()
}

// Clear clear password box
func (box *Box) Clear() ([]string, error) {
	box.Lock()
	defer box.Unlock()
	ids := make([]string, 0, len(box.passwords))
	for _, pw := range box.passwords {
		ids = append(ids, pw.ID)
		delete(box.passwords, pw.ID)
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

// List writes all passwords to specified writer
func (box *Box) List(w io.Writer, noHeader, showHidden bool) error {
	box.RLock()
	defer box.RUnlock()
	if box.masterPassword == "" {
		return errEmptyMasterPassword
	}
	var table textutil.Table
	table = passwordSlice(box.sortedPasswords(showHidden))
	if !noHeader {
		table = textutil.AddTableHeader(table, passwordHeader)
	}
	textutil.WriteTable(w, table, box.colorID(w, !noHeader))
	return nil
}

// Inspect show low-level information of password
func (box *Box) Inspect(w io.Writer, ids []string, all bool) error {
	passwords, err := box.findPasswords(ids, all)
	if err != nil {
		return err
	}
	sort.Sort(passwordPtrSlice(passwords))
	prefix := "    "
	fmt.Fprintf(w, "[\n%s", prefix)
	for i, pw := range passwords {
		if i != 0 {
			fmt.Fprintf(w, ",\n%s", prefix)
		}
		pw.inspect(w, prefix)
	}
	fmt.Fprintf(w, "\n]\n")
	return nil
}

// Find finds password by word
func (box *Box) Find(w io.Writer, word string, justPassword, justFirst bool) error {
	box.RLock()
	defer box.RUnlock()

	if box.masterPassword == "" {
		return errEmptyMasterPassword
	}
	table := passwordPtrSlice(box.find(func(pw *Password) bool {
		return pw.match(word)
	}))
	if len(table) == 0 {
		return nil
	}
	sort.Sort(table)
	if justFirst {
		table = table[:1]
	}
	if justPassword {
		for _, pw := range table {
			fmt.Fprintf(w, "%s\n", pw.PlainPassword)
		}
		return nil
	}
	var t textutil.Table
	t = table
	if !justFirst {
		t = textutil.AddTableHeader(table, passwordHeader)
	}
	textutil.WriteTable(w, t, box.colorID(w, !justFirst))
	return nil
}

// color ID style
type colorIDStyle struct {
	textutil.DefaultStyle
	hasHeader bool
}

func (style colorIDStyle) CellRender(row, col int, cell string, cw *textutil.ColorWriter) {
	if col != 0 || (row == 0 && style.hasHeader) {
		style.DefaultStyle.CellRender(row, col, cell, cw)
	} else {
		fmt.Fprint(cw, cw.Cyan(cell))
	}
}

func (box *Box) colorID(w io.Writer, hasHeader bool) textutil.TableStyler {
	return colorIDStyle{hasHeader: hasHeader}
}

func (box *Box) sortedPasswords(showHidden bool) []Password {
	passwords := make([]Password, 0, len(box.passwords))
	for _, pw := range box.passwords {
		if showHidden || !pw.Hidden {
			passwords = append(passwords, *pw)
		}
	}
	sort.Sort(passwordSlice(passwords))
	return passwords
}

func (box *Box) allocID() (string, error) {
	count := 0
	for count < 10 {
		id := md5sum(rand.Int63())
		if _, ok := box.passwords[id]; !ok {
			return id, nil
		}
		count++
	}
	return "", errAllocateID
}

func (box *Box) marshal() ([]byte, error) {
	if err := box.encryptAll(); err != nil {
		return nil, err
	}
	box.store.Passwords = box.sortedPasswords(true)
	return json.MarshalIndent(box.store, "", "    ")
}

func (box *Box) unmarshal(data []byte) error {
	if data == nil || len(data) == 0 {
		return nil
	}
	box.store.clear()
	err := json.Unmarshal(data, box.store)
	if err != nil {
		box.store.Version = 0
		err = json.Unmarshal(data, &box.store.Passwords)
		if err != nil {
			return err
		}
	}

	for i := range box.store.Passwords {
		pw := &(box.store.Passwords[i])
		box.passwords[pw.ID] = pw
	}
	if box.masterPassword != "" {
		if err := box.decryptAll(); err != nil {
			return err
		}
	}
	return nil
}

func derivedKey(password string, salt []byte) ([]byte, error) {
	if salt == nil || len(salt) == 0 {
		// Deprecated: insecure
		return []byte(md5sum([]byte(password))), nil
	}
	return scrypt.Key([]byte(password), salt, 4096, 8, 1, 32)
}

func (box *Box) initSalt(reinit bool) error {
	salt := box.store.Salt
	if salt == nil || len(salt) == 0 || reinit {
		salt = make([]byte, 64)
		if _, err := crand.Read(salt); err != nil {
			return err
		}
		box.store.Salt = salt
	}
	return nil
}

func (box *Box) encryptAll() error {
	dk, err := derivedKey(box.masterPassword, box.store.Salt)
	if err != nil {
		return err
	}
	for _, pw := range box.passwords {
		if err = box.encrypt(pw, dk); err != nil {
			return err
		}
	}
	return nil
}

func (box *Box) encrypt(pw *Password, dk []byte) error {
	if dk == nil {
		var err error
		dk, err = derivedKey(box.masterPassword, box.store.Salt)
		if err != nil {
			return err
		}
	}
	block, err := aes.NewCipher(dk)
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

func (box *Box) decryptAll() error {
	dk, err := derivedKey(box.masterPassword, box.store.Salt)
	if err != nil {
		return err
	}
	for _, pw := range box.passwords {
		if err = box.decrypt(pw, dk); err != nil {
			return err
		}
	}
	return nil
}

func (box *Box) decrypt(pw *Password, dk []byte) error {
	if dk == nil {
		var err error
		dk, err = derivedKey(box.masterPassword, box.store.Salt)
		if err != nil {
			return err
		}
	}
	dk, err := derivedKey(box.masterPassword, box.store.Salt)
	if err != nil {
		return err
	}
	block, err := aes.NewCipher(dk)
	if err != nil {
		return err
	}
	if len(pw.AccountIV) != block.BlockSize() {
		debug.Panicf("%s: AccountIV.length=%d, want %d", pw.ID, len(pw.AccountIV), block.BlockSize())
		return errLengthOfIV
	}
	if len(pw.PasswordIV) != block.BlockSize() {
		debug.Panicf("%s: PasswordIV.length=%d, want %d", pw.ID, len(pw.PasswordIV), block.BlockSize())
		return errLengthOfIV
	}
	pw.PlainAccount = string(cfbDecrypt(block, pw.AccountIV, pw.CipherAccount))
	pw.PlainPassword = string(cfbDecrypt(block, pw.PasswordIV, pw.CipherPassword))
	return nil
}

func shorten(s string, n int) string {
	var padding = " ..."
	var paddingSize = len(padding)
	if n <= paddingSize {
		n = paddingSize + 1
	}
	if len(s) > n {
		return s[:n-paddingSize] + padding
	}
	return s
}

// sort passwords by Id
type passwordSlice []Password

func (ps passwordSlice) Len() int           { return len(ps) }
func (ps passwordSlice) Less(i, j int) bool { return ps[i].ID < ps[j].ID }
func (ps passwordSlice) Swap(i, j int)      { ps[i], ps[j] = ps[j], ps[i] }
func (ps passwordSlice) RowCount() int      { return ps.Len() }
func (ps passwordSlice) ColCount() int {
	if ps.Len() == 0 {
		return 0
	}
	return ps[0].colCount()
}
func (ps passwordSlice) Get(i, j int) string {
	return shorten(ps[i].get(j), 32)
}

type passwordPtrSlice []*Password

func (ps passwordPtrSlice) Len() int           { return len(ps) }
func (ps passwordPtrSlice) Less(i, j int) bool { return ps[i].ID < ps[j].ID }
func (ps passwordPtrSlice) Swap(i, j int)      { ps[i], ps[j] = ps[j], ps[i] }
func (ps passwordPtrSlice) RowCount() int      { return ps.Len() }
func (ps passwordPtrSlice) ColCount() int {
	if ps.Len() == 0 {
		return 0
	}
	return ps[0].colCount()
}
func (ps passwordPtrSlice) Get(i, j int) string {
	return shorten(ps[i].get(j), 32)
}

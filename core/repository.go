package core

import (
	"io/ioutil"
)

// fileRepository implements BoxRepository interface
type fileRepository struct {
	filename string
}

// NewFileRepository creates a FileRepository
func NewFileRepository(filename string) BoxRepository {
	return &fileRepository{filename: filename}
}

// Load implements BoxRepository.Load method
func (repo *fileRepository) Load() ([]byte, error) {
	return ioutil.ReadFile(repo.filename)
}

// Save implements BoxRepository.Save method
func (repo *fileRepository) Save(data []byte) error {
	return ioutil.WriteFile(repo.filename, data, 0666)
}

// memRepository implements BoxRepository interface
// NOTE: only used to test
type memRepository struct {
	data []byte
}

func NewMemRepository(data []byte) BoxRepository {
	return &memRepository{data: data}
}

func (repo *memRepository) Load() ([]byte, error) {
	return repo.data, nil
}

func (repo *memRepository) Save(data []byte) error {
	repo.data = data
	return nil
}

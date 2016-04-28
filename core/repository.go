package core

import (
	"io/ioutil"
)

// FileRepository implements BoxRepository interface
type FileRepository struct {
	Filename string
}

// NewFileRepository creates a FileRepository
func NewFileRepository(filename string) *FileRepository {
	return &FileRepository{Filename: filename}
}

// Load implements BoxRepository.Load method
func (repo *FileRepository) Load() ([]byte, error) {
	return ioutil.ReadFile(repo.Filename)
}

// Save implements BoxRepository.Save method
func (repo *FileRepository) Save(data []byte) error {
	return ioutil.WriteFile(repo.Filename, data, 0666)
}

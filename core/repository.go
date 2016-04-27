package core

import (
	"io/ioutil"
)

type FileRepository struct {
	Filename string
}

func NewFileRepository(filename string) *FileRepository {
	return &FileRepository{Filename: filename}
}

func (repo *FileRepository) Load() ([]byte, error) {
	return ioutil.ReadFile(repo.Filename)
}

func (repo *FileRepository) Save(data []byte) error {
	return ioutil.WriteFile(repo.Filename, data, 0666)
}

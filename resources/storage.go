package resources

import (
	"archive/zip"
	"bytes"
	"io/fs"
)

var data []byte

type Storage struct {
	r *zip.Reader
}

func NewStorage() *Storage {
	reader := bytes.NewReader(data)

	zipReader, err := zip.NewReader(reader, int64(reader.Len()))
	if err != nil {
		panic(err)
	}

	return &Storage{
		r: zipReader,
	}
}

func (r *Storage) Get(path string) (fs.File, error) {
	return r.r.Open(path)
}

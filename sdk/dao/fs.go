package dao

import (
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"
)

// FileSystemDAO handles basic file operations in a given base directory
type FileSystemDAO struct {
	BaseDir string
}

// NewFileSystemDAO creates a new instance with the given base directory
func NewFileSystemDAO(baseDir string) (*FileSystemDAO, error) {
	if _, err := os.Stat(baseDir); os.IsNotExist(err) {
		return nil, errors.New("base directory does not exist")
	}
	return &FileSystemDAO{BaseDir: baseDir}, nil
}

// ListFiles returns a list of all files (not folders) in the base directory (recursive)
func (dao *FileSystemDAO) List() ([]string, error) {
	var files []string
	err := filepath.Walk(dao.BaseDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			relPath, _ := filepath.Rel(dao.BaseDir, path)
			files = append(files, relPath)
		}
		return nil
	})
	return files, err
}

// GetFileContent returns the content of the given relative file path
func (dao *FileSystemDAO) Instace(fileName string) (string, error) {
	fullPath := filepath.Join(dao.BaseDir, fileName)
	data, err := os.ReadFile(fullPath)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// CreateFile creates a new file with the given content. Fails if file exists.
func (dao *FileSystemDAO) Create(fileName string, content string) error {
	fullPath := filepath.Join(dao.BaseDir, fileName)
	if _, err := os.Stat(fullPath); err == nil {
		return errors.New("file already exists")
	}
	err := os.MkdirAll(filepath.Dir(fullPath), os.ModePerm)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(fullPath, []byte(content), 0644)
}

// UpdateFile updates an existing file with new content. Fails if file does not exist.
func (dao *FileSystemDAO) Update(fileName string, content string) error {
	fullPath := filepath.Join(dao.BaseDir, fileName)
	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		return errors.New("file does not exist")
	}
	return ioutil.WriteFile(fullPath, []byte(content), 0644)
}

// DeleteFile deletes the specified file
func (dao *FileSystemDAO) Delete(fileName string) error {
	fullPath := filepath.Join(dao.BaseDir, fileName)
	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		return errors.New("file does not exist")
	}
	return os.Remove(fullPath)
}

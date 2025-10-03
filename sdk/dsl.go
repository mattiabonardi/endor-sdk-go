package sdk

import (
	"errors"
	"os"
	"path/filepath"
)

type DSLDAO struct {
	BasePath string
}

func NewDSLDAO(username string, development bool) *DSLDAO {
	homeDir, _ := os.UserHomeDir()
	var path string
	if development {
		path = filepath.Join(homeDir, "etc", "endor", "dsl", username)
	} else {
		path = filepath.Join(homeDir, "etc", "endor", "dsl", "_root")
	}
	os.MkdirAll(path, 0755)
	return &DSLDAO{
		BasePath: path,
	}
}

// ListAll returns a list of all files and folders (relative paths) in the base directory (recursive)
func (dao *DSLDAO) ListAll() ([]string, error) {
	var items []string
	err := filepath.Walk(dao.BasePath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if path == dao.BasePath {
			return nil
		}
		relPath, _ := filepath.Rel(dao.BasePath, path)
		items = append(items, relPath)
		return nil
	})
	return items, err
}

// ListFiles returns a list of all files (not folders) in the base directory (recursive)
func (dao *DSLDAO) ListFile() ([]string, error) {
	var files []string
	err := filepath.Walk(dao.BasePath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			relPath, _ := filepath.Rel(dao.BasePath, path)
			files = append(files, relPath)
		}
		return nil
	})
	return files, err
}

// ListFolders returns a list of all folders (not files) in the base directory (recursive)
func (dao *DSLDAO) ListFolders() ([]string, error) {
	var folders []string
	err := filepath.Walk(dao.BasePath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			// Skip the base directory itself if you don't want it in the list
			if path == dao.BasePath {
				return nil
			}
			relPath, _ := filepath.Rel(dao.BasePath, path)
			folders = append(folders, relPath)
		}
		return nil
	})
	return folders, err
}

// GetFileContent returns the content of the given relative file path
func (dao *DSLDAO) Instace(fileName string) (string, error) {
	fullPath := filepath.Join(dao.BasePath, fileName)
	data, err := os.ReadFile(fullPath)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// CreateFile creates a new file with the given content. Fails if file exists.
func (dao *DSLDAO) Create(fileName string, content string) error {
	fullPath := filepath.Join(dao.BasePath, fileName)
	if _, err := os.Stat(fullPath); err == nil {
		return errors.New("file already exists")
	}
	err := os.MkdirAll(filepath.Dir(fullPath), os.ModePerm)
	if err != nil {
		return err
	}
	return os.WriteFile(fullPath, []byte(content), 0644)
}

// UpdateFile updates an existing file with new content. Fails if file does not exist.
func (dao *DSLDAO) Update(fileName string, content string) error {
	fullPath := filepath.Join(dao.BasePath, fileName)
	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		return errors.New("file does not exist")
	}
	return os.WriteFile(fullPath, []byte(content), 0644)
}

// DeleteFile deletes the specified file
func (dao *DSLDAO) Delete(fileName string) error {
	fullPath := filepath.Join(dao.BasePath, fileName)
	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		return errors.New("file does not exist")
	}
	return os.Remove(fullPath)
}

// Rename renames a file or folder from oldName to newName (relative to BasePath)
func (dao *DSLDAO) Rename(oldName, newName string) error {
	oldPath := filepath.Join(dao.BasePath, oldName)
	newPath := filepath.Join(dao.BasePath, newName)

	if _, err := os.Stat(oldPath); os.IsNotExist(err) {
		return errors.New("source file or folder does not exist")
	}
	if _, err := os.Stat(newPath); err == nil {
		return errors.New("destination file or folder already exists")
	}

	err := os.MkdirAll(filepath.Dir(newPath), os.ModePerm)
	if err != nil {
		return err
	}

	return os.Rename(oldPath, newPath)
}

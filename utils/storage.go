package utils

import (
	"errors"
	"os"
	"path"
	"path/filepath"
	"strings"
)

type LocalStorage struct {
}

func (st LocalStorage) GetFiles(dirname string) ([]string, error) {
	fullPath := path.Join(rootDir, dirname)
	contents, err := os.ReadDir(fullPath)
	if err != nil {
		return nil, errors.New("directory does not exist")
	}
	var files []string
	for i := range contents {
		if !contents[i].IsDir() {
			files = append(files, contents[i].Name())
		}
	}
	return files, nil

}

func (st LocalStorage) TransformName(dirname, filename string) (string, error) {
	fullPath := dirname
	contents, err := os.ReadDir(path.Join(rootDir, fullPath))
	if err != nil {
		return "", err
	}

	for i := range contents {
		if contents[i].IsDir() && contents[i].Name() == convertedDir {
			fullPath = path.Join(fullPath, convertedDir)
			break
		}
	}

	files, err := st.GetFiles(fullPath)

	if err != nil {
		return "", err
	}

	namePart := strings.Split(filename, ".")[0]
	for i := range files {
		if strings.Contains(filepath.Base(files[i]), namePart) {
			return path.Join(rootDir, dirname, convertedDir, files[i]), nil
		}
	}

	resultName := path.Join(rootDir, dirname, dirname+".pdf")
	return resultName, nil
}

func (st LocalStorage) Retrieve(dirname string) (string, error) {
	ext := ".pdf"
	fullPathEntry := path.Join(rootDir, dirname)
	if directoryExists(path.Join(fullPathEntry, convertedDir)) {
		target := path.Join(fullPathEntry, convertedDir)
		ext = ".zip"
		output := path.Join(rootDir, dirname, dirname+ext)
		ZipEntry(target, output)
	}
	fullPathFileName := path.Join(fullPathEntry, dirname+ext)
	exists := fileExists(fullPathFileName)
	if !exists {
		return "", errors.New("document does not exist")
	}

	return fullPathFileName, nil
}
func (st LocalStorage) GetFilename(dirname, filename string) string {
	fullPathEntry := path.Join(rootDir, dirname, filename)
	return fullPathEntry

}

func directoryExists(dirPath string) bool {
	info, err := os.Stat(dirPath)
	if os.IsNotExist(err) {
		return false // Directory does not exist
	}
	return err == nil && info.IsDir() // Exists and is a directory
}

func fileExists(filePath string) bool {
	_, err := os.Stat(filePath)
	if os.IsNotExist(err) {
		return false
	}
	return err == nil
}

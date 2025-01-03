package utils

import (
	"archive/zip"
	"io"
	"os"
	"path/filepath"

	"golang.org/x/crypto/bcrypt"
)

const (
	rootDir      string = "storage"
	convertedDir string = "converted"
	uuidRegex    string = "[a-f0-9]{8}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{12}"
	issuer       string = "conversion_api"
)

type Generator interface {
	Generate(userId string, username, role string) (string, error)
	Validate(providedToken string) (*CustomClaims, error)
}

type Storage interface {
	Retrieve(dirname string) (string, error)
	GetFilename(dirname, filename string) string
	TransformName(dirname, filename string) (string, error)
	GetFiles(dirname string) ([]string, error)
	FileExists(filePath string) bool
	DirectoryExists(dirPath string) bool
}

type Converter interface {
	convert(name string, outputDir string) error
	Convert(filenames []string, outputDir string, done chan bool)
	Merge(filenames []string, outputFile string, done chan bool)
	GetPageCount(fullName string) (int32, error)
}


func HashedPassword(password string) (string, error) {
	hashed, err := bcrypt.GenerateFromPassword([]byte(password), 10)
	return string(hashed), err
}

func IsPasswordValid(providedPassword string, hashedPassword string) error {

	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(providedPassword))
}

func zipDir(source, target string) error {
	
	f, err := os.Create(target)
	if err != nil {
		return err
	}
	defer f.Close()

	writer := zip.NewWriter(f)
	defer writer.Close()

	return filepath.Walk(source, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		header, err := zip.FileInfoHeader(info)
		if err != nil {
			return err
		}
		header.Method = zip.Deflate
		header.Name, err = filepath.Rel(filepath.Dir(source), path)
		if err != nil {
			return err
		}
		if info.IsDir() {
			header.Name += "/"
		}
		headerWriter, err := writer.CreateHeader(header)
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		f, err := os.Open(path)
		if err != nil {
			return err
		}
		defer f.Close()

		_, err = io.Copy(headerWriter, f)
		return err
	})
}

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
}

type Converter interface {
	convert(name string, outputDir string, done chan error)
	Convert(filenames []string, outputDir string, done chan bool)
	Merge(filenames []string, outputFile string, done chan bool)
	GetPageCount(fullName string) (int32, error)
}

type ConversionModel struct {
	Name      string
	Content   []byte
	PageCount int
}

func HashedPassword(password string) (string, error) {
	hashed, err := bcrypt.GenerateFromPassword([]byte(password), 10)
	return string(hashed), err
}

func IsPasswordValid(providedPassword string, hashedPassword string) error {

	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(providedPassword))
}

func ZipEntry(source, target string) error {
	// 1. Create a ZIP file and zip.Writer
	f, err := os.Create(target)
	if err != nil {
		return err
	}
	defer f.Close()

	writer := zip.NewWriter(f)
	defer writer.Close()

	// 2. Go through all the files of the source
	return filepath.Walk(source, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// 3. Create a local file header
		header, err := zip.FileInfoHeader(info)
		if err != nil {
			return err
		}

		// set compression
		header.Method = zip.Deflate

		// 4. Set relative path of a file as the header name
		header.Name, err = filepath.Rel(filepath.Dir(source), path)
		if err != nil {
			return err
		}
		if info.IsDir() {
			header.Name += "/"
		}

		// 5. Create writer for the file header and save content of the file
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

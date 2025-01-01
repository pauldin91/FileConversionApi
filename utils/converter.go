package utils

import (
	"os"
	"os/exec"
	"path"
	"sync"

	"github.com/google/uuid"
	"github.com/pdfcpu/pdfcpu/pkg/api"
)

type Converter interface {
	convert(contents []byte, name string, entryId uuid.UUID, done chan error)
	Convert(files map[string][]byte, entryId uuid.UUID, done chan bool)
	Merge(files map[string][]byte, entryId uuid.UUID, done chan bool)
	GetPageCount(name string, entryId uuid.UUID) (int32, error)
}

type ConversionModel struct {
	Name      string
	Content   []byte
	PageCount int
}

type PdfConverter struct{}

func (conv PdfConverter) GetPageCount(name string, entryId uuid.UUID) (int32, error) {
	fullName := path.Join(entryId.String(), name)
	pageCount, err := api.PageCountFile(fullName)
	if err != nil {
		return 0, err
	}
	return int32(pageCount), nil
}

func (conv PdfConverter) Merge(files map[string][]byte, entryId uuid.UUID, done chan bool) {
	// Step 1: Write each []byte to a temporary file
	var tempFiles []string
	for name, content := range files {
		tempFileName := path.Join(entryId.String(), name)
		err := os.WriteFile(tempFileName, content, 0755)
		if err != nil {
			done <- false // Send error back via channel
			return
		}
		tempFiles = append(tempFiles, tempFileName)
	}

	// Step 2: Merge the PDF files
	mergedFile := path.Join(entryId.String(), entryId.String()+".pdf")
	err := api.MergeCreateFile(tempFiles, mergedFile, false, nil)
	if err != nil {
		done <- false
		return
	}

	done <- true // No error, successful
}

func (conv PdfConverter) Convert(files map[string][]byte, entryId uuid.UUID, done chan bool) {
	var wg sync.WaitGroup
	errs := make(chan error, len(files))

	// Step 1: Iterate over files and start goroutines for each conversion
	for name, contents := range files {
		wg.Add(1)
		go func(name string, contents []byte) {
			defer wg.Done()
			conv.convert(contents, name, entryId, errs)
		}(name, contents)
	}

	// Wait for all goroutines to finish
	wg.Wait()

	// Close error channel after all goroutines are done
	close(errs)

	// If any error occurred, return the first one
	for err := range errs {
		if err != nil {
			done <- false
			return
		}
	}

	done <- true // No error, successful
}

func (conv PdfConverter) convert(contents []byte, name string, entryId uuid.UUID, done chan error) {
	// Create a directory for the entryId if it doesn't exist
	err := os.MkdirAll(entryId.String(), 0755) // Use MkdirAll to ensure the path exists
	if err != nil {
		done <- err
		return
	}

	fullName := path.Join(entryId.String(), name)
	// Step 1: Save the []byte to a temporary file
	err = os.WriteFile(fullName, contents, 0644)
	if err != nil {
		done <- err
		return
	}
	defer os.Remove(fullName) // Clean up the temporary file

	// Step 2: Use LibreOffice to convert the file to PDF
	cmd := exec.Command("libreoffice", "--headless", "--convert-to", "pdf", fullName)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err = cmd.Run()
	if err != nil {
		done <- err
		return
	}

	// Notify that the conversion is successful
	done <- nil
}

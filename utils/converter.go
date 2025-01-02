package utils

import (
	"fmt"
	"os"
	"os/exec"
	"path"

	"github.com/pdfcpu/pdfcpu/pkg/api"
)

type PdfConverter struct {
}

func (conv PdfConverter) GetPageCount(fullName string) (int32, error) {
	pageCount, err := api.PageCountFile(fullName)
	if err != nil {
		return 0, err
	}
	return int32(pageCount), nil
}

func (conv PdfConverter) Merge(filenames []string, outputFile string, done chan bool) {

	err := api.MergeCreateFile(filenames, outputFile, false, nil)
	if err != nil {
		done <- false
		return
	}

	done <- true
}

func (conv PdfConverter) Convert(filenames []string, outputDir string, done chan bool) {

	doneChan := make([]chan bool, len(filenames))
	for i := range filenames {
		doneChan[i] = make(chan bool)
		go conv.convert(filenames[i], outputDir, doneChan[i])
	}

	for i := range doneChan {
		<-doneChan[i]
	}

	done <- true
}

func (conv PdfConverter) convert(name string, outputDir string, done chan bool) {

	finalOutputDir := path.Join(rootDir, outputDir, convertedDir)
	filename := path.Join(rootDir, outputDir, name)

	cmd := exec.Command("libreoffice", "--headless", "--convert-to", "pdf", "--outdir", finalOutputDir, filename)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Run()
	if err != nil {
		fmt.Println(err.Error())
		done <- false
		return
	}

	done <- true
}

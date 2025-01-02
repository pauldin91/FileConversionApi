package utils

import (
	"fmt"
	"os"
	"os/exec"
	"path"
	"sync"

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

	var wg sync.WaitGroup
	errs := make(chan error, len(filenames))

	for i := range filenames {
		wg.Add(1)
		go func(name string) {
			defer wg.Done()
			conv.convert(name, outputDir, errs)
		}(filenames[i])
	}

	wg.Wait()

	close(errs)

	for err := range errs {
		if err != nil {
			done <- false
			return
		}
	}

	done <- true
}

func (conv PdfConverter) convert(name string, outputDir string, done chan error) {

	finalOutputDir := path.Join(rootDir, outputDir, convertedDir)
	filename := path.Join(rootDir, outputDir, name)

	cmd := exec.Command("libreoffice", "--headless", "--convert-to", "pdf", "--outdir", finalOutputDir, filename)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Run()
	if err != nil {
		fmt.Println(err.Error())
		done <- err
		return
	}

	done <- nil
}

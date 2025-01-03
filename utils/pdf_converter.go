package utils

import (
	"os"
	"os/exec"
	"path"

	"github.com/pdfcpu/pdfcpu/pkg/api"
	"github.com/rs/zerolog/log"
)

type PdfConverter struct{}

func (conv PdfConverter) GetPageCount(fullName string) (int32, error) {
	pageCount, err := api.PageCountFile(fullName)
	if err != nil {
		return 0, err
	}
	return int32(pageCount), nil
}

func (conv PdfConverter) Merge(filenames []string, outputFile string, done chan bool) {
	finalOutputFile := path.Join(rootDir, outputFile, outputFile+".pdf")
	err := api.MergeCreateFile(filenames, finalOutputFile, false, nil)
	if err != nil {
		log.Error().Msgf("PDF merge failed: %v", err)
		done <- false
		return
	}
	log.Info().Msgf("Merged PDF successfully created: %s", outputFile)
	done <- true
}

func (conv PdfConverter) Convert(filenames []string, outputDir string, done chan bool) {
	errChan := make(chan error, len(filenames))

	for _, filename := range filenames {
		if err := conv.convert(filename, outputDir); err != nil {
			errChan <- err
		}
	}
	close(errChan)

	if len(errChan) > 0 {
		for err := range errChan {
			log.Error().Msgf("Error during conversion: %v", err)
		}
		done <- false
		return
	}

	log.Info().Msg("All files converted successfully")
	done <- true
}

func (conv PdfConverter) convert(name string, outputDir string) error {
	finalOutputDir := path.Join(rootDir, outputDir, convertedDir)

	// Ensure the output directory exists
	if err := os.MkdirAll(finalOutputDir, 0755); err != nil {
		log.Error().Msgf("Failed to create directory %s: %v", finalOutputDir, err)
		return err
	}

	cmd := exec.Command("libreoffice", "--headless", "--convert-to", "pdf", "--outdir", finalOutputDir, name)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// Run the command and capture errors
	if err := cmd.Run(); err != nil {
		log.Error().Msgf("LibreOffice conversion failed for %s: %v", name, err)
		return err
	}

	// Log success
	log.Info().Msgf("File converted successfully: %s", name)
	return nil
}

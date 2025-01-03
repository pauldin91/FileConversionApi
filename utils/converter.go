package utils

type Converter interface {
	convert(name string, outputDir string) error
	Convert(filenames []string, outputDir string, done chan bool)
	Merge(filenames []string, outputFile string, done chan bool)
	GetPageCount(fullName string) (int32, error)
}
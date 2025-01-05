package utils

type Storage interface {
	Retrieve(dirname string) (string, error)
	GetFilename(dirname, filename string) string
	GetConvertedFilename(dirname, filename string) (string, error)
	GetFiles(dirname string) ([]string, error)
	FileExists(filePath string) bool
	DirectoryExists(dirPath string) bool
}

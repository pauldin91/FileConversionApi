package utils

type Status string

const (
	rootDir      string = "storage"
	convertedDir string = "converted"
	uuidRegex    string = "[a-f0-9]{8}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{12}"
	issuer       string = "conversion_api"
	Failed       Status = "failed"
	Processing   Status = "processing"
	Success      Status = "success"
)
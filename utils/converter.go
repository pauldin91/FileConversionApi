package utils

type Converter interface {
	Convert(files []ConversionModel) ([]ConversionModel, error)
	Merge(files []ConversionModel) (ConversionModel, error)
}

type ConversionModel struct {
	Name    string
	Content []byte
	Pages   []int
}

/*
type PdfConverter struct {
}

func (conv *PdfConverter) Merge(files []ConversionModel) (ConversionModel, error) {
	file, err := model.NewPdfReaderFromFile("example.pdf", nil)
	if err != nil {
		log.Fatalf("Error opening PDF: %v", err)
	}

	// Extract text from each page
	numPages, err := file.GetNumPages()
	if err != nil {
		log.Fatalf("Error getting page count: %v", err)
	}

	for i := 1; i <= numPages; i++ {
		page, err := file.GetPage(i)
		if err != nil {
			log.Fatalf("Error getting page: %v", err)
		}

		text, err := page.GetText()
		if err != nil {
			log.Fatalf("Error extracting text: %v", err)
		}

		fmt.Printf("Page %d:\n%s\n", i, text)
	}
}*/

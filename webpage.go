package main

import "github.com/SebastiaanKlippert/go-wkhtmltopdf"

// webpageAsPDF fetches the html corresponding to a url
// and then generate a PDF file of the desired website.
func webpageAsPDF(url string) ([]byte, error) {
	pdfg, err := wkhtmltopdf.NewPDFGenerator()
	if err != nil {
		return nil, err
	}

	page := wkhtmltopdf.NewPage(url)

	pdfg.AddPage(page)

	// Create PDF document in internal buffer
	err = pdfg.Create()
	if err != nil {
		return nil, err
	}

	return pdfg.Bytes(), nil
}

package main

import (
	"fmt"
	"io"
	"path/filepath"

	"github.com/juruen/rmapi/archive"
)

// uploadToRm sends the file to the Remarkable device.
func (s server) uploadToRm(id string, file []byte, name string) error {
	zip := archive.NewZip()
	zip.UUID = id

	fname, ext := splitExt(name)

	switch ext {

	case ".epub":
		zip.Epub = file
		zip.Content.FileType = "epub"
	case ".pdf":
		zip.Pdf = file
		zip.Content.FileType = "pdf"
	default:
		return fmt.Errorf("file does not have a supported extension")

	}

	// We use an io.Pipe to connect the io.Writer of the archive
	// package to the io.Reader needed by the api package.
	pr, pw := io.Pipe()

	// Write the zip file to the pipe.
	// Writing without a reader will deadlock so write in a goroutine.
	go func() {
		pw.CloseWithError(zip.Write(pw))
	}()

	// Upload to the Remarkable from the pipe.
	if err := s.cli.Upload(id, fname, pr); err != nil {
		return err
	}

	return nil
}

// deleteFromRm removes the file from the Remarkable device.
func (s server) deleteFromRm(id string) error {
	return s.cli.Delete(id)
}

// splitExt splits the extension from a filename
func splitExt(name string) (string, string) {
	ext := filepath.Ext(name)
	return name[0 : len(name)-len(ext)], ext
}

package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/tom-hoover/darkroom/imaging"
)

// Convert decodes the HEIC file at src and writes it to dst as a JPEG at
// the given quality, carrying the source's EXIF metadata across.
//
// The output is written to a temporary file and renamed into place, so an
// interrupted run never leaves a truncated JPEG that a later run would
// mistake for finished work.
func Convert(src, dst string, quality int) error {
	if quality < 1 || quality > 100 {
		return fmt.Errorf("quality %d out of range (1-100)", quality)
	}

	img, exif, err := imaging.Decode(src)
	if err != nil {
		return err
	}
	if len(exif) > imaging.MaxExifPayload {
		warnf("%s: EXIF block of %d bytes does not fit in a JPEG, dropping it", src, len(exif))
		exif = nil
	}

	dir := filepath.Dir(dst)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	tmp, err := os.CreateTemp(dir, ".heic2jpg-*.tmp")
	if err != nil {
		return err
	}
	// Removing the temporary file is a no-op once it has been renamed away.
	defer func() {
		tmp.Close()
		os.Remove(tmp.Name())
	}()

	if err := imaging.WriteJPEG(tmp, img, quality, exif); err != nil {
		return fmt.Errorf("encode: %w", err)
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	// CreateTemp makes the file private; photos should be readable.
	if err := os.Chmod(tmp.Name(), 0o644); err != nil {
		return err
	}
	return os.Rename(tmp.Name(), dst)
}

package main

import (
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

// Job is one HEIC file and the JPEG it should become.
type Job struct {
	Src string
	Dst string
}

// Scan lists the conversions to perform for the HEIC files in dir. It
// descends into subdirectories only when recursive is set.
//
// With an empty outDir each JPEG lands beside its source; otherwise the
// source tree's shape is mirrored under outDir.
func Scan(dir, outDir string, recursive bool) ([]Job, error) {
	info, err := os.Stat(dir)
	if err != nil {
		return nil, err
	}
	if !info.IsDir() {
		return nil, &os.PathError{Op: "scan", Path: dir, Err: os.ErrInvalid}
	}

	// An output directory inside the scanned tree would otherwise feed
	// its own contents back in on a later run.
	skipDir := ""
	if outDir != "" {
		if skipDir, err = filepath.Abs(outDir); err != nil {
			return nil, err
		}
	}

	var jobs []Job
	add := func(path string) error {
		dst, err := destination(dir, outDir, path, recursive)
		if err != nil {
			return err
		}
		jobs = append(jobs, Job{Src: path, Dst: dst})
		return nil
	}

	if !recursive {
		entries, err := os.ReadDir(dir)
		if err != nil {
			return nil, err
		}
		for _, e := range entries {
			if e.IsDir() || !isHEIC(e.Name()) {
				continue
			}
			if err := add(filepath.Join(dir, e.Name())); err != nil {
				return nil, err
			}
		}
		return jobs, nil
	}

	err = filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			if abs, err := filepath.Abs(path); err == nil && abs == skipDir {
				return fs.SkipDir
			}
			return nil
		}
		if !isHEIC(d.Name()) {
			return nil
		}
		return add(path)
	})
	if err != nil {
		return nil, err
	}
	return jobs, nil
}

// destination is the JPEG path for a given source file.
func destination(dir, outDir, src string, recursive bool) (string, error) {
	name := strings.TrimSuffix(filepath.Base(src), filepath.Ext(src)) + ".jpg"
	if outDir == "" {
		return filepath.Join(filepath.Dir(src), name), nil
	}
	if !recursive {
		return filepath.Join(outDir, name), nil
	}
	rel, err := filepath.Rel(dir, filepath.Dir(src))
	if err != nil {
		return "", err
	}
	return filepath.Join(outDir, rel, name), nil
}

func isHEIC(name string) bool {
	switch strings.ToLower(filepath.Ext(name)) {
	case ".heic", ".heif":
		return true
	}
	return false
}

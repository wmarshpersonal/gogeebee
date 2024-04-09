package main

import (
	"archive/zip"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
)

func openROMFile(path string) ([]byte, error) {
	if zr, err := zip.OpenReader(path); err == nil {
		slog.Debug("looks like a zip file", "path", path)
		defer zr.Close()
		if romData, err := openZippedROM(zr); err != nil {
			return nil, err
		} else {
			return romData, nil
		}
	}

	// try to open as a regular file
	return os.ReadFile(path)
}

func openZippedROM(z *zip.ReadCloser) ([]byte, error) {
	var candidates []*zip.File
	for _, f := range z.File {
		if f.FileInfo().Size() >= 32*1024 {
			candidates = append(candidates, f)
		}
	}

	// more than one candidate? filter by extension
	if len(candidates) != 1 {
		var pruned []*zip.File
		for _, f := range candidates {
			ext := filepath.Ext(f.Name)
			if ext == ".gb" || ext == ".gbc" {
				pruned = append(pruned, f)
			}
		}
		candidates = pruned
	}

	// still more than one candidate? filter by multiple of 1KB
	if len(candidates) != 1 {
		var pruned []*zip.File
		for _, f := range candidates {
			if f.FileInfo().Size()%1024 == 0 {
				pruned = append(pruned, f)
			}
		}
		candidates = pruned
	}

	if len(candidates) != 1 {
		return nil, fmt.Errorf("no rom candidate in zip file")
	}

	var f io.ReadCloser
	var err error
	if f, err = candidates[0].Open(); err != nil {
		return nil, err
	}

	defer f.Close()
	return io.ReadAll(f)
}

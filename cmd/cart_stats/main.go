package main

import (
	"archive/zip"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path"
	"path/filepath"

	"github.com/wmarshpersonal/gogeebee/cartridge"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintf(os.Stderr, "usage: %s <directory>\n", path.Base(os.Args[0]))
		os.Exit(1)
	}

	var (
		dir string = os.Args[1]
	)

	if entries, err := os.ReadDir(dir); err != nil {
		panic(err)
	} else {
		for _, entry := range entries {
			if !entry.IsDir() {
				filePath := path.Join(dir, entry.Name())
				if file, err := openROMFile(filePath); err != nil {
					slog.Error("opening failed", "path", filePath, "err", err)
				} else {
					func() {
						defer file.Close()
						var cart cartridge.Cartridge = make(cartridge.Cartridge, 0x150)
						if _, err := io.ReadAtLeast(file, cart[:], len(cart)); err != nil {
							slog.Error("reading failed", "path", filePath, "err", err)
						} else {
							printStats(entry.Name(), cart)
						}
					}()
				}
			}
		}
	}
}

func openROMFile(path string) (io.ReadCloser, error) {
	if zr, err := zip.OpenReader(path); err == nil {
		slog.Debug("looks like a zip file", "path", path)
		if r, err := openZippedROM(zr); err != nil {
			defer zr.Close()
			return nil, err
		} else {
			return r, nil
		}
	}

	// try to open as a regular file
	return os.Open(path)
}

func openZippedROM(z *zip.ReadCloser) (io.ReadCloser, error) {
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

	return f, nil
}

func printStats(name string, cart cartridge.Cartridge) {
	if header, err := cartridge.ReadHeader(cart); err != nil {
		slog.Error("reading header failed", err)
	} else {
		// if header.MBC == 3 && header.ROM.Size() == 1024*1024 {
		fmt.Printf("%s: %v\n", name, header)
		// }
	}
}

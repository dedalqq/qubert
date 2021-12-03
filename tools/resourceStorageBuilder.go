package main

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/jessevdk/go-flags"
)

type Options struct {
	Path []string `long:"path" short:"p" description:"Add folder path in to storage" required:"true"`
	Dest string   `long:"dest" short:"d" description:"Path to save result" required:"true"`
}

func main() {
	var opt Options

	_, err := flags.ParseArgs(&opt, os.Args)
	if err != nil {
		os.Exit(1)
	}

	data := bytes.NewBuffer([]byte{})

	zipWriter := zip.NewWriter(data)

	for _, p := range opt.Path {
		var srcPath, dirName string

		pathData := strings.Split(p, ":")
		if len(pathData) != 2 {
			panic("incorrect path")
		}

		dirName = pathData[0]
		srcPath = pathData[1]

		err = filepath.Walk(srcPath, func(itemPath string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			if info.IsDir() {
				return nil
			}

			localPath := strings.TrimPrefix(itemPath, srcPath)

			destPath := path.Join(dirName, localPath)

			_, _ = fmt.Fprintf(os.Stdout, "[Resources] Add: %s\n", destPath)

			fileWriter, err := zipWriter.Create(destPath)
			if err != nil {
				return err
			}

			f, err := os.Open(itemPath)
			if err != nil {
				return err
			}

			defer func() { _ = f.Close() }()

			_, err = io.Copy(fileWriter, f)
			if err != nil {
				return err
			}

			return nil
		})

		if err != nil {
			panic(err)
		}
	}

	err = zipWriter.Close()
	if err != nil {
		panic(err)
	}

	goFile, err := os.Create(path.Join(opt.Dest, "data.go"))
	if err != nil {
		panic(err)
	}

	_, _ = fmt.Fprintf(goFile, "package resources\n\nvar zipBytes = [...]byte{")

	for i, b := range data.Bytes() {
		if i%16 == 0 {
			_, _ = fmt.Fprintf(goFile, "\n\t")
		}

		_, _ = fmt.Fprintf(goFile, "0x%02X, ", b)
	}

	_, _ = fmt.Fprintf(goFile, "\n}\n\nfunc init() {\n\tdata = zipBytes[:]\n}\n")

	err = goFile.Close()
	if err != nil {
		panic(err)
	}
}

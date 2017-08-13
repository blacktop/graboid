package main

import (
	"archive/tar"
	"compress/gzip"
	"io"
	"os"
	"path/filepath"
)

func tarFiles(srcDir, tarName string) error {
	tarfile, err := os.Create(tarName)
	if err != nil {
		return err
	}
	defer tarfile.Close()

	gw := gzip.NewWriter(tarfile)
	defer gw.Close()
	tarball := tar.NewWriter(gw)
	defer tarball.Close()

	return filepath.Walk(srcDir,
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			header, err := tar.FileInfoHeader(info, info.Name())
			if err != nil {
				return err
			}

			if err = tarball.WriteHeader(header); err != nil {
				return err
			}

			if info.IsDir() {
				return nil
			}

			file, err := os.Open(path)
			if err != nil {
				return err
			}
			defer file.Close()
			_, err = io.Copy(tarball, file)
			return err
		})
}

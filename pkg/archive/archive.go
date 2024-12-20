package archive

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"io"
)

type TarGzReader struct {
	gzipReader *gzip.Reader
	tarReader  *tar.Reader
}

type ArchiveHeader struct {
	Name   string
	IsDir  bool
	IsFile bool
}

func NewTarGzReader(r io.Reader) (*TarGzReader, error) {
	gzipReader, err := gzip.NewReader(r)
	if err != nil {
		return nil, err
	}
	return &TarGzReader{
		gzipReader: gzipReader,
		tarReader:  tar.NewReader(gzipReader),
	}, nil
}

func (r *TarGzReader) Close() error {
	return r.gzipReader.Close()
}

func (r *TarGzReader) ReadNext() (*ArchiveHeader, error) {
	header, err := r.tarReader.Next()
	if err != nil {
		return nil, err
	}
	return &ArchiveHeader{
		Name:   header.Name,
		IsDir:  header.Typeflag == tar.TypeDir,
		IsFile: header.Typeflag == tar.TypeReg,
	}, nil
}

func (r *TarGzReader) ReadContent() (string, error) {
	var buf bytes.Buffer
	if _, err := io.Copy(&buf, r.tarReader); err != nil {
		return "", err
	}
	return buf.String(), nil
}

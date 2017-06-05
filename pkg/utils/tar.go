// Author: lipixun
// Created Time : 日 12/18 23:30:24 2016
//
// File Name: tar.go
// Description:
//
//	The tar utility
//

package utils

import (
	"archive/tar"
	"io"
	"os"
)

// Write file to tar
// Parameters:
//  path        The source file path
//  name        The name of the file in tar
//  writer      The tar writer
func TarWriteFile(path string, name string, writer *tar.Writer) error {
	info, err := os.Stat(path)
	if err != nil {
		return err
	}
	return TarWriteFileWithInfo(path, info, name, writer)
}

// Write file to tar with info
// Parameters:
//  path        The source file path
//  info        The info the file
//  name        The name of the file in tar
//  writer      The tar writer
func TarWriteFileWithInfo(path string, info os.FileInfo, name string, writer *tar.Writer) error {
	// Write header
	hdr, err := tar.FileInfoHeader(info, "")
	if err != nil {
		return err
	}
	hdr.Name = name
	writer.WriteHeader(hdr)
	// Write data
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()
	_, err = io.Copy(writer, file)
	// Done
	return err
}

// Write data to tar
func TarWriteData(data []byte, name string, writer *tar.Writer) error {
	// Write data to tar
	hdr := tar.Header{
		Name: name,
		Mode: int64(os.ModePerm),
		Size: int64(len(data)),
	}
	writer.WriteHeader(&hdr)
	_, err := writer.Write(data)
	return err
}

// Author: lipixun
// Created Time : 日 12/18 16:35:39 2016
//
// File Name: file.go
// Description:
//	The file artifact
package artifact

import (
	"archive/tar"
	"compress/gzip"
	"errors"
	"fmt"
	"github.com/ops-openlight/openlight/pkg/util"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
)

const (
	ArtifactTypeFile = "file"

	FileArtifactAttrPath       = "path"
	FileArtifactAttrFiles      = "files"
	FileArtifactAttrCompressed = "compressed"
)

type FileArtifact struct {
	Name       string   `json:"name" yaml:"name"`             // The name of this artifact
	Path       string   `json:"path" yaml:"path"`             // The root path this file artifact. This path is the root directory path or the file path itself if the artifact is not compressed otherwise this path is the compressed file path
	Files      []string `json:"files" yaml:"files"`           // The files in the artifact, the relative file path. If not empty the path field will be the root directory of the artifact otherwise (this field is nil or has 0 length) means this artifact only contains a single file and the path field is the path of the file
	Compressed bool     `json:"compressed" yaml:"compressed"` // Whether the artifact is compressed
}

func NewFileArtifact(name, path string, files []string, compressed bool) *FileArtifact {
	return &FileArtifact{
		Name:       name,
		Path:       path,
		Files:      files,
		Compressed: compressed,
	}
}

func NewSingleFileArtifact(name, path string) *FileArtifact {
	return &FileArtifact{Name: name, Path: path}
}

func (this *FileArtifact) GetName() string {
	return this.Name
}

func (this *FileArtifact) GetType() string {
	return ArtifactTypeFile
}

func (this *FileArtifact) GetAttr(name string) interface{} {
	switch name {
	case FileArtifactAttrPath:
		return this.Path
	case FileArtifactAttrFiles:
		return this.Files
	case FileArtifactAttrCompressed:
		return this.Compressed
	default:
		return nil
	}
}

func (this *FileArtifact) String() string {
	if len(this.Files) == 0 {
		return fmt.Sprintf("%s: %s --> Single file itself", ArtifactTypeFile, this.Path)
	} else if this.Compressed {
		return fmt.Sprintf("%s: %s --> Compressed file contains %d files", ArtifactTypeFile, this.Path, len(this.Files))
	} else {
		return fmt.Sprintf("%s: %s --> Collected %d files", ArtifactTypeFile, this.Path, len(this.Files))
	}
}

// The collect options
type CollectFileArtifactOptions struct {
	Recursive     bool           // Recursive collect or not
	FollowLink    bool           // Follow the symbol link or not. It's dangerous to enable this feature and thus not encouraged
	Includes      *regexp.Regexp // The regexp to test the files to include
	Excludes      *regexp.Regexp // The regexp to test the files to exclude
	CompressLevel int            // The compress level when doing compress collect
}

// Create the default options
func NewDefaultCollectFileArtifactOptions() CollectFileArtifactOptions {
	return CollectFileArtifactOptions{CompressLevel: gzip.DefaultCompression}
}

// Collect file artifact
// Parameters:
// 	name 		The artifact name
// 	path 		The root path of the collecting files
//	options 	The collect options
// NOTE:
//	- Directory will not be collected as a file, so empty directory will be ignored
// 	- You can only either specify includes or excludes or neither of them but both
func CollectFileArtifact(name, path string, options CollectFileArtifactOptions) (*FileArtifact, error) {
	files, err := listPath(path, &options)
	if err != nil && err != pathIsAFileError {
		return nil, err
	} else if len(files) == 0 {
		// No file collected
		return nil, nil
	}
	// Done
	return NewFileArtifact(name, path, files, false), nil
}

// Collect and compress file artifact
// Parameters:
// 	name 		The artifact name
// 	path 		The root path of the collecting files
// 	pkg 		The path of generated compressed package file
//	options 	The collect options
// NOTE:
//	- Directory will not be collected as a file, so empty directory will be ignored
// 	- You can only either specify includes or excludes or neither of them but both
//	- The files wll be compressed by gzip method
func CompressCollectFileArtifact(name, path, pkg string, options CollectFileArtifactOptions) (*FileArtifact, error) {
	pkgFile, err := os.Create(pkg)
	if err != nil {
		return nil, err
	}
	defer pkgFile.Close()
	gzipWriter, err := gzip.NewWriterLevel(pkgFile, options.CompressLevel)
	if err != nil {
		return nil, err
	}
	defer gzipWriter.Close()
	tarWriter := tar.NewWriter(gzipWriter)
	defer tarWriter.Close()
	// Collect files
	files, err := listPath(path, &options)
	if err != nil && err != pathIsAFileError {
		return nil, err
	} else if err == pathIsAFileError {
		// A single file
		if err := util.TarWriteFile(path, filepath.Base(path), tarWriter); err != nil {
			return nil, err
		}
		files = append(files, filepath.Base(path))
	} else {
		// Add files to tar
		if len(files) == 0 {
			// No files to compress
			return nil, nil
		}
		// Compress each file
		for _, file := range files {
			if err := util.TarWriteFile(filepath.Join(path, file), file, tarWriter); err != nil {
				return nil, err
			}
			files = append(files, file)
		}
	}
	// Done
	return NewFileArtifact(name, pkg, files, true), nil
}

var (
	pathIsAFileError = errors.New("Path is a file")
)

// List the path
// Parameters:
// 	path 			The path to list
// 	options 		The options to list
// Returns:
// 	A tuple (files, error) which the files is a list of relative path of the file
func listPath(path string, options *CollectFileArtifactOptions) ([]string, error) {
	// Check options
	if options.Includes != nil && options.Excludes != nil {
		return nil, errors.New("Cannot both specify includes and excludes")
	}
	// Check link
	if options.FollowLink {
		var err error
		// Get the real, follow the symbol link
		path, err = util.GetRealPath(path)
		if err != nil {
			return nil, err
		}
	}
	// Check if is a directory
	info, err := os.Stat(path)
	if err != nil {
		return nil, err
	}
	if info.IsDir() {
		var files []string // The relative path of the files
		// List the dir
		infos, err := ioutil.ReadDir(path)
		if err != nil {
			return nil, err
		}
		for _, info := range infos {
			if info.IsDir() {
				// A directory
				if options.Recursive {
					// Continue list the directory
					_files, err := listPath(filepath.Join(path, info.Name()), options)
					if err != nil { // Directory will never return pathIsAFileError error, so we don't check it
						return nil, err
					}
					// Add files
					for _, _file := range _files {
						files = append(files, filepath.Join(info.Name(), _file))
					}
				}
			} else {
				// A file
				_files, err := listPath(filepath.Join(path, info.Name()), options)
				if err != nil && err != pathIsAFileError {
					return nil, err
				} else if err == pathIsAFileError {
					// Add info itself as a file
					files = append(files, info.Name())
				} else {
					// Add files
					for _, _file := range _files {
						files = append(files, filepath.Join(info.Name(), _file))
					}
				}
			}
		}
		// Done
		return files, nil
	} else {
		// Return the file itself
		return nil, pathIsAFileError
	}
}

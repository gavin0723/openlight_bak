// Author: lipixun
// Created Time : 二 10/18 22:02:35 2016
//
// File Name: docker.go
// Description:
//	Openlight docker builder
package target

import (
	"archive/tar"
	"bufio"
	"context"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/docker/docker/api/types"
	dockerClient "github.com/docker/docker/client"
	"github.com/ops-openlight/openlight/helper/iohelper"
	"github.com/ops-openlight/openlight/sourcecode/spec"
	"github.com/satori/go.uuid"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

const (
	DockerBuilderLogHeader = "DockerBuilder"
	BuilderTypeDocker      = "docker"

	DockerImageArtifactName     = "image"
	DockerImageArtifactFileName = "image"
	DockerBuildFileName         = "DOCKERIMAGE"
)

type DockerBuilder struct {
	builderSpec *spec.DockerBuilder
}

func NewDockerBuilder(builderSpec *spec.DockerBuilder) (Builder, error) {
	// Create a new docker builder
	if builderSpec == nil {
		return nil, errors.New("Require docker builder config")
	}
	// Done
	return &DockerBuilder{builderSpec: builderSpec}, nil
}

func DockerBuilderCreator(s *spec.Target) (Builder, error) {
	if s.Builder == nil || s.Builder.Docker == nil {
		return nil, errors.New("Require docker builder config")
	}
	return NewDockerBuilder(s.Builder.Docker)
}

func (this *DockerBuilder) Type() string {
	return BuilderTypeDocker
}

func (this *DockerBuilder) Build(ctx *TargetBuildContext) (*BuildResult, error) {
	// Do necessary check
	if this.builderSpec.Image == "" {
		return nil, errors.New("Require image name")
	}
	// Get the output path
	outputPath, err := ctx.Workspace.FileSystem.GetGeneratePath(hex.EncodeToString(uuid.NewV4().Bytes()), true)
	if err != nil {
		return nil, errors.New(fmt.Sprint("Failed to ensure generate path, error: ", err))
	}
	// Check dockerfile
	var dockerfile string // The docker file absolute path
	if this.builderSpec.Dockerfile != "" {
		dockerfile = filepath.Join(ctx.Target.Repository.Local.RootPath, this.builderSpec.Dockerfile)
	} else {
		dockerfile = filepath.Join(ctx.Target.Repository.Local.RootPath, "Dockerfile")
	}
	if _, err := os.Stat(dockerfile); err != nil {
		return nil, errors.New(fmt.Sprintf("Failed to check dockerfile [%s] error: %s", dockerfile, err))
	}
	// Build all the dependency
	var deps []*BuildDependency
	for _, dep := range ctx.Target.Deps {
		depCtx, err := ctx.Derive(dep)
		if err != nil {
			return nil, err
		}
		depBuildResult, err := depCtx.Build()
		if err != nil {
			return nil, err
		}
		// Good
		deps = append(deps, &BuildDependency{Spec: dep.Spec, Result: depBuildResult})
	}
	// Create files
	var files []DockerBuildFile
	for _, f := range this.builderSpec.Files {
		if f.Dep == nil && f.LocalName == "" {
			return nil, errors.New("Invalid docker config files, require either dep or localName")
		} else if f.Dep != nil && f.LocalName != "" {
			return nil, errors.New("Invalid docker config files, cannot specify both dep and localName")
		} else if f.Dep != nil {
			// Use dependency
			found := false
			for _, dep := range deps {
				if ((f.Dep.Repository == "" && dep.Spec.Repository == nil) || (dep.Spec.Repository != nil && dep.Spec.Repository.Uri == f.Dep.Repository)) &&
					f.Dep.Target == dep.Spec.Target {
					// Target match
					for _, artifact := range dep.Result.Artifacts {
						if artifact.Name == f.Dep.Artifact {
							// Artifact match
							found = true
							if artifact.Type != ArtifactTypeFile {
								return nil, errors.New(fmt.Sprintf("Artifact [%s] must be a file, got [%s]", artifact.Name, artifact.Type))
							}
							files = append(files, DockerBuildFile{LocalName: artifact.Uri, TargetName: f.TargetName})
							// Break artifacts
							break
						}
					}
					// Break deps
					break
				}
			}
			if !found {
				return nil, errors.New(fmt.Sprintf("Artifact [%s:%s:%s] not found", f.Dep.Repository, f.Dep.Target, f.Dep.Artifact))
			}
		} else {
			// Use local file
			localPath := filepath.Join(ctx.Target.Repository.Local.RootPath, f.LocalName)
			if _, err := os.Stat(localPath); err != nil {
				return nil, errors.New(fmt.Sprintf("Failed to check local file [%s], error: %s", localPath, err))
			}
			files = append(files, DockerBuildFile{LocalName: localPath, TargetName: f.TargetName})
		}
	}
	// Create docker client
	ctx.Workspace.Logger.WriteDebugHeaderln(DockerBuilderLogHeader, "Connect to docker server")
	// TODO: Support specify docker uri
	c, err := dockerClient.NewClient("unix:///var/run/docker.sock", "", nil, nil)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("Failed to create docker client, error: %s", err))
	}
	// Build the image
	image := DockerImage{
		Repository: this.builderSpec.Repository,
		ImageName:  this.builderSpec.Image,
		Tag:        fmt.Sprintf("%s%s", this.builderSpec.TagPrefix, ctx.Option.Tag),
		Dockerfile: dockerfile,
		Files:      files,
	}
	ctx.Workspace.Logger.WriteInfoHeaderln(DockerBuilderLogHeader, "Building image: ", image.Uri())
	if err := this.buildDockerImage(c, &image, this.builderSpec.MarkLatest, ctx); err != nil {
		return nil, err
	}
	// Mark lastest or not
	//var lastUri string
	//if this.config.MarkLatest {
	//	if err := c.ImageTag(context.Background(), image.Uri(), image.LatestUri()); err != nil {
	//		return nil, errors.New(fmt.Sprintf("Failed to tag image [%s] to [%s], error: %s", image.Uri(), image.LatestUri(), err))
	//	}
	//}
	// Push or not
	if this.builderSpec.Push {
		// Push the image
		ctx.Workspace.Logger.WriteInfoHeaderln(DockerBuilderLogHeader, "Push image: ", image.Uri())
		if err := this.pushDockerImage(c, image.Uri()); err != nil {
			return nil, errors.New(fmt.Sprintf("Failed to push image [%s], error: %s", image.Uri(), err))
		}
		// Push the latest or not
		if this.builderSpec.MarkLatest {
			ctx.Workspace.Logger.WriteInfoHeaderln(DockerBuilderLogHeader, "Push image: ", image.LatestUri())
			if err := this.pushDockerImage(c, image.LatestUri()); err != nil {
				return nil, errors.New(fmt.Sprintf("Failed to push image [%s], error: %s", image.LatestUri(), err))
			}
		}
	}
	// Write out image file to output path
	if data, err := json.Marshal(image); err != nil {
		return nil, errors.New(fmt.Sprintf("Failed to marshal image data, error: %s", err))
	} else if err := ioutil.WriteFile(filepath.Join(outputPath, DockerImageArtifactFileName), data, 0666); err != nil {
		return nil, errors.New(fmt.Sprintf("Failed to write image artifact file, error: %s", err))
	}
	// Done
	result := &BuildResult{
		OutputPath: outputPath,
		Metadata:   BuildMetadata{Tag: ctx.Option.Tag, Time: ctx.Option.Time},
		Artifacts: []*Artifact{
			{
				Name: ctx.Target.Name,
				Type: ArtifactTypeDockerImage,
				Uri:  image.Uri(),
				Attrs: map[string]string{
					ArtifactDockerImageAttrFullname:   image.Uri(),
					ArtifactDockerImageAttrRepository: image.Repository,
					ArtifactDockerImageAttrImageName:  image.ImageName,
					ArtifactDockerImageAttrTag:        image.Tag,
				},
			},
		},
		Deps: deps,
	}
	// Done
	return result, nil
}

type DockerImage struct {
	Repository string            `json:"repository"`
	ImageName  string            `json:"imageName"`
	Tag        string            `json:"tag"`
	Dockerfile string            `json:"dockerfile"`
	Files      []DockerBuildFile `json:"files"`
}

func (this *DockerImage) Uri() string {
	if this.Repository != "" {
		return fmt.Sprintf("%s/%s:%s", this.Repository, this.ImageName, this.Tag)
	} else {
		return fmt.Sprintf("%s:%s", this.ImageName, this.Tag)
	}
}

func (this *DockerImage) LatestUri() string {
	if this.Repository != "" {
		return fmt.Sprintf("%s/%s:latest", this.Repository, this.ImageName)
	} else {
		return fmt.Sprintf("%s:latest", this.ImageName)
	}
}

type DockerBuildFile struct {
	TargetName string `json:"targetName"` // The target filename in docker tar
	LocalName  string `json:"localName"`  // The local filename to add into docker
}

type DockerBuildResponseData struct {
	Error       string `json:"error"`
	Stream      string `json:"stream"`
	ErrorDetail struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	} `json:"errorDetail"`
}

func (this *DockerBuilder) buildDockerImage(c *dockerClient.Client, image *DockerImage, markLatest bool, ctx *TargetBuildContext) error {
	// Build docker image
	// Create the docker build context
	var tarError error
	contextReader, contextWriter := io.Pipe()
	defer func() {
		contextReader.Close()
		contextWriter.Close()
	}()
	runCtx, cancel := context.WithCancel(context.Background())
	defer cancel()
	// Run the tar builder in another goroutine
	go func() {
		// Create a tar writer
		tarWriter := tar.NewWriter(contextWriter)
		defer func() {
			tarWriter.Flush()
			tarWriter.Close()
			contextWriter.Close()
		}()
		// Write dockerfile
		if ctx.Workspace.Verbose() {
			ctx.Workspace.Logger.WriteInfoHeaderln(DockerBuilderLogHeader, "Add Dockerfile from ", image.Dockerfile)
		}
		if err := this.writePath2Tar(image.Dockerfile, "Dockerfile", tarWriter); err != nil {
			tarError = errors.New(fmt.Sprintf("Failed to write Dockerfile to tar, error: %s", err))
			return
		}
		// Write build file
		if ctx.Workspace.Verbose() {
			ctx.Workspace.Logger.WriteInfoHeaderln(DockerBuilderLogHeader, "Add ", DockerBuildFileName, " file")
		}
		if data, err := json.Marshal(image); err != nil {
			tarError = errors.New(fmt.Sprintf("Failed to marshal image to %s file, error: %s", DockerBuildFileName, err))
			return
		} else if err := this.writeData2Tar(data, DockerBuildFileName, tarWriter); err != nil {
			tarError = errors.New(fmt.Sprintf("Failed to write BUILD to tar, error: %s", err))
			return
		}
		// Write files
		for _, f := range image.Files {
			if ctx.Workspace.Verbose() {
				ctx.Workspace.Logger.WriteInfoHeaderln(DockerBuilderLogHeader, "Add file from [", f.LocalName, "] to [", f.TargetName, "]")
			}
			if err := this.writePath2Tar(f.LocalName, f.TargetName, tarWriter); err != nil {
				tarError = errors.New(fmt.Sprintf("Failed to write local path [%s] to target path [%s] in tar, error: %s", f.LocalName, f.TargetName, err))
				return
			}
		}
		if ctx.Workspace.Verbose() {
			ctx.Workspace.Logger.WriteInfoHeaderln(DockerBuilderLogHeader, "Tar completed")
		}
	}()
	// Get the build options
	tags := []string{image.Uri()}
	if markLatest {
		tags = append(tags, image.LatestUri())
	}
	imageBuildOptions := types.ImageBuildOptions{
		Dockerfile:  "Dockerfile",
		Tags:        tags,
		Remove:      true,
		ForceRemove: true,
		PullParent:  !this.builderSpec.NoPull,
		NoCache:     this.builderSpec.NoCache, // Please use "ADD BUILD /BUILD" before any commands that should not be cached instead of "nocache: true"
	}
	// Check tar error
	if tarError != nil {
		return tarError
	}
	// Run docker build
	if rsp, err := c.ImageBuild(runCtx, contextReader, imageBuildOptions); err != nil {
		return err
	} else {
		defer rsp.Body.Close()
		scanner := bufio.NewScanner(rsp.Body)
		for scanner.Scan() {
			// Decode the response
			var data DockerBuildResponseData
			text := scanner.Text()
			if err = json.Unmarshal([]byte(text), &data); err != nil {
				ctx.Workspace.Logger.WriteWarningHeaderln(DockerBuilderLogHeader, "Docker --> Decode docker response failed, raw: ", text)
			} else {
				if data.Error == "" {
					// No error happend
					if ctx.Workspace.Verbose() {
						message := strings.Trim(data.Stream, "\n")
						if message != "" {
							ctx.Workspace.Logger.WriteInfoHeaderln(DockerBuilderLogHeader, "Docker --> ", message)
						}
					}
				} else {
					// Error happened
					ctx.Workspace.Logger.WriteErrorHeaderln(DockerBuilderLogHeader, "Docker --> Error ", data.Error, " Code: ", data.ErrorDetail.Code, " Message: ", data.ErrorDetail.Message)
					return errors.New(fmt.Sprint("Docker build error ", data.Error, " code ", data.ErrorDetail.Code, " message ", data.ErrorDetail.Message))
				}
			}
		}
		if err := scanner.Err(); err != nil {
			return errors.New(fmt.Sprint("Failed to read docker build response, error: ", err))
		}
	}
	// Done
	return tarError
}

func (this *DockerBuilder) pushDockerImage(c *dockerClient.Client, uri string) error {
	// Push docker image
	raw, _ := json.Marshal(types.AuthConfig{})
	if rsp, err := c.ImagePush(context.Background(), uri, types.ImagePushOptions{RegistryAuth: base64.URLEncoding.EncodeToString(raw)}); err != nil {
		return err
	} else {
		defer rsp.Close()
		buffer := make([]byte, 4096)
		for {
			if _, err := rsp.Read(buffer); err != nil {
				if err == io.EOF {
					break
				} else {
					return err
				}
			}
		}
		// Good
		return nil
	}
}

func (this *DockerBuilder) writePath2Tar(p string, targetPath string, writer *tar.Writer) error {
	// Write a path to tar
	// Get the real path and info
	p, err := iohelper.GetRealPath(p)
	if err != nil {
		return err
	}
	info, err := os.Stat(p)
	if err != nil {
		return err
	}
	// Check a directory or file
	if info.IsDir() {
		// A directory
		var walkThroughPath func(dirPath, targetDirPath string) error
		walkThroughPath = func(dirPath, targetDirPath string) error {
			// The filepath.Walk will not follow the symbol link
			infos, err := ioutil.ReadDir(dirPath)
			if err != nil {
				return err
			}
			for _, info := range infos {
				// The original name
				name := info.Name()
				// Get real path
				subPath, err := iohelper.GetRealPath(filepath.Join(dirPath, info.Name()))
				if err != nil {
					return err
				}
				info, err := os.Stat(subPath)
				if err != nil {
					return err
				}
				if info.IsDir() {
					// Walk through this dir
					if err := walkThroughPath(subPath, filepath.Join(targetDirPath, name)); err != nil {
						return err
					}
				} else {
					// A file
					if err := this.writeFile2Tar(subPath, info, filepath.Join(targetDirPath, name), writer); err != nil {
						return err
					}
				}
			}
			// Done
			return nil
		}
		// Walk through path
		return walkThroughPath(p, targetPath)
	} else {
		// A file
		return this.writeFile2Tar(p, info, targetPath, writer)
	}
}

func (this *DockerBuilder) writeFile2Tar(p string, info os.FileInfo, name string, writer *tar.Writer) error {
	// Write file to tar
	// Write header
	hdr, err := tar.FileInfoHeader(info, "")
	if err != nil {
		return errors.New(fmt.Sprintf("Failed to create tar header for file [%s] error: %s", p, err))
	}
	hdr.Name = name
	writer.WriteHeader(hdr)
	// Write data
	file, err := os.Open(p)
	if err != nil {
		return err
	}
	defer file.Close()
	if _, err = io.Copy(writer, file); err != nil {
		return err
	}
	// Done
	return nil
}

func (this *DockerBuilder) writeData2Tar(data []byte, name string, writer *tar.Writer) error {
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

// Author: lipixun
// Created Time : 二 10/18 22:02:35 2016
//
// File Name: docker.go
// Description:
//	Openlight docker builder
package builder

import (
	"archive/tar"
	"bufio"
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/docker/docker/api/types"
	dockerClient "github.com/docker/docker/client"
	"github.com/ops-openlight/openlight/pkg/artifact"
	"github.com/ops-openlight/openlight/pkg/log"
	"github.com/ops-openlight/openlight/pkg/sourcecode/spec"
	"github.com/ops-openlight/openlight/pkg/util"
	"github.com/ops-openlight/openlight/pkg/workspace"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"text/template"
	"time"
)

const (
	DockerBuilderLogHeader = "DockerBuilder"
	BuilderTypeDocker      = "docker"

	DefaultDockerFilename = "Dockerfile"

	DockerImageArtifactName     = "image"
	DockerImageArtifactFileName = "image"
	DockerImageSummaryFileName  = "DOCKERIMAGE"
)

type DockerSourceCodeBuilder struct{}

func NewDockerSourceCodeBuilder() *DockerSourceCodeBuilder {
	return new(DockerSourceCodeBuilder)
}

// Create new environment for the builder
func (this *DockerSourceCodeBuilder) NewEnviron(builder *Builder) (Environment, error) {
	return NewGeneralEnvironment(filepath.Join(builder.EnvironmentPath(), BuilderTypeDocker))
}

// Prepare for the target
func (this *DockerSourceCodeBuilder) Prepare(target *spec.Target, env Environment, context *BuilderContext) error {
	// We have to do
	return nil
}

// Build the target
func (this *DockerSourceCodeBuilder) Build(target *spec.Target, env Environment, context *BuilderContext) error {
	startBuildTime := time.Now()
	logger := context.Workspace.Logger.GetLoggerWithHeader(DockerBuilderLogHeader)
	// Get docker spec
	dockerSpec := target.Spec.Build.Docker
	if dockerSpec == nil {
		return errors.New("Docker build spec not defined")
	}
	// Do necessary check
	if dockerSpec.Image == "" {
		return errors.New("Require image name")
	}
	dockerArtifactName := dockerSpec.Name
	if dockerArtifactName == "" {
		dockerArtifactName = BuilderDefaultArtifactName
	}
	// Get the output path
	outputPath, err := context.Builder.EnsureTargetOutputPath(target)
	if err != nil {
		return err
	}
	// Format the dockerfile
	var dockerfilePath string // The docker file absolute path
	if dockerSpec.Dockerfile == "" {
		dockerfilePath = filepath.Join(target.Path(), DefaultDockerFilename)
	} else {
		dockerfilePath = filepath.Join(target.Path(), dockerSpec.Dockerfile)
	}
	dockerfileData, err := ioutil.ReadFile(dockerfilePath)
	if err != nil {
		return errors.New(fmt.Sprintf("Failed to load dockerfile [%s] error: %s", dockerfilePath, err))
	}
	dockerfileContent, err := this.formatDockerfile(string(dockerfileData), target, context, logger)
	if err != nil {
		return err
	}
	if context.Workspace.Verbose {
		logger.LeveledPrintf(log.LevelDebug, "Formatted dockerfile:\n%s\n", dockerfileContent)
	}
	// Get docker build files
	var files []DockerBuildFile
	for _, f := range dockerSpec.Files {
		if f.Source.Local != nil && f.Source.Dep != nil {
			return errors.New("Cannot define local and dep at the same time")
		} else if f.Source.Local == nil && f.Source.Dep == nil {
			return errors.New("Require either define local or dep")
		} else if f.Source.Local != nil {
			// A local file / dir
			localPath := filepath.Join(target.Path(), f.Source.Local.Path)
			if _, err := os.Stat(localPath); err != nil {
				return errors.New(fmt.Sprintf("Failed to check local file [%s], error: %s", localPath, err))
			}
			files = append(files, DockerBuildFile{Target: f.Target, Path: localPath})
		} else {
			// A dependency file / dir
			depSpec, ok := target.Spec.Deps[f.Source.Dep.Name]
			if !ok {
				return errors.New(fmt.Sprintf("Dependency [%s] not found", f.Source.Dep.Name))
			}
			buildResult := context.Builder.Results[depSpec.Key()]
			if buildResult == nil {
				return errors.New(fmt.Sprintf("Build result of [%s] that is referenced by dependency [%s] not found", depSpec.Key(), f.Source.Dep.Name))
			}
			art := buildResult.Artifacts[f.Source.Dep.Artifact]
			if art == nil {
				return errors.New(fmt.Sprintf("Artifact [%s] of target [%s] that is referenced by dependency [%s] not found", f.Source.Dep.Artifact, depSpec.Key(), f.Source.Dep.Name))
			}
			// Add this artifact
			files, err = this.AddFileFromArtifact(f.Target, art, files)
			if err != nil {
				return err
			}
		}
	}
	// Create docker client
	c, err := this.createDockerClient(context.Workspace, logger)
	if err != nil {
		return errors.New(fmt.Sprintf("Failed to create docker client, error: %s", err))
	}
	// Build the image
	image := DockerImage{
		Repository: dockerSpec.Repository,
		ImageName:  dockerSpec.Image,
		Tag:        fmt.Sprintf("%s%s", dockerSpec.TagPrefix, context.Builder.Options.Tag),
		Dockerfile: dockerfileContent,
		Files:      files,
	}
	logger.LeveledPrintf(log.LevelDebug, "Start to build the image [%s]\n", image.Uri())
	if err := this.buildDockerImage(c, &image, dockerSpec, context); err != nil {
		return err
	}
	// Push
	if context.Builder.Options.ThirdParty.Docker.Push {
		// Push the image
		logger.LeveledPrintf(log.LevelDebug, "Start to push the image [%s]\n", image.Uri())
		if err := this.pushDockerImage(c, image.Uri()); err != nil {
			return errors.New(fmt.Sprintf("Failed to push image [%s], error: %s", image.Uri(), err))
		}
		// Push the latest or not
		if dockerSpec.MarkLatest {
			logger.LeveledPrintf(log.LevelDebug, "Start to push the image [%s]\n", image.LatestUri())
			if err := this.pushDockerImage(c, image.LatestUri()); err != nil {
				return errors.New(fmt.Sprintf("Failed to push image [%s], error: %s", image.LatestUri(), err))
			}
		}
	}
	// Write out a image file to output path
	imageSummaryFile := filepath.Join(outputPath, fmt.Sprintf("%s.json", dockerArtifactName))
	if data, err := json.Marshal(image); err != nil {
		return errors.New(fmt.Sprintf("Failed to marshal docker image summary, error: %s", err))
	} else if err := ioutil.WriteFile(imageSummaryFile, data, 0666); err != nil {
		return errors.New(fmt.Sprintf("Failed to write docker image summary file, error: %s", err))
	}
	// Create artifacts
	imageArtifact := artifact.NewDockerArtifact(dockerArtifactName, image.Uri(), image.Repository, image.ImageName, image.Tag)
	imageSummaryArtifact := artifact.NewSingleFileArtifact(fmt.Sprintf("%s.summary", dockerArtifactName), imageSummaryFile)
	// Create the build result
	buildResult := spec.NewBuildResult(target, context.Builder.NewBuildMetadata(target))
	buildResult.Metadata.Builder = BuilderTypeDocker
	buildResult.Metadata.BuildTimeUsage = time.Now().Sub(startBuildTime).Seconds()
	buildResult.Metadata.LinkedPath = env.GetTargetPath(target)
	buildResult.Metadata.OutputPath = outputPath
	buildResult.Artifacts[imageArtifact.GetName()] = imageArtifact
	buildResult.Artifacts[imageSummaryArtifact.GetName()] = imageSummaryArtifact
	context.Builder.SetBuildResultDependency(target, buildResult)
	context.Builder.AddResult(target, buildResult)
	// Done
	return nil
}

// Add file from artifact
func (this *DockerSourceCodeBuilder) AddFileFromArtifact(target string, art artifact.Artifact, files []DockerBuildFile) ([]DockerBuildFile, error) {
	switch t := art.(type) {
	default:
		// Unknown type
		return nil, errors.New(fmt.Sprintf("Unsupported artifact type [%s]", t))
	case *artifact.FileArtifact:
		// File artifact
		fileArtifact := art.(*artifact.FileArtifact)
		if _, err := os.Stat(fileArtifact.Path); err != nil {
			return nil, errors.New(fmt.Sprintf("Failed to check local file [%s], error: %s", fileArtifact.Path, err))
		}
		files = append(files, DockerBuildFile{Target: target, Path: fileArtifact.Path})
		return files, nil
	}
}

type DockerfileRecipient struct {
	Tag    string
	Time   string
	Branch string
	Commit string
}

// Format the docker file
func (this *DockerSourceCodeBuilder) formatDockerfile(tempStr string, target *spec.Target, context *BuilderContext, logger log.Logger) (string, error) {
	temp, err := template.New("dockerfile").Parse(tempStr)
	if err != nil {
		return "", errors.New(fmt.Sprintf("Failed to parse dockerfile as go template, error: %s", err))
	}
	// Create the recipient
	recipient := DockerfileRecipient{
		Tag:    context.Builder.Options.Tag,
		Time:   context.Builder.Options.Time.Format(time.RFC3339),
		Branch: target.Repository.Metadata.Branch,
		Commit: target.Repository.Metadata.Commit,
	}
	buf := new(bytes.Buffer)
	if err := temp.Execute(buf, recipient); err != nil {
		return "", errors.New(fmt.Sprintf("Failed to execute dockerfile template, error: %s", err))
	}
	return buf.String(), nil
}

func (this *DockerSourceCodeBuilder) createDockerClient(ws *workspace.Workspace, logger log.Logger) (*dockerClient.Client, error) {
	dockerUri := ws.Options.ThirdService.Docker.Uri
	logger.LeveledPrintf(log.LevelDebug, "Connect to docker daemon via: %s", dockerUri)
	return dockerClient.NewClient(dockerUri, "", nil, nil)
}

type DockerImage struct {
	Repository string            `json:"repository"`
	ImageName  string            `json:"imageName"`
	Tag        string            `json:"tag"`
	Dockerfile string            `json:"dockerfile"` // The dockerfile content, NOT the dockerfile path!!!!
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
	Target string `json:"target"` // The target filename in docker tar
	Path   string `json:"path"`   // The local filename to add into docker
}

type DockerBuildResponseData struct {
	Error       string `json:"error"`
	Stream      string `json:"stream"`
	ErrorDetail struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	} `json:"errorDetail"`
}

// Build docker image
func (this *DockerSourceCodeBuilder) buildDockerImage(c *dockerClient.Client, image *DockerImage, dockerSpec *spec.DockerBuildSpec, ctx *BuilderContext) error {
	logger := ctx.Workspace.Logger.GetLoggerWithHeader(DockerBuilderLogHeader)
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
		logger.LeveledPrintf(log.LevelDebug, "Add dockerfile from %s\n", image.Dockerfile)
		if err := this.writeData2Tar([]byte(image.Dockerfile), "Dockerfile", tarWriter); err != nil {
			tarError = errors.New(fmt.Sprintf("Failed to write Dockerfile to tar, error: %s", err))
			cancel()
			return
		}
		// Write build file
		logger.LeveledPrintf(log.LevelDebug, "Add image summary file from %\n", image.Dockerfile)
		if data, err := json.Marshal(image); err != nil {
			tarError = errors.New(fmt.Sprintf("Failed to marshal image to [%s] file, error: %s", DockerImageSummaryFileName, err))
			cancel()
			return
		} else if err := this.writeData2Tar(data, DockerImageSummaryFileName, tarWriter); err != nil {
			tarError = errors.New(fmt.Sprintf("Failed to write [%s] to tar, error: %s", DockerImageSummaryFileName, err))
			cancel()
			return
		}
		// Write files
		for _, f := range image.Files {
			logger.LeveledPrintf(log.LevelDebug, "Add [%s] from: %s\n", f.Target, f.Path)
			if err := this.writePath2Tar(f.Path, f.Target, tarWriter); err != nil {
				tarError = errors.New(fmt.Sprintf("Failed to write target [%s] from [%s] in tar, error: %s", f.Target, f.Path, err))
				cancel()
				return
			}
		}
		logger.LeveledPrintln(log.LevelDebug, "Docker tar completed")
	}()
	// Get the build options
	tags := []string{image.Uri()}
	if dockerSpec.MarkLatest {
		tags = append(tags, image.LatestUri())
	}
	imageBuildOptions := types.ImageBuildOptions{
		Dockerfile:  "Dockerfile",
		Tags:        tags,
		Remove:      true,
		ForceRemove: true,
		PullParent:  !dockerSpec.NoPull,
		NoCache:     dockerSpec.NoCache, // Please use "ADD BUILD /BUILD" before any commands that should not be cached instead of "nocache: true"
	}
	// Check tar error
	if tarError != nil {
		cancel()
		return tarError
	}
	// Run docker build
	rsp, err := c.ImageBuild(runCtx, contextReader, imageBuildOptions)
	if err != nil {
		return err
	}
	defer rsp.Body.Close()
	scanner := bufio.NewScanner(rsp.Body)
	for scanner.Scan() {
		// Decode the response
		var data DockerBuildResponseData
		text := scanner.Text()
		err := json.Unmarshal([]byte(text), &data)
		if err != nil {
			logger.LeveledPrintf(log.LevelWarn, "Docker --> Decode docker response failed, raw: %s\n", text)
		}
		if data.Error == "" {
			// No error happend
			if ctx.Workspace.Verbose {
				message := strings.Trim(data.Stream, "\n")
				if message != "" {
					logger.LeveledPrintf(log.LevelDebug, "Docker --> %s\n", message)
				}
			}
		} else {
			// Error happened
			logger.LeveledPrintf(log.LevelError, "Docker --> Error [%s] Code [%d] Message: %s\n", data.Error, data.ErrorDetail.Code, data.ErrorDetail.Message)
			return errors.New(fmt.Sprint("Docker build error ", data.Error, " code ", data.ErrorDetail.Code, " message ", data.ErrorDetail.Message))
		}
	}
	if err := scanner.Err(); err != nil {
		return errors.New(fmt.Sprint("Failed to read docker build response, error: ", err))
	}
	// Done
	return nil
}

func (this *DockerSourceCodeBuilder) pushDockerImage(c *dockerClient.Client, uri string) error {
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

func (this *DockerSourceCodeBuilder) writePath2Tar(p string, targetPath string, writer *tar.Writer) error {
	// Write a path to tar
	// Get the real path and info
	p, err := util.GetRealPath(p)
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
				subPath, err := util.GetRealPath(filepath.Join(dirPath, info.Name()))
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

func (this *DockerSourceCodeBuilder) writeFile2Tar(p string, info os.FileInfo, name string, writer *tar.Writer) error {
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

func (this *DockerSourceCodeBuilder) writeData2Tar(data []byte, name string, writer *tar.Writer) error {
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

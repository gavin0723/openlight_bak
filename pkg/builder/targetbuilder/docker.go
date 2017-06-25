// Author: lipixun
// Created Time : 二 10/18 22:02:35 2016
//
// File Name: docker.go
// Description:
//

package targetbuilder

import (
	"archive/tar"
	"bufio"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/docker/docker/api/types"
	dockerClient "github.com/docker/docker/client"
	"github.com/satori/go.uuid"

	pbSpec "github.com/ops-openlight/openlight/protoc-gen-go/spec"

	"github.com/ops-openlight/openlight/pkg/artifact"
	"github.com/ops-openlight/openlight/pkg/builder/buildcontext"
	"github.com/ops-openlight/openlight/pkg/repository"
	"github.com/ops-openlight/openlight/pkg/utils"
)

// DockerImageTargetBuilder builds docker image target
type DockerImageTargetBuilder struct {
	target *repository.Target
	spec   *pbSpec.DockerImageTarget
}

// NewDockerImageTargetBuilder creates a new DockerImageTargetBuilder
func NewDockerImageTargetBuilder(target *repository.Target, spec *pbSpec.DockerImageTarget) (*DockerImageTargetBuilder, error) {
	if target == nil {
		return nil, errors.New("Require target")
	}
	if spec == nil {
		return nil, errors.New("Require spec")
	}
	return &DockerImageTargetBuilder{target, spec}, nil
}

// Build the target
func (builder *DockerImageTargetBuilder) Build(ctx buildcontext.Context) (artifact.Artifact, error) {
	if builder.spec.GetImage() == "" {
		return nil, errors.New("Require image")
	}
	if len(builder.spec.GetCommands()) == 0 {
		return nil, errors.New("Require commands")
	}

	// The output path
	outputPath, err := ctx.GetTargetOutputDir(builder.target, true)
	if err != nil {
		log.Errorf("Failed to get target output dir: %v", err)
		return nil, err
	}

	//
	// Build dockerfile and feed stream
	//

	log.Debugln("DockerImageTargetBuilder.Build: Generating dockerfile and docker feed stream")

	var dockerfile []string
	var feedStream _DockerImageFeedStream
	for _, cmdSpec := range builder.spec.Commands {
		if cmd := cmdSpec.GetFrom(); cmd != nil {
			// From
			if cmd.GetName() == "" {
				log.Errorln("Invalid docker command. FROM requires name")
				return nil, errors.New("Invalid docker command")
			}
			dockerfile = append(dockerfile, fmt.Sprintf("FROM %v", cmd.GetName()))
		} else if cmd := cmdSpec.GetLabel(); cmd != nil {
			// Label
			if cmd.GetKey() == "" {
				log.Errorln("Invalid docker command. LABLE requires key")
				return nil, errors.New("Invalid docker command")
			}
			dockerfile = append(dockerfile, fmt.Sprintf("LABEL %v=%v", cmd.GetKey(), cmd.GetValue()))
		} else if cmd := cmdSpec.GetAdd(); cmd != nil {
			// Add
			if cmd.GetFile() == nil {
				log.Errorln("Invalid docker command. ADD requires file")
				return nil, errors.New("Invalid docker command")
			}
			if cmd.GetPath() == "" {
				log.Errorln("Invalid docker command. ADD requires path")
				return nil, errors.New("Invalid docker command")
			}
			dirname := uuid.NewV4().String()
			log.Debugf("DockerImageTargetBuilder.Build: Generate tar dirname [%v] for dest path [%v]", dirname, cmd.GetPath())
			if err := builder.feedStreamForAddCopy(&feedStream, cmd.GetFile(), dirname, ctx); err != nil {
				return nil, err
			}
			dockerfile = append(dockerfile, fmt.Sprintf("ADD %v %v", dirname, cmd.GetPath()))
		} else if cmd := cmdSpec.GetCopy(); cmd != nil {
			// Copy
			if cmd.GetFile() == nil {
				log.Errorln("Invalid docker command. COPY requires file")
				return nil, errors.New("Invalid docker command")
			}
			if cmd.GetPath() == "" {
				log.Errorln("Invalid docker command. COPY requires path")
				return nil, errors.New("Invalid docker command")
			}
			dirname := uuid.NewV4().String()
			log.Debugf("DockerImageTargetBuilder.Build: Generate tar dirname [%v] for dest path [%v]", dirname, cmd.GetPath())
			if err := builder.feedStreamForAddCopy(&feedStream, cmd.GetFile(), dirname, ctx); err != nil {
				return nil, err
			}
			dockerfile = append(dockerfile, fmt.Sprintf("COPY %v %v", dirname, cmd.GetPath()))
		} else if cmd := cmdSpec.GetRun(); cmd != nil {
			// Run
			if cmd.GetCommand() == "" {
				log.Errorln("Invalid docker command. RUN requires command")
				return nil, errors.New("Invalid docker command")
			}
			dockerfile = append(dockerfile, fmt.Sprintf("RUN %v", cmd.GetCommand()))
		} else if cmd := cmdSpec.GetEntrypoint(); cmd != nil {
			// Entrypoint
			if len(cmd.GetArgs()) == 0 {
				log.Errorln("Invalid docker command. ENTRYPOINT requires args")
				return nil, errors.New("Invalid docker command")
			}
			argBytes, err := json.Marshal(cmd.GetArgs())
			if err != nil {
				log.Errorln("Invalid docker command. Failed to dump ENTRYPOINT args")
				return nil, errors.New("Invalid docker command")
			}
			dockerfile = append(dockerfile, fmt.Sprintf("ENTRYPOINT %v", string(argBytes)))
		} else if cmd := cmdSpec.GetExpose(); cmd != nil {
			// Expose
			if len(cmd.GetPorts()) == 0 {
				log.Errorln("Invalid docker command. EXPOSE requires ports")
				return nil, errors.New("Invalid docker command")
			}
			var portStrs []string
			for _, port := range cmd.GetPorts() {
				portStrs = append(portStrs, fmt.Sprintf("%v", port))
			}
			dockerfile = append(dockerfile, fmt.Sprintf("EXPOSE %v", strings.Join(portStrs, " ")))
		} else if cmd := cmdSpec.GetVolume(); cmd != nil {
			// Volume
			if len(cmd.GetPaths()) == 0 {
				log.Errorln("Invalid docker command. VOLUME requires paths")
				return nil, errors.New("Invalid docker command")
			}
			pathBytes, err := json.Marshal(cmd.GetPaths())
			if err != nil {
				log.Errorln("Invalid docker command. Failed to dump VOLUME paths")
				return nil, errors.New("Invalid docker command")
			}
			dockerfile = append(dockerfile, fmt.Sprintf("VOLUME %v", string(pathBytes)))
		} else if cmd := cmdSpec.GetUser(); cmd != nil {
			// User
			if cmd.GetName() == "" {
				log.Errorln("Invalid docker command. USER requires name")
				return nil, errors.New("Invalid docker command")
			}
			dockerfile = append(dockerfile, fmt.Sprintf("USER %v", cmd.GetName()))
		} else if cmd := cmdSpec.GetWorkdir(); cmd != nil {
			// Workdir
			if cmd.GetPath() == "" {
				log.Errorln("Invalid docker command. WORKDIR requires path")
				return nil, errors.New("Invalid docker command")
			}
			dockerfile = append(dockerfile, fmt.Sprintf("WORKDIR %v", cmd.GetPath()))
		} else if cmd := cmdSpec.GetEnv(); cmd != nil {
			// Env
			if cmd.GetKey() == "" {
				log.Errorln("Invalid docker command. ENV requires key")
				return nil, errors.New("Invalid docker command")
			}
			dockerfile = append(dockerfile, fmt.Sprintf("ENV %v=%v", cmd.GetKey(), cmd.GetValue()))
		} else {
			log.Errorln("Invalid docker command. No specific command is specified")
			return nil, errors.New("Invalid docker command")
		}
	}

	// Write dockerfile
	if ctx.Verbose() {
		log.Debugln("Generated dockerfile:\n", strings.Join(dockerfile, "\n"))
	}
	if err := feedStream.FeedBytes([]byte(strings.Join(dockerfile, "\n")), "Dockerfile"); err != nil {
		log.Errorln("Failed to feed dockerfile to docker feed stream:", err)
		return nil, errors.New("Failed to feed dockerfile to docker feed stream")
	}

	// Write BUILD file
	gitRepoInfo, err := utils.GetGitRepositoryInfo(builder.target.Path())
	if err != nil {
		log.Warnf("Failed to get git repository info: %v", err)
	}
	buildlines := []string{
		fmt.Sprintf("CI_BUILD_TIME=\"%v\"", time.Now().Format(time.RFC3339)),
		fmt.Sprintf("CI_TAG=\"%v\"", ctx.Tag()),
	}
	if gitRepoInfo != nil {
		buildlines = append(buildlines,
			fmt.Sprintf("CI_BRANCH=\"%v\"", gitRepoInfo.Branch),
			fmt.Sprintf("CI_COMMIT=\"%v\"", gitRepoInfo.Commit),
			fmt.Sprintf("CI_COMMIT_TIME=\"%v\"", gitRepoInfo.CommitTime),
		)
	}
	if err := feedStream.FeedBytes([]byte(strings.Join(buildlines, "\n")), "BUILD"); err != nil {
		log.Errorln("Failed to feed BUILD to docker feed stream:", err)
		return nil, errors.New("Failed to feed BUILD to docker feed stream")
	}

	//
	// Run docker build
	//

	log.Debugln("DockerImageTargetBuilder.Build: Run docker build")

	// TODO: Create docker client by configs
	docker, err := dockerClient.NewEnvClient()
	if err != nil {
		log.Errorln("Failed to create docker client:", err)
		return nil, errors.New("Failed to create docker client")
	}

	// Create docker image artifact

	art := artifact.NewDockerImageArtifact(outputPath, builder.spec.Repository, builder.spec.Image, ctx.Tag())
	if err := builder.buildDockerImage(docker, art, &feedStream, ctx); err != nil {
		return nil, err
	}

	// Push
	if builder.spec.Push {
		// Push the image
		log.Debugln("DockerImageTargetBuilder.Build: Push docker image")

		if err := builder.pushDockerImage(docker, art.FullName()); err != nil {
			return nil, err
		}
		// Push the latest or not
		if builder.spec.SetLatestTag {
			log.Debugln("DockerImageTargetBuilder.Build: Set docker image latest tag")

			if err := builder.pushDockerImage(docker, art.LatestName()); err != nil {
				return nil, err
			}
		}
	}

	// Write artifact file to output path
	if err := art.Dump(); err != nil {
		log.Errorln("Failed to dump docker artifact:", err)
		return nil, errors.New("Failed to dump docker artifact")
	}

	// Done
	return art, nil
}

func (builder *DockerImageTargetBuilder) feedStreamForAddCopy(feedStream *_DockerImageFeedStream, fileSource *pbSpec.FileSource, destPath string, ctx buildcontext.Context) error {
	// Get the local file from file source
	var localFile string
	if f := fileSource.GetFile(); f != nil {
		// From file
		if f.Reference != "" {
			ref := builder.target.Package().GetReference(f.Reference)
			if ref == nil {
				log.Errorf("Failed to add referenced local file. Reference [%v] not found in package", f.Reference)
				return errors.New("Failed to add referenced local file")
			}
			repo := ctx.GetRemoteRepository(ref.Spec().GetRemote())
			if repo == nil {
				log.Errorf("Failed to add referenced local file. Reference [%v] is not resolved", ref.Spec().GetRemote())
				return errors.New("Failed to add referenced local file")
			}
			localFile = filepath.Join(builder.target.Path(), f.Filename)
		} else {
			localFile = filepath.Join(builder.target.Path(), f.Filename)
		}
	} else if art := fileSource.GetArtifact(); art != nil {
		// From artifact
		if art.Reference != "" {
			ref := builder.target.Package().GetReference(art.Reference)
			if ref == nil {
				log.Errorf("Failed to add referenced artifact. Reference [%v] not found in package", art.Reference)
				return errors.New("Failed to add referenced artifact")
			}
			repo := ctx.GetRemoteRepository(ref.Spec().GetRemote())
			if repo == nil {
				log.Errorf("Failed to add referenced artifact. Reference [%v] is not resolved", ref.Spec().GetRemote())
				return errors.New("Failed to add referenced artifact")
			}
			target, err := repo.GetTarget(art.Path, art.Target)
			if err != nil {
				log.Errorf("Failed to add referenced artifact. Failed to get target [%v] in path [%v] in reference [%v]", art.Target, art.Path, ref.Spec().Remote)
				return errors.New("Failed to add referenced artifact")
			}
			if target == nil {
				log.Errorf("Failed to add referenced artifact. Target [%v] in path [%v] in reference [%v] not found", art.Target, art.Path, ref.Spec().Remote)
				return errors.New("Failed to add referenced artifact")
			}
			result := ctx.GetBuildResult(target)
			if result == nil {
				log.Errorf("Failed to add referenced artifact. Target [%v] in path [%v] in reference [%v] has not been built yet", art.Target, art.Path, ref.Spec().Remote)
				return errors.New("Failed to add referenced artifact")
			}
			localFile = result.Artifact().GetPath()
		} else {
			target, err := builder.target.Package().GetRelativeTarget(art.Path, art.Target)
			if err != nil {
				log.Errorf("Failed to add artifact. Failed to get target [%v] in path [%v]: %v", art.Target, art.Path, err)
				return errors.New("Failed to add artifact")
			}
			if target == nil {
				log.Errorf("Failed to add artifact. Failed to get target [%v] in path [%v] not found", art.Target, art.Path)
				return errors.New("Failed to add artifact")
			}
			result := ctx.GetBuildResult(target)
			if result == nil {
				log.Errorf("Failed to add artifact. Target [%v] in path [%v] has not been built", art.Target, art.Path)
				return errors.New("Failed to add artifact")
			}
			localFile = result.Artifact().GetPath()
		}
	} else {
		// No defined
		return errors.New("Invalid docker command. Add / Copy requires specific file source")
	}
	// Add this local file
	if err := feedStream.FeedFile(localFile, destPath); err != nil {
		log.Errorf("Failed to add local file [%v]: %v", localFile, err)
		return errors.New("Failed to add local file")
	}
	// Done
	return nil
}

type _DockerBuildResponseData struct {
	Error       string `json:"error"`
	Stream      string `json:"stream"`
	ErrorDetail struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	} `json:"errorDetail"`
}

// Build docker image
func (builder *DockerImageTargetBuilder) buildDockerImage(docker *dockerClient.Client, art *artifact.DockerImageArtifact, feedStream *_DockerImageFeedStream, ctx buildcontext.Context) error {
	// Create the docker build context
	contextReader, contextWriter := io.Pipe()
	defer func() {
		contextReader.Close()
	}()
	// Run the context builder
	go func() {
		// Create a tar writer
		tarWriter := tar.NewWriter(contextWriter)
		var err error
		defer func() {
			tarWriter.Flush()
			tarWriter.Close()
			if err != nil {
				contextWriter.CloseWithError(err)
			} else {
				contextWriter.Close()
			}
		}()
		// Write
		err = feedStream.WriteTar(tarWriter)
	}()
	// Get the build options
	tags := []string{art.FullName()}
	if builder.spec.SetLatestTag {
		tags = append(tags, art.LatestName())
	}
	imageBuildOptions := types.ImageBuildOptions{
		Dockerfile:  "Dockerfile",
		Tags:        tags,
		Remove:      true,
		ForceRemove: true,
		PullParent:  true,
	}
	// Run docker build
	rsp, err := docker.ImageBuild(context.Background(), contextReader, imageBuildOptions)
	if err != nil {
		return err
	}
	defer rsp.Body.Close()
	// Read response
	scanner := bufio.NewScanner(rsp.Body)
	for scanner.Scan() {
		var data _DockerBuildResponseData
		text := scanner.Text()
		err := json.Unmarshal([]byte(text), &data)
		if err != nil {
			log.Warnf("Failed to decode docker response [%v]: %v", err, text)
			continue
		}
		// Check response data
		if data.Error == "" {
			if ctx.Verbose() {
				message := strings.Trim(data.Stream, " \t\t\n")
				if message != "" {
					log.Debugln("Docker -->", message)
				}
			}
		} else {
			// Error happened
			log.Errorf("Failed to run docker build. Error [%v] Code [%v]: Message: %v", data.Error, data.ErrorDetail.Code, data.ErrorDetail.Message)
			return errors.New("Docker build failed")
		}
	}
	if err := scanner.Err(); err != nil {
		log.Errorln("Failed to run docker build:", err)
		return errors.New("Docker build failed")
	}
	// Done
	return nil
}

func (builder *DockerImageTargetBuilder) pushDockerImage(c *dockerClient.Client, uri string) error {
	raw, _ := json.Marshal(types.AuthConfig{})
	rsp, err := c.ImagePush(context.Background(), uri, types.ImagePushOptions{RegistryAuth: base64.URLEncoding.EncodeToString(raw)})
	if err != nil {
		return err
	}
	defer rsp.Close()
	// Discard any data
	_, err = io.Copy(ioutil.Discard, rsp)
	return err
}

//
//	Docker image feed stream
//

type _DockerImageFeedStream struct {
	Files []_DockerImageFeedFile
	Bytes []_DockerImageFeedBytes
}

type _DockerImageFeedFile struct {
	Source string
	Dest   string
}

type _DockerImageFeedBytes struct {
	Bytes []byte
	Dest  string
}

func (s *_DockerImageFeedStream) FeedFile(localPath, destPath string) error {
	// Check local path
	if _, err := os.Stat(localPath); err != nil {
		log.Errorf("Failed to check local file [%v] when feeding docker stream: %v", localPath, err)
		return errors.New("Require to feed file")
	}
	// Add it
	if destPath == "" {
		log.Errorf("Require dest path for feeding docker stream")
		return errors.New("Require to feed file")
	}
	s.Files = append(s.Files, _DockerImageFeedFile{localPath, destPath})
	// Done
	return nil
}

func (s *_DockerImageFeedStream) FeedBytes(bytes []byte, destPath string) error {
	if destPath == "" {
		log.Errorf("Require dest path for feeding docker stream")
		return errors.New("Require to feed file")
	}
	s.Bytes = append(s.Bytes, _DockerImageFeedBytes{bytes, destPath})
	// Done
	return nil
}

// Write the stream to tar writer
func (s *_DockerImageFeedStream) WriteTar(writer *tar.Writer) error {
	// Write bytes
	log.Debugln("_DockerImageFeedStream.WriteTar: Write bytes")
	for _, feedBytes := range s.Bytes {
		log.Debugf("_DockerImageFeedStream.WriteTar: Write bytes: %v", feedBytes.Dest)
		if err := s.writeData2Tar([]byte(feedBytes.Bytes), feedBytes.Dest, writer); err != nil {
			return fmt.Errorf("Failed to write bytes to [%v]: %v", feedBytes.Dest, err)
		}
	}
	// Write file
	log.Debugln("_DockerImageFeedStream.WriteTar: Write files")
	for _, feedFile := range s.Files {
		log.Debugf("_DockerImageFeedStream.WriteTar: Write files: %v to %v", feedFile.Source, feedFile.Dest)
		if err := s.writePath2Tar(feedFile.Source, feedFile.Dest, writer); err != nil {
			return fmt.Errorf("Failed to write file [%v] to [%v]: %v", feedFile.Source, feedFile.Dest, err)
		}
	}
	// Done
	return nil
}

func (s *_DockerImageFeedStream) writePath2Tar(p string, targetPath string, writer *tar.Writer) error {
	p, err := utils.GetRealPath(p)
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
				subPath, err := utils.GetRealPath(filepath.Join(dirPath, info.Name()))
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
					if err := s.writeFile2Tar(subPath, info, filepath.Join(targetDirPath, name), writer); err != nil {
						return err
					}
				}
			}
			// Done
			return nil
		}
		// Walk through path
		return walkThroughPath(p, targetPath)
	}
	// A file
	return s.writeFile2Tar(p, info, targetPath, writer)
}

func (s *_DockerImageFeedStream) writeFile2Tar(p string, info os.FileInfo, name string, writer *tar.Writer) error {
	log.Debugf("_DockerImageFeedStream.writeFile2Tar: Write [%v] to [%v]", p, name)
	// Write header
	hdr, err := tar.FileInfoHeader(info, "")
	if err != nil {
		return fmt.Errorf("Failed to create tar header for file [%v] error: %v", p, err)
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

func (s *_DockerImageFeedStream) writeData2Tar(data []byte, name string, writer *tar.Writer) error {
	log.Debugf("_DockerImageFeedStream.writeData2Tar: Write to [%v]", name)
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

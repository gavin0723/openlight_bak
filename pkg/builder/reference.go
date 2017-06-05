// Author: lipixun
// File Name: reference.go
// Description:

package builder

import (
	"errors"

	log "github.com/Sirupsen/logrus"

	"path/filepath"

	"github.com/ops-openlight/openlight/pkg/repository"
)

// GetTargetByReference gets target by reference, path and target name
func (builder *Builder) GetTargetByReference(ref *repository.Reference, path, target string) (*repository.Target, error) {
	if ref == nil {
		return nil, errors.New("Require reference")
	}
	if target == "" {
		return nil, errors.New("Require target")
	}
	// Resolve reference
	repo, err := builder.resolveReference(ref)
	if err != nil {
		return nil, err
	}
	// Get target
	return repo.GetTarget(path, target)
}

func (builder *Builder) resolveReference(ref *repository.Reference) (*repository.LocalRepository, error) {
	if repo := builder.remoteRepos[ref.Spec().Remote]; repo != nil {
		return repo, nil
	}

	// Cache not found, resolve this reference
	log.Infof("Resolve: %v --> %v", ref.Name(), ref.Spec().Remote)

	// Try to find
	log.Debugln("Builder.resolveReference: Try to find reference by finders")
	if path := ref.Find(); path != "" {
		// Find it
		log.Infof("Resolve: %v --> %v by finder", ref.Name(), path)
		if ref.Spec().Path != "" {
			path = filepath.Join(path, ref.Spec().Path)
		}
		// Load repository
		repo, err := repository.NewLocalRepository(path)
		if err != nil {
			return nil, err
		}
		// Add to cache
		if builder.remoteRepos == nil {
			builder.remoteRepos = make(map[string]*repository.LocalRepository)
		}
		builder.remoteRepos[ref.Spec().Remote] = repo
		// Done
		return repo, nil
	}

	// TODO: Implement resolve from remote
	return nil, errors.New("Resolve reference from remote is not implemented yet")
}

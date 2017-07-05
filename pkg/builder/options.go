// Author: lipixun
// File Name: options.go
// Description:
//
//	Define all options
//

package builder

// Ensure the interfaces are implements
var _ BuildTargetOption = (*_DonotBuildTargetOption)(nil)
var _ BuildTargetOption = (*_UpdateDependencyOption)(nil)
var _ BuildTargetOption = (*_DonotBuildGoDependencyOption)(nil)
var _ BuildTargetOption = (*_DonotBuildPipDependencyOption)(nil)
var _ BuildTargetOption = (*_DonotBuildTargetDependencyOption)(nil)

//
//	Do not build target option
//

type _DonotBuildTargetOption struct{}

func (o *_DonotBuildTargetOption) set(options *_BuildTargetOption) {
	options.noBuild = true
}

func (o *_DonotBuildTargetOption) setdep(options *_BuildTargetDependencyOptions) {}

// WithDonotBuildTargetOption creates an option to avoid building target
func WithDonotBuildTargetOption() BuildTargetOption {
	return new(_DonotBuildTargetOption)
}

type _DonotInstallGoBinaryOption struct{}

func (o *_DonotInstallGoBinaryOption) set(options *_BuildTargetOption) {
	options.goBinaryTarget.NoInstall = true
}

func (o *_DonotInstallGoBinaryOption) setdep(options *_BuildTargetDependencyOptions) {}

// WithDonotInstallGoBinaryOption creates an option to avoid to install go binary
func WithDonotInstallGoBinaryOption() BuildTargetOption {
	return new(_DonotInstallGoBinaryOption)
}

type _IgnoreInstallGoBinaryErrorOption struct{}

func (o *_IgnoreInstallGoBinaryErrorOption) set(options *_BuildTargetOption) {
	options.goBinaryTarget.IgnoreInstallError = true
}

func (o *_IgnoreInstallGoBinaryErrorOption) setdep(options *_BuildTargetDependencyOptions) {}

// WithIgnoreInstallGoBinaryErrorOption creates an option to ignore go install error
func WithIgnoreInstallGoBinaryErrorOption() BuildTargetOption {
	return new(_IgnoreInstallGoBinaryErrorOption)
}

//
//	Update dependency option
//

type _UpdateDependencyOption struct{}

func (o *_UpdateDependencyOption) set(options *_BuildTargetOption) {}

func (o *_UpdateDependencyOption) setdep(options *_BuildTargetDependencyOptions) {
	options.Update = true
}

// WithUpdateDependencyOption creates an option to update dependency
func WithUpdateDependencyOption() BuildTargetOption {
	return new(_UpdateDependencyOption)
}

//
//	Do not build target dependency
//

type _DonotBuildTargetDependencyOption struct{}

func (o *_DonotBuildTargetDependencyOption) set(options *_BuildTargetOption) {}

func (o *_DonotBuildTargetDependencyOption) setdep(options *_BuildTargetDependencyOptions) {
	options.NoTarget = true
}

// WithDonotBuildTargetDependencyOption creates an option to avoid building target dependency
func WithDonotBuildTargetDependencyOption() BuildTargetOption {
	return new(_DonotBuildTargetDependencyOption)
}

//
//	Do not build pip dependency
//

type _DonotBuildPipDependencyOption struct{}

func (o *_DonotBuildPipDependencyOption) set(options *_BuildTargetOption) {}

func (o *_DonotBuildPipDependencyOption) setdep(options *_BuildTargetDependencyOptions) {
	options.NoPip = true
}

// WithDonotBuildPipDependencyOption creates an option to avoid building target dependency
func WithDonotBuildPipDependencyOption() BuildTargetOption {
	return new(_DonotBuildPipDependencyOption)
}

//
//	Do not build go dependency
//

type _DonotBuildGoDependencyOption struct{}

func (o *_DonotBuildGoDependencyOption) set(options *_BuildTargetOption) {}

func (o *_DonotBuildGoDependencyOption) setdep(options *_BuildTargetDependencyOptions) {
	options.NoGo = true
}

// WithDonotBuildGoDependencyOption creates an option to avoid building target dependency
func WithDonotBuildGoDependencyOption() BuildTargetOption {
	return new(_DonotBuildGoDependencyOption)
}

package poetry

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"

	"github.com/paketo-buildpacks/packit"
	"github.com/paketo-buildpacks/packit/postal"
)

//go:generate faux --interface DependencyManager --output fakes/dependency_manager.go
//go:generate faux --interface EntryResolver --output fakes/entry_resolver.go
//go:generate faux --interface InstallProcess --output fakes/install_process.go
//go:generate faux --interface SitePackageProcess --output fakes/site_package_process.go

// DependencyManager defines the interface for picking the best matching
// dependency and installing it.
type DependencyManager interface {
	Resolve(path, id, version, stack string) (postal.Dependency, error)
	Deliver(dependency postal.Dependency, cnbPath, destinationPath, platformPath string) error
	GenerateBillOfMaterials(dependencies ...postal.Dependency) []packit.BOMEntry
}

// EntryResolver defines the interface for picking the most relevant entry from
// the Buildpack Plan entries.
type EntryResolver interface {
	MergeLayerTypes(name string, entries []packit.BuildpackPlanEntry) (launch, build bool)
}

// InstallProcess defines the interface for installing the poetry dependency into a layer.
type InstallProcess interface {
	Execute(srcPath, targetLayerPath string) error
}

// SitePackageProcess defines the interface for looking site packages within a layer.
type SitePackageProcess interface {
	Execute(targetLayerPath string) (string, error)
}

func Build(dependencyManager DependencyManager, entryResolver EntryResolver, installProcess InstallProcess, siteProcess SitePackageProcess) packit.BuildFunc {
	return func(context packit.BuildContext) (packit.BuildResult, error) {

		dependency, err := dependencyManager.Resolve(filepath.Join(context.CNBPath, "buildpack.toml"), "poetry", "*", context.Stack)
		if err != nil {
			return packit.BuildResult{}, err
		}

		bom := dependencyManager.GenerateBillOfMaterials(dependency)

		poetryLayer, err := context.Layers.Get("poetry")
		if err != nil {
			panic(err)
		}

		poetryLayer.Launch, poetryLayer.Build = entryResolver.MergeLayerTypes("poetry", context.Plan.Entries)
		poetryLayer.Cache = poetryLayer.Build

		var buildMetadata = packit.BuildMetadata{}
		var launchMetadata = packit.LaunchMetadata{}

		// Install the poetry source to a temporary dir, since we only need access to
		// it as an intermediate step when installing poetry.
		// It doesn't need to go into a layer, since we won't need it in future builds.
		poetrySrcDir, err := ioutil.TempDir("", "poetry-source")
		if err != nil {
			return packit.BuildResult{}, fmt.Errorf("failed to create temp poetry-source dir: %w", err)
		}

		err = dependencyManager.Deliver(dependency, context.CNBPath, poetrySrcDir, context.Platform.Path)
		if err != nil {
			panic(err)
		}

		err = installProcess.Execute(poetrySrcDir, poetryLayer.Path)
		if err != nil {
			panic(err)
		}

		// Look up the site packages path and prepend it onto $PYTHONPATH
		sitePackagesPath, err := siteProcess.Execute(poetryLayer.Path)
		if err != nil {
			return packit.BuildResult{}, fmt.Errorf("failed to locate site packages in poetry layer: %w", err)
		}
		if sitePackagesPath == "" {
			return packit.BuildResult{}, fmt.Errorf("poetry installation failed: site packages are missing from the poetry layer")
		}

		poetryLayer.SharedEnv.Prepend("PYTHONPATH", strings.TrimRight(sitePackagesPath, "\n"), ":")

		if poetryLayer.Build {
			buildMetadata = packit.BuildMetadata{BOM: bom}
		}

		if poetryLayer.Launch {
			launchMetadata = packit.LaunchMetadata{BOM: bom}
		}
		return packit.BuildResult{
			Layers: []packit.Layer{poetryLayer},
			Launch: launchMetadata,
			Build:  buildMetadata,
		}, nil
	}
}

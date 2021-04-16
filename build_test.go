package poetry_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/paketo-buildpacks/packit"
	"github.com/paketo-buildpacks/packit/postal"
	"github.com/paketo-community/poetry"
	"github.com/paketo-community/poetry/fakes"
	"github.com/sclevine/spec"

	. "github.com/onsi/gomega"
)

func testBuild(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect = NewWithT(t).Expect

		layersDir string
		cnbDir    string

		dependencyManager *fakes.DependencyManager
		entryResolver     *fakes.EntryResolver
		installProcess    *fakes.InstallProcess
		siteProcess       *fakes.SitePackageProcess

		build packit.BuildFunc
	)

	it.Before(func() {
		var err error
		layersDir, err = os.MkdirTemp("", "layers")
		Expect(err).NotTo(HaveOccurred())

		cnbDir, err = os.MkdirTemp("", "cnb")
		Expect(err).NotTo(HaveOccurred())

		dependencyManager = &fakes.DependencyManager{}
		dependencyManager.ResolveCall.Returns.Dependency = postal.Dependency{
			ID:      "poetry",
			Name:    "poetry-dependency-name",
			SHA256:  "poetry-dependency-sha",
			Stacks:  []string{"some-stack"},
			URI:     "poetry-dependency-uri",
			Version: "poetry-dependency-version",
		}

		dependencyManager.GenerateBillOfMaterialsCall.Returns.BOMEntrySlice = []packit.BOMEntry{
			{
				Name: "poetry",
				Metadata: map[string]interface{}{
					"version": "poetry-dependency-version",
					"name":    "poetry-dependency-name",
					"sha256":  "poetry-dependency-sha",
					"stacks":  []string{"some-stack"},
					"uri":     "poetry-dependency-uri",
				},
			},
		}

		entryResolver = &fakes.EntryResolver{}
		installProcess = &fakes.InstallProcess{}
		siteProcess = &fakes.SitePackageProcess{}
		siteProcess.ExecuteCall.Returns.String = filepath.Join(layersDir, "poetry", "lib", "python1.23", "site-packages")

		build = poetry.Build(dependencyManager, entryResolver, installProcess, siteProcess)
	})

	it("returns a result that installs poetry", func() {
		result, err := build(packit.BuildContext{
			BuildpackInfo: packit.BuildpackInfo{
				Name:    "Some Buildpack",
				Version: "some-version",
			},
			CNBPath: cnbDir,
			Plan: packit.BuildpackPlan{
				Entries: []packit.BuildpackPlanEntry{
					{Name: "poetry"},
				},
			},
			Platform: packit.Platform{Path: "some-platform-path"},
			Layers:   packit.Layers{Path: layersDir},
			Stack:    "some-stack",
		})
		Expect(err).NotTo(HaveOccurred())

		Expect(result).To(Equal(packit.BuildResult{
			Layers: []packit.Layer{
				{
					Name: "poetry",
					Path: filepath.Join(layersDir, "poetry"),
					SharedEnv: packit.Environment{
						"PYTHONPATH.delim":   ":",
						"PYTHONPATH.prepend": filepath.Join(layersDir, "poetry", "lib/python1.23/site-packages"),
					},
					BuildEnv:         packit.Environment{},
					LaunchEnv:        packit.Environment{},
					ProcessLaunchEnv: map[string]packit.Environment{},
					Build:            false,
					Launch:           false,
					Cache:            false,
					// Metadata: map[string]interface{}{
					// 	miniconda.DepKey: "miniconda3-dependency-sha",
					// 	"built_at":       timeStamp.Format(time.RFC3339Nano),
					// },
				},
			},
		}))

		Expect(dependencyManager.ResolveCall.Receives.Path).To(Equal(filepath.Join(cnbDir, "buildpack.toml")))
		Expect(dependencyManager.ResolveCall.Receives.Id).To(Equal("poetry"))
		Expect(dependencyManager.ResolveCall.Receives.Version).To(Equal("*"))
		Expect(dependencyManager.ResolveCall.Receives.Stack).To(Equal("some-stack"))

		Expect(dependencyManager.GenerateBillOfMaterialsCall.Receives.Dependencies).To(Equal([]postal.Dependency{
			{
				ID:      "poetry",
				Name:    "poetry-dependency-name",
				SHA256:  "poetry-dependency-sha",
				Stacks:  []string{"some-stack"},
				URI:     "poetry-dependency-uri",
				Version: "poetry-dependency-version",
			},
		}))

		Expect(entryResolver.MergeLayerTypesCall.Receives.Name).To(Equal("poetry"))
		Expect(entryResolver.MergeLayerTypesCall.Receives.Entries).To(Equal([]packit.BuildpackPlanEntry{
			{Name: "poetry"},
		}))

		Expect(dependencyManager.DeliverCall.Receives.Dependency).To(Equal(
			postal.Dependency{
				ID:      "poetry",
				Name:    "poetry-dependency-name",
				SHA256:  "poetry-dependency-sha",
				Stacks:  []string{"some-stack"},
				URI:     "poetry-dependency-uri",
				Version: "poetry-dependency-version",
			}))
		Expect(dependencyManager.DeliverCall.Receives.CnbPath).To(Equal(cnbDir))
		Expect(dependencyManager.DeliverCall.Receives.DestinationPath).To(ContainSubstring("poetry-source"))
		Expect(dependencyManager.DeliverCall.Receives.PlatformPath).To(Equal("some-platform-path"))

		Expect(installProcess.ExecuteCall.Receives.SrcPath).To(Equal(dependencyManager.DeliverCall.Receives.DestinationPath))
		Expect(installProcess.ExecuteCall.Receives.TargetLayerPath).To(Equal(filepath.Join(layersDir, "poetry")))

		// Expect(buffer.String()).To(ContainSubstring("Some Buildpack some-version"))
		// Expect(buffer.String()).To(ContainSubstring("Executing build process"))
		// Expect(buffer.String()).To(ContainSubstring("Installing Miniconda"))
	})

	context("when buildplan entries require poetry at build/launch", func() {
		it.Before(func() {
			entryResolver.MergeLayerTypesCall.Returns.Launch = true
			entryResolver.MergeLayerTypesCall.Returns.Build = true
		})
		it("returns a layer with build and launch set true and the BOM is set for build and launch", func() {
			result, err := build(packit.BuildContext{
				BuildpackInfo: packit.BuildpackInfo{
					Name:    "Some Buildpack",
					Version: "some-version",
				},
				CNBPath: cnbDir,
				Plan: packit.BuildpackPlan{
					Entries: []packit.BuildpackPlanEntry{
						{
							Name: "poetry",
							Metadata: map[string]interface{}{
								"build": true,
							},
						},
						{
							Name: "poetry",
							Metadata: map[string]interface{}{
								"launch": true,
							},
						},
					},
				},
				Platform: packit.Platform{Path: "some-platform-path"},
				Layers:   packit.Layers{Path: layersDir},
				Stack:    "some-stack",
			})
			Expect(err).NotTo(HaveOccurred())

			Expect(result).To(Equal(packit.BuildResult{
				Layers: []packit.Layer{
					{
						Name: "poetry",
						Path: filepath.Join(layersDir, "poetry"),
						SharedEnv: packit.Environment{
							"PYTHONPATH.delim":   ":",
							"PYTHONPATH.prepend": filepath.Join(layersDir, "poetry", "lib/python1.23/site-packages"),
						},
						BuildEnv:         packit.Environment{},
						LaunchEnv:        packit.Environment{},
						ProcessLaunchEnv: map[string]packit.Environment{},
						Build:            true,
						Launch:           true,
						Cache:            true,
						// Metadata: map[string]interface{}{
						// 	poetry.DepKey: "poetry-dependency-sha",
						// 	"built_at":       timeStamp.Format(time.RFC3339Nano),
						// },
					},
				},
				Launch: packit.LaunchMetadata{
					BOM: []packit.BOMEntry{
						{
							Name: "poetry",
							Metadata: map[string]interface{}{
								"version": "poetry-dependency-version",
								"name":    "poetry-dependency-name",
								"sha256":  "poetry-dependency-sha",
								"stacks":  []string{"some-stack"},
								"uri":     "poetry-dependency-uri",
							},
						},
					},
				},
				Build: packit.BuildMetadata{
					BOM: []packit.BOMEntry{
						{
							Name: "poetry",
							Metadata: map[string]interface{}{
								"version": "poetry-dependency-version",
								"name":    "poetry-dependency-name",
								"sha256":  "poetry-dependency-sha",
								"stacks":  []string{"some-stack"},
								"uri":     "poetry-dependency-uri",
							},
						},
					},
				},
			}))
		})
	})
}

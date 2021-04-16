package poetry

import "github.com/paketo-buildpacks/packit"

//go:generate faux --interface ProjectParser --output fakes/project_parser.go
type ProjectParser interface {
	Parse(path string) (detected bool, version string, err error)
}

type BuildPlanMetadata struct {
	Version       string `toml:"version,omitempty"`
	VersionSource string `toml:"version-source,omitempty"`
	Build         bool   `toml:"build"`
}

func Detect(pyProjParser ProjectParser) packit.DetectFunc {
	return func(context packit.DetectContext) (packit.DetectResult, error) {

		pythonRequirement := packit.BuildPlanRequirement{
			Name: "cpython",
			Metadata: BuildPlanMetadata{
				Build: true,
			},
		}

		detected, pythonVersion, err := pyProjParser.Parse(context.WorkingDir)
		if err != nil {
			panic(err)
		}

		if !detected {
			return packit.DetectResult{}, packit.Fail
		}

		if pythonVersion != "" {
			pythonRequirement = packit.BuildPlanRequirement{
				Name: "cpython",
				Metadata: BuildPlanMetadata{
					Version:       pythonVersion,
					VersionSource: "pyproject.toml",
					Build:         true,
				},
			}
		}

		return packit.DetectResult{
			Plan: packit.BuildPlan{
				Provides: []packit.BuildPlanProvision{
					{Name: "poetry"},
				},
				Requires: []packit.BuildPlanRequirement{
					pythonRequirement,
					{
						Name: Pip,
						Metadata: BuildPlanMetadata{
							Build: true,
						},
					},
				},
			},
		}, nil
	}
}

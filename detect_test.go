package poetry_test

import (
	"testing"

	"github.com/paketo-buildpacks/packit"
	"github.com/paketo-community/poetry"
	"github.com/paketo-community/poetry/fakes"
	"github.com/sclevine/spec"

	. "github.com/onsi/gomega"
)

func testDetect(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect = NewWithT(t).Expect

		pyProjParser *fakes.ProjectParser
		detect       packit.DetectFunc
	)

	it.Before(func() {
		pyProjParser = &fakes.ProjectParser{}
		pyProjParser.ParseCall.Returns.Detected = true

		detect = poetry.Detect(pyProjParser)
	})

	it("returns a plan that provides poetry", func() {
		result, err := detect(packit.DetectContext{
			WorkingDir: "/working-dir",
		})
		Expect(err).NotTo(HaveOccurred())
		Expect(result).To(Equal(packit.DetectResult{
			Plan: packit.BuildPlan{
				Provides: []packit.BuildPlanProvision{
					{Name: poetry.Poetry},
				},
				Requires: []packit.BuildPlanRequirement{
					{
						Name: poetry.CPython,
						Metadata: poetry.BuildPlanMetadata{
							Build: true,
						},
					},
					{
						Name: poetry.Pip,
						Metadata: poetry.BuildPlanMetadata{
							Build: true,
						},
					},
				},
			},
		}))
		Expect(pyProjParser.ParseCall.Receives.Path).To(Equal("/working-dir"))
	})

	context("when pyproject.toml provides a Python version", func() {
		it.Before(func() {
			pyProjParser.ParseCall.Returns.Version = "3.8"
		})

		it("returns a plan that provides poetry and requires specific Python version", func() {
			result, err := detect(packit.DetectContext{
				WorkingDir: "/working-dir",
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(result).To(Equal(packit.DetectResult{
				Plan: packit.BuildPlan{
					Provides: []packit.BuildPlanProvision{
						{Name: "poetry"},
					},
					Requires: []packit.BuildPlanRequirement{
						{
							Name: "cpython",
							Metadata: poetry.BuildPlanMetadata{
								Version:       "3.8",
								VersionSource: "pyproject.toml",
								Build:         true,
							},
						},
						{
							Name: poetry.Pip,
							Metadata: poetry.BuildPlanMetadata{
								Build: true,
							},
						},
					},
				},
			}))
		})
	})

	context("when poetry is not detected", func() {
		it.Before(func() {
			pyProjParser.ParseCall.Returns.Detected = false
		})
		it("fails detection", func() {
			result, err := detect(packit.DetectContext{
				WorkingDir: "/working-dir",
			})
			Expect(err).To(MatchError(packit.Fail))
			Expect(result).To(Equal(packit.DetectResult{}))
		})
	})
}

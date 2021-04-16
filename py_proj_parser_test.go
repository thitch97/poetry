package poetry_test

import (
	"io/ioutil"
	"path/filepath"
	"testing"

	"github.com/paketo-community/poetry"
	"github.com/sclevine/spec"

	. "github.com/onsi/gomega"
)

func testPyProjParser(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect = NewWithT(t).Expect

		workingDir   string
		pyProjParser poetry.PyProjParser
	)

	it.Before(func() {
		var err error
		workingDir, err = ioutil.TempDir("", "working-dir")
		Expect(err).NotTo(HaveOccurred())

		pyProjParser = poetry.NewPyProjParser()
	})

	context("Parse", func() {

		context("when pyproject.toml contains poetry configuration", func() {
			it.Before(func() {
				Expect(ioutil.WriteFile(filepath.Join(workingDir, "pyproject.toml"), []byte(`[tool.poetry]
name = "some-app"
version = "some-version"

[tool.poetry.dependencies]
python = "*"`), 0644)).To(Succeed())
			})

			it("parses configuration", func() {
				detected, pyVersion, err := pyProjParser.Parse(workingDir)
				Expect(err).NotTo(HaveOccurred())
				Expect(detected).To(Equal(true))
				Expect(pyVersion).To(Equal("*"))
			})
		})

	})
}

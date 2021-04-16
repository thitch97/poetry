package poetry_test

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"testing"

	"github.com/paketo-buildpacks/packit/pexec"
	"github.com/paketo-community/poetry"
	"github.com/paketo-community/poetry/fakes"
	"github.com/sclevine/spec"

	. "github.com/onsi/gomega"
)

func testPoetryInstallProcess(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect = NewWithT(t).Expect

		srcLayerPath    string
		targetLayerPath string
		executable      *fakes.Executable

		poetryInstallProcess poetry.PoetryInstallProcess
	)

	it.Before(func() {
		var err error
		srcLayerPath, err = ioutil.TempDir("", "poetry-source")
		Expect(err).NotTo(HaveOccurred())

		targetLayerPath, err = ioutil.TempDir("", "poetry")
		Expect(err).NotTo(HaveOccurred())

		executable = &fakes.Executable{}

		poetryInstallProcess = poetry.NewPoetryInstallProcess(executable)
	})

	context("Execute", func() {
		context("there is a poetry dependency to install", func() {
			it("installs it to the poetry layer", func() {
				err := poetryInstallProcess.Execute(srcLayerPath, targetLayerPath)
				Expect(err).NotTo(HaveOccurred())

				Expect(executable.ExecuteCall.Receives.Execution.Env).To(Equal(append(os.Environ(), fmt.Sprintf("PYTHONUSERBASE=%s", targetLayerPath))))
				Expect(executable.ExecuteCall.Receives.Execution.Args).To(Equal([]string{"install", "poetry", "--user", fmt.Sprintf("--find-links=%s", srcLayerPath)}))
			})
		})

		context("failure cases", func() {
			context("the poetry install process fails", func() {
				it.Before(func() {
					executable.ExecuteCall.Stub = func(execution pexec.Execution) error {
						fmt.Fprintln(execution.Stdout, "stdout output")
						fmt.Fprintln(execution.Stderr, "stderr output")
						return errors.New("installing poetry failed")
					}
				})

				it("returns an error", func() {
					err := poetryInstallProcess.Execute(srcLayerPath, targetLayerPath)
					Expect(err).To(MatchError(ContainSubstring("installing poetry failed")))
					Expect(err).To(MatchError(ContainSubstring("stdout output")))
					Expect(err).To(MatchError(ContainSubstring("stderr output")))
				})
			})
		})
	})
}

package poetry_test

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/paketo-buildpacks/packit/pexec"
	poetry "github.com/paketo-community/poetry"
	"github.com/paketo-community/poetry/fakes"
	"github.com/sclevine/spec"

	. "github.com/onsi/gomega"
)

func testSiteProcess(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect = NewWithT(t).Expect

		targetLayerPath string
		executable      *fakes.Executable

		siteProcess poetry.SiteProcess
	)

	it.Before(func() {
		var err error
		targetLayerPath, err = ioutil.TempDir("", "poetry")
		Expect(err).NotTo(HaveOccurred())

		executable = &fakes.Executable{}
		executable.ExecuteCall.Stub = func(execution pexec.Execution) error {
			if execution.Stdout != nil {
				fmt.Fprint(execution.Stdout, targetLayerPath, "/poetry/lib/python/site-packages")
			}
			return nil
		}

		siteProcess = poetry.NewSiteProcess(executable)
	})

	it.After(func() {
		Expect(os.RemoveAll(targetLayerPath)).To(Succeed())
	})

	context("Execute", func() {
		context("there are site packages in the poetry layer", func() {
			it("returns the full path to the packages", func() {
				sitePackagesPath, err := siteProcess.Execute(targetLayerPath)
				Expect(err).NotTo(HaveOccurred())

				Expect(executable.ExecuteCall.Receives.Execution.Env).To(Equal(append(os.Environ(), fmt.Sprintf("PYTHONUSERBASE=%s", targetLayerPath))))
				Expect(executable.ExecuteCall.Receives.Execution.Args).To(Equal([]string{"-m", "site", "--user-site"}))

				Expect(sitePackagesPath).To(Equal(filepath.Join(targetLayerPath, "poetry", "lib", "python", "site-packages")))
			})
		})

		context("failure cases", func() {
			context("site package lookup fails", func() {
				it.Before(func() {
					executable.ExecuteCall.Stub = func(execution pexec.Execution) error {
						fmt.Fprintln(execution.Stdout, "stdout output")
						fmt.Fprintln(execution.Stderr, "stderr output")
						return errors.New("locating site packages failed")
					}
				})

				it("returns an error", func() {
					_, err := siteProcess.Execute(targetLayerPath)
					Expect(err).To(MatchError(ContainSubstring("failed to locate site packages:")))
					Expect(err).To(MatchError(ContainSubstring("stderr output")))
					Expect(err).To(MatchError(ContainSubstring("error: locating site packages failed")))
				})
			})
		})
	})
}

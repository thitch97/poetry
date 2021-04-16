package poetry_test

import (
	"testing"

	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"
)

func TestUnitPoetry(t *testing.T) {
	suite := spec.New("pipenv", spec.Report(report.Terminal{}))
	suite("Detect", testDetect)
	suite("Build", testBuild)
	suite("PyProjParser", testPyProjParser)
	suite("InstallProcess", testPoetryInstallProcess)
	suite("SiteProcess", testSiteProcess)
	suite.Run(t)
}

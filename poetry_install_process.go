package poetry

import (
	"bytes"
	"fmt"
	"os"

	"github.com/paketo-buildpacks/packit/pexec"
)

//go:generate faux --interface Executable --output fakes/executable.go

// Executable defines the interface for invoking an executable.
type Executable interface {
	Execute(pexec.Execution) error
}

// PoetryInstallProcess implements the InstallProcess interface.
type PoetryInstallProcess struct {
	executable Executable
}

// NewPoetryInstallProcess creates an instance of the PoetryInstallProcess given an Executable that runs `python`.
func NewPoetryInstallProcess(executable Executable) PoetryInstallProcess {
	return PoetryInstallProcess{
		executable: executable,
	}
}

// Execute installs the pip binary from source code located in the given srcPath into the a layer path designated by targetLayerPath.
func (p PoetryInstallProcess) Execute(srcPath, targetLayerPath string) error {
	buffer := bytes.NewBuffer(nil)

	err := p.executable.Execute(pexec.Execution{
		// Install pip from source with the pip that comes pre-installed with cpython
		Args: []string{"install", "poetry", "--user", fmt.Sprintf("--find-links=%s", srcPath)},
		// Set the PYTHONUSERBASE to ensure that pip is installed to the newly created target layer.
		Env:    append(os.Environ(), fmt.Sprintf("PYTHONUSERBASE=%s", targetLayerPath)),
		Stdout: buffer,
		Stderr: buffer,
	})

	if err != nil {
		return fmt.Errorf("failed to configure poetry:\n%s\nerror: %w", buffer.String(), err)
	}
	return nil
}

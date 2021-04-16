package poetry

import (
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

type PyProjParser struct{}

func NewPyProjParser() PyProjParser {
	return PyProjParser{}
}

func (p PyProjParser) Parse(path string) (bool, string, error) {

	pyProject, err := os.Open(filepath.Join(path, PyProject))
	if err != nil {
		panic(err)
	}

	var pyProjectTOML struct {
		Tool struct {
			Poetry struct {
				Name         string `toml:"name"`
				Dependencies struct {
					Python string `toml:"python"`
				} `toml:"dependencies"`
			} `toml:"poetry"`
		} `toml:"tool"`
	}

	_, err = toml.DecodeReader(pyProject, &pyProjectTOML)
	if err != nil {
		panic(err)
	}

	var detected bool
	pyVersion := pyProjectTOML.Tool.Poetry.Dependencies.Python

	if pyProjectTOML.Tool.Poetry.Name != "" {
		detected = true
	}

	return detected, pyVersion, nil
}

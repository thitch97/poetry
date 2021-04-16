package main

import (
	"github.com/paketo-buildpacks/packit"
	"github.com/paketo-buildpacks/packit/cargo"
	"github.com/paketo-buildpacks/packit/draft"
	"github.com/paketo-buildpacks/packit/pexec"
	"github.com/paketo-buildpacks/packit/postal"
	"github.com/paketo-community/poetry"
)

func main() {
	pyProjectParser := poetry.NewPyProjParser()
	dependencyManager := postal.NewService(cargo.NewTransport())
	entryResolver := draft.NewPlanner()
	installProcess := poetry.NewPoetryInstallProcess(pexec.NewExecutable("pip"))
	siteProcess := poetry.NewSiteProcess(pexec.NewExecutable("python"))

	packit.Run(
		poetry.Detect(pyProjectParser),
		poetry.Build(dependencyManager, entryResolver, installProcess, siteProcess),
	)
}

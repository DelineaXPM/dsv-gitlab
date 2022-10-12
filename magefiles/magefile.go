// ⚡ Core Mage Tasks.
package main

import (
	"os"

	"github.com/DelineaXPM/dsv-gitlab/magefiles/constants"

	"github.com/bitfield/script"
	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/sh"
	"github.com/pterm/pterm"
	"github.com/sheldonhull/magetools/ci"
	"github.com/sheldonhull/magetools/pkg/magetoolsutils"

	// mage:import
	"github.com/sheldonhull/magetools/gotools"
	//mage:import
	_ "github.com/sheldonhull/magetools/secrets"
)

// createDirectories creates the local working directories for build artifacts and tooling.
func createDirectories() error {
	magetoolsutils.CheckPtermDebug()
	for _, dir := range []string{constants.ArtifactDirectory, constants.CacheDirectory} {
		if err := os.MkdirAll(dir, constants.PermissionUserReadWriteExecute); err != nil {
			pterm.Error.Printf("failed to create dir: [%s] with error: %v\n", dir, err)

			return err
		}
		pterm.Success.Printf("✅ [%s] dir created\n", dir)
	}

	return nil
}

// Init runs multiple tasks to initialize all the requirements for running a project for a new contributor.
func Init() error { //nolint:deadcode // Not dead, it's alive.
	magetoolsutils.CheckPtermDebug()
	pterm.DefaultHeader.Println("running Init()")

	mg.SerialDeps(
		Clean,
		createDirectories,
		(gotools.Go{}.Tidy),
	)

	if ci.IsCI() {
		installArgs := []string{}

		if mg.Verbose() {
			installArgs = append(installArgs, "--log-level")
			installArgs = append(installArgs, "debug")
		}
		installArgs = append(installArgs, "install")
		installArgs = append(installArgs, "aqua")
		pterm.DefaultSection.Printfln("aqua install ci dependencies")
		if err := sh.RunWithV(map[string]string{"AQUA_CONFIG": "aqua.ci.yaml"}, "aqua", installArgs...); err != nil {
			pterm.Error.Printfln("aqua-ci%v", err)
			return err
		}
		pterm.Success.Println("aqua install ci dependencies")
		pterm.Debug.Println("CI detected, done with init")
		return nil
	}
	pterm.DefaultSection.Printfln("aqua install dev dependencies")
	pterm.DefaultSection.Println("Aqua install")
	if err := sh.RunV("aqua", "install"); err != nil {
		return err
	}
	pterm.Success.Println("aqua install dev dependencies")
	return nil
}

// InstallTrunk installs trunk.io tooling.
func InstallTrunk() error {
	magetoolsutils.CheckPtermDebug()
	_, err := script.Exec("curl https://get.trunk.io -fsSL").Exec("bash").Stdout()
	if err != nil {
		return err
	}
	return nil
}

// InstallSyft installs SBOM tooling for goreleaser.
func InstallSyft() error {
	magetoolsutils.CheckPtermDebug()
	_, err := script.Exec("curl -sSfL https://raw.githubusercontent.com/anchore/syft/main/install.sh").Exec("sh -s -- -b /usr/local/bin").Stdout()
	if err != nil {
		return err
	}
	return nil
}

// Clean up after yourself.
func Clean() {
	magetoolsutils.CheckPtermDebug()
	pterm.Success.Println("Cleaning...")
	for _, dir := range []string{constants.ArtifactDirectory, constants.CacheDirectory} {
		err := os.RemoveAll(dir)
		if err != nil {
			pterm.Error.Printf("failed to removeall: [%s] with error: %v\n", dir, err)
		}
		pterm.Success.Printf("🧹 [%s] dir removed\n", dir)
	}
	mg.Deps(createDirectories)
}

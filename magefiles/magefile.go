// ⚡ Core Mage Tasks.
package main

import (
	"os"
	"os/exec"

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

// Test contains mage tasks for testing.
type Test mg.Namespace

// createDirectories creates the local working directories for build artifacts and tooling.
func createDirectories() error {
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

	pterm.Success.Println("running Init()...")

	mg.SerialDeps(
		Clean,
		createDirectories,
		(gotools.Go{}.Tidy),
	)

	_, err := exec.LookPath("aqua")
	if err != nil && os.IsNotExist(err) {
		pterm.Error.Printfln("unable to resolve aqua cli tool, please install for automated project tooling setup: https://aquaproj.github.io/docs/tutorial-basics/quick-start#install-aqua")
		return err
	}

	if ci.IsCI() {
		installArgs := []string{}

		if mg.Verbose() {
			installArgs = append(installArgs, "--log-level")
			installArgs = append(installArgs, "debug")
		}
		installArgs = append(installArgs, "install")
		installArgs = append(installArgs, "aqua")
		if err := sh.RunWithV(map[string]string{"AQUA_CONFIG": "aqua.ci.yaml"}, "aqua", installArgs...); err != nil {
			pterm.Error.Printfln("aqua-ci%v", err)
			return err
		}
		pterm.Debug.Println("CI detected, done with init")
		return nil
	}

	pterm.DefaultSection.Println("Aqua install")
	if err := sh.RunV("aqua", "install"); err != nil {
		return err
	}
	// These can run in parallel as different toolchains.
	mg.Deps(
		(InstallTrunk),
	)
	return nil
}

// Clean up after yourself.
func Clean() {
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

// InstallTrunk installs trunk.io tooling.
func InstallTrunk() error {
	_, err := script.Exec("curl https://get.trunk.io -fsSL").Exec("bash -s -- -y").Stdout()
	if err != nil {
		return err
	}
	return nil
}

// Unit runs go unit tests.
func (Test) Unit() {
	mg.Deps(gotools.Go{}.TestSum("./..."))
}

// VulnCheck runs Go's vulncheck tool to scan dependencies.
func VulnCheck() error {
	magetoolsutils.CheckPtermDebug()
	return sh.RunV("govulncheck", "./...")
}

// ⚡ Core Mage Tasks.
package main

import (
	"errors"
	"os"
	"runtime"

	survey "github.com/AlecAivazis/survey/v2"
	"github.com/DelineaXPM/dsv-gitlab/magefiles/constants"
	"github.com/bitfield/script"
	"github.com/caarlos0/env/v8"
	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/sh"
	"github.com/pterm/pterm"
	"github.com/sheldonhull/magetools/ci"
	"github.com/sheldonhull/magetools/pkg/magetoolsutils"
	yaml "gopkg.in/yaml.v3"

	// mage:import
	"github.com/sheldonhull/magetools/gotools"
)

// Test contains tasks related to testing.
type Test mg.Namespace

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
		// installArgs := []string{}

		// if mg.Verbose() {
		// 	installArgs = append(installArgs, "--log-level")
		// 	installArgs = append(installArgs, "debug")
		// }
		// installArgs = append(installArgs, "install")
		// installArgs = append(installArgs, "aqua")
		// pterm.DefaultSection.Printfln("aqua install ci dependencies")
		// if err := sh.RunV("aqua", installArgs...); err != nil {
		// 	pterm.Error.Printfln("aqua-ci%v", err)
		// 	return err
		// }
		// pterm.Success.Println("aqua install ci dependencies")
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

// InstallAqua runs bash installer for aqua.
func InstallAqua() error {
	magetoolsutils.CheckPtermDebug()
	_, err := script.Exec("curl -sSfL https://raw.githubusercontent.com/aquaproj/aqua-installer/v1.1.2/aqua-installer").Exec("bash").Stdout()
	if err != nil {
		return err
	}
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

// InstallGitLabCILocal installs `gitlab-ci-local` tool for validating pipeline in local testing without running Gitlab separately.
func InstallGitLabCILocal() (err error) {
	magetoolsutils.CheckPtermDebug()

	switch currentGOOS := runtime.GOOS; currentGOOS {
	case "linux":
		_, err = script.Exec("bash .devcontainer/library-scripts/install-gitlab-ci.sh").Stdout()
	case "darwin":
		os.Setenv("HOMEBREW_NO_AUTO_UPDATE", "1")
		_, err = script.Exec("brew install gitlab-ci-local").Stdout()
	case "windows":
		pterm.Warning.Printfln("not sure if this will run right in windows, you should use WSL2 😀")
		pterm.Warning.Printfln("See git bash directions if you aren't certain you want to try that: https://github.com/firecow/gitlab-ci-local#windows-git-bash")
		err = errors.New("can't automatically setup windows")
	default:
		err = errors.New("unsupported")
	}

	if err != nil {
		return err
	}
	return nil
}

// SetupLocalTest creates the required secrets file for the local integration tests to run.
func (Test) SetupLocal() error {
	response, err := pterm.DefaultInteractiveConfirm.Show("Do you want to manually inject the environment variables?")
	if err != nil {
		return err
	}
	pterm.Info.Printfln("response: %v", response)
	if response {
		answers := struct {
			DSVDomain       string
			DSVClientID     string
			DSVClientSecret string
		}{}
		qs := []*survey.Question{
			{
				Name:     "DSVDomain",
				Prompt:   &survey.Input{Message: "DSV_DOMAIN"},
				Validate: survey.Required,
			},
			{
				Name:   "DSVClientID",
				Prompt: &survey.Input{Message: "DSV_CLIENT_ID"},
			},
			{
				Name:   "DSVClientSecret",
				Prompt: &survey.Password{Message: "DSV_CLIENT_SECRET"},
			},
		}
		err := survey.Ask(qs, &answers)
		if err != nil {
			pterm.Error.Println("issue collecting input")
			return err
		}
		os.Setenv("DSV_DOMAIN", answers.DSVDomain)
		os.Setenv("DSV_CLIENT_ID", answers.DSVClientID)
		os.Setenv("DSV_CLIENT_SECRET", answers.DSVClientSecret)
	}
	//nolint:tagliatelle // environment variables
	type TestConfig struct {
		DSVDomain       string `yaml:"DSV_DOMAIN" env:"DSV_DOMAIN"`
		DSVClientID     string `yaml:"DSV_CLIENT_ID" env:"DSV_CLIENT_ID"`
		DSVClientSecret string `yaml:"DSV_CLIENT_SECRET" env:"DSV_CLIENT_SECRET"`
	}
	cfg := &TestConfig{}
	opts := env.Options{RequiredIfNoDef: true}
	// Load env vars.
	if err := env.ParseWithOptions(cfg, opts); err != nil {
		pterm.Error.Printfln("unable to parse required environment variables: %v", err)
		return err
	}
	bites, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}
	if err := os.WriteFile(constants.GitLabCILocalVariablesFile, bites, constants.PermissionUserReadWriteExecute); err != nil {
		return err
	}
	pterm.Success.Printfln("generated file: %s", constants.GitLabCILocalVariablesFile)
	return nil
}

// Integration test runs gitlab-ci-local and verifies the functionality of the pipeline locally.
func (Test) Integration() error {
	return sh.RunWithV(map[string]string{}, "gitlab-ci-local", "--shell-isolation")
}

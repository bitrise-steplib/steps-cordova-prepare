package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/bitrise-community/steps-cordova-archive/cordova"
	"github.com/bitrise-community/steps-ionic-archive/jsdependency"
	"github.com/bitrise-io/go-utils/colorstring"
	"github.com/bitrise-io/go-utils/errorutil"
	"github.com/bitrise-io/go-utils/log"
	"github.com/bitrise-io/go-utils/pathutil"
	"github.com/bitrise-tools/go-steputils/stepconf"
)

type config struct {
	Platform string `env:"platform,opt[ios,android,'ios,android']"`
	Version  string `env:"cordova_version"`
	WorkDir  string `env:"workdir,dir"`
}

func failf(format string, v ...interface{}) {
	log.Errorf(format, v...)
	os.Exit(1)
}

func installDependency(packageManager jsdependency.Tool, name string, version string) error {
	cmdSlice, err := jsdependency.InstallGlobalDependencyCommand(packageManager, name, version)
	if err != nil {
		return fmt.Errorf("Failed to update %s version, error: %s", name, err)
	}
	for i, cmd := range cmdSlice {
		fmt.Println()
		log.Donef("$ %s", cmd.PrintableCommandArgs())
		fmt.Println()

		// Yarn returns an error if the package is not added before removal, ignoring
		if out, err := cmd.RunAndReturnTrimmedCombinedOutput(); err != nil && !(packageManager == jsdependency.Yarn && i == 0) {
			if errorutil.IsExitStatusError(err) {
				return fmt.Errorf("Failed to update %s version: %s failed, output: %s", name, cmd.PrintableCommandArgs(), out)
			}
			return fmt.Errorf("Failed to update %s version: %s failed, error: %s", name, cmd.PrintableCommandArgs(), err)
		}
	}
	return nil
}

func main() {
	var cfg config
	if err := stepconf.Parse(&cfg); err != nil {
		failf("Error: %s\n", err)
	}
	stepconf.Print(cfg)

	// Change dir to working directory
	workDir, err := pathutil.AbsPath(cfg.WorkDir)
	log.Debugf("New work dir: %s", workDir)
	if err != nil {
		failf("Failed to expand WorkDir (%s), error: %s", cfg.WorkDir, err)
	}

	currentDir, err := pathutil.CurrentWorkingDirectoryAbsolutePath()
	if err != nil {
		failf("Failed to get current directory, error: %s", err)
	}

	if workDir != currentDir {
		fmt.Println()
		log.Infof("Switch working directory to: %s", workDir)

		revokeFunc, err := pathutil.RevokableChangeDir(workDir)
		if err != nil {
			failf("Failed to change working directory, error: %s", err)
		}
		defer func() {
			fmt.Println()
			log.Infof("Reset working directory")
			if err := revokeFunc(); err != nil {
				failf("Failed to reset working directory, error: %s", err)
			}
		}()
	}

	// Update cordova version
	if cfg.Version != "" {
		log.Printf("\n")
		log.Infof("Updating cordova version to: %s", cfg.Version)
		packageName := "cordova"
		packageName += "@" + cfg.Version

		packageManager, err := jsdependency.DetectTool(workDir)
		if err != nil {
			log.Warnf("%s", err)
		}
		log.Printf("Js package manager used: %s", packageManager)

		if err := installDependency(packageManager, "cordova", cfg.Version); err != nil {
			failf("Updating cordova failed, error: %s", err)
		}
	}
	// Print cordova and ionic version
	cordovaVersion, err := cordova.CurrentVersion()
	if err != nil {
		failf(err.Error())
	}

	fmt.Println()
	log.Printf("Using cordova version:\n%s", colorstring.Green(cordovaVersion))

	// Fulfill cordova builder
	builder := cordova.New()

	platforms := []string{}
	if cfg.Platform != "" {
		platformsSplit := strings.Split(cfg.Platform, ",")
		for _, platform := range platformsSplit {
			platforms = append(platforms, strings.TrimSpace(platform))
		}

		builder.SetPlatforms(platforms...)
	}

	// cordova prepare
	fmt.Println()
	log.Infof("Preparing project")
	platformPrepareCmd := builder.PrepareCommand()
	platformPrepareCmd.SetStdout(os.Stdout).SetStderr(os.Stderr)
	log.Donef("$ %s", platformPrepareCmd.PrintableCommandArgs())

	if err := platformPrepareCmd.Run(); err != nil {
		failf("cordova prepare failed, error: %s", err)
	}
}

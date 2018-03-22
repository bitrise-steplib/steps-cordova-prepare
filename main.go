package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/bitrise-io/go-utils/log"
	"github.com/bitrise-tools/go-steputils/stepconf"
)

type config struct {
	Platform string `env:"platform,opt[ios,android,'ios,android']"`
	Readd    bool   `env:"readd_platform,opt[true,false]"`
	Version  string `env:"cordova_version"`
	WorkDir  string `env:"workdir,dir"`
}

func runCommand(name string, arg ...string) error {
	log.Infof("$ %s %s", name, strings.Join(arg, " "))
	cmd := exec.Command(name, arg...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func failf(format string, v ...interface{}) {
	log.Errorf(format, v...)
	os.Exit(1)
}

func update(version string) error {
	args := []string{"install", "-g"}
	if version == "latest" {
		args = append(args, "cordova")
	} else {
		args = append(args, "cordova@"+version)
	}
	return runCommand("npm", args...)
}

func main() {
	var cfg config
	if err := stepconf.Parse(&cfg); err != nil {
		failf("Error: %s\n", err)
	}
	stepconf.Print(cfg)

	if cfg.Version != "" {
		if err := update(cfg.Version); err != nil {
			failf("Failed to update cordova: %v", err)
		}
	}

	out, err := exec.Command("cordova", "-v").Output()
	if err != nil {
		failf("Failed to get cordova version: %v", err)
	}
	fmt.Printf("using cordova version %s\n", out)

	if err := os.Chdir(cfg.WorkDir); err != nil {
		failf("Failed to change to directory (%s), error: %s", cfg.WorkDir, err)
	}

	platforms := strings.Split(cfg.Platform, ",")
	if cfg.Readd {
		if err := runCommand("cordova", append([]string{"platform", "rm"}, platforms...)...); err != nil {
			failf("Failed to remove platform, error: %s", err)
		}
	}
	if err := runCommand("cordova", append([]string{"platform", "add"}, platforms...)...); err != nil {
		failf("Failed to add platform, error: %s", err)
	}
}

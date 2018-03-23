package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/bitrise-io/go-utils/command"
	"github.com/bitrise-io/go-utils/log"
	"github.com/bitrise-tools/go-steputils/stepconf"
)

type config struct {
	Platform string `env:"platform,opt[ios,android,'ios,android']"`
	Readd    bool   `env:"readd_platform,opt[true,false]"`
	Version  string `env:"cordova_version"`
	WorkDir  string `env:"workdir,dir"`
}

func failf(format string, v ...interface{}) {
	log.Errorf(format, v...)
	os.Exit(1)
}

func main() {
	var cfg config
	if err := stepconf.Parse(&cfg); err != nil {
		failf("Error: %s\n", err)
	}
	stepconf.Print(cfg)

	if cfg.Version != "" {
		args := []string{"install", "-g", "cordova",
			map[bool]string{true: "@" + cfg.Version}[cfg.Version != "latest"]}
		cmd := command.NewWithStandardOuts("npm", args...)
		log.Infof(cmd.PrintableCommandArgs())
		if err := cmd.Run(); err != nil {
			failf("Failed to update cordova: %v", err)
		}
	}

	out, err := exec.Command("cordova", "-v").Output()
	if err != nil {
		failf("Failed to get cordova version: %v", err)
	}
	fmt.Printf("\nusing cordova version %s\n", out)

	platforms := strings.Split(cfg.Platform, ",")
	if cfg.Readd {
		args := append([]string{"platform", "rm"}, platforms...)
		cmd := command.NewWithStandardOuts("cordova", args...).SetDir(cfg.WorkDir)
		log.Infof(cmd.PrintableCommandArgs())
		if err := cmd.Run(); err != nil {
			failf("Failed to remove platform, error: %s", err)
		}
	}
	args := append([]string{"platform", "add"}, platforms...)
	cmd := command.NewWithStandardOuts("cordova", args...).SetDir(cfg.WorkDir)
	log.Infof(cmd.PrintableCommandArgs())
	if err := cmd.Run(); err != nil {
		failf("Failed to add platform, error: %s", err)
	}
}

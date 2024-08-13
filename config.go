package main

import (
	"errors"
	"io/fs"
	"log"
	"os"
	"path"

	"golang.org/x/mod/module"
	"gopkg.in/yaml.v3"
)

type Config struct {
	AllowedLicenses  []string
	GoModFiles       []string
	LicenseOverrides []LicenseOverride
}

type LicenseOverride struct {
	ModuleVersion module.Version
	SPDX          string
}

func ReadConfig(filename string) Config {
	var config Config
	// if no config file is provided, look for one
	if filename == "" {
		executableName := path.Base(os.Args[0])
		var possibleFileNames = []string{
			"." + executableName + ".yaml",
			executableName + ".yaml",
			"." + executableName + ".yml",
			executableName + ".yml",
			"." + executableName + ".json",
			executableName + ".json",
		}
		for _, t := range possibleFileNames {
			if _, err := os.Stat(t); !errors.Is(err, fs.ErrNotExist) {
				filename = t
				break
			}
		}
	}
	// parse config file, if one was provided or detected
	if filename != "" {
		configFileBytes, err := os.ReadFile(filename)
		if err != nil {
			log.Fatalf("cannot read configuration file: %v", err)
		}
		err = yaml.Unmarshal(configFileBytes, &config)
		if err != nil {
			log.Fatalf("cannot parse configuration file: %v", err)
		}
	}
	return config
}

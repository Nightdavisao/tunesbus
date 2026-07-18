package main

import (
	"errors"
	"os"
	"path"
	"path/filepath"

	"github.com/charmbracelet/log"
	"github.com/pelletier/go-toml/v2"
)

type MprisSection struct {
	BusNameSuffix string
	Identity      string
	DesktopEntry  string
}

type ProgramConfig struct {
	MPRIS MprisSection
}

const CONFIG_FILE = "config.toml"

var programConfig = &ProgramConfig{
	MPRIS: MprisSection{
		Identity: "iTunes",
		BusNameSuffix: "iTunes",
		DesktopEntry: "iTunes",
	},
}

func ParseConfigFile() (error) {
	configFilePath := CONFIG_FILE
	executablePath, err := os.Executable()
	if err == nil {
		configFilePath = path.Join(filepath.Dir(executablePath), CONFIG_FILE)
	}
	
	_, err = os.Stat(configFilePath)
	if err != nil && errors.Is(err, os.ErrNotExist) {
		log.Info("creating a default config file", "reason", err)
		defaults, err := toml.Marshal(programConfig)
		if err != nil {
			return err
		}
		err = os.WriteFile(configFilePath, defaults, 0664)
		if err != nil {
			return err
		}
		return nil
	} else if err != nil {
		return err
	}
	bytes, err := os.ReadFile(configFilePath)
	if err != nil {
		return err
	}
	err = toml.Unmarshal(bytes, programConfig)
	if err != nil {
		return err
	}
	return nil
}
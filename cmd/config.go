package main

import (
	"errors"
	"os"
	"github.com/pelletier/go-toml/v2"
	"github.com/charmbracelet/log"
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
	_, err := os.Stat(CONFIG_FILE)
	if err != nil && errors.Is(err, os.ErrNotExist) {
		log.Info("creating a default config file", "reason", err)
		defaults, err := toml.Marshal(programConfig)
		if err != nil {
			return err
		}
		err = os.WriteFile(CONFIG_FILE, defaults, 0664)
		if err != nil {
			return err
		}
		return nil
	} else if err != nil {
		return err
	}
	bytes, err := os.ReadFile(CONFIG_FILE)
	if err != nil {
		return err
	}
	err = toml.Unmarshal(bytes, programConfig)
	if err != nil {
		return err
	}
	return nil
}
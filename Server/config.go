package main

import (
	"code.google.com/p/gcfg"
	"log"
)

var build_version string

type Config struct {
	Network struct {
		BindAddress string
		BindPort    string
	}
}

func LoadConfiguration(config_path string) *Config {
	kc := new(Config)
	err := gcfg.ReadFileInto(kc, config_path)
	if err != nil {
		log.Fatal("Failed to parse gcfg data: ", err)
	}
	return kc
}

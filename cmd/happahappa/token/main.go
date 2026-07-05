package main

import (
	"flag"
	"fmt"
	"log"
	"log/slog"
	"os"

	"github.com/jhachmer/happahappa/pkg/client/matrix"
	"github.com/jhachmer/happahappa/pkg/config"
)

func main() {
	var configPath string
	flag.StringVar(&configPath, "config", "settings.yaml", "path to config file")
	flag.Parse()

	log.SetPrefix("[HappaHappa] ")

	cfg, err := config.ReadConfig(configPath)
	if err != nil {
		slog.Error("could not read cfg", "msg", err.Error())
		os.Exit(1)
	}
	if cfg.Matrix == nil {
		slog.Error("no matrix config was given")
		os.Exit(1)
	}
	token, err := TokenGen(cfg)
	if err != nil {
		slog.Error("could not create matrix credentials", "msg", err.Error())
		os.Exit(1)
	}
	fmt.Printf("Generated Token: %s\n", token)
	os.Exit(0)
}

func TokenGen(config *config.Config) (string, error) {
	loginResponse, err := matrix.GenerateAccessToken(config)
	if err != nil {
		return "", err
	}
	return loginResponse.AccessToken, nil
}

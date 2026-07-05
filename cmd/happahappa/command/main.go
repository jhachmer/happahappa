package main

import (
	"flag"
	"log"
	"log/slog"
	"os"

	"github.com/jhachmer/happahappa/pkg/client/matrix"
	"github.com/jhachmer/happahappa/pkg/config"
	"github.com/jhachmer/happahappa/pkg/data/station"
)

func main() {
	var configPath string
	flag.StringVar(&configPath, "config", "settings.yaml", "path to config file")
	flag.Parse()

	log.SetPrefix("[StationDepartures] ")

	cfg, err := config.ReadConfig(configPath)
	if err != nil {
		slog.Error("could not read cfg", "msg", err.Error())
		os.Exit(1)
	}

	if cfg.Matrix != nil {
		credentials, err := matrix.NewCredentials(cfg)
		if err != nil {
			slog.Error("could not create matrix credentials", "msg", err.Error())
			os.Exit(1)
		}

		matrixClient := matrix.NewMatrixClient(cfg.Matrix.BaseURL, credentials)
		commandHandler, err := matrix.NewCommandHandler(cfg, matrixClient)
		if err != nil {
			slog.Error("could not construct command handler")
			os.Exit(1)
		}
		departureHandler, err := station.NewDepartureCommand(cfg, &matrixClient)
		if err != nil {
			slog.Error("could not add departure handler to commands", "err", err)
		}

		commandHandler.Register(departureHandler)
		commandHandler.HandleCommands()
	} else {
		slog.Error("no matrix config")
		os.Exit(1)
	}
	os.Exit(0)
}

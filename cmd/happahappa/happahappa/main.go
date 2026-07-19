package main

import (
	"flag"
	"log"
	"log/slog"
	"os"

	"github.com/jhachmer/happahappa/pkg/client/matrix"
	"github.com/jhachmer/happahappa/pkg/config"
	"github.com/jhachmer/happahappa/pkg/data/canteen"
	"github.com/jhachmer/happahappa/pkg/data/weather"
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

	RoomID := cfg.Canteen.RoomID

	canteenScraper, err := canteen.NewCanteenScraper(cfg)
	if err != nil {
		slog.Error("unable to scrape canteen data", "msg", err.Error())
		os.Exit(1)
	}

	todaysMenu := canteenScraper.Scrape()

	curWeather, err := weather.GetCurrentWeather(cfg.Weather.Lat, cfg.Weather.Lon)
	if err != nil {
		log.Fatal(err)
	}

	if cfg.Matrix != nil {
		credentials, err := matrix.NewCredentials(cfg)
		if err != nil {
			slog.Error("could not create matrix credentials", "msg", err.Error())
			os.Exit(1)
		}

		matrixClient := matrix.NewMatrixClient(cfg.Matrix.BaseURL, credentials)

		canteenMessage := matrix.NewMatrixMessageFromSender(todaysMenu, RoomID)
		weatherMessage := matrix.NewMatrixMessageFromSender(curWeather, RoomID)

		matrixClient.Register(canteenMessage)
		matrixClient.Register(weatherMessage)

		matrixClient.SendRegistered()
	}
}

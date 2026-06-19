package station

import (
	"net/url"
	"strconv"
	"time"

	"github.com/jhachmer/happahappa/pkg/config"
)

type Info struct {
	Type          string
	Priority      string
	IncidentStart time.Time
	IncidentEnd   time.Time
	Title         string
	Content       string
}

type Event struct {
	PlannedTime   time.Time
	EstimatedTime time.Time
}

type Departure struct {
	Line        string
	LineNumber  string
	Destination string
	Events      []Event
	Infos       []Info
}

type DepartureBoard struct {
	StationName string
	Departures  []Departure
}

type StationScraper struct {
	url *url.URL
}

func NewStationScraper(cfg *config.Config) (*StationScraper, error) {
	requestURL, err := buildDepartureURL(cfg)
	if err != nil {
		return nil, err
	}
	return &StationScraper{
		url: requestURL,
	}, nil
}

func buildDepartureURL(cfg *config.Config) (*url.URL, error) {
	u, err := url.Parse(cfg.Departure.URL)
	if err != nil {
		return nil, err
	}

	params := url.Values{}
	params.Add("outputFormat", "rapidJSON")
	params.Add("name_dm", strconv.Itoa(cfg.Departure.StationID))
	params.Add("type_dm", "any")
	params.Add("mode", "direct")
	params.Add("useRealtime", "1")

	u.RawQuery = params.Encode()
	return u, nil
}

func (s *StationScraper) BuildDepartureBoard() *DepartureBoard {
	return &DepartureBoard{}
}

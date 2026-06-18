package station

import (
	"net/url"
	"time"
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

func (s *StationScraper) Scrape() *DepartureBoard {
	return &DepartureBoard{}
}

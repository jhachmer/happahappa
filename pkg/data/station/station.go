package station

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/jhachmer/happahappa/pkg/client/matrix"
	"github.com/jhachmer/happahappa/pkg/config"
)

func Loc() *time.Location {
	tz, _ := time.LoadLocation("Europe/Berlin")
	return tz
}

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

func (e Event) String() string {
	var sb strings.Builder
	fmt.Fprintf(&sb, "%s %s | ", e.PlannedTime.In(Loc()).Format("15:04"), e.TimeDifference())

	return sb.String()
}

func (e Event) Body() string {
	var sb strings.Builder
	fmt.Fprintf(&sb, "%s %s | ", e.PlannedTime.In(Loc()).Format("15:04"), e.TimeDifference())

	return sb.String()
}

func (e Event) HTML() string {
	var sb strings.Builder
	fmt.Fprintf(&sb, "%s %s | ", e.PlannedTime.In(Loc()).Format("15:04"), e.TimeDifference())

	return sb.String()
}

func (e Event) TimeDifference() string {
	diff := e.EstimatedTime.Sub(e.PlannedTime).Minutes()
	positive := diff >= 0.0
	sign := ""
	if positive {
		sign = "+"
	}
	return fmt.Sprintf("%s%.0f", sign, diff)
}

type Departure struct {
	Line        string
	LineNumber  string
	Destination string
	Events      []Event
	Infos       []Info
}

func (d Departure) Body() string {
	var sb strings.Builder
	fmt.Fprintf(&sb, "(%s) %s\n", d.LineNumber, d.Destination)
	for _, event := range d.Events {
		fmt.Fprintf(&sb, "%s", event)
	}
	return sb.String()
}

func (d Departure) HTML() string {
	var sb strings.Builder
	fmt.Fprintf(&sb, "<b>(%s) %s</b><br>", d.LineNumber, d.Destination)
	for _, event := range d.Events {
		fmt.Fprintf(&sb, "%s", event)
	}
	return sb.String()
}

func NewDeparture(departure DeparturesResponse) Departure {
	lineEvents := make([]Event, 0)
	for _, event := range departure.Events {
		lineEvents = append(lineEvents, Event{
			PlannedTime:   event.PlannedTime,
			EstimatedTime: event.EstimatedTime,
		})
	}
	lineInfos := make([]Info, 0)
	for _, info := range departure.Infos {
		lineInfos = append(lineInfos, Info{
			Type:          info.Type,
			Priority:      info.Priority,
			IncidentStart: info.IncidentStart,
			IncidentEnd:   info.IncidentEnd,
			Title:         info.Title,
			Content:       info.Content,
		})
	}

	return Departure{
		Line:        departure.Line,
		LineNumber:  departure.Number,
		Destination: departure.Destination,
		Events:      lineEvents,
		Infos:       lineInfos,
	}
}

type DepartureBoard struct {
	StationName string
	Departures  []Departure
}

func (db DepartureBoard) Body() string {
	var sb strings.Builder
	fmt.Fprintf(&sb, "%s\n", db.StationName)
	for _, departure := range db.Departures {
		fmt.Fprintf(&sb, "%s\n", departure.Body())
	}
	return sb.String()
}

func (db DepartureBoard) HTML() string {
	var sb strings.Builder
	fmt.Fprintf(&sb, "<h1>%s</h1>", db.StationName)
	for _, departure := range db.Departures {
		fmt.Fprintf(&sb, "%s<br>", departure.HTML())
	}
	return sb.String()
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

func (s *StationScraper) getResponse() (*DepatureResponse, error) {
	req, err := http.NewRequest("GET", s.url.String(), nil)
	if err != nil {
		return nil, err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	decoded := DepatureResponse{}
	err = json.NewDecoder(resp.Body).Decode(&decoded)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return &decoded, nil
}

func (s *StationScraper) BuildDepartureBoard() *DepartureBoard {
	apiResponse, err := s.getResponse()
	if err != nil {
		return &DepartureBoard{}
	}
	departures := make([]Departure, 0)
	for _, departureResponse := range apiResponse.Departures {
		departures = append(departures, NewDeparture(departureResponse))
	}

	return &DepartureBoard{
		StationName: apiResponse.Stop,
		Departures:  departures,
	}
}

type DepatureResponse struct {
	Stop        string               `json:"stop"`
	Coordinates []int                `json:"coordinates"`
	Departures  []DeparturesResponse `json:"departures"`
}

type DeparturesResponse struct {
	Line        string `json:"line"`
	LineID      string `json:"lineId"`
	Bon         string `json:"bon"`
	Destination string `json:"destination"`
	Number      string `json:"number"`
	Events      []struct {
		PlannedTime   time.Time `json:"plannedTime"`
		EstimatedTime time.Time `json:"estimated_time"`
	} `json:"events"`
	Info  struct{} `json:"info"`
	Infos []struct {
		ID            string    `json:"id"`
		Priority      string    `json:"priority"`
		Type          string    `json:"type"`
		IncidentStart time.Time `json:"incidentStart"`
		IncidentEnd   time.Time `json:"incidentEnd"`
		Title         string    `json:"titel"`
		Content       string    `json:"content"`
	} `json:"infos"`
	Hints []struct {
		Content      string `json:"content"`
		ProviderCode string `json:"providerCode"`
		Type         string `json:"type"`
		Properties   struct {
			Subnet string `json:"subnet"`
		} `json:"properties"`
	} `json:"hints"`
}

type DepartureCommand struct {
	scraper *StationScraper
	client  *matrix.Client
}

func NewDepartureCommand(cfg *config.Config, client *matrix.Client) (*DepartureCommand, error) {
	stationScraper, err := NewStationScraper(cfg)
	if err != nil {
		return nil, err
	}
	return &DepartureCommand{
		scraper: stationScraper,
		client:  client,
	}, nil
}

func (dc *DepartureCommand) Name() string {
	return "abfahrt"
}

func (dc *DepartureCommand) Execute(roomID string, args []string) error {
	db := dc.scraper.BuildDepartureBoard()
	message := matrix.NewMatrixMessageFromSender(db, roomID)
	err := dc.client.SendMessage(message)
	if err != nil {
		return err
	}
	return nil
}

package station

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
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
	fmt.Fprintf(&sb, "%s %s", e.PlannedTime.Format("15:04"), e.TimeDifference())

	return sb.String()
}

func (e Event) Body() string {
	var sb strings.Builder
	fmt.Fprintf(&sb, "%s %s", e.PlannedTime.Format("15:04"), e.TimeDifference())

	return sb.String()
}

func (e Event) HTML() string {
	var sb strings.Builder
	fmt.Fprintf(&sb, "%s %s", e.PlannedTime.Format("15:04"), e.TimeDifference())

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
		fmt.Fprintf(&sb, "%s | ", event)
	}
	return sb.String()
}

func (d Departure) HTML() string {
	var sb strings.Builder
	fmt.Fprintf(&sb, "<b>(%s) %s</b><br>", d.LineNumber, d.Destination)
	for _, event := range d.Events {
		fmt.Fprintf(&sb, "%s | ", event)
	}
	if len(d.Infos) > 0 {
		for _, info := range d.Infos {
			fmt.Fprintf(&sb, "<br> &#x26A0 %s (%s - %s)", info.Title, info.IncidentStart.Format("02 Jan 06 15:04"), info.IncidentEnd.Format("02 Jan 06 15:04"))
		}
	}
	return sb.String()
}

func NewDeparture(departure DeparturesResponse) Departure {
	lineEvents := make([]Event, 0)
	for _, event := range departure.Events {
		lineEvents = append(lineEvents, Event{
			PlannedTime:   event.PlannedTime.In(Loc()),
			EstimatedTime: event.EstimatedTime.In(Loc()),
		})
	}
	lineInfos := make([]Info, 0)
	for _, info := range departure.Infos {
		lineInfos = append(lineInfos, Info{
			Type:          info.Type,
			Priority:      info.Priority,
			IncidentStart: info.IncidentStart.In(Loc()),
			IncidentEnd:   info.IncidentEnd.In(Loc()),
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
	fmt.Fprintf(&sb, "%s (%s)\n", db.StationName, time.Now().Format("15:04"))
	for _, departure := range db.Departures {
		fmt.Fprintf(&sb, "%s\n", departure.Body())
	}
	return sb.String()
}

func (db DepartureBoard) HTML() string {
	var sb strings.Builder
	fmt.Fprintf(&sb, "<h1>%s (%s)</h1>", db.StationName, time.Now().Format("02 Jan 15:04"))
	for _, departure := range db.Departures {
		fmt.Fprintf(&sb, "%s<br>", departure.HTML())
	}
	return sb.String()
}

type Scraper struct {
	BaseURL    string
	stationMap map[string]string
}

func NewStationScraper(cfg *config.Config) (*Scraper, error) {
	return &Scraper{
		BaseURL:    cfg.Departure.URL,
		stationMap: cfg.Departure.Stations,
	}, nil
}

func (s *Scraper) buildDepartureURL(stationID string) (*url.URL, error) {
	u, err := url.Parse(s.BaseURL)
	if err != nil {
		return nil, err
	}

	params := url.Values{}
	params.Add("outputFormat", "rapidJSON")
	params.Add("name_dm", stationID)
	params.Add("type_dm", "any")
	params.Add("mode", "direct")
	params.Add("useRealtime", "1")

	u.RawQuery = params.Encode()
	return u, nil
}

func (s *Scraper) getResponse(stationID string) (*DepartureResponse, error) {
	requestURL, err := s.buildDepartureURL(stationID)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest("GET", requestURL.String(), nil)
	if err != nil {
		return nil, err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	decoded := DepartureResponse{}
	err = json.NewDecoder(resp.Body).Decode(&decoded)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return &decoded, nil
}

func (s *Scraper) BuildDepartureBoard(stationID string) *DepartureBoard {
	apiResponse, err := s.getResponse(stationID)
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

type DepartureResponse struct {
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
	scraper *Scraper
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

type DepartureHelp struct {
	helpText string
}

func (cmd *DepartureHelp) Body() string {
	return cmd.helpText
}

func (cmd *DepartureHelp) HTML() string {
	return cmd.helpText
}

func (dc *DepartureCommand) Help() string {
	var sb strings.Builder
	for abbr, station := range dc.scraper.stationMap {
		fmt.Fprintf(&sb, "%s: %s\n", abbr, station)
	}
	return sb.String()
}

func (dc *DepartureCommand) Name() string {
	return "abfahrt"
}

func (dc *DepartureCommand) Execute(roomID string, args []string) error {
	if len(args) < 1 || len(args) >= 2 {
		return fmt.Errorf("invalid argument count: %d", len(args))
	}
	stationName := args[0]
	if stationName == "help" {
		return dc.send(roomID, &DepartureHelp{helpText: dc.Help()})
	}
	id, ok := dc.scraper.stationMap[stationName]
	if !ok {
		return fmt.Errorf("unknown station: %s", stationName)
	}

	return dc.send(roomID, dc.scraper.BuildDepartureBoard(id))
}

func (dc *DepartureCommand) send(roomID string, sender matrix.MatrixHTMLSender) error {
	message := matrix.NewMatrixMessageFromSender(sender, roomID)
	if err := dc.client.SendMessage(message); err != nil {
		return err
	}

	slog.Info("Message sent", "message", message)
	return nil
}

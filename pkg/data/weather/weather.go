package weather

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"lab.it.hs-hannover.de/8mg-y3w-u2/happahappa/pkg/util"
)

const BaseURL = "https://api.met.no/weatherapi/locationforecast/2.0/complete"

type WeatherResponse struct {
	Type       string     `json:"type"`
	Geometry   Geometry   `json:"geometry"`
	Properties Properties `json:"properties"`
}

type Geometry struct {
	Type        string    `json:"type"`
	Coordinates []float64 `json:"coordinates"`
}

type Properties struct {
	Meta       Meta         `json:"meta"`
	Timeseries []Timeseries `json:"timeseries"`
}

type Timeseries struct {
	Time time.Time `json:"time"`
	Data Data      `json:"data"`
}

type Data struct {
	Instant    Instant    `json:"instant"`
	Next1Hours Next1Hours `json:"next_1_hours,omitempty"`
}

type Instant struct {
	Details Details `json:"details"`
}

type Next1Hours struct {
	Summary struct {
		SymbolCode string `json:"symbol_code"`
	} `json:"summary"`
	Details Details `json:"details"`
}

type Details struct {
	AirPressureAtSeaLevel    float64 `json:"air_pressure_at_sea_level"`
	AirTemperature           float64 `json:"air_temperature"`
	CloudAreaFraction        float64 `json:"cloud_area_fraction"`
	RelativeHumidity         float64 `json:"relative_humidity"`
	WindFromDirection        float64 `json:"wind_from_direction"`
	WindSpeed                float64 `json:"wind_speed"`
	PrecipitationAmount      float64 `json:"precipitation_amount,omitempty"`
	CloudAreaFractionHigh    float64 `json:"cloud_area_fraction_high"`
	CloudAreaFractionLow     float64 `json:"cloud_area_fraction_low"`
	CloudAreaFractionMedium  float64 `json:"cloud_area_fraction_medium"`
	DewPointTemperature      float64 `json:"dew_point_temperature"`
	FogAreaFraction          float64 `json:"fog_area_fraction,omitempty"`
	UltravioletIndexClearSky float64 `json:"ultraviolet_index_clear_sky,omitempty"`
}

type Meta struct {
	UpdatedAt time.Time `json:"updated_at"`
	Units     Units     `json:"units"`
}

type Units struct {
	AirPressureAtSeaLevel    string `json:"air_pressure_at_sea_level"`
	AirTemperature           string `json:"air_temperature"`
	CloudAreaFraction        string `json:"cloud_area_fraction"`
	PrecipitationAmount      string `json:"precipitation_amount"`
	RelativeHumidity         string `json:"relative_humidity"`
	WindFromDirection        string `json:"wind_from_direction"`
	WindSpeed                string `json:"wind_speed"`
	AirTemperatureMax        string `json:"air_temperature_max"`
	AirTemperatureMin        string `json:"air_temperature_min"`
	CloudAreaFractionHigh    string `json:"cloud_area_fraction_high"`
	CloudAreaFractionLow     string `json:"cloud_area_fraction_low"`
	CloudAreaFractionMedium  string `json:"cloud_area_fraction_medium"`
	DewPointTemperature      string `json:"dew_point_temperature"`
	FogAreaFraction          string `json:"fog_area_fraction"`
	UltravioletIndexClearSky string `json:"ultraviolet_index_clear_sky"`
}

func GetCurrentWeather(lat, lon string) (*WeatherResponse, error) {
	slog.Info("Retrieving current weather information", "lat", lat, "lon", lon)
	var err error
	_, err = strconv.ParseFloat(lat, 64)
	if err != nil {
		slog.Error("could not parse latitude as float64", "lat", lat)
		return nil, err
	}
	_, err = strconv.ParseFloat(lon, 64)
	if err != nil {
		slog.Error("could not parse longitude as float64", "lon", lon)
		return nil, err
	}
	URL, err := url.Parse(BaseURL)
	if err != nil {
		return nil, err
	}
	values := URL.Query()
	values.Set("lat", lat)
	values.Set("lon", lon)
	URL.RawQuery = values.Encode()
	req, err := http.NewRequest("GET", URL.String(), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("User-Agent", "HappaHappa")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, errors.New(resp.Status)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	var weatherResponse WeatherResponse
	err = json.Unmarshal(body, &weatherResponse)
	if err != nil {
		return nil, err
	}
	return &weatherResponse, nil
}

func (wr *WeatherResponse) Body() string {
	temp := wr.Properties.Timeseries[0].Data.Instant.Details.AirTemperature
	tempUnit := util.UnitToSymbol(wr.Properties.Meta.Units.AirTemperature)
	precipitationAmount := wr.Properties.Timeseries[0].Data.Next1Hours.Details.PrecipitationAmount
	precipitationUnit := wr.Properties.Meta.Units.PrecipitationAmount

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Temp: %.2f %s\n", temp, tempUnit))
	sb.WriteString(fmt.Sprintf("Rain (1h): %.2f %s", precipitationAmount, precipitationUnit))
	return sb.String()
}

func (wr *WeatherResponse) HTML() string {
	temp := wr.Properties.Timeseries[0].Data.Instant.Details.AirTemperature
	tempUnit := util.UnitToSymbol(wr.Properties.Meta.Units.AirTemperature)
	precipitationAmount := wr.Properties.Timeseries[0].Data.Next1Hours.Details.PrecipitationAmount
	precipitationUnit := wr.Properties.Meta.Units.PrecipitationAmount

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("<h1> Weather at %.2f° N, %.2f° O:</h1>", wr.Geometry.Coordinates[1], wr.Geometry.Coordinates[0]))
	sb.WriteString(fmt.Sprintf("<h2> %s %.1f%s", wr.weatherSymbol(), temp, tempUnit))
	if precipitationAmount > 0.0 {
		sb.WriteString(fmt.Sprintf(" &#x2602 %.2f%s</h2>", precipitationAmount, precipitationUnit))
	} else {
		sb.WriteString("</h2>")
	}
	return sb.String()
}

func (wr *WeatherResponse) weatherSymbol() string {
	code := wr.Properties.Timeseries[0].Data.Next1Hours.Summary.SymbolCode
	symbol := util.WeatherSymbolToEmoji(code)
	return symbol
}

package util

import (
	"log/slog"
)

const Sun string = " &#x1F31E "
const Cloud string = " &#x2601 "
const SunWithCloud string = " &#x26C5 "
const Fog string = " &#x1F32B "
const Rain string = " &#x1F327 "
const Shower string = " &#x1F326 "
const Thunder string = " &#x26C8 "
const Snow string = " &#x1F328 "
const QuestionMark string = " &#x2753 "

func UnitToSymbol(unit string) string {
	UnitMap := map[string]string{
		"celsius": "°C",
		"degrees": "°",
	}
	Unit, ok := UnitMap[unit]
	if !ok {
		return ""
	}
	return Unit
}

func WeatherSymbolToEmoji(symbolCode string) string {
	symbolMap := map[string]string{
		"clearsky_day":                     Sun,
		"cloudy_day":                       Cloud,
		"fair_day":                         SunWithCloud,
		"fog_day":                          Fog,
		"heavyrain_day":                    Rain,
		"heavyrainandthunder_day":          Thunder,
		"heavyrainshowers_day":             Shower,
		"heavyrainshowersandthunder_day":   Thunder,
		"heavysleet_day":                   Rain,
		"heavysleetandthunder_day":         Thunder,
		"heavysleetshowers_day":            Shower,
		"heavysleetshowersandthunder_day":  Thunder,
		"heavysnow_day":                    Snow,
		"heavysnowandthunder_day":          Thunder,
		"heavysnowshowers_day":             Snow,
		"heavysnowshowersandthunder_day":   Snow,
		"lightrain_day":                    Rain,
		"lightrainandthunder_day":          Thunder,
		"lightrainshowers_day":             Shower,
		"lightrainshowersandthunder_day":   Thunder,
		"lightsleet_day":                   Rain,
		"lightsleetandthunder_day":         Thunder,
		"lightsleetshowers_day":            Shower,
		"lightsnow_day":                    Snow,
		"lightsnowandthunder_day":          Thunder,
		"lightsnowshowers_day":             Snow,
		"lightssleetshowersandthunder_day": Thunder,
		"lightssnowshowersandthunder_day":  Thunder,
		"partlycloudy_day":                 SunWithCloud,
		"rain_day":                         Rain,
		"rainandthunder_day":               Thunder,
		"rainshowers_day":                  Shower,
		"rainshowersandthunder_day":        Thunder,
		"sleet_day":                        Rain,
		"sleetandthunder_day":              Thunder,
		"sleetshowers_day":                 Shower,
		"sleetshowersandthunder_day":       Thunder,
		"snow_day":                         Snow,
		"snowandthunder_day":               Thunder,
		"snowshowers_day":                  Snow,
		"snowshowersandthunder_day":        Thunder,
		"clearsky":                         Sun,
		"cloudy":                           Cloud,
		"cl":                               Cloud,
		"fair":                             SunWithCloud,
		"fog":                              Fog,
		"heavyrain":                        Rain,
		"heavyrainandthunder":              Thunder,
		"heavyrainshowers":                 Shower,
		"heavyrainshowersandthunder":       Thunder,
		"heavysleet":                       Rain,
		"heavysleetandthunder":             Thunder,
		"heavysleetshowers":                Shower,
		"heavysleetshowersandthunder":      Thunder,
		"heavysnow":                        Snow,
		"heavysnowandthunder":              Thunder,
		"heavysnowshowers":                 Snow,
		"heavysnowshowersandthunder":       Snow,
		"lightrain":                        Rain,
		"lightrainandthunder":              Thunder,
		"lightrainshowers":                 Shower,
		"lightrainshowersandthunder":       Thunder,
		"lightsleet":                       Rain,
		"lightsleetandthunder":             Thunder,
		"lightsleetshowers":                Shower,
		"lightsnow":                        Snow,
		"lightsnowandthunder":              Thunder,
		"lightsnowshowers":                 Snow,
		"lightssleetshowersandthunder":     Thunder,
		"lightssnowshowersandthunder":      Thunder,
		"partlycloudy":                     SunWithCloud,
		"rain":                             Rain,
		"rainandthunder":                   Thunder,
		"rainshowers":                      Shower,
		"rainshowersandthunder":            Thunder,
		"sleet":                            Rain,
		"sleetandthunder":                  Thunder,
		"sleetshowers":                     Shower,
		"sleetshowersandthunder":           Thunder,
		"snow":                             Snow,
		"snowandthunder":                   Thunder,
		"snowshowers":                      Snow,
		"snowshowersandthunder":            Thunder,
	}
	symbol, ok := symbolMap[symbolCode]
	if !ok {
		slog.Warn("could not find symbol code", "symbolCode", symbolCode)
		return QuestionMark
	}
	return symbol
}

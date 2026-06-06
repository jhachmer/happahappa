package util

import "testing"

func TestUnitToSymbol(t *testing.T) {
	type args struct {
		unit string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "Celsius",
			args: args{unit: "celsius"},
			want: "°C",
		},
		{
			name: "Degrees",
			args: args{
				unit: "degrees",
			},
			want: "°",
		},
		{
			name: "Not in Map",
			args: args{
				unit: "notinmap",
			},
			want: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := UnitToSymbol(tt.args.unit); got != tt.want {
				t.Errorf("UnitToSymbol() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestWeatherSymbolToEmoji(t *testing.T) {
	type args struct {
		symbolCode string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "Sunny",
			args: args{
				symbolCode: "clearsky_day"},
			want: Sun,
		},
		{
			name: "Cloudy",
			args: args{
				symbolCode: "cloudy_day",
			},
			want: Cloud,
		},
		{
			name: "Rain",
			args: args{
				symbolCode: "heavyrain_day",
			},
			want: Rain,
		},
		{
			name: "Snow",
			args: args{
				symbolCode: "heavysnow_day",
			},
			want: Snow,
		},
		{
			name: "Fair",
			args: args{
				symbolCode: "fair_day",
			},
			want: SunWithCloud,
		},
		{
			name: "Fog",
			args: args{
				symbolCode: "fog_day",
			},
			want: Fog,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := WeatherSymbolToEmoji(tt.args.symbolCode); got != tt.want {
				t.Errorf("WeatherSymbolToEmoji() = %v, want %v", got, tt.want)
			}
		})
	}
}

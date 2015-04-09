package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

const KELVIN = 273.15

type Weather_s struct {
	Id          int
	Main        string
	Description string
}

type Main_s struct {
	Temp     float64
	Humidity float64
	Temp_min float64
	Temp_max float64
}

type WeatherResponse struct {
	Weather []Weather_s
	Main    Main_s
	Rain    map[string]float64
	Clouds  map[string]float64
}

func getWeatherLine() (string, error) {
	res, err := http.Get("http://api.openweathermap.org/data/2.5/weather?q=Boston,us")
	if err != nil {
		return "Weather Not Avail", err
	}
	json_s, err := ioutil.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {
		return "Weather Not Avail", err
	}
	var resp WeatherResponse
	err = json.Unmarshal(json_s, &resp)
	if err != nil {
		return "Weather Not Avail", err
	}
	temp_c := resp.Main.Temp - KELVIN
	temp_low := resp.Main.Temp_min - KELVIN
	temp_high := resp.Main.Temp_max - KELVIN

	if len(resp.Weather) > 0 {

		return fmt.Sprintf(
			`%s %.1fC(Hi %.1fC, Lo %.1fC) Humid %.0f%% Cloud %.0f%% Precip %.1f `,
			resp.Weather[0].Main,
			temp_c, temp_high, temp_low,
			resp.Main.Humidity,
			resp.Clouds["all"],
			resp.Rain["1h"],
		), nil
	} else {
		return fmt.Sprintf(
			`%.1fC(Hi %.1fC, Lo %.1fC) Humid %.0f%% Cloud %.0f%% Precip %.1f `,
			temp_c, temp_high, temp_low,
			resp.Main.Humidity,
			resp.Clouds["all"],
			resp.Rain["1h"],
		), nil
	}
}

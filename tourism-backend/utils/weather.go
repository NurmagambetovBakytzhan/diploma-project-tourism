package utils

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"tourism-backend/internal/entity"
)

type ForecastResponse struct {
	Forecast struct {
		ForecastDay []struct {
			Hour []entity.WeatherInfo `json:"hour"`
		} `json:"forecastday"`
	} `json:"forecast"`
}

func GetWeatherInfo(tourWeatherRQ *entity.WeatherInfoRQ) (*entity.WeatherInfo, error) {
	client := &http.Client{}

	url := fmt.Sprintf("https://api.weatherapi.com/v1/forecast.json?days=1&dt=%s&hour=%s&key=%s&q=%f,%f",
		tourWeatherRQ.Date,
		tourWeatherRQ.Time,
		os.Getenv("WEATHER_API_KEY"),
		tourWeatherRQ.Longitude,
		tourWeatherRQ.Latitude,
	)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("get weather info: %w", err)
	}

	req.Header.Set("Accept", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("get Weather Info error sending request: %w", err)
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("get Weather Info error reading request: %w", err)
	}
	var forecast ForecastResponse
	err = json.Unmarshal(body, &forecast)
	if err != nil {
		return nil, fmt.Errorf("get weather info: Error parsing JSON: %w", err)
	}
	if len(forecast.Forecast.ForecastDay) > 0 && len(forecast.Forecast.ForecastDay[0].Hour) > 0 {
		weatherInfo := forecast.Forecast.ForecastDay[0].Hour[0]
		return &weatherInfo, nil
	} else {
		fmt.Errorf("No weather data available.")
	}

	return nil, nil
}

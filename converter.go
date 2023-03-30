package main

import (
	_ "embed"
	"encoding/xml"
	"errors"
	"strconv"
	"time"
)

//go:embed conversion_rate.xml
var ratesDaily []byte

func newRatesDefault() USD2EURRates {
	type rates struct {
		DataSet struct {
			Series struct {
				Obs []struct {
					Date string  `xml:"TIME_PERIOD,attr"`
					Rate float32 `xml:"OBS_VALUE,attr"`
				} `xml:"Obs"`
			} `xml:"Series"`
		} `xml:"DataSet"`
	}

	var tmp rates
	if err := xml.Unmarshal(ratesDaily, &tmp); err != nil {
		panic(err)
	}

	o := USD2EURRates{}
	for _, rate := range tmp.DataSet.Series.Obs {
		date, err := time.Parse("2006-01-02", rate.Date)
		if err != nil {
			panic(err)
		}
		o[date] = rate.Rate
	}

	return o
}

type Converter interface {
	USD2EUR(date time.Time, v float32) (float32, error)
	EUR2USD(date time.Time, v float32) (float32, error)
}

type USD2EURRates map[time.Time]float32

func (r USD2EURRates) GetRate(date time.Time) (float32, error) {
	for isWeekend(date) {
		date = date.Add(-24 * time.Hour)
	}
	v, ok := r[date]
	if !ok {
		return 0, errors.New("no rate found for " + date.Format("2006-01-02"))
	}
	return v, nil
}

func isWeekend(date time.Time) bool {
	switch date.Format("Mon") {
	case "Sat", "Sun":
		return true
	default:
		return false
	}
}

type converter struct {
	rates USD2EURRates
}

func (c converter) USD2EUR(date time.Time, v float32) (float32, error) {
	r, err := c.rates.GetRate(date)
	if err != nil {
		return 0, err
	}
	return v / r, nil
}

func (c converter) EUR2USD(date time.Time, v float32) (float32, error) {
	r, err := c.rates.GetRate(date)
	if err != nil {
		return 0, err
	}
	return v * r, nil
}

func NewConverterDaily(rates USD2EURRates) (Converter, error) {
	if rates == nil {
		rates = newRatesDefault()
	}

	if err := validateRates(rates); err != nil {
		return nil, err
	}

	return &converter{rates: rates}, nil
}

func validateRates(rates USD2EURRates) error {
	if len(rates) == 0 {
		return errors.New("rates shall not be empty")
	}

	f := func(date time.Time, rate float32) string {
		return "date=" + date.Format("2006-01-02") +
			", rate=" + strconv.FormatFloat(float64(rate), 'f', -1, 32)
	}

	for date, rate := range rates {
		if date.After(time.Now().UTC()) {
			return errors.New("historical rates are expected only, got: " + f(date, rate))
		}
		if rate <= 0 {
			return errors.New("rate shall be positive, got: " + f(date, rate))
		}
	}

	return nil
}

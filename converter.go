package converter

import (
	_ "embed"
	"encoding/xml"
	"errors"
	"strconv"
	"time"
)

// NewConverter initialises the daily converter.
func NewConverter(rates DailyRates) (Converter, error) {
	if rates == nil {
		rates = newRatesDefault()
	}

	if err := validateRates(rates); err != nil {
		return nil, err
	}

	return &converter{rates: rates}, nil
}

// Converter defines the converter to convert EURO to USD and vise-versa.
type Converter interface {
	// A2B converts currency A to currency B.
	A2B(date time.Time, v float64) (float64, error)
	// B2A converts currency B to currency A.
	B2A(date time.Time, v float64) (float64, error)
}

// DailyRates defines the daily conversion rates object.
// The map's value is the rate of the currency B-to-A,
// i.e. it is equal to the ratio of the currency B to the currency A.
//
// Example:
//
//		ratesA2B := DailyRates{
//			time.Date(2000, 1, 1, 1, 0, 0, 0, time.UTC): 2.,
//	 }
//	 // is equivalent of
//
// currencyB := 50
// // corresponds to
// currencyA := 100
type DailyRates map[time.Time]float64

// GetRate returns the rate for a given date.
func (r DailyRates) GetRate(date time.Time) (float64, error) {
	initialDate := date
	for isWeekend(date) {
		date = date.Add(-24 * time.Hour)
	}

	const lookBackDays = 7
	var (
		v  float64
		ok bool
	)
	for i := 0; i < lookBackDays; i++ {
		v, ok = r[date]
		if ok {
			break
		}
		date = date.Add(-24 * time.Hour)
	}

	if !ok {
		return 0, errors.New("no rate found for " + initialDate.Format("2006-01-02"))
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
	rates DailyRates
}

func (c converter) A2B(date time.Time, v float64) (float64, error) {
	r, err := c.rates.GetRate(date)
	if err != nil {
		return 0, err
	}
	return v / r, nil
}

func (c converter) B2A(date time.Time, v float64) (float64, error) {
	r, err := c.rates.GetRate(date)
	if err != nil {
		return 0, err
	}
	return v * r, nil
}

func validateRates(rates DailyRates) error {
	if len(rates) == 0 {
		return errors.New("rates shall not be empty")
	}

	f := func(date time.Time, rate float64) string {
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

//go:embed conversion_rate_usd2eur.xml
var ratesDaily []byte

func newRatesDefault() DailyRates {
	type rates struct {
		DataSet struct {
			Series struct {
				Obs []struct {
					Date string  `xml:"TIME_PERIOD,attr"`
					Rate float64 `xml:"OBS_VALUE,attr"`
				} `xml:"Obs"`
			} `xml:"Series"`
		} `xml:"DataSet"`
	}

	var tmp rates
	if err := xml.Unmarshal(ratesDaily, &tmp); err != nil {
		panic(err)
	}

	o := DailyRates{}
	for _, rate := range tmp.DataSet.Series.Obs {
		date, err := time.Parse("2006-01-02", rate.Date)
		if err != nil {
			panic(err)
		}
		o[date] = rate.Rate
	}

	return o
}

type MockConverter struct {
	V   float64
	Err error
}

func (m MockConverter) A2B(_ time.Time, _ float64) (float64, error) {
	return m.V, m.Err
}

func (m MockConverter) B2A(_ time.Time, _ float64) (float64, error) {
	return m.V, m.Err
}

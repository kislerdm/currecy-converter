package converter

import (
	"errors"
	"reflect"
	"testing"
	"time"
)

func Test_newRatesDefault(t *testing.T) {
	t.Run(
		"parse file", func(t *testing.T) {
			// GIVEN
			// the xml file collocated with the package

			// WHEN
			got := newRatesDefault()

			// THEN

			// the test date is Sunday
			const testDate = "2023-01-01"
			// the value is taken from Friday of the prev. week
			const wantValue float64 = 1.0666

			wantDate, err := time.Parse("2006-01-02", testDate)
			if err != nil {
				panic(err)
			}

			v, err := got.GetRate(wantDate)
			if err != nil {
				t.Errorf("unexpected error")
				return
			}

			if v != wantValue {
				t.Errorf("unexpected rate for the test date: " + testDate)
			}
		},
	)
}

func TestNewConverterDaily(t *testing.T) {
	type args struct {
		rates DailyRates
	}
	var customRates = DailyRates{
		time.Date(2022, 12, 30, 0, 0, 0, 0, &time.Location{}): 1.0666,
	}
	tests := []struct {
		name    string
		args    args
		want    Converter
		wantErr bool
	}{
		{
			name: "happy path: default rates",
			args: args{
				rates: nil,
			},
			want:    &converter{newRatesDefault()},
			wantErr: false,
		},
		{
			name: "happy path: custom rates",
			args: args{
				rates: customRates,
			},
			want: &converter{
				customRates,
			},
			wantErr: false,
		},
		{
			name: "unhappy path: empty rates",
			args: args{
				rates: DailyRates{},
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				got, err := NewConverter(tt.args.rates)
				if (err != nil) != tt.wantErr {
					t.Errorf("NewConverter() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
				if !reflect.DeepEqual(got, tt.want) {
					t.Errorf("NewConverter() got = %v, want %v", got, tt.want)
				}
			},
		)
	}
}

func Test_validateRates(t *testing.T) {
	type args struct {
		rates DailyRates
	}
	tests := []struct {
		name    string
		args    args
		wantErr error
	}{
		{
			name: "no error",
			args: args{
				rates: newRatesDefault(),
			},
			wantErr: nil,
		},
		{
			name: "error: rate from future",
			args: args{
				rates: DailyRates{
					time.Date(
						9999, 12, 31, 23, 59, 59, 999999999,
						&time.Location{},
					): 1.0666,
				},
			},
			wantErr: errors.New("historical rates are expected only, got: date=9999-12-31, rate=1.0666"),
		},
		{
			name: "error: negative rate",
			args: args{
				rates: DailyRates{
					time.Date(
						2000, 12, 31, 23, 59, 59, 999999999,
						&time.Location{},
					): -1.0666,
				},
			},
			wantErr: errors.New("rate shall be positive, got: date=2000-12-31, rate=-1.0666"),
		},
		{
			name:    "error: empty rates",
			args:    args{},
			wantErr: errors.New("rates shall not be empty"),
		},
	}
	t.Parallel()
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				if err := validateRates(tt.args.rates); !reflect.DeepEqual(err, tt.wantErr) {
					t.Errorf("validateRates() error = %v, wantErr %v", err, tt.wantErr)
				}
			},
		)
	}
}

func TestEUR2USDRates_GetRate(t *testing.T) {
	type args struct {
		date time.Time
	}
	tests := []struct {
		name    string
		r       DailyRates
		args    args
		want    float64
		wantErr bool
	}{
		{
			name: "happy path: weekend",
			r:    newRatesDefault(),
			args: args{
				date: time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
			},
			want:    1.0666,
			wantErr: false,
		},
		{
			name: "unhappy path: rate not found",
			r: DailyRates{
				time.Date(2000, 12, 31, 0, 0, 0, 0, time.UTC): -1.0666,
			},
			args: args{
				date: time.Date(2020, 12, 31, 0, 0, 0, 0, time.UTC),
			},
			want:    0,
			wantErr: true,
		},
	}
	t.Parallel()
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				got, err := tt.r.GetRate(tt.args.date)
				if (err != nil) != tt.wantErr {
					t.Errorf("GetRate() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
				if got != tt.want {
					t.Errorf("GetRate() got = %v, want %v", got, tt.want)
				}
			},
		)
	}
}

func Test_converter_USD2EUR(t *testing.T) {
	type fields struct {
		rates DailyRates
	}
	type args struct {
		date time.Time
		v    float64
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    float64
		wantErr bool
	}{
		{
			name: "happy path",
			fields: fields{
				rates: newRatesDefault(),
			},
			args: args{
				date: time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
				v:    10666,
			},
			want:    10000,
			wantErr: false,
		},
		{
			name: "unhappy path",
			fields: fields{
				rates: newRatesDefault(),
			},
			args: args{
				date: time.Time{},
				v:    10666,
			},
			want:    0,
			wantErr: true,
		},
	}
	t.Parallel()
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				c := converter{
					rates: tt.fields.rates,
				}
				got, err := c.A2B(tt.args.date, tt.args.v)
				if (err != nil) != tt.wantErr {
					t.Errorf("A2B() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
				if got != tt.want {
					t.Errorf("A2B() got = %v, want %v", got, tt.want)
				}
			},
		)
	}
}

func Test_converter_EUR2USD(t *testing.T) {
	type fields struct {
		rates DailyRates
	}
	type args struct {
		date time.Time
		v    float64
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    float64
		wantErr bool
	}{
		{
			name: "happy path",
			fields: fields{
				rates: newRatesDefault(),
			},
			args: args{
				date: time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
				v:    10000,
			},
			want:    10666,
			wantErr: false,
		},
		{
			name: "unhappy path",
			fields: fields{
				rates: newRatesDefault(),
			},
			args: args{
				date: time.Time{},
				v:    10000,
			},
			want:    0,
			wantErr: true,
		},
	}
	t.Parallel()
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				c := converter{
					rates: tt.fields.rates,
				}
				got, err := c.B2A(tt.args.date, tt.args.v)
				if (err != nil) != tt.wantErr {
					t.Errorf("B2A() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
				if got != tt.want {
					t.Errorf("B2A() got = %v, want %v", got, tt.want)
				}
			},
		)
	}
}

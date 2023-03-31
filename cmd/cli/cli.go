package main

import (
	"bytes"
	"encoding/csv"
	"errors"
	"flag"
	"io"
	"os"
	"strconv"
	"strings"
	"time"

	converter "github.com/kislerdm/usd2eur"
)

func isHelp(args []string) bool {
	for _, arg := range args {
		switch strings.ToLower(
			strings.TrimLeftFunc(
				arg, func(r rune) bool {
					return r == '-'
				},
			),
		) {
		case "h", "help":
			return true
		}
	}
	return false
}

type CLI struct {
	in        io.Reader
	out       io.Writer
	flags     *flag.FlagSet
	cfg       *cliConfig
	converter converter.Converter
}

type cliConfig struct {
	pathCSV string
	pathOut string
	b2a     bool
}

func writeStrings(buf *bytes.Buffer, strings ...string) {
	for _, s := range strings {
		_, _ = buf.WriteString(s)
	}
}

func (c CLI) help() error {
	const help = `Tool to convert USD to EUR.

Options:
`
	var buf bytes.Buffer
	_, _ = buf.WriteString(help)

	c.flags.VisitAll(
		func(f *flag.Flag) {
			writeStrings(&buf, "\t-", f.Name, ": ", f.Usage, " [default: \"", f.DefValue, "\"]\n")
		},
	)

	_, err := c.out.Write(buf.Bytes())
	return err
}

func (c CLI) Run(args []string) error {
	if isHelp(args) {
		if err := c.help(); err != nil {
			return err
		}
		return nil
	}

	if err := c.flags.Parse(args); err != nil {
		return err
	}

	var (
		r   = c.in
		w   = c.out
		err error
	)

	if c.cfg.pathCSV != "" {
		w, err = os.Open(c.cfg.pathCSV)
	}

	if c.cfg.pathOut != "" {
		w, err = os.Open(c.cfg.pathOut)
	}

	if err != nil {
		return err
	}

	return c.convertRow(r, w)
}

func (c CLI) convertRow(r io.Reader, w io.Writer) error {
	csvReader := csv.NewReader(r)
	csvWriter := csv.NewWriter(w)

	var rowIndex uint16
	for {
		row, err := csvReader.Read()
		if err != nil {
			if err == io.EOF {
				break
			}
			return errors.New("error reading csv row " + strconv.FormatUint(uint64(rowIndex), 10))
		}

		dateRaw := row[0]
		date, err := time.Parse("2006-01-02", dateRaw)
		if err != nil {
			return errors.New("error parsing date in the row " + strconv.FormatUint(uint64(rowIndex), 10))
		}

		amount, err := strconv.ParseFloat(row[1], 16)
		if err != nil {
			return errors.New("error parsing amount in the row " + strconv.FormatUint(uint64(rowIndex), 10))
		}

		var conversionFn = c.converter.A2B
		if c.cfg.b2a {
			conversionFn = c.converter.B2A
		}

		amountOut, err := conversionFn(date, amount)
		if err != nil {
			return errors.New(
				"conversion error in the row " + strconv.FormatUint(uint64(rowIndex), 10) + ". " +
					err.Error(),
			)
		}

		if err := csvWriter.Write(
			[]string{
				dateRaw,
				strconv.FormatFloat(amountOut, 'f', -1, 64),
			},
		); err != nil {
			return errors.New("error writing the row " + strconv.FormatUint(uint64(rowIndex), 10))
		}
		csvWriter.Flush()
		rowIndex++
	}
	return nil
}

type devNull struct{}

func (d devNull) Write(p []byte) (n int, err error) {
	return len(p), nil
}

func NewCLI(stdIn io.Reader, writerOut io.Writer, converter converter.Converter) CLI {
	c := cliConfig{}

	f := flag.NewFlagSet("", flag.ContinueOnError)
	f.StringVar(
		&c.pathCSV, "i", "",
		`Path to input headless csv with the structure col0:date, col1:amount). If empty, stdin will be read.`,
	)
	f.StringVar(
		&c.pathOut, "o", "",
		`Path to output csv. If empty, result will be printed to stdout.`,
	)
	f.BoolVar(&c.b2a, "b2a", false, `Flag to convert EUR to USD.`)

	// to prevent default output of flag parsing errors
	f.Usage = func() {}
	f.SetOutput(devNull{})

	return CLI{
		in:        stdIn,
		out:       writerOut,
		flags:     f,
		cfg:       &c,
		converter: converter,
	}
}

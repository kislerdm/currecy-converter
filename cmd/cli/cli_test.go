package main

import (
	"errors"
	"io"
	"reflect"
	"strings"
	"testing"

	converter "github.com/kislerdm/usd2eur"
)

type writer struct {
	V []byte
}

func (w *writer) Write(p []byte) (n int, err error) {
	w.V = p
	return len(p), nil
}

func Test_help(t *testing.T) {
	t.Parallel()

	t.Run(
		"happy path", func(t *testing.T) {
			// GIVEN
			wOut := writer{}
			wantOut := []byte(`Tool to convert USD to EUR.

Options:
	-b2a: Flag to convert EUR to USD. [default: "false"]
	-i: Path to input headless csv with the structure col0:date, col1:amount). If empty, stdin will be read. [default: ""]
	-o: Path to output csv. If empty, result will be printed to stdout. [default: ""]
`)

			cli := NewCLI(nil, &wOut, nil)

			// WHEN
			err := cli.help()

			// THEN
			if err != nil {
				t.Error(err)
				return
			}

			if !reflect.DeepEqual(wOut.V, wantOut) {
				t.Error("unexpected result")
				return
			}

		},
	)
}

func TestCLI_Run(t *testing.T) {
	type fields struct {
		in   io.Reader
		out  io.Writer
		conv converter.Converter
	}
	type args struct {
		args []string
	}
	tests := []struct {
		name              string
		fields            fields
		args              args
		wantStdOutContent []byte
		wantErr           error
	}{
		{
			name: "happy path: help",
			fields: fields{
				out: &writer{},
			},
			args: args{
				args: []string{"-help"},
			},
			wantStdOutContent: []byte(`Tool to convert USD to EUR.

Options:
	-b2a: Flag to convert EUR to USD. [default: "false"]
	-i: Path to input headless csv with the structure col0:date, col1:amount). If empty, stdin will be read. [default: ""]
	-o: Path to output csv. If empty, result will be printed to stdout. [default: ""]
`),
			wantErr: nil,
		},
		{
			name: "happy path: default options",
			fields: fields{
				in:   strings.NewReader(`2023-01-01,1.1`),
				out:  &writer{},
				conv: converter.MockConverter{V: 11},
			},
			wantStdOutContent: []byte("2023-01-01,11\n"),
			wantErr:           nil,
		},
		{
			name: "unhappy path: unknown option",
			fields: fields{
				out: &writer{},
			},
			args: args{
				args: []string{"-foo"},
			},
			wantErr: errors.New("flag provided but not defined: -foo"),
		},
	}

	t.Parallel()
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				c := NewCLI(tt.fields.in, tt.fields.out, tt.fields.conv)
				if err := c.Run(tt.args.args); !reflect.DeepEqual(err, tt.wantErr) {
					t.Errorf("Run() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
				if !reflect.DeepEqual(tt.fields.out.(*writer).V, tt.wantStdOutContent) {
					t.Error("unexpected output to out")
					return
				}
			},
		)
	}
}

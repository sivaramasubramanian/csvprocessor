package csvprocessor_test

import (
	"context"
	"encoding/csv"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/sivaramasubramanian/csvprocessor"
)

func TestNew(t *testing.T) {
	var buffer = make([]strings.Builder, 4)

	type args struct {
		opts []csvprocessor.Option
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "error on invalid input",
			args: args{
				opts: nil,
			},
			wantErr: true,
		},
		{
			name: "error on invalid input - without output generator",
			args: args{
				opts: []csvprocessor.Option{
					csvprocessor.WithReader(csv.NewReader(strings.NewReader("a,b,c"))),
				},
			},
			wantErr: true,
		},
		{
			name: "error on invalid input - with invalid chunk size",
			args: args{
				opts: []csvprocessor.Option{
					csvprocessor.WithReader(csv.NewReader(strings.NewReader("a,b,c"))),
					csvprocessor.WithOutputFileFormat("abcd.csv"),
					csvprocessor.WithChunkSize(-1),
				},
			},
			wantErr: true,
		},
		{
			name: "error on invalid input - with writer generator but invalid chunk size",
			args: args{
				opts: []csvprocessor.Option{
					csvprocessor.WithReader(csv.NewReader(strings.NewReader("a,b,c"))),
					csvprocessor.WithWriterGenerator(func(i int) (io.WriteCloser, error) {
						return csvprocessor.NoOpCloser(&buffer[i-1]), nil
					}),
					csvprocessor.WithChunkSize(-1),
				},
			},
			wantErr: true,
		},
		{
			name: "error on invalid input - with invalid output format",
			args: args{
				opts: []csvprocessor.Option{
					csvprocessor.WithReader(csv.NewReader(strings.NewReader("a,b,c"))),
					csvprocessor.WithOutputFileFormat(""),
					csvprocessor.WithChunkSize(2),
				},
			},
			wantErr: true,
		},
		{
			name: "error on invalid input - with invalid output format",
			args: args{
				opts: []csvprocessor.Option{
					csvprocessor.WithReader(csv.NewReader(strings.NewReader("a,b,c"))),
					csvprocessor.WithOutputFileFormat("abc.csv"),
					csvprocessor.WithChunkSize(2),
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := csvprocessor.New(tt.args.opts...)
			if (err != nil) != tt.wantErr {
				t.Errorf("New() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && got == nil {
				t.Errorf("New() returned nil; expected non-nil value")
				return
			}
		})
	}
}

func TestNewFileReader(t *testing.T) {
	tempOutputFile, err := os.CreateTemp(t.TempDir(), "test_new_file_reader_*.csv")
	if err != nil {
		t.Errorf("NewFileReader() unable to create temp file for testing; error = %v", err)
	}

	type args struct {
		inputFile        string
		chunkSize        int
		outputFileFormat string
		rowTransformer   func(context.Context, []string) []string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "test file reader with valid input",
			args: args{
				inputFile:        tempOutputFile.Name(),
				chunkSize:        100,
				outputFileFormat: "output.csv",
				rowTransformer:   nil,
			},
			wantErr: false,
		},
		{
			name: "test file reader with row transformer",
			args: args{
				inputFile:        tempOutputFile.Name(),
				chunkSize:        100,
				outputFileFormat: "output.csv",
				rowTransformer:   csvprocessor.AddChunkRowNoTransformer("new_column"),
			},
			wantErr: false,
		},
		{
			name: "test file reader with invalid input - invalid input file",
			args: args{
				inputFile:        "non-existent-file",
				chunkSize:        100,
				outputFileFormat: "output.csv",
				rowTransformer:   nil,
			},
			wantErr: true,
		},
		{
			name: "test file reader with invalid input - invalid chunk size",
			args: args{
				inputFile:        tempOutputFile.Name(),
				chunkSize:        -1,
				outputFileFormat: "output.csv",
				rowTransformer:   nil,
			},
			wantErr: true,
		},
		{
			name: "test file reader with valid input - invalid output format",
			args: args{
				inputFile:        tempOutputFile.Name(),
				chunkSize:        100,
				outputFileFormat: "",
				rowTransformer:   nil,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := csvprocessor.NewFileReader(tt.args.inputFile, tt.args.chunkSize, tt.args.outputFileFormat, tt.args.rowTransformer)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewFileReader() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && got == nil {
				t.Errorf("NewFileReader() returned nil; expected non-nil value")
				return
			}
		})
	}
}

func TestNewBufferReader(t *testing.T) {
	type args struct {
		inputReader  io.Reader
		outputWriter io.WriteCloser
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "test new buffer reader - valid input",
			args: args{
				inputReader:  strings.NewReader("a.b,c.d,e.f"),
				outputWriter: csvprocessor.NoOpCloser(&strings.Builder{}),
			},
			wantErr: false,
		},
		{
			name: "test new buffer reader - invalid reader",
			args: args{
				inputReader:  nil,
				outputWriter: csvprocessor.NoOpCloser(&strings.Builder{}),
			},
			wantErr: true,
		},
		{
			name: "test new buffer reader - invalid writer",
			args: args{
				inputReader:  strings.NewReader("a.b,c.d,e.f"),
				outputWriter: nil,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := csvprocessor.NewBufferReader(tt.args.inputReader, tt.args.outputWriter)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewBufferReader() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && got == nil {
				t.Errorf("NewBufferReader() returned nil; expected non-nil value")
				return
			}
		})
	}
}

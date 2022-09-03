// Package csvprocessor can be used to efficiently transform and split CSV files.
//
// See README.md for more details
package csvprocessor_test

import (
	"encoding/csv"
	"io"
	"strings"
	"testing"

	"github.com/sivaramasubramanian/csvprocessor"
)

const verySmallCSV = `a,b,c
d,e,f
g,h,i
j,k,l
`

// noOpLogger is used for disabling logging when running benchmarks.
var noOpLogger csvprocessor.Logger = func(string, ...any) {

}

type args struct {
	reader         io.Reader
	opt            []csvprocessor.Option
	expectedChunks int
}

type expect struct {
	wantErr       bool
	bufferLength  int
	elementLength int
}

func TestProcessor_Process(t *testing.T) {
	tests := []struct {
		name   string
		args   args
		expect expect
	}{
		{
			name: "Test With Headers - verySmallCSV",
			args: args{
				reader:         strings.NewReader(verySmallCSV),
				expectedChunks: 3,
				opt: []csvprocessor.Option{
					csvprocessor.WithLogger(t.Logf),
				},
			},
			expect: expect{
				wantErr:       false,
				bufferLength:  3,
				elementLength: 12,
			},
		},
		{
			name: "Test Without Headers - verySmallCSV",
			args: args{
				reader:         strings.NewReader(verySmallCSV),
				expectedChunks: 4,
				opt: []csvprocessor.Option{
					csvprocessor.SkipHeaders(true),
					csvprocessor.WithLogger(t.Logf),
				},
			},
			expect: expect{
				wantErr:       false,
				bufferLength:  4,
				elementLength: 6,
			},
		},
		{
			name: "Test Without Headers - smallCSV",
			args: args{
				reader:         strings.NewReader(strings.Repeat(verySmallCSV, 100)),
				expectedChunks: 4 * 100,
				opt: []csvprocessor.Option{
					csvprocessor.SkipHeaders(true),
					csvprocessor.WithLogger(t.Logf),
				},
			},
			expect: expect{
				wantErr:       false,
				bufferLength:  400,
				elementLength: 6,
			},
		},
		{
			name: "Test - smallCSV - chunk 10",
			args: args{
				reader: strings.NewReader(strings.Repeat(verySmallCSV, 100)),
				// 400 lines = 1 header + 399 lines
				expectedChunks: 399 / 3,
				opt: []csvprocessor.Option{
					csvprocessor.SkipHeaders(false),
					csvprocessor.WithLogger(t.Logf),
					csvprocessor.WithChunkSize(3),
				},
			},
			expect: expect{
				wantErr:       false,
				bufferLength:  399 / 3,
				elementLength: (6 * 3) + 6,
			},
		},
		{
			name: "Test Without Headers - smallCSV - chunk 10",
			args: args{
				reader: strings.NewReader(strings.Repeat(verySmallCSV, 100)),
				// 400 lines = 0 header + 400 lines
				expectedChunks: ((4 * 100) / 10),
				opt: []csvprocessor.Option{
					csvprocessor.SkipHeaders(true),
					csvprocessor.WithLogger(t.Logf),
					csvprocessor.WithChunkSize(10),
				},
			},
			expect: expect{
				wantErr:       false,
				bufferLength:  ((4 * 100) / 10),
				elementLength: 6 * 10,
			},
		},
		{
			name: "Test With Headers - verySmallCSV - with transformer",
			args: args{
				reader:         strings.NewReader(verySmallCSV),
				expectedChunks: 3,
				opt: []csvprocessor.Option{
					csvprocessor.WithLogger(t.Logf),
					csvprocessor.WithTransformer(
						csvprocessor.AddChunkRowNoTransformer("id"),
					),
				},
			},
			expect: expect{
				wantErr:      false,
				bufferLength: 3,
				// existing content + id header + id value
				elementLength: (6 * 2) + (2 + 1) + (1 + 1),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buffer = make([]strings.Builder, tt.args.expectedChunks)
			proc := newProcessor(t, tt.args.reader, buffer, tt.args.opt...)
			if err := proc.Process(); (err != nil) != tt.expect.wantErr {
				t.Errorf("Processor.Process() error = %v, wantErr %v", err, tt.expect.wantErr)
			}

			if len(buffer) != tt.expect.bufferLength {
				t.Errorf("Processor.Process() bufferLength = %v, expect %v", len(buffer), tt.expect.bufferLength)
			}

			for index, element := range buffer {
				if element.Len() != tt.expect.elementLength {
					t.Errorf("Processor.Process() index = %v, elementLength = %v, expect %v", index, element.Len(), tt.expect.elementLength)
				}
			}
		})
	}
}

func BenchmarkProcessor(b *testing.B) {
	csv4mRows := strings.NewReader(strings.Repeat(verySmallCSV, 1_000_000))
	benches := []struct {
		name   string
		args   args
		expect expect
	}{

		{
			name: "BenchmarkProcessor_Process_4000_rows",
			args: args{
				// 4 * 1000
				reader:         strings.NewReader(strings.Repeat(verySmallCSV, 1000)),
				expectedChunks: ((4 * 1000) / 10),
				opt: []csvprocessor.Option{
					csvprocessor.WithLogger(noOpLogger),
					csvprocessor.WithChunkSize(10),
				},
			},
			expect: expect{
				wantErr: false,
			},
		},
		{
			name: "BenchmarkProcessor_Process_4000_rows_cz_100",
			args: args{
				reader:         strings.NewReader(strings.Repeat(verySmallCSV, 1000)),
				expectedChunks: ((4 * 1000) / 100),
				opt: []csvprocessor.Option{
					csvprocessor.WithLogger(noOpLogger),
					csvprocessor.WithChunkSize(100),
				},
			},
			expect: expect{
				wantErr: false,
			},
		},
		{
			name: "BenchmarkProcessor_ProcessWithoutHeader_4000_rows",
			args: args{
				reader:         strings.NewReader(strings.Repeat(verySmallCSV, 1000)),
				expectedChunks: ((4 * 1000) / 10),
				opt: []csvprocessor.Option{
					csvprocessor.WithLogger(noOpLogger),
					csvprocessor.WithChunkSize(10),
					csvprocessor.SkipHeaders(true),
				},
			},
			expect: expect{
				wantErr: false,
			},
		},
		{
			name: "BenchmarkProcessor_ProcessWithoutHeader_4_000_000_rows_cz_100",
			args: args{
				reader:         csv4mRows,
				expectedChunks: ((4 * 1_000_000) / 1000),
				opt: []csvprocessor.Option{
					csvprocessor.WithLogger(noOpLogger),
					csvprocessor.WithChunkSize(1000),
					csvprocessor.SkipHeaders(true),
				},
			},
			expect: expect{
				wantErr: false,
			},
		},
	}

	for _, bench := range benches {
		b.Run(bench.name, func(b *testing.B) {
			runBenchmark(b, bench.args, bench.expect.wantErr)
		})
	}
}

func runBenchmark(b *testing.B, arg args, wantErr bool) {
	b.Helper()

	for i := 0; i < b.N; i++ {
		var buffer = make([]strings.Builder, arg.expectedChunks)
		proc := newProcessor(b, arg.reader, buffer, arg.opt...)
		if err := proc.Process(); (err != nil) != wantErr {
			b.Errorf(" error = %v, wantErr %v", err, wantErr)
		}
	}
}

// Util Functions for test.
func newProcessor(tb testing.TB, reader io.Reader, bytesArr []strings.Builder, opt ...csvprocessor.Option) *csvprocessor.Processor {
	tb.Helper()

	bufferOpt := []csvprocessor.Option{
		csvprocessor.WithReader(csv.NewReader(reader)),
		csvprocessor.WithWriterGenerator(func(i int) (io.WriteCloser, error) {
			return csvprocessor.NoOpCloser(&bytesArr[i-1]), nil
		}),
		csvprocessor.WithChunkSize(1),
	}

	bufferOpt = append(bufferOpt, opt...)
	c, err := csvprocessor.New(
		bufferOpt...,
	)
	if err != nil {
		tb.Errorf("Processor.Process() buffer creation error = %v", err)
	}
	return c
}

type any = interface{} //nolint:predeclared

package csvprocessor_test

import (
	"context"
	"reflect"
	"testing"

	"github.com/sivaramasubramanian/csvprocessor"
)

func TestPanicSafe(t *testing.T) {
	type args struct {
		transformer csvprocessor.CsvRowTransformer
		log         csvprocessor.Logger
	}
	tests := []struct {
		name string
		args args
		want csvprocessor.CsvRowTransformer
	}{
		{
			name: "Test panic is recovered",
			args: args{
				transformer: func(ctx context.Context, s []string) []string {
					panic("test")
				},
				log: t.Logf,
			},
		},
		{
			name: "Test non-panic transformer",
			args: args{
				transformer: csvprocessor.NoOpTransformer(),
				log:         t.Logf,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			csvprocessor.PanicSafe(tt.args.transformer, tt.args.log)(context.TODO(), []string{})
		})
	}
}

func TestDebugWrapper(t *testing.T) {
	var calls = 0
	type args struct {
		transformer csvprocessor.CsvRowTransformer
		log         csvprocessor.Logger
		callCount   int
	}
	tests := []struct {
		name string
		args args
		want csvprocessor.CsvRowTransformer
	}{
		{
			name: "Test debug logs are called",
			args: args{
				transformer: csvprocessor.NoOpTransformer(),
				log:         callCount(t.Logf, &calls),
				callCount:   2,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			csvprocessor.DebugWrapper(tt.args.transformer, tt.args.log)(context.TODO(), []string{})
			if tt.args.callCount > 0 {
				if !reflect.DeepEqual(calls, tt.args.callCount) {
					t.Errorf("DebugWrapper() = %v, want %v", calls, tt.want)
				}
			}
		})
	}
}

func callCount(log csvprocessor.Logger, count *int) csvprocessor.Logger {
	return func(format string, args ...any) {
		*count++
		log(format, args...)
	}
}

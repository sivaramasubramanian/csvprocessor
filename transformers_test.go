package csvprocessor_test

import (
	"context"
	"reflect"
	"testing"

	"github.com/sivaramasubramanian/csvprocessor"
)

func TestAddRowNoTransformer(t *testing.T) {
	transformer := csvprocessor.AddRowNoTransformer("test column")

	type args struct {
		ctx      context.Context //nolint:containedctx
		inputRow []string
	}
	tests := []struct {
		name string
		args args
		want []string
	}{
		{
			name: "Test non-header row",
			args: args{
				ctx:      context.WithValue(context.TODO(), csvprocessor.CtxRowNum, 1),
				inputRow: []string{"a", "b"},
			},
			want: []string{"1", "a", "b"},
		},
		{
			name: "Test header row",
			args: args{
				ctx:      context.WithValue(context.WithValue(context.TODO(), csvprocessor.CtxRowNum, 1), csvprocessor.CtxIsHeader, true),
				inputRow: []string{"a", "b"},
			},
			want: []string{"test column", "a", "b"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := transformer(tt.args.ctx, tt.args.inputRow)
			if !reflect.DeepEqual(actual, tt.want) {
				t.Errorf("AddRowNoTransformer() = %v, want %v", actual, tt.want)
			}
		})
	}
}

func TestAddChunkRowNoTransformer(t *testing.T) {
	transformer := csvprocessor.AddChunkRowNoTransformer("test column")

	type args struct {
		ctx      context.Context //nolint:containedctx
		inputRow []string
	}
	tests := []struct {
		name string
		args args
		want []string
	}{
		{
			name: "Test non-header chunk row no",
			args: args{
				ctx:      context.WithValue(context.WithValue(context.TODO(), csvprocessor.CtxRowNum, 1), csvprocessor.CtxChunkSize, 100),
				inputRow: []string{"a", "b"},
			},
			want: []string{"1", "a", "b"},
		},
		{
			name: "Test last row in chunk",
			args: args{
				ctx:      context.WithValue(context.WithValue(context.TODO(), csvprocessor.CtxRowNum, 100), csvprocessor.CtxChunkSize, 100),
				inputRow: []string{"b", "c"},
			},
			want: []string{"100", "b", "c"},
		},
		{
			name: "Test row no in 2nd chunk",
			args: args{
				ctx:      context.WithValue(context.WithValue(context.TODO(), csvprocessor.CtxRowNum, 202), csvprocessor.CtxChunkSize, 100),
				inputRow: []string{"b", "c"},
			},
			want: []string{"2", "b", "c"},
		},
		{
			name: "Test header chunk row no",
			args: args{
				ctx:      context.WithValue(context.WithValue(context.WithValue(context.TODO(), csvprocessor.CtxRowNum, 1), csvprocessor.CtxChunkSize, 100), csvprocessor.CtxIsHeader, true),
				inputRow: []string{"a", "b"},
			},
			want: []string{"test column", "a", "b"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := transformer(tt.args.ctx, tt.args.inputRow)
			if !reflect.DeepEqual(actual, tt.want) {
				t.Errorf("TestAddChunkRowNoTransformer() = %v, want %v", actual, tt.want)
			}
		})
	}
}

func TestAddConstantColumnTransformer(t *testing.T) {
	transformer := csvprocessor.AddConstantColumnTransformer("const column", "hello", 2)

	type args struct {
		ctx      context.Context //nolint:containedctx
		inputRow []string
	}
	tests := []struct {
		name string
		args args
		want []string
	}{
		{
			name: "Test non-header chunk row no",
			args: args{
				ctx:      context.WithValue(context.TODO(), csvprocessor.CtxIsHeader, false),
				inputRow: []string{"a", "b"},
			},
			want: []string{"a", "b", "hello"},
		},
		{
			name: "Test header row const column",
			args: args{
				ctx:      context.WithValue(context.TODO(), csvprocessor.CtxIsHeader, true),
				inputRow: []string{"a", "b"},
			},
			want: []string{"a", "b", "const column"},
		},
		{
			name: "Test duplicate header row const column",
			args: args{
				ctx:      context.WithValue(context.TODO(), csvprocessor.CtxIsHeader, true),
				inputRow: []string{"a", "const column"},
			},
			want: []string{"a", "const column", "const column"},
		},
		{
			name: "Test row has more columns",
			args: args{
				ctx:      context.WithValue(context.TODO(), csvprocessor.CtxIsHeader, true),
				inputRow: []string{"a", "b", "c", "d"},
			},
			want: []string{"a", "b", "const column", "c", "d"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := transformer(tt.args.ctx, tt.args.inputRow)
			if !reflect.DeepEqual(actual, tt.want) {
				t.Errorf("AddConstantColumnTransformer() = %v, want %v", actual, tt.want)
			}
		})
	}
}

func TestReplaceValuesTransformer(t *testing.T) {
	replacements := make(map[string]string)
	replacements["NULL"] = ""
	transformer := csvprocessor.ReplaceValuesTransformer(replacements)

	type args struct {
		ctx      context.Context //nolint:containedctx
		inputRow []string
	}
	tests := []struct {
		name string
		args args
		want []string
	}{
		{
			name: "Test replacements",
			args: args{
				ctx:      context.WithValue(context.TODO(), csvprocessor.CtxIsHeader, false),
				inputRow: []string{"a", "b", "NULL"},
			},
			want: []string{"a", "b", ""},
		},
		{
			name: "Test row without any match",
			args: args{
				ctx:      context.WithValue(context.TODO(), csvprocessor.CtxIsHeader, false),
				inputRow: []string{"a", "b", "c"},
			},
			want: []string{"a", "b", "c"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := transformer(tt.args.ctx, tt.args.inputRow)
			if !reflect.DeepEqual(actual, tt.want) {
				t.Errorf("ReplaceValuesTransformer() = %v, want %v", actual, tt.want)
			}
		})
	}
}

func TestChainTransformers(t *testing.T) {
	replacements := make(map[string]string)
	replacements["NULL"] = ""
	replaceValues := csvprocessor.ReplaceValuesTransformer(replacements)
	addConstColumn := csvprocessor.AddConstantColumnTransformer("const column", "hello", 2)

	transformer := csvprocessor.ChainTransformers(replaceValues, addConstColumn)

	type args struct {
		ctx      context.Context //nolint:containedctx
		inputRow []string
	}
	tests := []struct {
		name string
		args args
		want []string
	}{
		{
			name: "Test replacements",
			args: args{
				ctx:      context.WithValue(context.TODO(), csvprocessor.CtxIsHeader, false),
				inputRow: []string{"a", "b", "NULL"},
			},
			want: []string{"a", "b", "hello", ""},
		},
		{
			name: "Test row without any match",
			args: args{
				ctx:      context.WithValue(context.TODO(), csvprocessor.CtxIsHeader, false),
				inputRow: []string{"a", "b", "c"},
			},
			want: []string{"a", "b", "hello", "c"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := transformer(tt.args.ctx, tt.args.inputRow)
			if !reflect.DeepEqual(actual, tt.want) {
				t.Errorf("ChainTransformers() = %v, want %v", actual, tt.want)
			}
		})
	}
}

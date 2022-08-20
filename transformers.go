package csvprocessor

import (
	"context"
	"fmt"
)

// NoOpTransformer applies no transformations on the rows.
// Can be used when only splitting is needed without any modifications to the rows.
func NoOpTransformer() CsvRowTransformer {
	return func(ctx context.Context, row []string) []string {
		return row
	}
}

// AddRowNoTransformer adds a row number to each row.
// This uses the overall row number across chunks, For chunk-wise row number use AddChunkRowNoTransformer().
// If SkipHeaders is false, it will add a header column for the row number with the given columnName.
func AddRowNoTransformer(columnName string) CsvRowTransformer {
	return func(ctx context.Context, row []string) []string {
		isHeader := (ctx.Value(CtxIsHeader)).(bool)
		if isHeader {
			return addToSliceAtIndex(row, columnName, 0)
		}

		rowID := ctx.Value(CtxRowNum).(int)
		_ = (ctx.Value(CtxChunkSize)).(int)
		return addToSliceAtIndex(row, fmt.Sprintf("%d", rowID), 0)
	}
}

// AddChunkRowNoTransformer adds the row number within current chunk to each row.
// If SkipHeaders is false, it will add a header column for the row number with the given columnName.
func AddChunkRowNoTransformer(columnName string) CsvRowTransformer {
	return func(ctx context.Context, row []string) []string {
		isHeader := (ctx.Value(CtxIsHeader)).(bool)
		if isHeader {
			return addToSliceAtIndex(row, columnName, 0)
		}

		rowID := ctx.Value(CtxRowNum).(int)
		chunkSize := (ctx.Value(CtxChunkSize)).(int)
		chunkRowID := (rowID % chunkSize)
		if chunkRowID == 0 {
			chunkRowID = chunkSize
		}

		return addToSliceAtIndex(row, fmt.Sprintf("%d", chunkRowID), 0)
	}
}

// ReplaceValuesTransformer adds a row number to each row.
// If SkipHeaders is false, it will add a header column for the row number with the given columnName.
func ReplaceValuesTransformer(replacements map[string]string) CsvRowTransformer {
	return func(ctx context.Context, row []string) []string {
		isHeader := (ctx.Value(CtxIsHeader)).(bool)
		if isHeader {
			return row
		}

		for i, element := range row {
			if val, ok := replacements[element]; ok {
				row[i] = val
			}
		}

		return row
	}
}

// AddConstantColumnTransformer adds a new column with the given constant value.
// If SkipHeaders is false, it will add a header column for the row number with the given columnName.
func AddConstantColumnTransformer(columnName string, val string, columIndex int) CsvRowTransformer {
	return func(ctx context.Context, row []string) []string {
		isHeader := (ctx.Value(CtxIsHeader)).(bool)
		if isHeader {
			return addToSliceAtIndex(row, columnName, columIndex)
		}

		return addToSliceAtIndex(row, val, columIndex)
	}
}

// ChainTransformers can be used to chain multiple transformers and run them one after another for each row.
// Eg: csvprocessor.ChainTransformers(csvprocessor.AddRowNoTransformer("S.no"), csvprocessor.ReplaceValuesTransformer(valsMap))
// Will add a 'S.no' row and then replace value based on the valsMap
func ChainTransformers(transformers ...CsvRowTransformer) CsvRowTransformer {
	return func(ctx context.Context, row []string) []string {
		for _, transformer := range transformers {
			row = transformer(ctx, row)
		}

		return row
	}
}

func addToSliceAtIndex(slice []string, val string, index int) []string {
	slice = append(slice, "")
	copy(slice[(index+1):], slice[index:])
	slice[index] = val
	return slice
}

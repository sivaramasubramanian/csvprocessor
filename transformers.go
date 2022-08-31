package csvprocessor

import (
	"context"
	"strconv"
)

// CsvRowTransformer represents the transformer function that modifies each row in csv.
// It takes a row as a slice of strings as input and produces the transformed row that will be written to the output file.
// A context.Context with extra metadata about the row is also passed.
type CsvRowTransformer func(context.Context, []string) []string

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
		isHeader, isBool := (ctx.Value(CtxIsHeader)).(bool)
		if isBool && isHeader {
			return addToSliceAtIndex(row, columnName, 0)
		}

		rowID, _ := ctx.Value(CtxRowNum).(int) //nolint:errcheck

		return addToSliceAtIndex(row, strconv.Itoa(rowID), 0)
	}
}

// AddChunkRowNoTransformer adds the row number within current chunk to each row.
// If SkipHeaders is false, it will add a header column for the row number with the given columnName.
func AddChunkRowNoTransformer(columnName string) CsvRowTransformer {
	return func(ctx context.Context, row []string) []string {
		isHeader, isBool := (ctx.Value(CtxIsHeader)).(bool)
		if isBool && isHeader {
			return addToSliceAtIndex(row, columnName, 0)
		}

		rowID, _ := ctx.Value(CtxRowNum).(int)          //nolint:errcheck
		chunkSize, _ := (ctx.Value(CtxChunkSize)).(int) //nolint:errcheck
		chunkRowID := (rowID % chunkSize)
		if chunkRowID == 0 {
			chunkRowID = chunkSize
		}

		return addToSliceAtIndex(row, strconv.Itoa(chunkRowID), 0)
	}
}

// ReplaceValuesTransformer adds a row number to each row.
func ReplaceValuesTransformer(replacements map[string]string) CsvRowTransformer {
	return func(ctx context.Context, row []string) []string {
		for i, element := range row {
			if val, ok := replacements[element]; ok {
				row[i] = val
			}
		}

		return row
	}
}

// AddConstantColumnTransformer adds a new column with the given constant value.
func AddConstantColumnTransformer(columnName, val string, columIndex int) CsvRowTransformer {
	return func(ctx context.Context, row []string) []string {
		isHeader, isBool := (ctx.Value(CtxIsHeader)).(bool)
		if isBool && isHeader {
			return addToSliceAtIndex(row, columnName, columIndex)
		}

		return addToSliceAtIndex(row, val, columIndex)
	}
}

// ChainTransformers can be used to chain multiple transformers and run them one after another for each row.
// Eg: csvprocessor.ChainTransformers(csvprocessor.AddRowNoTransformer("S.no"), csvprocessor.ReplaceValuesTransformer(valsMap))
// Will add a 'S.no' row and then replace value based on the valsMap.
func ChainTransformers(transformers ...CsvRowTransformer) CsvRowTransformer {
	return func(ctx context.Context, row []string) []string {
		for _, transformer := range transformers {
			row = transformer(ctx, row)
		}

		return row
	}
}

// addToSliceAtIndex adds the given value at particular index and shifts the remaining elements to the left.
func addToSliceAtIndex(slice []string, val string, index int) []string {
	slice = append(slice, "")
	copy(slice[(index+1):], slice[index:])
	slice[index] = val

	return slice
}

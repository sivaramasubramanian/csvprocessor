package csvprocessor

import "context"

// TransformerWrapper functions can be used to wrap transformer executions.
type TransformerWrapper func(CsvRowTransformer) CsvRowTransformer

// PanicSafe wraps the transformer execution in a recover block.
func PanicSafe(transformer CsvRowTransformer, log Logger) CsvRowTransformer {
	return func(ctx context.Context, row []string) []string {
		defer func() {
			if r := recover(); r != nil {
				log("csvprocessor: Recovered from panic in transformer ", r)
			}
		}()

		transformedRow := transformer(ctx, row)
		return transformedRow
	}
}

// DebugWrapper can be used to print log statements during transformer execution.
func DebugWrapper(transformer CsvRowTransformer, log Logger) CsvRowTransformer {
	return func(ctx context.Context, row []string) []string {
		log("csvprocessor: before transformation : %v", row)
		transformedRow := transformer(ctx, row)
		log("csvprocessor: after transformation : %v", transformedRow)

		return transformedRow
	}
}

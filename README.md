
# CsvProcessor

CsvProcessor is a simple and fast library to transform and split CSV files in Go.

[![GoDoc](https://godoc.org/github.com/sivaramasubramanian/csvprocessor?status.svg)](https://godoc.org/github.com/sivaramasubramanian/csvprocessor)
[![Go](https://github.com/sivaramasubramanian/csvprocessor/actions/workflows/go.yml/badge.svg?branch=main)](https://github.com/sivaramasubramanian/csvprocessor/actions/workflows/go.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/sivaramasubramanian/csvprocessor)](https://goreportcard.com/report/github.com/sivaramasubramanian/csvprocessor)
[![MIT License](https://img.shields.io/badge/License-MIT-green.svg)](https://choosealicense.com/licenses/mit/)

## Installation
Install

```shell
go get -u github.com/sivaramasubramanian/csvprocessor
```

Import:

```go
import "github.com/sivaramasubramanian/csvprocessor"
```
## Usage

### Splitting a single CSV file into multiple files
```go
	// To split a file into multiple files
	inputFile := "/path/to/input.csv"
	rowsPerFile := 100_000
	outputFilenameFormat := "/path/to/output_%03d.csv" // %03d will be replaced by split id - 001, 002, etc.
	transformer := csvprocessor.NoOpTransformer()      // no-op transformer does not transform the rows

	c := csvprocessor.New(inputFile, rowsPerFile, outputFilenameFormat, transformer)
	err := c.Process()
	if err != nil {
		log.Printf("error while splitting csv %v ",err)
	}
```


### Transforming the content of the CSV
For example, to convert all the values in the 3rd column to Upper case,
```go
	inputFile := "/path/to/input.csv"
	rowsPerFile := 100_000
	outputFilenameFormat := "/path/to/output_%03d.csv" 
	upperCaseTransformer := func(ctx context.Context, row []string) []string {
		isHeader, _ := (ctx.Value(csvprocessor.CtxIsHeader)).(bool)
		if isHeader {
            // ignoring header rows
			return row
		}

		if len(row) > 2 {
            // convert the 3rd column value to Upper case
			row[2] = strings.ToUpper(row[2])
		}

        // return the modified row.
		return row
	}

	c := csvprocessor.New(inputFile, rowsPerFile, outputFilenameFormat, upperCaseTransformer)
	err := c.Process()
	if err != nil {
		log.Printf("error while splitting csv %v ",err)
	}
```

### Predefined Transformer Functions
For some often used cases, there are pre-defined transformer functions that can be used.
For example, to add row number column to a CSV.
```go
	inputFile := "/path/to/input.csv"
	rowsPerFile := 100_000
	outputFilenameFormat := "/path/to/output_%03d.csv" 
	addRowNumberTransformer := csvprocessor.AddRowNoTransformer()

	c := csvprocessor.New(inputFile, rowsPerFile, outputFilenameFormat, addRowNumberTransformer)
	err := c.Process()
	if err != nil {
		log.Printf("error while splitting csv %v ",err)
	}
```
See [transformers.go](./transformers.go) for more pre-defined transformer functions.

### Transform without splitting
To do just transformation without splitiing the CSV into multiple parts,
Give the rowsPerFile value to be equal to or greater than the total rows in file.
```go
	inputFile := "/path/to/input-with-1500-rows.csv"
	rowsPerFile := 1500 // entire input file has only 1500 so only one output file will be generated
	outputFilenameFormat := "/path/to/output.csv" // we can omit the %d format as there will be only one output file
	addRowNumberTransformer := csvprocessor.AddRowNoTransformer()

	c := csvprocessor.New(inputFile, rowsPerFile, outputFilenameFormat, addRowNumberTransformer)
	err := c.Process()
	if err != nil {
		log.Printf("error while splitting csv %v ",err)
	}
```

## Contributing

Contributions are always welcome!

See `contributing.md` for ways to get started.

Please adhere to this project's `code of conduct`.


## License

[![MIT License](https://img.shields.io/badge/License-MIT-green.svg)](https://choosealicense.com/licenses/mit/) See [LICENSE](./LICENSE)

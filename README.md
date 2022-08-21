
# CsvProcessor

CsvProcessor is a simple and fast library to transform and split CSV files in Go. <br>
The file is streamed and processed so it can handle large files that are multiple Gigabytes in size.

[![GoDoc](https://godoc.org/github.com/sivaramasubramanian/csvprocessor?status.svg)](https://godoc.org/github.com/sivaramasubramanian/csvprocessor)
[![Go](https://github.com/sivaramasubramanian/csvprocessor/actions/workflows/go.yml/badge.svg?branch=main)](https://github.com/sivaramasubramanian/csvprocessor/actions/workflows/go.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/sivaramasubramanian/csvprocessor)](https://goreportcard.com/report/github.com/sivaramasubramanian/csvprocessor)
[![MIT License](https://img.shields.io/badge/License-MIT-green.svg)](https://choosealicense.com/licenses/mit/)

## Installation
Install:

Use go get to install the latest version of the library.
```shell
go get -u github.com/sivaramasubramanian/csvprocessor
```

Import:
```go
import "github.com/sivaramasubramanian/csvprocessor"
```
## Usage
- [Simple Usage](#simple-usage)
    - [Splitting a single CSV file into multiple files](#splitting-a-single-csv-file-into-multiple-files)


### Simple Usage
#### Splitting a single CSV file into multiple files
```go
// To split a file into multiple files

inputFile := "/path/to/input.csv"
rowsPerFile := 100_000
// %03d will be replaced by split id - 001, 002, etc.
outputFilenameFormat := "/path/to/output_%03d.csv"
// no-op transformer does not transform the rows
transformer := csvprocessor.NoOpTransformer()

c := csvprocessor.New(inputFile, rowsPerFile, outputFilenameFormat, transformer)
// processes and splits the file
err := c.Process()
if err != nil {
    log.Printf("error while splitting csv %v ",err)
}
```


#### Transforming the content of the CSV
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

#### Predefined Transformer Functions
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

#### Transform without splitting
To do just transformation without splitiing the CSV into multiple parts,
Give the rowsPerFile value to be equal to or greater than the total rows in file.
```go
inputFile := "/path/to/input-with-1500-rows.csv"
rowsPerFile := 1500 // entire input file has only 1500 so only one output file will be generated
outputFilenameFormat := "/path/to/output.csv" // we can omit the %d format as there will be only one output file
addRowNumberTransformer := csvprocessor.AddRowNoTransformer("S.no") // pass the column name for the row number colum

c := csvprocessor.New(inputFile, rowsPerFile, outputFilenameFormat, addRowNumberTransformer)
err := c.Process()
if err != nil {
    log.Printf("error while splitting csv %v ",err)
}
```

#### Combining multiple Transformers
To do a series of transformations for each row, we can chain the transformations,
```go
inputFile := "/path/to/input.csv"
rowsPerFile := 100_000
outputFilenameFormat := "/path/to/output_%03d.csv"
// add row number within current chunk
addChunkRowNumber := csvprocessor.AddRowNoTransformer("Chunk Row no.")
// add overall row number
addRowNumber := csvprocessor.AddRowNoTransformer("Row no.")
// add column 'User' with value 'siva' at index 4; 0-based indexing.
addAConstantColumn := csvprocessor.AddConstantColumnTransformer("User", "siva", 4) 
// Replace all values 'Madras' with 'Chennai'
replacements := make(map[string]string)
replacements["Madras"]="Chennai"
replaceValues := csvprocessor.ReplaceValuesTransformer(replacements)

// chain all these transformations
combinedTransformer := csvprocessor.ChainTransformers(
    addChunkRowNumber,
    addRowNumber,
    addAConstantColumn,
    replaceValues
)

// performs all the transformations
c := csvprocessor.New(inputFile, rowsPerFile, outputFilenameFormat, combinedTransformer)
err := c.Process()
if err != nil {
    log.Printf("error while splitting csv %v ",err)
}
```

#### Using custom logger
To provide your own logger,
```go
// any custom logger that implements `func(format string, args ...any)`
logger = logrus.New().Debugf

c := csvprocessor.New("input.csv", 100, "output.csv", nil)

c.LoggerFunc = logger // set logger to csvprocessor

err := c.Process()
if err != nil {
    log.Printf("error while splitting csv %v ",err)
}
```

#### Wrapping transformer executions
For example, To wrap the transformer with debug statements
```go
logFunc := logrus.WithField("user", "1234").Infof

inputFile := "/path/to/input.csv"
rowsPerFile := 5
outputFilenameFormat := "/path/to/input_%03d.csv"
// add row number within current chunk
addChunkRowNumber := csvprocessor.AddRowNoTransformer("Chunk Row no.")
// print debug statements for each row
debug := csvprocessor.DebugWrapper
// recover from panic in the wrapped transformer function
panicSafe := csvprocessor.PanicSafe

// chain all these transformations
wrappedTransformer := panicSafe(debug(addChunkRowNumber, logFunc), logFunc)

// performs all the transformations
c := csvprocessor.New(inputFile, rowsPerFile, outputFilenameFormat, wrappedTransformer)
c.LoggerFunc = logFunc
err := c.Process()
if err != nil {
    log.Printf("error while splitting csv %v ", err)
}
```

## Roadmap
- [x] csvprocessor
- [x] Transforemr
- [x] Wrapper
- [ ] Helper Functions including merger
- [ ] Unit tests
- [ ] Benchmarking

## Contributing

Contributions are always welcome!

See [contributing.md](./CONTRIBUTING.md) for ways to get started.

Please adhere to this project's [code of conduct](/CODE_OF_CONDUCT.md).


## License

[![MIT License](https://img.shields.io/badge/License-MIT-green.svg)](https://choosealicense.com/licenses/mit/) See [LICENSE](./LICENSE)

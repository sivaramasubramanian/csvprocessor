// Package csvprocessor can be used to efficiently transform and split CSV files.
//
// See README.md for more details
package csvprocessor

import (
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"io/fs"
	"log"
	"os"
	"strings"
)

// CsvProcessor represents the interface for transforming and splitting CSVs.
type CsvProcessor interface {
	Process() error
}

type Processor struct {
	// InputFile location as string
	InputFile string

	// ChunkSize represents the no of rows per each file when splitting the CSV into multiple files.
	// To prevent splitting, set this value to be greater than the total no. of rows.
	ChunkSize int

	// RowTransformer represents the transformer function that will be applied to each row in the input.
	// If no transformation is needed, use NoOpTransformer.
	// This will be called for header rows too. For header rows, CtxIsHeader value in ctx will be true.
	RowTransformer CsvRowTransformer

	// OutputFileFormat represents the format with which the output file names are generated.
	//
	// Only %d format specifier is supported and it will be replaced with the chunk number (0,1,2 etc) for each chunk.
	// Eg: "output_%02d.csv" will generate output files like "output_00.csv", "output_01.csv" etc. where each file contains a single chunk of the file with no of rows as specified in ChunkSize.
	//
	// If there will be only a single output file, we can omit the '%d'. Eg: "only_output.csv"
	// In case the file already exists, then the content will be appended to it.
	OutputFileFormat string

	// SkipHeaders controls whether header should be written in the output file.
	// If true, no header rows are written in any of the split files,
	// else the first row of the input file will be written as header in each split chunk.
	SkipHeaders bool

	// LoggerFunc represents the logger used by the processor to print info and diagnostics.
	LoggerFunc Logger

	// Unexported fields
	header               []string                          // contains the header row
	input                io.ReadCloser                     // reader from which input content is read.
	outputChunkGenerator func(int) (io.WriteCloser, error) // function to generate output chunk files
}

type ctxKey string

type streamElement struct {
	Row []string
	Err error
}

var (
	// CtxChunkNum represents the context.Context() key which contains the current Chunk ID being processed by the Processor.
	CtxChunkNum ctxKey = "_csvproc_chunknum"

	// CtxRowNum represents the context.Context() key which contains the current row ID being processed by the Processor.
	// This is the overall row id and not the one within this chunk.
	// For headers, this value will be -1.
	CtxRowNum ctxKey = "_csvproc_rownum"

	// CtxRowNum represents the context.Context() key which contains whether the current row is a header or not.
	CtxIsHeader ctxKey = "_csvproc_isheader"

	// CtxRowNum represents the context.Context() key which contains the Chunk size for this processor.
	CtxChunkSize ctxKey = "_csvproc_chunksize"

	// noOpTransformer is the default transformer, it does not modify the rows.
	noOpTransformer CsvRowTransformer = NoOpTransformer()
)

const (
	// file permission for the output files.
	permission fs.FileMode = 0o644
)

// New creates a new instance of CsvProcessor.
//
// Parameters:
// 	inputFile - specifies the input file location.
// 	chunkSize - no. of rows per file (if you do want to split the output file give the total row count here).
// 	outputFileFormat - the format with which the output file names are generated.
// 	rowTransformer - function to modify/transform the row.
func New(inputFile string, chunkSize int, outputFileFormat string, rowTransformer func(context.Context, []string) []string) *Processor {
	if rowTransformer == nil {
		rowTransformer = noOpTransformer
	}

	return &Processor{
		InputFile:        inputFile,
		ChunkSize:        chunkSize,
		OutputFileFormat: outputFileFormat,
		LoggerFunc:       log.Default().Printf,
		RowTransformer:   rowTransformer,
	}
}

// Process performs the transformation and splitting and writes the output to the given location.
func (c *Processor) Process() error {
	if c.input == nil {
		inputFile, err := os.Open(c.InputFile)
		if err != nil {
			return err
		}
		defer inputFile.Close()

		c.input = inputFile
	}

	if c.outputChunkGenerator == nil {
		c.outputChunkGenerator = c.createNewSplitFile
	}

	return c.process()
}

func (c *Processor) process() error {
	currentRow := 0
	currentSplit := 1
	var err error
	var fileWriter *csv.Writer
	var outputFile io.WriteCloser
	var addHeaders = !c.SkipHeaders
	var needNewChunk = true
	defer func() {
		if outputFile != nil {
			// close last output chunk file
			outputFile.Close()
		}
	}()

	for element := range streamRows(c.input) {
		row := element.Row
		if element.Err != nil {
			return element.Err
		}

		if needNewChunk {
			if fileWriter != nil {
				// close previous chunk file
				c.LoggerFunc("%d rows processed \n", currentRow)
				if err = c.flushToFile(fileWriter); err != nil {
					return err
				}

				if err = outputFile.Close(); err != nil {
					return err
				}
				currentSplit++
				addHeaders = !c.SkipHeaders
			}

			outputFile, err = c.outputChunkGenerator(currentSplit)
			if err != nil {
				return err
			}

			fileWriter = csv.NewWriter(outputFile)
		}

		if addHeaders {
			addHeaders = false

			// first row in first chunk
			if c.header == nil {
				c.header = row
			}

			// transform and write header
			ctx := c.getCtx(currentSplit, -1, true)
			if err := fileWriter.Write(c.RowTransformer(ctx, c.header)); err != nil {
				return err
			}

			if currentRow == 0 {
				needNewChunk = false
				continue
			}
		}

		currentRow++
		// transform the row
		ctx := c.getCtx(currentSplit, currentRow, false)
		if err := fileWriter.Write(c.RowTransformer(ctx, row)); err != nil {
			return err
		}

		needNewChunk = (currentRow % c.ChunkSize) == 0
	}

	if (currentRow%c.ChunkSize) != 0 && fileWriter != nil {
		if err := c.flushToFile(fileWriter); err != nil {
			return err
		}
	}

	c.LoggerFunc("%d total rows updated", currentRow)
	return nil
}

func (c *Processor) getCtx(chunkID, rowID int, isHeader bool) context.Context {
	ctx := context.WithValue(context.TODO(), CtxChunkSize, c.ChunkSize)
	ctx = context.WithValue(ctx, CtxChunkNum, chunkID)
	ctx = context.WithValue(ctx, CtxRowNum, rowID)
	ctx = context.WithValue(ctx, CtxIsHeader, isHeader)
	return ctx
}

func (c *Processor) createNewSplitFile(split int) (io.WriteCloser, error) {
	c.LoggerFunc("createNewSplitfile called %d", split)
	filename := fmt.Sprintf(c.OutputFileFormat, split)
	filename = strings.Split(filename, "%!")[0]

	return os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, permission)
}

func (c *Processor) flushToFile(w *csv.Writer) error {
	w.Flush()
	return w.Error()
}

func streamRows(rc io.Reader) (ch chan streamElement) {
	buffer := 10
	ch = make(chan streamElement, buffer)
	go func() {
		r := csv.NewReader(rc)
		r.LazyQuotes = true
		r.TrimLeadingSpace = true
		r.FieldsPerRecord = -1
		defer close(ch)
		for {
			rec, err := r.Read()
			if err == io.EOF {
				break
			}

			ch <- streamElement{rec, err}
		}
	}()

	return
}

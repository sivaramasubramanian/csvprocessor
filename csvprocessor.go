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

	// Logger represents the logger used by the processor to print info and diagnostics.
	Logger interface {
		Printf(string, ...any)
	}

	// Unexported fields
	header []string // contains the header row
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

	// file permission for the output files
	permission fs.FileMode = 0o644
)

func New(inputFile string, chunkSize int, outputFileFormat string, rowTransformer func(context.Context, []string) []string) *Processor {
	if rowTransformer == nil {
		rowTransformer = NoOpTransformer()
	}

	return &Processor{
		InputFile:        inputFile,
		ChunkSize:        chunkSize,
		OutputFileFormat: outputFileFormat,
		Logger:           log.Default(),
		RowTransformer:   rowTransformer,
	}
}

// Process performs the transformation and splitting and writes the output to the given location.
func (c *Processor) Process() error {
	rows := 0
	split := 1

	inputFile, err := os.Open(c.InputFile)
	if err != nil {
		return err
	}
	defer inputFile.Close()

	var fileWriter *csv.Writer
	var outputFile *os.File
	var addHeaders = !c.SkipHeaders
	var needNewChunk = true
	defer func() {
		if outputFile != nil {
			outputFile.Close()
		}
	}()

	for element := range streamRows(inputFile) {
		row := element.Row
		if element.Err != nil {
			return element.Err
		}

		if needNewChunk {
			if fileWriter != nil {
				// close previous chunk file
				c.Logger.Printf("%d rows processed \n", rows)
				if err = c.flushToFile(fileWriter); err != nil {
					return err
				}

				if err = outputFile.Close(); err != nil {
					return err
				}
				split++
				addHeaders = !c.SkipHeaders
			}

			outputFile, err = c.createNewSplitFile(split)
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
			ctx := c.getCtx(split, -1, true)
			if err := fileWriter.Write(c.RowTransformer(ctx, c.header)); err != nil {
				return err
			}

			if rows == 0 {
				needNewChunk = false
				continue
			}
		}

		rows++
		// transform the row
		ctx := c.getCtx(split, rows, false)
		if err := fileWriter.Write(c.RowTransformer(ctx, row)); err != nil {
			return err
		}

		needNewChunk = (rows % c.ChunkSize) == 0
	}

	if (rows%c.ChunkSize) != 0 && fileWriter != nil {
		if err := c.flushToFile(fileWriter); err != nil {
			return err
		}
	}

	c.Logger.Printf("%d total rows updated", rows)
	return nil
}

func (c *Processor) getCtx(chunkID, rowID int, isHeader bool) context.Context {
	ctx := context.WithValue(context.TODO(), CtxChunkSize, c.ChunkSize)
	ctx = context.WithValue(ctx, CtxChunkNum, chunkID)
	ctx = context.WithValue(ctx, CtxRowNum, rowID)
	ctx = context.WithValue(ctx, CtxIsHeader, isHeader)
	return ctx
}

func (c *Processor) createNewSplitFile(split int) (*os.File, error) {
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

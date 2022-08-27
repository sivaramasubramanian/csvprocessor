// Package csvprocessor can be used to efficiently transform and split CSV files.
//
// See README.md for more details
package csvprocessor

import (
	"bufio"
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"strings"
)

// CsvProcessor represents the interface for transforming and splitting CSVs.
type CsvProcessor interface {
	Process() error
}

// CsvWriter represents the writer on to which CSV content can written.
// This abstracts the csv.Writer struct from encoding/csv.
type CsvWriter interface {
	Flush()
	Error() error
	Write(record []string) error
}

// CsvReader represents the reader from which CSV content can be read.
type CsvReader interface {
	Read() ([]string, error)
}

// Processor is the default implementation for CsvProcessor.
type Processor struct {
	// chunkSize represents the no of rows per each file when splitting the CSV into multiple files.
	// To prevent splitting, set this value to be greater than the total no. of rows.
	chunkSize int

	// rowTransformer represents the transformer function that will be applied to each row in the input.
	// If no transformation is needed, use NoOpTransformer.
	// This will be called for header rows too. For header rows, CtxIsHeader value in ctx will be true.
	rowTransformer CsvRowTransformer

	// skipHeaders controls whether header should be written in the output file.
	// If true, no header rows are written in any of the split files,
	// else the first row of the input file will be written as header in each split chunk.
	skipHeaders bool

	// log represents the logger used by the processor to print info and diagnostics.
	log Logger

	WriteBufferSize int

	// Unexported fields
	header               []string             // contains the header row
	reader               CsvReader            // reader from which input content is read.
	outputChunkGenerator OutputChunkGenerator // function to generate output chunk files
}

type ctxKey string

// OutputChunkGenerator generates an output writer io.WriteCloser given a chunkID.
// chunkID will always be > 0.
//
// You can customize the output file location, type etc by providing a OutputChunkGenerator to the processor.
// See 'csvprocessor.WithWriterGenerator()' for more details.
type OutputChunkGenerator func(chunkID int) (io.WriteCloser, error)

func NoOpCloser(w io.Writer) io.WriteCloser {
	return nopCloser{w}
}

type nopCloser struct {
	io.Writer
}

func (nopCloser) Close() error { return nil }

var (
	// CtxChunkNum represents the context.Context() key which contains the current Chunk ID being processed by the Processor.
	CtxChunkNum ctxKey = "_csvproc_chunknum"

	// CtxRowNum represents the context.Context() key which contains the current row ID being processed by the Processor.
	// This is the overall row id and not the one within this chunk.
	// For headers, this value will be -1.
	CtxRowNum ctxKey = "_csvproc_rownum"

	// CtxIsHeader represents the context.Context() key which contains whether the current row is a header or not.
	CtxIsHeader ctxKey = "_csvproc_isheader"

	// CtxChunkSize represents the context.Context() key which contains the Chunk size for this processor.
	CtxChunkSize ctxKey = "_csvproc_chunksize"

	// noOpTransformer is the default transformer, it does not modify the rows.
	noOpTransformer CsvRowTransformer = NoOpTransformer()
)

const (
	// file permission for the output files.
	permission fs.FileMode = 0o644

	// DefaultWriteBufferSize represents the default write buffer size of CsvWriter implementation used by the Processor.
	DefaultWriteBufferSize = 10 * 1024 * 1024

	// DefaultReadBufferSize represents the default read buffer size of CsvReader implementation used by the Processor.
	DefaultReadBufferSize = 10 * 1024 * 1024
)

// Process performs the transformation and splitting and writes the output to the given location.
func (c *Processor) Process() error {
	return c.process()
}

func (c *Processor) process() error {
	var fileWriter CsvWriter
	var outputFile io.WriteCloser

	currentRow := 0
	currentSplit := 0
	addHeaders := !c.skipHeaders
	needNewChunk := true
	ctx := newCtx()

	ctx.setValue(CtxChunkSize, c.chunkSize)

	for {
		row, err := c.reader.Read()
		if errors.Is(err, io.EOF) {
			break
		}

		if needNewChunk {
			// close previous chunk file
			c.log("%d rows processed \n", currentRow)
			err := flushAndCloseFile(fileWriter, outputFile)
			if err != nil {
				return err
			}

			// update split id
			currentSplit++
			ctx.setValue(CtxChunkNum, currentSplit)
			addHeaders = !c.skipHeaders

			// create next chunk file
			outputFile, err = c.outputChunkGenerator(currentSplit)
			if err != nil {
				return err
			}

			fileWriter = c.getCsvWriter(outputFile)
		}

		if addHeaders {
			// transform and write header
			err = c.writeHeaders(row, ctx, fileWriter)
			if err != nil {
				return err
			}

			addHeaders = false

			if currentRow == 0 {
				needNewChunk = false
				continue
			}
		}

		currentRow++
		// transform the row
		ctx.setValue(CtxIsHeader, false)
		ctx.setValue(CtxRowNum, currentRow)
		if err := fileWriter.Write(c.rowTransformer(ctx, row)); err != nil {
			return err
		}

		needNewChunk = (currentRow % c.chunkSize) == 0
	}

	c.log("%d total rows updated", currentRow)
	return flushAndCloseFile(fileWriter, outputFile)
}

func flushAndCloseFile(fileWriter CsvWriter, outputFile io.WriteCloser) error {
	if fileWriter != nil {
		if err := flushToFile(fileWriter); err != nil {
			return fmt.Errorf("csprocessor: error while flushing to output file: %w", err)
		}
	}

	if outputFile != nil {
		if err := outputFile.Close(); err != nil {
			return fmt.Errorf("csprocessor: error while closing output file: %w", err)
		}
	}
	return nil
}

func (c *Processor) writeHeaders(row []string, ctx *csvCtx, fileWriter CsvWriter) error {
	if c.header == nil {
		c.header = row
	}

	ctx.setValue(CtxIsHeader, true)
	ctx.setValue(CtxRowNum, -1)
	return fileWriter.Write(c.rowTransformer(ctx, c.header))
}

func (c *Processor) getCsvWriter(outputFile io.WriteCloser) CsvWriter {
	return csv.NewWriter(bufio.NewWriterSize(outputFile, c.WriteBufferSize))
}

func splitFileGenerator(outputFileFormat string) func(int) (io.WriteCloser, error) {
	return func(split int) (io.WriteCloser, error) {
		filename := fmt.Sprintf(outputFileFormat, split)
		filename = strings.Split(filename, "%!")[0]

		return os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, permission) //nolint:nosnakecase
	}
}

func flushToFile(w CsvWriter) error {
	w.Flush()
	return w.Error()
}

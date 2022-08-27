package csvprocessor

import (
	"bufio"
	"context"
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"log"
	"math"
	"os"
)

// NewFileReader creates a new instance of CsvProcessor.
func NewFileReader(inputFile string, chunkSize int, outputFileFormat string, rowTransformer func(context.Context, []string) []string) (*Processor, error) {
	return New(
		WithFileReader(inputFile),
		WithOutputFileFormat(outputFileFormat),
		WithTransformer(rowTransformer),
		WithChunkSize(chunkSize),
	)
}

func NewBufferReader(inputReader io.Reader, outputWriter io.WriteCloser) (*Processor, error) {
	return New(
		WithReader(csv.NewReader(inputReader)),
		WithWriterGenerator(func(int) (io.WriteCloser, error) {
			return outputWriter, nil
		}),
		WithChunkSize(math.MaxInt64),
	)
}

// Option represents a customization option for the Processor.
// A list if such customizations can be passed to New().
type Option func(*Processor) error

var defaultProcessor Processor = Processor{
	WriteBufferSize: DefaultWriteBufferSize,
	rowTransformer:  noOpTransformer,
	log:             log.Default().Printf,
}

func New(opts ...Option) (*Processor, error) {
	newProcessor := defaultProcessor
	for _, opt := range opts {
		if err := opt(&newProcessor); err != nil {
			return nil, err
		}
	}

	processor, err := validate(&newProcessor)
	if err != nil {
		return nil, fmt.Errorf("csvprocessor: invalid input for New(): %w", err)
	}

	return processor, err
}

// WithReader sets the given CsvReader as the reader for the processor.
func WithReader(reader CsvReader) Option {
	return func(c *Processor) error {
		c.reader = reader
		return nil
	}
}

// WithFileReader sets the filename from which the processor will read the data.
func WithFileReader(inputFile string) Option {
	return func(c *Processor) error {
		input, err := os.Open(inputFile)
		if err != nil {
			return err
		}

		var csvReader = csv.NewReader(bufio.NewReaderSize(input, DefaultReadBufferSize))
		csvReader.LazyQuotes = true
		csvReader.TrimLeadingSpace = true
		csvReader.FieldsPerRecord = -1
		csvReader.ReuseRecord = true

		c.reader = csvReader
		return nil
	}
}

// WithTransformer allows to set a custom row transformer.
// To use multiple transformers, chain them using ChainTransformers().
func WithTransformer(t CsvRowTransformer) Option {
	return func(c *Processor) error {
		if t == nil {
			t = noOpTransformer
		}

		c.rowTransformer = t
		return nil
	}
}

// WithOutputFileFormat sets the output file format used to generate output file names.
func WithOutputFileFormat(format string) Option {
	return func(c *Processor) error {
		c.outputChunkGenerator = splitFileGenerator(format)
		return nil
	}
}

// WithWriterGenerator sets the OutputChunkGenerator that generates output io.WriteCloser instances for each split.
func WithWriterGenerator(generator OutputChunkGenerator) Option {
	return func(c *Processor) error {
		c.outputChunkGenerator = generator
		return nil
	}
}

// WithChunkSize sets the chunk size (in no. of rows) for each split.
func WithChunkSize(size int) Option {
	return func(c *Processor) error {
		c.chunkSize = size
		return nil
	}
}

// WithLogger sets the logger for processor.
func WithLogger(logger Logger) Option {
	return func(c *Processor) error {
		c.log = logger
		return nil
	}
}

// SkipHeaders determines whether the processor should write header rows in output files.
func SkipHeaders(skip bool) Option {
	return func(c *Processor) error {
		c.skipHeaders = skip
		return nil
	}
}

var (
	ErrInputReaderNil             = errors.New("csvprocessor: input reader cannot be nil")
	ErrOutputChunkGeneratorNotSet = errors.New("csvprocessor: function to generate output chunks not set")
	ErrInvalidChunkSize           = errors.New("csvprocessor: ChunkSize for splitting must be >= 0, to prevent splitting use math.MaxInt as ChunkSize")
)

func validate(c *Processor) (*Processor, error) {
	if c.reader == nil {
		return nil, ErrInputReaderNil
	}

	if c.outputChunkGenerator == nil {
		return nil, ErrOutputChunkGeneratorNotSet
	}

	if c.chunkSize <= 0 {
		return nil, ErrInvalidChunkSize
	}

	return c, nil
}

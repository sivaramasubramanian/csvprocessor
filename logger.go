package csvprocessor

// Logger defines the logging interface used by csvprocessor.
type Logger logFunc

type logFunc func(string, ...any)

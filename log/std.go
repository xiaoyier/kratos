package log

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"sync"
)

var _ Logger = (*stdLogger)(nil)

type stdLogger struct {
	log  *log.Logger
	pool *sync.Pool
}

// NewStdLogger new a logger with writer.
func NewStdLogger(w io.Writer) Logger {
	return &stdLogger{
		log: log.New(w, "", 0),
		pool: &sync.Pool{
			New: func() interface{} {
				return new(bytes.Buffer)
			},
		},
	}
}

// Log print the kv pairs log.
func (l *stdLogger) Log(level Level, keyvals ...interface{}) error {
	if len(keyvals) == 0 {
		return nil
	}
	if len(keyvals)%2 != 0 {
		keyvals = append(keyvals, "")
	}
	buf := l.pool.Get().(*bytes.Buffer)
	buf.WriteString(level.String())
	for i := 0; i < len(keyvals); i += 2 {
		fmt.Fprintf(buf, " %s=%v", keyvals[i], keyvals[i+1])
	}
	l.log.Output(4, buf.String())
	buf.Reset()
	l.pool.Put(buf)
	return nil
}

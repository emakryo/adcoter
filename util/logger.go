package util

import (
	"io"
	"log"
)

type Logger struct {
	*log.Logger
}

func NewLogger(out io.Writer, prefix string, flag int) *Logger {
	var logger *Logger = &Logger{log.New(out, prefix, flag)}
	return logger
}

func (logger *Logger)Write(p []byte) (int, error) {
	var buf []byte
	for _, c := range p {
		buf = append(buf, c)
		if c == '\n' {
			logger.Printf(string(buf[:]))
			buf = []byte{}
		}
	}

	if len(buf) > 0 {
		logger.Printf(string(buf[:]))
	}
	return len(p), nil
}

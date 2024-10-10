package main

import (
	"errors"
	"os"
	"sync"

	"github.com/rs/zerolog"
)

var (
	errLevelTooLow = errors.New("Level of msg is lower than cap level")
	maxCapacity    = 16
)

type logger struct {
	minLoggingLevel int8
	mu              *sync.Mutex
	first           *log
	last            *log
	count           int
	zlog            zerolog.Logger
}

type log struct {
	level int8
	next  *log
	msg   string
}

func (l *log) addNext(level int8, msg string) *log {
	new := &log{
		level: level,
		next:  nil,
		msg:   msg,
	}

	l.next = new
	return new
}

func (l *logger) add(level int8, msg string) {
	l.mu.Lock()
	defer l.mu.Unlock()

	if level < l.minLoggingLevel {
		return
	}

	if l.last == nil {
		firstLog := &log{
			level: level,
			next:  nil,
			msg:   msg,
		}

		l.first = firstLog
		l.last = firstLog
	} else {
		l.last = l.last.addNext(level, msg)
	}
	l.count++

	if l.count == maxCapacity {
		l.startLogging()
	}
}

func NewLogger() *logger {
	return &logger{
		minLoggingLevel: -1,
		mu:              new(sync.Mutex),
		first:           nil,
		last:            nil,
		count:           0,
		zlog:            zerolog.New(zerolog.ConsoleWriter{Out: os.Stdout}).With().Timestamp().Logger(),
	}
}

func main() {
	l := NewLogger()
}

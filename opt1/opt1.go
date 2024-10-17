package opt1

import (
	"os"
	"sync"
	"time"

	"github.com/rs/zerolog"
	//"logger/constants"
)

type logger struct {
	minLoggingLevel   int8
	loggingLevelShift int8
	maxCapacity       int
	mu                *sync.Mutex
	first             *log
	last              *log
	count             int
	zlog              zerolog.Logger
	chIn              chan log
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

	if level < l.minLoggingLevel+l.loggingLevelShift {
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

	if l.count == l.maxCapacity {
		l.ThrowUp()
	}
}

func (l *logger) printFirst() {
	l.zlog.WithLevel(zerolog.Level(l.first.level)).Msg(l.first.msg)
	l.first = l.first.next
	l.count--
}

func (l *logger) adder() {
	for {
		select {
		case log := <-l.chIn:
			l.add(log.level, log.msg)
		}
	}
}

func (l *logger) worker(timeout time.Duration) {
	for {
		time.Sleep(timeout)

		if l.count != 0 {
			l.mu.Lock()

			l.printFirst()

			l.mu.Unlock()
		}
	}
}

func NewLogger(timeout time.Duration, minLoggingLevel int, maxCapacity int) *logger {
	l := logger{
		minLoggingLevel:   int8(minLoggingLevel),
		loggingLevelShift: 0,
		maxCapacity:       maxCapacity,
		mu:                new(sync.Mutex),
		first:             nil,
		last:              nil,
		count:             0,
		zlog:              zerolog.New(zerolog.ConsoleWriter{Out: os.Stdout}).With().Timestamp().Logger(),
		chIn:              make(chan log, maxCapacity),
	}

	go l.adder()
	go l.worker(timeout)

	return &l
}

func (l *logger) ThrowUp() {
	l.mu.Lock()
	defer l.mu.Unlock()

	next := l.first
	for next != nil {
		l.zlog.WithLevel(zerolog.Level(next.level)).Msg(next.msg)
		next = next.next
	}

	l.first = nil
	l.last = nil
	l.count = 0
}

func (l *logger) Log(level int, msg string) {
	l.mu.Lock()

	if len(l.chIn) > l.maxCapacity/2 && l.loggingLevelShift < 1 {
		l.loggingLevelShift = 1
	}

	l.mu.Unlock()

	l.chIn <- log{
		level: int8(level),
		msg:   msg,
	}
}

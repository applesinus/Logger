package opt2

import (
	"logger/constants"
	"os"
	"sync"
	"time"

	"github.com/rs/zerolog"
)

type logger struct {
	minLoggingLevel   int8
	loggingLevelShift int8

	timeout      time.Duration
	workersCount int
	//TODO
	extraWorkersCount int

	levelsCap map[int8]int

	logs       chan log
	stopWorker chan struct{}

	mu *sync.Mutex

	zlog zerolog.Logger
}

type log struct {
	level int8
	msg   string
}

func StandartLevelsCap() map[int8]int {
	capLevels := make(map[int8]int, 7)
	capLevels[-1] = 10
	capLevels[0] = 20
	capLevels[1] = 30
	capLevels[2] = 50
	capLevels[3] = 80
	capLevels[4] = 100
	capLevels[5] = 101
	return capLevels
}

func NewLogger(minLoggingLevel, maxCapacity int, timeout time.Duration, levelsCapPercentage map[int8]int) *logger {
	l := logger{
		minLoggingLevel:   int8(minLoggingLevel),
		loggingLevelShift: 0,

		logs:       make(chan log, maxCapacity),
		stopWorker: make(chan struct{}),

		levelsCap: levelsCapPercentage,

		timeout:      timeout,
		workersCount: 0,

		mu: new(sync.Mutex),

		zlog: zerolog.New(zerolog.ConsoleWriter{Out: os.Stdout}).With().Timestamp().Logger(),
	}

	go l.worker()

	return &l
}

func (l *logger) startNewWorker() {
	l.mu.Lock()
	defer l.mu.Unlock()

	go l.worker()
	l.workersCount++
}

func (l *logger) stopOneWorker() {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.workersCount == 1 {
		return
	}

	l.stopWorker <- struct{}{}
	l.workersCount--
}

func (l *logger) worker() {
	for {
		select {
		case <-l.stopWorker:
			return

		case toPrint := <-l.logs:
			l.zlog.WithLevel(zerolog.Level(toPrint.level)).Msg(toPrint.msg)

		default:
		}

		time.Sleep(l.timeout)
	}
}

func (l *logger) sendLog(level int8, msg string) error {
	if level < -1 || level > 6 {
		return constants.ErrInvalidLevel
	}

	if level < l.minLoggingLevel+l.loggingLevelShift {
		return constants.ErrLevelTooLow
	}

	l.logs <- log{
		level: level,
		msg:   msg,
	}

	return nil
}

func (l *logger) workersManager() {
	fillnessPercentage := 0
	currentLevel := l.minLoggingLevel + l.loggingLevelShift

	fullClock := 0

	for {
		l.mu.Lock()

		fillnessPercentage = (len(l.logs) * 100) / cap(l.logs)

		if fillnessPercentage > l.levelsCap[currentLevel] && l.loggingLevelShift < 5 {
			l.loggingLevelShift++
			currentLevel++

			if currentLevel > 3 {
				l.startNewWorker()
			}
		} else if l.loggingLevelShift > 0 && fillnessPercentage <= l.levelsCap[currentLevel-1] {
			l.loggingLevelShift--
			currentLevel--

			if currentLevel > 2 {
				l.stopOneWorker()
			}
		}

		l.mu.Unlock()

		if currentLevel == 4 {
			fullClock++
			if fullClock == 10 {
				l.zlog.WithLevel(constants.MtzlLevel).Msg("Logger buffer is full")
			}
		} else {
			fullClock = 0
		}

		time.Sleep(l.timeout * 10)
	}
}
